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

	// 1. �ֱ��� DB Polling
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
				cnt = m_AidIndex2Socket.count(it->first); //APPID �� connection count
			}
			if (it->second.m_AlarmFlag == 1 || it->second.m_AlarmFlag == 3)
			{
				if (cnt == 0) // �ڽſ��� ���� ���ῡ ���Ͽ�
				{
					if ((it->second.m_SessID & myCertID) == myCertID) // �ڽ��� ID�� ��ϵǾ��ִ� ��� ����(������ ���� case)
					{
						int iRet = rDB.SetSpDisConnect(it->first, false, myCertID);
						if (RET_DB_NOT_FOUND == iRet)
							EELOG(FN << "DB NOT_FOUND.");
						else if (RET_DB_SUCCESS != iRet)
							EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");
						else
							it->second.m_SessID -= myCertID;
					}
					else if (it->second.m_SessID == 0 && it->second.m_NowAlarm == 0) // � SIS�� ��ϵǾ����� �ʰ� �˶��� ���ٸ� �˶� �߻�
					{
						int iRet = rDB.SetSpAlarm(it->first);
						if (RET_DB_NOT_FOUND == iRet)
							EELOG(FN << "DB NOT_FOUND.");
						else if (RET_DB_SUCCESS != iRet)
							EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");
						else
							SisAlarmFault::AlarmDisConnSP(it->first, "Session ALL Disconnected");
					}
					else if (it->second.m_SessID != 0 && it->second.m_NowAlarm == 1) // � SIS�� ��ϵǾ� �ִµ� �˶��߻����̸� ����
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
					if (it->second.m_SessID != 0 && it->second.m_NowAlarm == 1) // �ڽſ��� ������ ���� ��� �˶��߻����̶�� ����
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
			else // ALARM_FLAG�� ���ǰ����� �ƴ� ���
			{
				if (it->second.m_NowAlarm == 1) // ������ ������� �˶� ����
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
			int iRet = rDB.CaCertCommit(); // DB Lock ������ ���� �ܼ� Commit
			if (RET_DB_NOT_FOUND == iRet)
				EELOG(FN << "DB NOT_FOUND.");
			else if (RET_DB_SUCCESS != iRet)
				EELOG(FN << "DB FAIL. Reason[" << rDB.GetReasonStr() << "]");

		}

		if (b_first) // ���� �����Ϳ� ���� �����͸� ���ϱ����� ����
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
		rDB.CaCertCommit(); // DB Lock ������ ���� �ܼ� Commit

							// ���� �����Ϳ� ���� �����͸� ���Ͽ� Ÿ���翡 �߻��� �˶��� �ڱ��翡�� �ؼ�
		for (it = ChkList.begin(); it != ChkList.end(); it++)
		{
			if ((it_old = ChkList_old.find(it->first)) != ChkList_old.end())
			{
				if (it_old->second.m_NowAlarm == 0 && it->second.m_NowAlarm == 1)
				{
					// � SIS�� �˶��� �߻� ��Ų ���. �ڱ��� �ϼ��� �ְ�, Ÿ���� �� ���� �ִ�.
					// ������̵� �˶��� �߻��Ͽ��ٸ� �Ǿ��� ��� �����ϰ� �Ѿ��...
				}
				else if (it_old->second.m_NowAlarm == 1 && it->second.m_NowAlarm == 0)
				{
					// � SIS�� �˶��� ���� ��Ų ���. �ڱ��� �ϼ��� �ְ�, Ÿ���� �� ���� �ִ�.
					// �ڱ����� ��� �������� Ÿ���翡�� ���� �� ���, �ڱ��翡 �˶��� ���� �� �ִ�.
					// �׷��Ƿ� �ֵ� ���� AlarmClear�� �ѹ� �õ��Ѵ�.
					SisAlarmFault::AlarmDisConnSPClear(it->first);
				}
			}
			else
			{
				// Maria DB ��ȯ ���� �� ������ �߻� �� �� �ִ� ���� ReadOnly Repl�� �ߴ� �� ���.
				EELOG(RED(FN << "AppID Mismatch when check the SP Connection Alarm"));
			}
		}

		ChkList_old = ChkList;
	}
}