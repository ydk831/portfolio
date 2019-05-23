#include "CommIh/MutexLock.h"
#include "CommIh/DebugLogger.h"
#include "AxLib/AxLib.h"

#include "Mysql/DBManager.h"

#ifndef _DB_TEST_
#include "Oracle/OraSessionPool.h"
#endif

#include "Mysql/MySessionPool.h"

bool DBManager::AddRange(const char *pStart,
	const char *pEnd,
	const char *pUser,
	const char *pPasswd,
	const char *pConnectString,
	const char *pMyDatabase,
	const char *pMyHost, const char *pMyRDHost,                    // R430 추가
	const char *pMyFailOverHost, const char *pMyFailOverRDHost,    // MARIA DB 연결 정보 및 
	__u_short MyPort, __u_short MyRDPort,                           //
	__u_short MyFailOverPort, __u_short MyFailOverRDPort,          // ORACLE & MARIA 사용시
	DBMODE DBMode, DBMODE ChoiceDB,                                // CONFIG 정보
	int maxpool,
#ifndef _DB_TEST_
	map<int32_t, StatementInfo*> *pSQLDB,
#endif
	int32_t pSessHangupTimer)
{
	setDBMode(DBMode, ChoiceDB);
	stDBServer *pDBServer = new stDBServer;

	pDBServer->m_id = ++m_index;
	pDBServer->m_Range_Start = pStart;
	pDBServer->m_Range_End = pEnd;

#ifndef _DB_TEST_
	pDBServer->m_OraDBPool = new OraSessionPool(this, pSessHangupTimer);
	pDBServer->m_OraDBPool->Initialize(pUser, pPasswd, pConnectString, m_bOraUse, maxpool, pSQLDB);
#endif

	//R430
	pDBServer->m_MyDBPool = new MySessionPool(this, pSessHangupTimer);
	pDBServer->m_MyDBPool->Initialize(pMyHost, MyPort, pUser, pPasswd, pMyDatabase,
		pMyRDHost, MyRDPort,
		pMyFailOverHost, MyFailOverPort,
		pMyFailOverRDHost, MyFailOverRDPort,
		m_bMyUse, maxpool, pSessHangupTimer);
	m_DBServerTable.push_back(pDBServer);
	EPRINT("DBManager::AddRange] create DB. " << pDBServer->m_Range_Start << "-" << pDBServer->m_Range_End << " size : " << m_DBServerTable.size());
	return true;
}

#ifndef _DB_TEST_
void DBManager::FindDBServer(const char *pPrefix, OraSessionPool *&pOraPool, MySessionPool *&pMyPool)
#else
void DBManager::FindDBServer(const char *pPrefix, MySessionPool *&pMyPool)
#endif
{
	list<stDBServer*>::iterator iter = m_DBServerTable.begin();
	if (*iter == NULL)
	{
		EPRINT(YELLOW("DBServerTable NULL...???????????????????"));
		return;
	}
	for (; iter != m_DBServerTable.end(); iter++) {
		if ((axStrcmpRange((*iter)->m_Range_Start.c_str(), pPrefix) >= 0)
			&& (axStrcmpRange((*iter)->m_Range_End.c_str(), pPrefix) <= 0))
		{
			//IPRINT("DBManager::FindDBServer] " << pPrefix << " DBPool:" << (*iter)->m_id);
			break;
		}
	}

	if (iter == m_DBServerTable.end()) {
		EPRINT("DBManager::FindDBServer] not found. " << pPrefix);
		iter = m_DBServerTable.begin();
	}

	if (iter == m_DBServerTable.end())
		//return NULL;
	{
#ifndef _DB_TEST_
		pOraPool = NULL;
#endif
		pMyPool = NULL;
		return;
	}

	//return (*iter)->m_OraDBPool;
#ifndef _DB_TEST_ 
	pOraPool = (*iter)->m_OraDBPool;
#endif
	pMyPool = (*iter)->m_MyDBPool;
}

void DBManager::ChangeDBMode(DBMODE pDBMode, DBMODE pChoiceDB, const char* pPrefix)
{
	DPRINT(GREEN("[DBManager::ChangeDBMode()] " << GETDBSTR(m_DBMode) << "-> " << GETDBSTR(pDBMode) << ", " << GETDBSTR(m_ChoiceDB) << "->" << GETDBSTR(pChoiceDB)));
	if (pDBMode == m_DBMode)
	{
		if ((pDBMode == T_BOTH) && (pChoiceDB != m_ChoiceDB))
		{
			EPRINT("[DBManager::ChangeDBMode()] Changed ChoiceDB value. Before(" << GETDBSTR(m_ChoiceDB) << "), After(" << GETDBSTR(pChoiceDB) << ")");
		}
	}

#ifndef _DB_TEST_
	OraSessionPool *pOraPool = NULL;
#endif
	MySessionPool *pMyPool = NULL;

#ifndef _DB_TEST_
	FindDBServer(pPrefix, pOraPool, pMyPool);
#else
	FindDBServer(pPrefix, pMyPool);
#endif

#ifndef _DB_TEST_
	switch (pDBMode)
	{
	case T_ORACLE:
		m_DBMode = pDBMode;
		m_ChoiceDB = pDBMode;
		if (pOraPool != NULL) pOraPool->SetDBUse(true);
		if (pMyPool != NULL)
		{
			pMyPool->SetDBUse(false);
			pMyPool->DisconnectAllSession();
		}
		break;
	case T_MYSQL:
		m_DBMode = pDBMode;
		m_ChoiceDB = pDBMode;
		if (pMyPool != NULL) pMyPool->SetDBUse(true);
		if (pOraPool != NULL)
		{
			pOraPool->SetDBUse(false);
			pOraPool->DisconnectAllSession();
		}
		break;
	case T_BOTH:
		if (pChoiceDB != T_ORACLE && pChoiceDB != T_MYSQL)
		{
			EPRINT("[DBManager::ChangeDBMode()] invalide ChoiceDB value.(" << pChoiceDB << ")");
			return;
		}

		m_DBMode = pDBMode;
		m_ChoiceDB = pChoiceDB;
		if (pMyPool != NULL) pMyPool->SetDBUse(true);
		if (pOraPool != NULL) pOraPool->SetDBUse(true);
		break;
	default:
		EPRINT("[DBManager::ChangeDBMode()] invaliad DBMODE value.(" << pDBMode << ")");
		break;
	}
#endif
}


