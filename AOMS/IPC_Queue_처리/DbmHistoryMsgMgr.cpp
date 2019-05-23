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
					isFail = true;// DB Operation �ϳ��� ���н� �����
				}
			}
			else if (rst == RET_DB_NOT_FOUND) {
				DDLOG(FN << "SelectSisHpush Not Found");
			}
			else {
				EELOG(FN << "SelectSisHpush FAIL");
				isFail = true;// DB Operation �ϳ��� ���н� �����
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
					isFail = true;// DB Operation �ϳ��� ���н� �����
				}
			}
			else if (rst == RET_DB_NOT_FOUND) {
				DDLOG(FN << "SelectHpush Not Found");
			}
			else {
				IILOG(FN << "SelectHpush FAIL");
				isFail = true;// DB Operation �ϳ��� ���н� �����
			}
		}

		IILOG(FN << "Select List Size:" << oList.size());
		cnt = oList.size();
		if (cnt) {
			ProcessingHstatData(oList); // List�� �ִ� �����͵��� ��� ���̺��� key���� key�� �ϴ� map �ڷᱸ���� list ������ ����
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

		// 1. �۾� ��� �ð� (IhData) Ȯ��
		// 2. �۾� ��� ���̺� ����
		// 3. �ڷᱸ���� �ִ� ������ �� �۾� ��� �ð��� �͵� ����
		//  - ���� �� �� ����
		//   > ���� �ڷᱸ���� ������ ���� ���� key�� �ִ��� ����
		//   > ������ �״�� �ְ� ��
		//   > ������ ������ ������ �����ְ� �ٽ� �ְ� ��
		// 4. ����� ��� ���ؼ� DB�� ������
		//  - key�� ���� �����Ͱ� �ִ��� select �� ���� ������ ������ �����ְ� update
		//  - ������ insert
		// 

		if (isJobTime)
		{
			int rst;

			{
				OraDBApi sDB(DBMIS_STATISTICS_DB);
				rst = sDB.CheckHstatTbl(jobDate);
			}

			if (rst == RET_DB_NOT_FOUND) { // ��� ���̺��� ������ ������
				OraDBApi rDB(DBMIS_STATISTICS_DB);
				rst = rDB.CreateHstatTbl(jobDate);
				if (rst == RET_DB_SUCCESS || rst == RET_DB_EXIST_TBL || rst == RET_DB_EXIST_TBL_MYSQL) { // ������ų� ������ ��赥���� Insert
					isJobTime = false;
					InsertHstatData(jobDate, CsNum);
				}
				else { // ��� ���̺��� ���ų� ��������ٸ� ó������ �ٽ� ����
					continue;
				}
			}
			else if (rst == RET_DB_SUCCESS) { // �̹� ��� ���̺��� �ִٸ� ��赥���� Insert
				isJobTime = false;
				InsertHstatData(jobDate, CsNum);
			}
			else { // ��� ���̺� �˻� �����ߴٸ� ó������ �ٽ� ����
				continue;
			}
		}
		else // JobTime ����
		{
			jobDate = GetStringPastTime(DbmIhData::instance()->GetStatProcPastTime(), true);
			if (SttDate != jobDate) { // ���� ���� �ð��� ������ �ٸ��ٸ� ��� ���̺� ���� �� Insert �ϵ���
				isJobTime = true;
				SttDate = jobDate;
				continue;
			}
			else { // ���� ����ð��� �����ϴٸ� Insert �� �����͸� �˻��ؼ� Insert ����
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
			* �ؼ� ���� ���̺� ����Ʈ�� �߿� strTmStt���� ���� �͸� ���� �ϰ�
			* ������ ���̺���� ����
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