#ifndef COMMLIB_ORACLE_ORA_SESSION_POOL_H_
#define COMMLIB_ORACLE_ORA_SESSION_POOL_H_

#include <list>
#include <string>
#ifndef _OCCI_H_INC_
//#define _OCCI_H_INC_
#include <occi.h>
using namespace oracle::occi;
#endif
#include "AxLib/AxLib.h"
#include "CommIh/ThreadService.h"
#include "Oracle/DBManager.h"
#include "Oracle/OraSession.h"


using namespace std;

class StatementInfo;
class DBManager;

typedef map<OraSession*, time_t>    OraSessTimeMap;
typedef OraSessTimeMap::value_type     OraSessTimeMapVt;

class OraSessionPool {
public:
	OraSession     *Obtain();
	void            Release(OraSession *pDB);

	bool            Initialize(const char *pUser, const char *pPasswd, const char *pConnectString, bool bOraUse, int iPoolSize = 30, map<int32_t, StatementInfo*> *pSQLDB = NULL);
	void            RunPoolManager();
	void            RunCheckKeepAlive();
	void            RunCheckerSessHangup();     // 2013.02.22
	void            WaitingAllSessionIdle();
	void            DisconnectAllSession();
	void            SetReConnSleepTime(int st) { m_ReConnSleepTime = st; }
	int             GetReConnSleepTime() { return m_ReConnSleepTime; }
	bool            isNormalSessionStatus(int rate);
	void            SetSessHangupTimer(int pValue) { m_SessHangupTimer = pValue; }
	void            SetThreadServicePtr(Ih::ThreadService* pIhService) { m_pIhService = pIhService; }
	void            SetDBUse(bool dbuse) { m_bOraUse = dbuse; }

	OraSessionPool();
	OraSessionPool(DBManager *pOwner, int pSessHangupTimer);
	~OraSessionPool();

private:
	OraSession                 *PopDownSession();
	OraSession                 *PopKeepAliveCheckSession();
	void                        PushIdleSession(OraSession *pDB);
	void                        PushDownSession(OraSession *pDB);

	AxWorker                   *m_pPoolManager;
	AxWorker                   *m_pKeepAliveChecker;
	AxWorker                   *m_pSessHangupChecker;

	string                      m_UserName;
	string                      m_Password;
	string                      m_ConnectString;
	int                         m_iPoolSize;
	int                         m_KeepAliveTime;
	int                         m_ReConnSleepTime;

	AxMutex                     m_lock;
	AxMutex                     m_lockDisConn;
	list<OraSession*>           m_IdleSessList;
	list<OraSession*>           m_DownSessList;
	OraSessTimeMap                 m_BusySessMap;
	bool                        m_bSQLDB;

	int                         m_SessHangupTimer;

	Ih::ThreadService*          m_pIhService;
	DBManager                  *m_pOwner;
	friend class                OraSession;

	bool                        m_bOraUse;

};

#endif 
