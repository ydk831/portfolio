#include "Oracle/OraSessionPool.h"

#include "CommIh/DebugLogger.h"
#include "CommIh/IhUtil.h"
#include "Oracle/OraStateCache.h"
#include "Miscellaneous/TimeUtil.h"
#include "Primitive/GenAlarmFault.h"
#include "Oracle/DBManager.h"

OraSessionPool::OraSessionPool(DBManager *pOwner, int pSessHangupTimer)
{
	m_pKeepAliveChecker = NULL;
	m_pPoolManager = NULL;
	m_pSessHangupChecker = NULL;
	m_iPoolSize = 0;
	m_KeepAliveTime = 15;       // 15 Sec
	m_ReConnSleepTime = 0;        // 15 Sec
	m_bSQLDB = false;
	m_SessHangupTimer = pSessHangupTimer;
	m_pIhService = NULL;
	m_pOwner = pOwner;
}


OraSessionPool::~OraSessionPool()
{
	if (m_pKeepAliveChecker) {
		delete m_pKeepAliveChecker;
		m_pKeepAliveChecker = NULL;
	}

	if (m_pPoolManager) {
		delete m_pPoolManager;
		m_pPoolManager = NULL;
	}

	if (m_pSessHangupChecker) {
		delete m_pSessHangupChecker;
		m_pSessHangupChecker = NULL;
	}

	while (!m_IdleSessList.empty()) {
		if (m_IdleSessList.front())
			delete m_IdleSessList.front();
		m_IdleSessList.pop_front();
	}
	while (!m_DownSessList.empty()) {
		if (m_DownSessList.front())
			delete m_DownSessList.front();
		m_DownSessList.pop_front();
	}
}

bool OraSessionPool::Initialize(const char *pUser,
	const char *pPasswd,
	const char *pConnectString,
	bool bOraUse,
	int iPoolSize,
	map<int32_t, StatementInfo*> *pSQLDB)
{
	m_UserName = pUser;
	m_Password = pPasswd;
	m_ConnectString = pConnectString;
	m_iPoolSize = iPoolSize;
	m_ReConnSleepTime = 0;
	m_bOraUse = bOraUse;

	// R330
	if (pSQLDB != NULL) m_bSQLDB = true;
	else                m_bSQLDB = false;

	if (iPoolSize < 1)
		m_iPoolSize = 1;
	else if (m_iPoolSize > MAX_DB_POOL_SIZE)
		m_iPoolSize = MAX_DB_POOL_SIZE;

	for (int i = 0; i < m_iPoolSize; i++) {
		m_DownSessList.push_back(new OraSession(this, i, pSQLDB));
	}

	if (m_pPoolManager == NULL) {
		m_pPoolManager = AxCreateJob(this, &OraSessionPool::RunPoolManager);
		if (NULL == m_pPoolManager)
			return false;
		EPRINT("OraSessionPool::Initialize] m_pPoolManager Start!!");
	}
	else
		EPRINT("OraSessionPool::Initialize] m_pPoolManager Already Started!!");

	if (m_pKeepAliveChecker == NULL) {
		m_pKeepAliveChecker = AxCreateJob(this, &OraSessionPool::RunCheckKeepAlive);
		if (NULL == m_pKeepAliveChecker)
			return false;
		EPRINT("OraSessionPool::Initialize] m_pKeepAliveChecker Start!!");
	}
	else
		EPRINT("OraSessionPool::Initialize] m_pKeepAliveChecker Already Started!!");

	if (m_pSessHangupChecker == NULL) {
		m_pSessHangupChecker = AxCreateJob(this, &OraSessionPool::RunCheckerSessHangup);
		if (NULL == m_pSessHangupChecker)
			return false;
		EPRINT("OraSessionPool::Initialize] m_pSessHangupChecker Start!!");
	}
	else
		EPRINT("OraSessionPool::Initialize] m_pSessHangupChecker Already Started!!");

	return true;
}

void OraSessionPool::RunPoolManager()
{
	while (1)
	{
		//------------------------------------------------------------------------------------
		// DisconnectAllSession() 호출 후 바로 재접속하는 것을 막기 위하여 Sleep 시간을 줌
		// SharePlex 의 DB 동기화 시간이 Max 10초 이하이므로 해당 초 이상 Sleep 을 주기 위함 
		//------------------------------------------------------------------------------------

		if (m_ReConnSleepTime > 0) {
			int SleepTime = m_ReConnSleepTime;
			m_ReConnSleepTime = 0;
			sleep(SleepTime);
		}

		if (!m_bOraUse)
		{
			sleep(1);
			continue;
		}

		m_lockDisConn.Lock();

		OraSession *pDB = PopDownSession();
		if (pDB)
		{
			if (true)
			{
				AxLock lock(m_lock);
				m_BusySessMap[pDB] = time(0);
			}

			if (pDB->Connect(m_UserName, m_Password, m_ConnectString) == true)
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
		else
		{
			m_lockDisConn.Unlock();
			sleep(1);
		}
	}
}

void OraSessionPool::RunCheckKeepAlive()
{
	static const char FN[] = "OraSessionPool::RunCheckKeepAlive] ";
	static const string SQL = "select to_char(sysdate) from dual";
	OraSession *pDB;
	string sNow;

	while (1) {
		sleep(1);
		if (!m_bOraUse)
			continue;

		pDB = PopKeepAliveCheckSession();
		if (pDB) {
			bool bIsOk = true;

			try {
				int  iCount = 0;
				char sysdate[64];
				ub2  len[1];

				if (m_bSQLDB == true) {
					OraStatement stmt(0, pDB);
					ResultSet    *rs = stmt.ExecuteQuery();
					rs->setDataBuffer(1, sysdate, OCCI_SQLT_CHR, sizeof(sysdate), &len[0], NULL, NULL);
					while (rs->next()) {
						++iCount;
						sysdate[len[0]] = 0x00;
						sNow = sysdate;
					}
				}
				else {
					OraStatement stmt(SQL, pDB);
					ResultSet    *rs = stmt.ExecuteQuery();
					rs->setDataBuffer(1, sysdate, OCCI_SQLT_CHR, sizeof(sysdate), &len[0], NULL, NULL);
					while (rs->next()) {
						++iCount;
						sysdate[len[0]] = 0x00;
						sNow = sysdate;
					}
				}

				if (iCount) {
					bIsOk = true;
					DPRINT(FN << "ID:" << pDB->GetID() << " OK. time:" << sNow);
				}
				else {
					bIsOk = false;
					EPRINT(FN << "ID:" << pDB->GetID() << " sysdate OOPS!!");
				}
			}
			catch (SQLException ex) {
				bIsOk = false;
				EPRINT(FN << "ID:" << pDB->GetID() << " code:" << ex.getErrorCode() << "," << ex.what() << SQL << "\n");
			}

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

void OraSessionPool::RunCheckerSessHangup()
{
	static const char FN[] = "OraSessionPool::RunCheckerSessHangup] ";

	while (true)
	{
		sleep(2);
		if (!m_bOraUse)
			continue;

		OraSessTimeMap::iterator   it, ittmp;
		OraSession              *pSess;
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
					if ((pSess->m_pConn != NULL) && (pSess->m_pEnv != NULL))
					{
						EPRINT("[DEADLOCK-DETECTED] ORA-SESSION IS PENDING FOR " << m_SessHangupTimer << " SECONDS (QRY-TM : " << TimeUtil::ConverTimeString(pSess->GetQrySttTime(), 0) << "). BREAK QUERY.");
						pSess->SetAfterDown();
						OCIBreak((dvoid*)pSess->m_pConn->getOCIServiceContext(), (OCIError*)pSess->GetOCIError());
						ittmp = it;
						++it;
						m_BusySessMap.erase(ittmp);
						bDel = true;
					}
					else
					{
						EPRINT("[DEADLOCK-DETECTED] ORA-SESSION IS PENDING FOR " << m_SessHangupTimer << " SECONDS (QRY-TM : " << TimeUtil::ConverTimeString(pSess->GetQrySttTime(), 0) << "). BREAK QUERY. CONNECT EXCEPTION.");
					}

					if (m_pIhService != NULL) {
						GenAlarmFault *pIhPrim = new GenAlarmFault();
						pIhPrim->m_nAlarmType = GEN_ALARM_FAULT_DB_PENDING;
						pIhPrim->m_strParam1 = Ih::IhUtil::itos(pSess->GetID());
						if (m_pIhService->insert(pIhPrim) == false)
							delete pIhPrim;
						else
							EPRINT("[DEADLOCK-DETECTED] ORA-SESSION IS PENDING. NOTICE ALARM ID " << pSess->GetID());
					}
					else
						EPRINT("[DEADLOCK-DETECTED] ORA-SESSION IS PENDING. CAN'T NOTICE ALARM ID " << pSess->GetID());
				}
			}
			if (bDel == false)
				++it;
		}
	}
}