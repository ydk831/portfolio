#ifndef DB_MYSQLAPI_SESSION_POOL_H_
#define DB_MYSQLAPI_SESSION_POOL_H_

#include <list>
#include <string>

#include "AxLib/AxLib.h"
#include "CommIh/ThreadService.h"
#include "Mysql/DBManager.h"
#include "Mysql/MySession.h"

using namespace std;

class StatementInfo;
class DBManager;
class MySession;

typedef map<MySession*, time_t> SessTimeMap;
typedef SessTimeMap::value_type SessTimeMapVt;

class MySessionPool {
public:
	MySessionPool();
	MySessionPool(DBManager *pOwner, int SessHangupTimer);
	~MySessionPool();

	//MySession            *Obtain();
	MySession            *Obtain(bool isSelectSQL = false, bool bUseRD = false);
	void                 Release(MySession *pDB);
	bool                 IsExistIdleSession();

	bool                 Initialize(
		const char *pMyHost, __u_short nMyPort
		, const char *pUser, const char *pPasswd, const char *pDatabase
		, const char *pRDHost, __u_short nRDPort
		, const char *pFailOverHost, __u_short nFailOverPort
		, const char *pFailOverRDHost, __u_short nFailOverRDPort
		, bool bMyUse
		, int nPoolSize = 30
		, int nTimeout = 0);

	void                 RunPoolManager();
	void                 RunCheckKeepAlive();
	void                 RunCheckSessHangup();

	void                 RunRDCheckKeepAlive();

	void                 GetPoolInfo(__uint32_t &rTotal, __uint32_t &rIdle, __uint32_t &rRDIdle, __uint32_t &rDown);
	void                 SetThreadServicePtr(Ih::ThreadService* pIhService) { m_pIhService = pIhService; }

	void                 SetReConnSleepTime(int st) { m_ReConnSleepTime = st; }
	int                  GetReConnSleepTime() { return m_ReConnSleepTime; }

	void                 SetSessHangupTimer(int pValue) { m_SessHangupTimer = pValue; }

	void                 DisconnectAllSession();
	void                 DisconnectRDSession();
	void                 WaitingAllSessionIdle();
	bool                 isNormalSessionStatus(int rate, bool &RDStatus);
	void                 SetDBUse(bool dbuse) { m_bMyUse = dbuse; }
	void                 SetSiteFailover(bool bSiteFailOver) { m_bSiteFailOver = bSiteFailOver; }
	void                 setMyRDUsed(bool bRdUsed) { m_bMyRDUsed = bRdUsed; }
	bool                 getMyRDUsed() { return m_bMyRDUsed; }

	//int                GetID()         { return m_ID; }

private:
	DBManager            *m_pOwner;
	Ih::ThreadService*   m_pIhService;           /* R360. */
	friend class         MySession;

	MySession            *PopDownSession();
	MySession            *PopKeepAliveCheckSession();
	MySession            *PopKeepAliveCheckRDSession();
	void                 PushIdleSession(MySession *pDB);
	void                 PushIdleRDSession(MySession *pDB);
	void                 PushDownSession(MySession *pDB);

	AxWorker             *m_pPoolManager;
	AxWorker             *m_pKeepAliveChecker;
	AxWorker             *m_pRDKeepAliveChecker;
	AxWorker             *m_pSessHangupChecker;

	std::string          m_LTag;

	std::string          m_ConnMyHost;
	__u_short            m_ConnMyPort;
	std::string          m_ConnRDHost;
	__u_short            m_ConnRDPort;
	std::string          m_ConnFailOverHost;
	__u_short            m_ConnFailOverPort;
	std::string          m_ConnFailOverRDHost;
	__u_short            m_ConnFailOverRDPort;
	std::string          m_ConnUser;
	std::string          m_ConnPasswd;
	std::string          m_ConnDatabase;
	int                  m_ConnTimeout;
	bool                 m_bMyUse;
	bool                 m_bSiteFailOver;

	int                  m_iPoolSize;
	int                  m_KeepAliveTime;
	int                  m_SessHangupTimer;

	AxMutex                m_lock;
	AxMutex                m_lockDisConn;
	std::list<MySession*>  m_IdleSessList;
	std::list<MySession*>  m_IdleRDSessList;
	std::list<MySession*>  m_DownSessList;

	SessTimeMap            m_BusySessMap;          // 2013.02.22
	int                    m_ReConnSleepTime;
	bool                   m_bMyRDUsed;
};

#endif 
