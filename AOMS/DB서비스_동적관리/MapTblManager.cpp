#include "MappingTable/MapTblManager.h"

#include <iostream>
#include <vector>
#include <algorithm>

#include "Common/AomDef.h"
#include "CommIh/IhUtil.h"
#include "Miscellaneous/TimeUtil.h"
#include "CommIh/DebugLogger.h"

#define CFN typeid(*this).name() << "::" << __func__ << "] "

using namespace AomDef;

#if 0
#define TRACE(A) FLOG(AREA_PAS, "[MapTblManager] Function : " << A)
#define DPRINT(A) DLOG(AREA_PAS, "[MapTblManager] " << A)
#define IPRINT(A) ILOG(AREA_PAS, "[MapTblManager] " << A)
#define EPRINT(A) ELOG(AREA_PAS, "[MapTblManager] " << RED(A))
#endif

#define IS_MA_SUBS(id) (id[1]=='M'?true:false)
#define IS_HOST_SUBS(id) (id[1]!='M'?true:false)

//------------------------------------------------------------------------------
// global variables
//------------------------------------------------------------------------------

MapTblManager* MapTblManager::m_pInstance = NULL;

void MapTblManager::RunMapTblInfoPolling()
{
	static const char *FN = "[MapTblManager::RunMapTblInfoPolling] ";

	mtiList mlist;
	mtiLIt it;
	int ret = 0;
	int cnt;
	time_t tmStt, tmEnd, tmCur;
	bool initFlag = true;

	while (true) {

		axSleep(500);

		if (true == initFlag) {
			tmCur = time(0);
			tmEnd = tmCur - m_PollingMargin;
			initFlag = false;
			// DB전체 Load 기능 변경에 따라 필요 없어짐
			//tmStt = 0;

			/* T_MDN_MAP_TBL 전체 Loading의 부하때문에 고려되는 사항.
			* 초기 기동 시 전체 Loading 하지 않고 Update Time에 따라 Loading 되도록 할 수 있게 수정된 내용이 아래.
			*/
			tmStt = tmEnd - m_PollingMargin;

		}
		else {
			time(&tmCur);
		}

		if (tmCur % m_PollingInterval)
			continue;

		bool isFail = false;

		{
			MapTblInfoDbApi db(m_UdbId.c_log(), true, true);

			if ((ret = db.selectChageSubsInfo(mlist, tmStt, tmEnd, m_isCS)) != MRET_DB_SUCCESS) {
				if (ret == MRET_DB_NOT_FOUND)
					goto SleepPollingInterval;

				isFail = true;

				EPRINT(FN << " selectChageSubsInfo() Fail Polling Inteval : " << m_PollingInterval);
				goto SleepPollingInterval;
			}
		}

		{
			int hash;
			mtiMIt fn;
			it = mlist.begin();
			MapTblInfo tmp;
			while (it != mlist.end()) {
				tmp = *it;
				if (m_isCS) {
					DPRINT(FN << " list mdn : " << tmp.m_Mdn << ", aomcid : " << tmp.m_HostAomcId << ", delFlag : "
						<< (tmp.m_delFlag ? "true" : "false") << ", HDV_SUPP : " << (tmp.m_bHDVSupp ? "true" : "false")
						<< ", UpdateTime : " << TimeUtil::ConverTimeString(tmp.m_UpdTime, TimeUtil::FMT_YYYYMMDDHH24MISS) << ",PDD_SVC: " << tmp.m_PddSvcType);
				}
				else {
					DPRINT(FN << " list mdn : " << tmp.m_Mdn << ", aomcid : " << tmp.m_HostAomcId << ", delFlag : "
						<< (tmp.m_delFlag ? "true" : "false") << ", HDV_SUPP : " << (tmp.m_bHDVSupp ? "true" : "false")
						<< ", UpdateTime : " << TimeUtil::ConverTimeString(tmp.m_UpdTime, TimeUtil::FMT_YYYYMMDDHH24MISS));
				}

				hash = Hash(tmp.m_Mdn);

				if (hash < 0 || hash >= MAP_TBL_HASH_SIZE)
					goto nextMapProc;

				{
					AxLock lock(m_lock_map_mdn[hash]);

					/* MAP 에서 찾자. */
					fn = m_map_mdn[hash].find(tmp.m_Mdn);
					if (fn == m_map_mdn[hash].end()) {
						/* MAP에 없는 경우 일반적인 가입자 추가.
						PDD 지원 단말 추가되는 경우. */

						//R431-p1. Oracle의 PddSvcType Select 이슈 보완
						//R431에서 주기적 select 시 오라클은 PddSvcType을 select 하지 않아서 지속적인 udpate 발생 이슈가 있음.
						//DB에서 select 한 결과 m_PddSvcTyep이 없는 경우 (오라클) PddSvcType을 위한 select query 수행
						//R420에서는 이 방법으로 PddSvcType을 Load하였다.
						//R431-p1. CS 외의 다른 프로세스들은 PDD 가져오지 못하도록 처리..

						if ((tmp.m_PddSvcType == "") && (m_isCS == true)) {
							int Ret;
							MapTblInfo mapinfo;
							MapTblInfoDbApi db(m_UdbId.c_log());
							if ((Ret = db.selectPddSvcSubsInfo(tmp.m_Mdn, mapinfo)) == MRET_DB_SUCCESS) {
								tmp.m_PddSvcType = mapinfo.m_PddSvcType;
								IPRINT("T_MDN_MAP_TBL Select Found. mdn: " << tmp.m_Mdn << ", PDD: " << tmp.m_PddSvcType);
							}
							else if (Ret == MRET_DB_NOT_FOUND)
								IPRINT("T_MDN_MAP_TBL Select Not Found. MDN: " << tmp.m_Mdn);
							else
								ILOG("T_MDN_MAP_TBL Select Error. MDN: " << tmp.m_Mdn);
						}

						m_map_mdn[hash].insert(pair<string, MapTblInfo>(tmp.m_Mdn, tmp));
						if (0 != tmStt) {
							EPRINT(CYAN("PDD-MDN INSERT [" << tmp.m_Mdn << ", " << tmp.m_HostAomcId << ", " << (tmp.m_delFlag ? "true" : "false") << ", " << (tmp.m_bHDVSupp ? "true" : "false") << ", " << tmp.m_PddSvcType << "]"));
						}
					}
					else {    // MAP 해당 MDN에 대한 정보가 이미 있는 경우.
						MapTblInfo fnTmp = fn->second;

						/** R402. CS의 경우 Insert Map에서 Cache Map 구성이 Account별도 된다. */
						fn->second.m_Mdn = tmp.m_Mdn;
						fn->second.m_AomcId = tmp.m_AomcId;
						fn->second.m_HostAomcId = tmp.m_HostAomcId;
						fn->second.m_delFlag = tmp.m_delFlag;
						fn->second.m_bHDVSupp = tmp.m_bHDVSupp;

						if ((tmp.m_PddSvcType == "") && (m_isCS == true)) {
							int Ret;
							MapTblInfo mapinfo;
							MapTblInfoDbApi db(m_UdbId.c_log());
							if ((Ret = db.selectPddSvcSubsInfo(tmp.m_Mdn, mapinfo)) == MRET_DB_SUCCESS) {
								tmp.m_PddSvcType = mapinfo.m_PddSvcType;
								fn->second.m_PddSvcType = mapinfo.m_PddSvcType;
								IPRINT("T_MDN_MAP_TBL Select Found. mdn: " << tmp.m_Mdn << ", PDD: " << mapinfo.m_PddSvcType);
							}
							else if (Ret == MRET_DB_NOT_FOUND)
								IPRINT("T_MDN_MAP_TBL Select Not Found. MDN: " << tmp.m_Mdn);
							else
								EPRINT("T_MDN_MAP_TBL Select Error. MDN: " << tmp.m_Mdn);
						}
						else {
							fn->second.m_PddSvcType = tmp.m_PddSvcType;
						}

						fn->second.m_UpdTime = tmp.m_UpdTime;
						if (0 != tmStt) {
							EPRINT(CYAN("PDD-MDN UPDATE [" << tmp.m_Mdn << ", " << tmp.m_HostAomcId << ", " << (tmp.m_delFlag ? "true" : "false") << ", " << (tmp.m_bHDVSupp ? "true" : "false") << ", " << tmp.m_PddSvcType << "]"));
						}
					}
				}   //AxLock End

				hash = Hash(tmp.m_HostAomcId);
				if (hash < 0 || hash >= MAP_TBL_HASH_SIZE)
					goto nextMapProc;

#if 0
				{
					AxLock lock(m_lock_map_aomcid[hash]);

					/* AOMCID로도 찾을수 있는 Map 구성. */
					fn = m_map_aomcid[hash].find(tmp.m_HostAomcId);
					if (fn == m_map_aomcid[hash].end()) {
						m_map_aomcid[hash].insert(pair<string, MapTblInfo>(tmp.m_HostAomcId, tmp));
					}
					else {    // MAP 해당 AOMCID에 대한 정보가 이미 있는 경우.
						MapTblInfo fnTmp = fn->second;
						if (fnTmp.m_Mdn != tmp.m_Mdn) {
							/* MDN가 서로 다르다면 업데이트라고 판단 하자.
							DB 정보를 최우선으로 선택. */
							m_map_aomcid[hash].erase(fn);
							m_map_aomcid[hash].insert(pair<string, MapTblInfo>(tmp.m_HostAomcId, tmp));
							goto nextMapProc;
						}
						/* 이미 있으면 무조건 Update임. */
						m_map_aomcid[hash].erase(fn);
						m_map_aomcid[hash].insert(pair<string, MapTblInfo>(tmp.m_HostAomcId, tmp));
					}
				}
#endif
			nextMapProc:
				it++;
			}

			if (mlist.size() > 0) {
				IPRINT(GREEN(FN << " update subs map info count : " << mlist.size()));
			}
		}

	SleepPollingInterval:
		cnt = mlist.size();
		mlist.clear();
		axSleep(600);
		SetQueryTime(tmStt, tmEnd, isFail, cnt);
	}
}