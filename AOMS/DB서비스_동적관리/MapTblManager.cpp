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
			// DB��ü Load ��� ���濡 ���� �ʿ� ������
			//tmStt = 0;

			/* T_MDN_MAP_TBL ��ü Loading�� ���϶����� ����Ǵ� ����.
			* �ʱ� �⵿ �� ��ü Loading ���� �ʰ� Update Time�� ���� Loading �ǵ��� �� �� �ְ� ������ ������ �Ʒ�.
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

					/* MAP ���� ã��. */
					fn = m_map_mdn[hash].find(tmp.m_Mdn);
					if (fn == m_map_mdn[hash].end()) {
						/* MAP�� ���� ��� �Ϲ����� ������ �߰�.
						PDD ���� �ܸ� �߰��Ǵ� ���. */

						//R431-p1. Oracle�� PddSvcType Select �̽� ����
						//R431���� �ֱ��� select �� ����Ŭ�� PddSvcType�� select ���� �ʾƼ� �������� udpate �߻� �̽��� ����.
						//DB���� select �� ��� m_PddSvcTyep�� ���� ��� (����Ŭ) PddSvcType�� ���� select query ����
						//R420������ �� ������� PddSvcType�� Load�Ͽ���.
						//R431-p1. CS ���� �ٸ� ���μ������� PDD �������� ���ϵ��� ó��..

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
					else {    // MAP �ش� MDN�� ���� ������ �̹� �ִ� ���.
						MapTblInfo fnTmp = fn->second;

						/** R402. CS�� ��� Insert Map���� Cache Map ������ Account���� �ȴ�. */
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

					/* AOMCID�ε� ã���� �ִ� Map ����. */
					fn = m_map_aomcid[hash].find(tmp.m_HostAomcId);
					if (fn == m_map_aomcid[hash].end()) {
						m_map_aomcid[hash].insert(pair<string, MapTblInfo>(tmp.m_HostAomcId, tmp));
					}
					else {    // MAP �ش� AOMCID�� ���� ������ �̹� �ִ� ���.
						MapTblInfo fnTmp = fn->second;
						if (fnTmp.m_Mdn != tmp.m_Mdn) {
							/* MDN�� ���� �ٸ��ٸ� ������Ʈ��� �Ǵ� ����.
							DB ������ �ֿ켱���� ����. */
							m_map_aomcid[hash].erase(fn);
							m_map_aomcid[hash].insert(pair<string, MapTblInfo>(tmp.m_HostAomcId, tmp));
							goto nextMapProc;
						}
						/* �̹� ������ ������ Update��. */
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