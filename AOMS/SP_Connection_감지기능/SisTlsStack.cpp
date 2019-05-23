#include <poll.h>
#include <occi.h>
#include <algorithm>
#include "cyassl/error-ssl.h"

#include "SisDef.h"
#include "SisIhUtil.h"
#include "SisIhData.h"
#include "SisTlsStack.h"
#include "SisTransactionMgr.h"
#include "OraDBApi.h"
#include "SisAlarmFault.h"

SisTlsStack::SisTlsStack()
{
	m_name = "SisTlsStack";

	m_bEnable = false;
	m_PacketBufferSize = 1024 * 1024;//16;    // default 16k   R350. for multi-push
	for (__u_int i = 0; i < CPUs; ++i) {
		m_pPollingTask[i] = NULL;
	}
	m_socketPoolCount = 0;
	m_pWorkerPool = NULL;

	m_SocketSNDBUFSize = 1024 * 1024;//16;    // default 16k   R350. for multi-push
	m_SocketRCVBUFSize = 1024 * 1024;//16;    // default 16k   R350. for multi-push

	m_FDIndex2Socket = NULL;
	m_MaxFD = 0;
	m_iSeqNo = 0;

	m_pSessChecker = NULL;
	m_SessionCount = 0;
#ifdef _SUPP_IPV6_
	m_SessionCountIpv6 = 0;
#endif

	m_pRunTlsListener = NULL;
	m_pTlsRegiChecker = NULL;
#ifdef _SUPP_IPV6_
	m_pRunTlsListenerIpv6 = NULL;
#endif

	m_pServerSslCtx = NULL;

	m_SpConnLimit = 10;

	//    InitCyaSSL();
	CyaSSL_Init();

	m_AidIndex2Socket.clear();
}

SisTlsStack::~SisTlsStack()
{
	Clear();

	//    FreeCyaSSL();
	CyaSSL_Cleanup();
}


void SisTlsStack::RunSpConnChecker()
{
	char FN[32]; snprintf(FN, sizeof(FN), "%s::RunSpConnChecker]", m_name.c_str());

	// 1. 주기적 DB Polling
	// 2. Alarm & Clear

	bool b_first = true;
	SpExtMap    ChkList;
	SpExtMap    ChkList_old;
	SpExtMapIt  it;
	SpExtMapIt  it_old;

	int myCertID = SisIhData::instance()->getCertID();
	int cnt = 0;


	while (1)
	{
		sleep(SisIhData::instance()->getSpConnCheckTimer());

		ChkList.clear();

		OraDBApi rDB(SIS_SERVICE_DB);
		rDB.SetAutocommit(false);

		int iRet = rDB.GetSpAlarmInfo(ChkList, true);
		if (RET_DB_NOT_FOUND == iRet)
			EELOG(FN << "DB NOT_FOUND.");
		else if (RET_DB_SUCCESS != iRet)
			EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");

		if (ChkList.empty())
		{
			EELOG(FN << "SpWatchTbl is EMPTY");
			int iRet = rDB.CaCertCommit();
			if (RET_DB_NOT_FOUND == iRet)
				EELOG(FN << "DB NOT_FOUND.");
			else if (RET_DB_SUCCESS != iRet)
				EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");
			continue;
		}

		for (it = ChkList.begin(); it != ChkList.end(); it++)
		{
			{
				AxLock lock(m_Aid2SockLock);
				cnt = m_AidIndex2Socket.count(it->first); //APPID 별 connection count
			}
			if (it->second.m_AlarmFlag == 1 || it->second.m_AlarmFlag == 3)
			{
				if (cnt == 0) // 자신에게 없는 연결에 대하여
				{
					if ((it->second.m_SessID & myCertID) == myCertID) // 자신의 ID가 등록되어있는 경우 해제(비정상 종료 case)
					{
						int iRet = rDB.SetSpDisConnect(it->first, false, myCertID);
						if (RET_DB_NOT_FOUND == iRet)
							EELOG(FN << "DB NOT_FOUND.");
						else if (RET_DB_SUCCESS != iRet)
							EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");
						else
							it->second.m_SessID -= myCertID;
					}
					else if (it->second.m_SessID == 0 && it->second.m_NowAlarm == 0) // 어떤 SIS도 등록되어있지 않고 알람도 없다면 알람 발생
					{
						int iRet = rDB.SetSpAlarm(it->first);
						if (RET_DB_NOT_FOUND == iRet)
							EELOG(FN << "DB NOT_FOUND.");
						else if (RET_DB_SUCCESS != iRet)
							EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");
						else
							SisAlarmFault::AlarmDisConnSP(it->first, "Session ALL Disconnected");
					}
					else if (it->second.m_SessID != 0 && it->second.m_NowAlarm == 1) // 어떤 SIS가 등록되어 있는데 알람발생중이면 해제
					{
						int iRet = rDB.SetSpAlarmClear(it->first);
						if (RET_DB_NOT_FOUND == iRet)
							EELOG(FN << "DB NOT_FOUND.");
						else if (RET_DB_SUCCESS != iRet)
							EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");
						else
							SisAlarmFault::AlarmDisConnSPClear(it->first);
					}
				}
				else
				{
					if (it->second.m_SessID != 0 && it->second.m_NowAlarm == 1) // 자신에게 연동이 있을 경우 알람발생중이라면 해제
					{
						int iRet = rDB.SetSpAlarmClear(it->first);
						if (RET_DB_NOT_FOUND == iRet)
							EELOG(FN << "DB NOT_FOUND.");
						else if (RET_DB_SUCCESS != iRet)
							EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");
						else
							SisAlarmFault::AlarmDisConnSPClear(it->first);
					}
				}
			}
			else // ALARM_FLAG가 세션감지가 아닐 경우
			{
				if (it->second.m_NowAlarm == 1) // 연동에 상관없이 알람 해제
				{
					int iRet = rDB.SetSpAlarmClear(it->first);
					if (RET_DB_NOT_FOUND == iRet)
						EELOG(FN << "DB NOT_FOUND.");
					else if (RET_DB_SUCCESS != iRet)
						EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");
					else
						SisAlarmFault::AlarmDisConnSPClear(it->first);
				}
			}
		}
		{
			int iRet = rDB.CaCertCommit(); // DB Lock 해제를 위해 단순 Commit
			if (RET_DB_NOT_FOUND == iRet)
				EELOG(FN << "DB NOT_FOUND.");
			else if (RET_DB_SUCCESS != iRet)
				EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");

		}

		if (b_first) // 이전 데이터와 현재 데이터를 비교하기위한 설정
		{
			b_first = false;

			int iRet = rDB.GetSpAlarmInfo(ChkList, false);
			if (RET_DB_NOT_FOUND == iRet)
				EELOG(FN << "DB NOT_FOUND.");
			else if (RET_DB_SUCCESS != iRet)
				EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");

			ChkList_old = ChkList;
			continue;
		}
		else
		{
			ChkList.clear();
			int iRet = rDB.GetSpAlarmInfo(ChkList, false);
			if (RET_DB_NOT_FOUND == iRet)
				EELOG(FN << "DB NOT_FOUND.");
			else if (RET_DB_SUCCESS != iRet)
				EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");

		}
		rDB.CaCertCommit(); // DB Lock 해제를 위해 단순 Commit

							// 이전 데이터와 현재 데이터를 비교하여 타국사에 발생한 알람도 자국사에서 해소
		for (it = ChkList.begin(); it != ChkList.end(); it++)
		{
			if ((it_old = ChkList_old.find(it->first)) != ChkList_old.end())
			{
				if (it_old->second.m_NowAlarm == 0 && it->second.m_NowAlarm == 1)
				{
					// 어떤 SIS가 알람을 발생 시킨 경우. 자국사 일수도 있고, 타국사 일 수도 있다.
					// 어느곳이든 알람이 발생하였다면 되었다 라고 생각하고 넘어간다...
				}
				else if (it_old->second.m_NowAlarm == 1 && it->second.m_NowAlarm == 0)
				{
					// 어떤 SIS가 알람을 해제 시킨 경우. 자국사 일수도 있고, 타국사 일 수도 있다.
					// 자국사라면 상관 없겠지만 타국사에서 해제 된 경우, 자국사에 알람이 남을 수 있다.
					// 그러므로 있든 없든 AlarmClear를 한번 시도한다.
					SisAlarmFault::AlarmDisConnSPClear(it->first);
				}
			}
			else
			{
				// Maria DB 전환 이후 이 에러가 발생 할 수 있는 경우는 ReadOnly Repl이 중단 된 경우.
				EELOG(RED(FN << "AppID Mismatch when check the SP Connection Alarm"));
			}
		}

		ChkList_old = ChkList;
	}
}