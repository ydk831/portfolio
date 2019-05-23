#include <errno.h>
#include <typeinfo>
#include <occi.h>
#include <cctype>
#include <algorithm>

#include "DbmHistoryMsgMgr.h"

#include "CommIh/DebugLogger.h"
#include "Miscellaneous/TimeUtil.h"
#include "Mysql/DBManager.h"
#include "IpcQ/IpcQ.h"
#include "AxLib/AxLib.h"

#include "DbmIhData.h"
#include "OraDBApi.h"

#include "Hmsg.h"
#include "HpushMsg.h"
#include "HnotiMsg.h"

#include "StringConverter.h"


using namespace Ih;

DbmHistoryMsgMgr *DbmHistoryMsgMgr::m_spInstance = NULL;

void DbmHistoryMsgMgr::RunHistoryGathering()
{
	static const char FN[] = "[DbmHistoryMsgMgr::RunHistoryGathering] ";

	int CsNum = DbmIhData::instance()->getServiceID();
	bool isJobTime = true;
	string SttDate, EndDate, jobDate;
	jobDate = SttDate;

	time_t tmStt, tmEnd, tmCur;
	int cnt = 0;
	bool initFlag = true, isFail = false;

	while (1)
	{
		axSleep(500); // 0.5sec

		if (true == initFlag) {
			tmCur = time(0);
			tmEnd = tmCur - m_PollingMargin;
			initFlag = false;
			tmStt = tmEnd - m_PollingMargin;
		}
		else
			time(&tmCur);

		if (tmCur % m_PollingInterval)
			continue;

		bool isFail = false;

		HDList oList;
		int rst;
		{
			OraDBApi rDB(DBMIS_STATISTICS_DB);
			rst = rDB.SelectSisHpush(tmStt, tmEnd, CsNum, oList);

			if (rst == RET_DB_SUCCESS) {
				OraDBApi rDB2(DBMIS_STATISTICS_DB);
				rst = rDB2.DeleteSisHpush(CsNum, tmStt, tmEnd);

				if ((rst == RET_DB_SUCCESS) || rst == RET_DB_NOT_FOUND) {
					DDLOG(FN << "DeleteSisHpush SUCCESS");
				}
				else {
					EELOG(FN << "DeleteSisHpush FAIL");
					isFail = true;// DB Operation 하나라도 실패시 재수행
				}
			}
			else if (rst == RET_DB_NOT_FOUND) {
				DDLOG(FN << "SelectSisHpush Not Found");
			}
			else {
				EELOG(FN << "SelectSisHpush FAIL");
				isFail = true;// DB Operation 하나라도 실패시 재수행
			}
		}

		{
			OraDBApi rDB(DBMIS_STATISTICS_DB);
			rst = rDB.SelectHpush(CsNum, tmStt, tmEnd, oList);
			if (rst == RET_DB_SUCCESS) {
				OraDBApi rDB2(DBMIS_STATISTICS_DB);
				rst = rDB2.DeleteHpush(tmStt, tmEnd, CsNum);

				if ((rst == RET_DB_SUCCESS) || rst == RET_DB_NOT_FOUND) {
					DDLOG(FN << "DeleteHpush SUCCESS");
				}
				else {
					EELOG(FN << "DeleteHpush FAIL");
					isFail = true;// DB Operation 하나라도 실패시 재수행
				}
			}
			else if (rst == RET_DB_NOT_FOUND) {
				DDLOG(FN << "SelectHpush Not Found");
			}
			else {
				IILOG(FN << "SelectHpush FAIL");
				isFail = true;// DB Operation 하나라도 실패시 재수행
			}
		}

		IILOG(FN << "Select List Size:" << oList.size());
		cnt = oList.size();
		if (cnt) {
			ProcessingHstatData(oList); // List에 있는 데이터들을 통계 테이블의 key값을 key로 하는 map 자료구조에 list 데이터 넣음
		}

	SleepPollingInterval:
		axSleep(600);
		SetQueryTime(tmStt, tmEnd, isFail, cnt);
	}
}

void DbmHistoryMsgMgr::RunHistoryStatCreating()
{
	static const char FN[] = "[RunHistoryStatCreating] ";

	bool isJobTime = true;
	int CsNum = DbmIhData::instance()->getServiceID();
	string SttDate, jobDate;
	SttDate = GetStringPastTime(DbmIhData::instance()->GetStatProcPastTime(), true);
	jobDate = SttDate;

	while (1)
	{
		sleep(1);

		// 1. 작업 대상 시간 (IhData) 확인
		// 2. 작업 대상 테이블 생성
		// 3. 자료구조에 있는 데이터 중 작업 대상 시간인 것들 추출
		//  - 추출 할 때 순서
		//   > 별도 자료구조에 추출한 대상과 같은 key가 있는지 보고
		//   > 없으면 그대로 넣고 끝
		//   > 있으면 꺼내서 데이터 합쳐주고 다시 넣고 끝
		// 4. 추출된 대상에 대해서 DB에 넣을때
		//  - key가 같은 데이터가 있는지 select 해 보고 있으면 데이터 합쳐주고 update
		//  - 없으면 insert
		// 

		if (isJobTime)
		{
			int rst;

			{
				OraDBApi sDB(DBMIS_STATISTICS_DB);
				rst = sDB.CheckHstatTbl(jobDate);
			}

			if (rst == RET_DB_NOT_FOUND) { // 통계 테이블이 없으면 만들자
				OraDBApi rDB(DBMIS_STATISTICS_DB);
				rst = rDB.CreateHstatTbl(jobDate);
				if (rst == RET_DB_SUCCESS || rst == RET_DB_EXIST_TBL || rst == RET_DB_EXIST_TBL_MYSQL) { // 만들었거나 있으면 통계데이터 Insert
					isJobTime = false;
					InsertHstatData(jobDate, CsNum);
				}
				else { // 통계 테이블이 없거나 못만들었다면 처음부터 다시 수행
					continue;
				}
			}
			else if (rst == RET_DB_SUCCESS) { // 이미 통계 테이블이 있다면 통계데이터 Insert
				isJobTime = false;
				InsertHstatData(jobDate, CsNum);
			}
			else { // 통계 테이블 검색 실패했다면 처음부터 다시 수행
				continue;
			}
		}
		else // JobTime 변경
		{
			jobDate = GetStringPastTime(DbmIhData::instance()->GetStatProcPastTime(), true);
			if (SttDate != jobDate) { // 이전 수행 시간과 지금이 다르다면 통계 테이블 생성 및 Insert 하도록
				isJobTime = true;
				SttDate = jobDate;
				continue;
			}
			else { // 이전 수행시간과 동일하다면 Insert 할 데이터만 검색해서 Insert 수행
				DDLOG(FN << " is not job time");
				InsertHstatData(jobDate, CsNum);
				continue;
			}
		}
	}
}

void DbmHistoryMsgMgr::RunHistoryStatRemoving()
{
	static const char FN[] = "[RunHistoryStatRemoving] ";

	bool isJobTime = true;
	int CsNum = DbmIhData::instance()->getServiceID();
	string SttDate, jobDate;
	SttDate = GetStringPastTime(DbmIhData::instance()->GetStatTblDropPastTime(), true);
	jobDate = SttDate;

	while (1)
	{
		sleep(5);

		if (isJobTime)
		{
			/* select table_name from information_schema.tables where table_name like 't_hpush_stat_%';
			* 해서 나온 테이블 리스트들 중에 strTmStt보다 작은 것만 선별 하고
			* 선별된 테이블들을 삭제
			*/
			int rst;

			{
				OraDBApi sDB(DBMIS_STATISTICS_DB);
				rst = sDB.CheckHstatTbl(jobDate);
			}

			if (rst == RET_DB_SUCCESS) {
				OraDBApi rDB(DBMIS_STATISTICS_DB);
				rst = rDB.DeleteHstatTbl(jobDate);

				if (rst != RET_DB_SUCCESS && rst != RET_DB_NOT_FOUND) {
					EELOG(FN << "Stat Table Drop Fail. Retry. " << jobDate.c_str() << ", ERR: " << rst);
					continue;
				}
				else {
					isJobTime = false;
				}
			}
			else {
				isJobTime = false;
			}
		}
		else
		{
			jobDate = GetStringPastTime(DbmIhData::instance()->GetStatTblDropPastTime(), true);
			if (SttDate != jobDate) {
				isJobTime = true;
				SttDate = jobDate;
				continue;
			}
			else {
				DDLOG(FN << "is not job time");
				continue;
			}
		}
	}
}