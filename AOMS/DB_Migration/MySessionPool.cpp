#include "AxLib/AxLock.h"
#include "AxLib/AxUtil.h"
#include "AxLib/AxString.h"
#include "CommIh/DebugLogger.h"
#include "CommIh/IhUtil.h"
#include "CommIh/IhGroupServer.h"
#include "Miscellaneous/TimeUtil.h"
#include "Primitive/GenAlarmFault.h"

#include "Mysql/DBManager.h"
#include "Mysql/MyStatement.h"
#include "Mysql/MySessionPool.h"



MySessionPool::MySessionPool(DBManager *pOwner, int pSessHangupTimer)
{
	m_pKeepAliveChecker = NULL;
	m_pRDKeepAliveChecker = NULL;
	m_pPoolManager = NULL;
	m_iPoolSize = 0;
	m_KeepAliveTime = 15;       // 10 Sec
	m_SessHangupTimer = pSessHangupTimer;
	m_ReConnSleepTime = 0;
	m_pIhService = NULL;
	m_pOwner = pOwner;
	m_bMyUse = false;
	m_bMyRDUsed = false;
	m_bSiteFailOver = false;
}

MySessionPool::~MySessionPool()
{
	if (m_pKeepAliveChecker) {
		delete m_pKeepAliveChecker;
		m_pKeepAliveChecker = NULL;
	}

	if (m_pPoolManager) {
		delete m_pPoolManager;
		m_pPoolManager = NULL;
	}

	while (!m_IdleSessList.empty()) {
		delete m_IdleSessList.front();
		m_IdleSessList.pop_front();
	}
	while (!m_IdleRDSessList.empty()) {
		delete m_IdleRDSessList.front();
		m_IdleRDSessList.pop_front();
	}
	while (!m_DownSessList.empty()) {
		delete m_DownSessList.front();
		m_DownSessList.pop_front();
	}
}

bool MySessionPool::Initialize(
	const char *pMyHost, __u_short nMyPort
	, const char *pUser, const char *pPasswd, const char *pDatabase
	, const char *pRDHost, __u_short nRDPort
	, const char *pFailOverHost, __u_short nFailOverPort
	, const char *pFailOverRDHost, __u_short nFailOverRDPort
	, bool bMyUse
	, int nPoolSize
	, int nTimeout
	)
{
	static const char *FN = "[MySessionPool::Init] ";

	if (AX_ISZERO(pMyHost) || AX_ISZERO(pRDHost) || AX_ISZERO(pFailOverHost) || AX_ISZERO(pFailOverRDHost)) {
		ELOG(FN << "Maria HOST Value FAIL. Master(" << pMyHost << "), ReadOnly(" << pRDHost << "), FailOver(" << pFailOverHost << "), FailOverReadOnly(" << pFailOverRDHost << ")");
		return false;
	}

	if ((0 == nMyPort) || (0 == nRDPort) || (0 == nFailOverPort) || (0 == nFailOverRDPort)) {
		ELOG(FN << "Maria PORT FAIL. Master(" << nMyPort << "), ReadOnly(" << nRDPort << "), FailOver(" << nFailOverPort << "), FailOverReadOnly(" << nFailOverRDPort << ")");
		return false;
	}

	if (AX_ISZERO(pUser)) {
		ELOG(FN << "User(null) FAIL");
		return false;
	}

	if (AX_ISZERO(pPasswd)) {
		ELOG(FN << "Passwd(null) FAIL");
		return false;
	}

	if (AX_ISZERO(pDatabase)) {
		ELOG(FN << "Database(null) FAIL");
		return false;
	}

	m_ConnMyHost = pMyHost;
	m_ConnMyPort = nMyPort;
	m_ConnRDHost = pRDHost;
	m_ConnRDPort = nRDPort;
	m_ConnFailOverHost = pFailOverHost;
	m_ConnFailOverPort = nFailOverPort;
	m_ConnFailOverRDHost = pFailOverRDHost;
	m_ConnFailOverRDPort = nFailOverRDPort;
	m_ConnUser = pUser;
	m_ConnPasswd = pPasswd;
	m_ConnDatabase = pDatabase;
	m_ConnTimeout = nTimeout;
	m_iPoolSize = nPoolSize;
	m_bMyUse = bMyUse;


	{
		AxString STR;
		STR.Csnprintf(256, "Master:%s:%u;%s, ReadOnly:%s:%u;%s , FailOver:%s:%u;%s, FailOverReadOnly:%s:%u;%s",
			m_ConnMyHost.c_str(), m_ConnMyPort, m_ConnUser.c_str(),
			m_ConnRDHost.c_str(), m_ConnRDPort, m_ConnUser.c_str(),
			m_ConnFailOverHost.c_str(), m_ConnFailOverPort, m_ConnUser.c_str(),
			m_ConnFailOverRDHost.c_str(), m_ConnFailOverRDPort, m_ConnUser.c_str()
			);
		m_LTag = STR.c_str();
	}

	if (1 > m_iPoolSize) {
		m_iPoolSize = 0;
	}
	else if (MAX_DB_POOL_SIZE < m_iPoolSize) {
		m_iPoolSize = MAX_DB_POOL_SIZE;
	}

	for (int i = 0; i < m_iPoolSize; i++) {
		m_DownSessList.push_back(new MySession(this, i));       // Master Session
		m_DownSessList.push_back(new MySession(this, i, true)); // ReadOnly Session
	}
	m_iPoolSize *= 2; //ReadOnly ���� ���Ƕ����� �ʱ�ȭ �� *2 ����. Disconnect �� ����count ���߱� ����.

	m_pPoolManager = AxCreateJob(this, &MySessionPool::RunPoolManager);
	if (NULL == m_pPoolManager) {
		return false;
	}
	EELOG(FN << m_LTag << " RunPoolManager Start!!");

	m_pKeepAliveChecker = AxCreateJob(this, &MySessionPool::RunCheckKeepAlive);
	if (NULL == m_pKeepAliveChecker) {
		return false;
	}
	EELOG(FN << m_LTag << " KeepAliveChecker Start!!");

	m_pSessHangupChecker = AxCreateJob(this, &MySessionPool::RunCheckSessHangup);
	if (NULL == m_pSessHangupChecker) {
		return false;
	}
	EELOG(FN << m_LTag << " SessHangupChecker Start!!");

	// Below ReadOnly Checker //
	m_pRDKeepAliveChecker = AxCreateJob(this, &MySessionPool::RunRDCheckKeepAlive);
	if (NULL == m_pRDKeepAliveChecker) {
		return false;
	}
	EELOG(FN << m_LTag << " ReadOnlyKeepAliveChecker Start!!");

	return true;
}

void MySessionPool::RunPoolManager()
{
	static const char *FN = "[MySessionPool::RunPoolManager] ";
	while (1)
	{
		if (m_ReConnSleepTime > 0) {
			int SleepTime = m_ReConnSleepTime;
			m_ReConnSleepTime = 0;
			sleep(SleepTime);
		}

		if (!m_bMyUse)
		{
			axSleep(100);
			continue;
		}

		string MasterHost, ReadOnlyHost;
		__u_short MasterPort, ReadOnlyPort;

		if (m_bSiteFailOver)
		{
			MasterHost = m_ConnFailOverHost;
			MasterPort = m_ConnFailOverPort;
			ReadOnlyHost = m_ConnFailOverRDHost;
			ReadOnlyPort = m_ConnFailOverRDPort;
		}
		else
		{
			MasterHost = m_ConnMyHost;
			MasterPort = m_ConnMyPort;
			ReadOnlyHost = m_ConnRDHost;
			ReadOnlyPort = m_ConnRDPort;
		}

		m_lockDisConn.Lock();

		MySession *pDB = PopDownSession();

		if (pDB) {
			if (!(pDB->isReadOnly())) // Master Session Connect First
			{
				if (true)
				{
					//2013.02.22 for connect deadlock
					AxLock lock(m_lock);
					m_BusySessMap[pDB] = time(0);
				}

				if (pDB->Connect(
					MasterHost.c_str(), MasterPort
					, m_ConnUser.c_str(), m_ConnPasswd.c_str(), m_ConnDatabase.c_str()
					, m_ConnTimeout))
				{
					PushIdleSession(pDB);
					m_lockDisConn.Unlock();
				}
				else
				{
					PushDownSession(pDB);
					m_lockDisConn.Unlock();
					sleep(1);
				}
			}
			else if (m_bMyRDUsed) // ReadOnly ����� ��� Connect
			{
				if (pDB->Connect(
					ReadOnlyHost.c_str(), ReadOnlyPort
					, m_ConnUser.c_str(), m_ConnPasswd.c_str(), m_ConnDatabase.c_str()
					, m_ConnTimeout))
				{
					PushIdleRDSession(pDB);
					m_lockDisConn.Unlock();
				}
				else
				{
					PushDownSession(pDB);
					m_lockDisConn.Unlock();
					sleep(1);
				}
			}
			else // ReadOnly ������� ���� ��� mutex Lock ���� �� Session ���Ÿ� ���� �κ�.
			{
				PushDownSession(pDB);
				m_lockDisConn.Unlock();
				usleep(100000);
				// 2018.09.21 YDK. ReadOnly ������� �ʴ� ���, �� �κ� ������ ���� �⵿ �� Master Connection�� 1�� �������� �����. sleep ����.
				// 2018.12.11 YDK. sleep ���� �� readonly �̻�� case���� cpu ���ϰ� �߻�(no sleep ���ѷ���), 0.1�� �������� �����ϵ��� ����.
			}
		}
		else
		{
			m_lockDisConn.Unlock();
			sleep(1);
		}
	}
}

void MySessionPool::RunCheckKeepAlive()
{
	static const char *FN = "[MySessionPool::KeepAlive] ";
	string SQL = "select now()";
	MySession  *pDB;
	char      NOW[64];
	string    now;

	while (1)
	{
		sleep(1);

		pDB = PopKeepAliveCheckSession();
		if (pDB) {
			bool bIsOk = false;
			do {
				MyStatement stmt(SQL, pDB);

				// Bind-ResultSet
				stmt.BindResultSetString(1, NOW, sizeof(NOW));

				// Execute-Query
				int dbRet = stmt.ExecuteQuery();
				if (RET_DB_SUCCESS != dbRet) {
					EEOUT(FN << pDB->GetSessInfo() << " ExecuteQuery OOPS!! " << stmt.ERRORSTR());
					break;
				}

				// Fetch-Data
				dbRet = stmt.Fetch();
				if (RET_DB_SUCCESS != dbRet) {
					EEOUT(FN << pDB->GetSessInfo() << " Fetch OOPS!! " << stmt.ERRORSTR());
					break;
				}

				bIsOk = true;
				now = NOW;
				DDOUT(FN << pDB->GetSessInfo() << " Fetch time: " << now);
			} while (0);

			if (bIsOk) {
				pDB->Used();
				PushIdleSession(pDB);
			}
			else {
				PushDownSession(pDB);
			}
		}
	}
}

void MySessionPool::RunRDCheckKeepAlive()
{
	static const char *FN = "[MySessionPool::RDKeepAlive] ";
	std::string SQL = "select now()";

	MySession  *pDB;
	char      NOW[64];
	string    now;

	while (1)
	{
		sleep(1);
		if (!m_bMyRDUsed) // ReadOnly ������� ���� ��� KA ���� ����.
			continue;

		pDB = PopKeepAliveCheckRDSession();
		if (pDB) {

			bool bIsOk = false;

			do {
				MyStatement stmt(SQL, pDB);

				// Bind-ResultSet
				stmt.BindResultSetString(1, NOW, sizeof(NOW));

				// Execute-Query
				int dbRet = stmt.ExecuteQuery();
				if (RET_DB_SUCCESS != dbRet) {
					EEOUT(FN << pDB->GetSessInfo() << " ExecuteQuery OOPS!! " << stmt.ERRORSTR());
					break;
				}

				// Fetch-Data
				dbRet = stmt.Fetch();
				if (RET_DB_SUCCESS != dbRet) {
					EEOUT(FN << pDB->GetSessInfo() << " Fetch OOPS!! " << stmt.ERRORSTR());
					break;
				}

				bIsOk = true;
				now = NOW;
				DDOUT(FN << pDB->GetSessInfo() << " Fetch time: " << now);
			} while (0);

			if (bIsOk) {
				pDB->Used();
				PushIdleRDSession(pDB);
			}
			else {
				PushDownSession(pDB);
			}
		}
	}
}

void MySessionPool::RunCheckSessHangup()
{
	static const char FN[] = "[MySessionPool::RunCheckSessHangup] ";

	while (true)
	{
		sleep(2);

		SessTimeMap::iterator   it, ittmp;
		MySession               *pSess;
		MySession               *KillSess;
		string                  MasterHost, ReadOnlyHost;
		__u_short               MasterPort, ReadOnlyPort;
		AxString                SQL;
		time_t                  tmcurr;
		AxLock                  lock(m_lock);

		time(&tmcurr);
		for (it = m_BusySessMap.begin(); it != m_BusySessMap.end(); )
		{
			bool bDel = false;
			if ((pSess = it->first) != NULL)
			{
				if (pSess->GetQrySttTime() == 0)
				{
					ittmp = it;
					++it;
					m_BusySessMap.erase(ittmp);
					bDel = true;
				}
				else if (m_SessHangupTimer == 0) {}
				else if (pSess->GetQrySttTime() + m_SessHangupTimer <= tmcurr)
				{
					if ((pSess->m_pConn != NULL)) // Hangup Session�� �����ϴ� ���
					{
						EPRINT("[DEADLOCK-DETECTED] MY-SESSION IS PENDING FOR " << m_SessHangupTimer << " SECONDS (QRY-TM : " << TimeUtil::ConverTimeString(pSess->GetQrySttTime(), 0) << "). BREAK QUERY SESSION. (ID:" << pSess->m_ConnID << ")" << ", SQL: " << pSess->m_SQL);

						if (m_bSiteFailOver) // DB ��ü �Ǵ�
						{
							MasterHost = m_ConnFailOverHost;
							MasterPort = m_ConnFailOverPort;
							ReadOnlyHost = m_ConnFailOverRDHost;
							ReadOnlyPort = m_ConnFailOverRDPort;
						}
						else
						{
							MasterHost = m_ConnMyHost;
							MasterPort = m_ConnMyPort;
							ReadOnlyHost = m_ConnRDHost;
							ReadOnlyPort = m_ConnRDPort;
						}

						if (pSess->isReadOnly()) // Master/Slave �Ǵ��Ͽ� Connect
						{
							KillSess = new MySession(this, 999, true);
							if (KillSess == NULL) // Allocate �����ϸ� �׳� Pass. ���� �Ͽ� ���� ���� ����Ѵ�.
							{
								EPRINT(FN << "Killing Session Allocate Fail");
								continue;
							}

							if (!KillSess->Connect(
								ReadOnlyHost.c_str(), ReadOnlyPort
								, m_ConnUser.c_str(), m_ConnPasswd.c_str(), m_ConnDatabase.c_str()
								, m_ConnTimeout)) // Connect �����ϸ� �׳� Pass. ���� �Ͽ� ���� ���� ����Ѵ�.
							{
								EPRINT(FN << "Killing Session Connect Fail.");
								delete KillSess;
								continue;
							}

						}
						else
						{
							KillSess = new MySession(this, 999);
							if (KillSess == NULL)
							{
								EPRINT(FN << "Killing Session Allocate Fail.");
								continue;
							}

							if (!KillSess->Connect(
								MasterHost.c_str(), MasterPort
								, m_ConnUser.c_str(), m_ConnPasswd.c_str(), m_ConnDatabase.c_str()
								, m_ConnTimeout))
							{
								EPRINT(FN << "Killing Session Allocate Fail.");
								delete KillSess;
								continue;
							}
						}

						SQL.Csnprintf(1024, "KILL %lu ", pSess->m_ConnID); // Session KILL�� ���� SQL
																		   //SQL.Csnprintf(1024, "KILL QUERY %lu ", pSess->m_ConnID); // KILL QUERY �� ���� SQL
						MyStatement mystmt(SQL.c_log(), KillSess);
						if (mystmt.ExecuteQuery() == RET_DB_SUCCESS)
							EPRINT("[DEADLOCK-DETECTED] Kill Query Success");
						else
						{
							EPRINT("[DEADLOCK-DETECTED] Kill Query Fail"); // KILL QUERY SQL�� �����ϸ� ������ ���Ǹ� Delete �ϰ� BusySessMap������ ������ �ʴ´�. ���� �����Ͽ� ������� ���.
							delete KillSess;
							continue;
						}

						//pSess->SetAfterDown();
						ittmp = it;
						++it;
						m_BusySessMap.erase(ittmp);
						bDel = true;
						delete KillSess;
					}
					else
					{
						EPRINT("[DEADLOCK-DETECTED] MY-SESSION IS PENDING FOR " << m_SessHangupTimer << " SECONDS (QRY-TM : " << TimeUtil::ConverTimeString(pSess->GetQrySttTime(), 0) << "). BREAK QUERY. CONNECT EXCEPTION.");
					}

					/* R360 DB Pending Alarm. */
					if (m_pIhService != NULL) {
						GenAlarmFault *pIhPrim = new GenAlarmFault();
						pIhPrim->m_nAlarmType = GEN_ALARM_FAULT_DB_PENDING;
						pIhPrim->m_strParam1 = Ih::IhUtil::itos(pSess->GetID());
						if (m_pIhService->insert(pIhPrim) == false)
							delete pIhPrim;
						else
						{
							EPRINT("[DEADLOCK-DETECTED] MY-SESSION IS PENDING. NOTICE ALARM ID " << pSess->GetID());
						}
					}
					else
					{
						EPRINT("[DEADLOCK-DETECTED] MY-SESSION IS PENDING. CAN'T NOTICE ALARM ID " << pSess->GetID());
					}
				}
			}
			if (bDel == false)
				++it;
		}
	}
}
