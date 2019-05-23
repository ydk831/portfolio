#ifndef COMMLIB_ORACLE_DBMANGER_H_
#define COMMLIB_ORACLE_DBMANGER_H_

#include <string>
#include <list>
#ifndef _OCCI_H_INC_
#include <occi.h>
#endif
#include <map>
#include "AxLib/AxLib.h"

#ifndef _MYSQL_H_INC_
#define _MYSQL_H_INC_
#include <mysql.h>
#endif

#define MAX_DB_POOL_SIZE 200
#define GETDBSTR(X)  ((X==T_BOTH)?"BOTH":((X==T_ORACLE)?"ORACLE":"MARIA"))
#define IS_ORA_TYPE(X) (X == T_ORACLE)
#define IS_MARIA_TYPE(X) (X == T_MYSQL)
#define IS_BOTH_TYPE(X) (X == T_BOTH)

using namespace std;
#ifndef _OCCI_H_INC_
using namespace oracle::occi;
#endif

#ifndef _DB_TEST_
class OraSession;
class OraSessionPool;
class StatementInfo;
#endif

class MySession;
class MySessionPool;
class MyStatement;

namespace Ih
{
	class ThreadService;
}

struct stDBServer {
	int             m_id;
	std::string     m_Range_Start;
	std::string     m_Range_End;
#ifndef _DB_TEST_
	OraSessionPool *m_OraDBPool;
#endif
	MySessionPool  *m_MyDBPool;

	stDBServer() {
#ifndef _DB_TEST_
		m_OraDBPool = NULL;
#endif
		m_MyDBPool = NULL;
		m_id = 0;
	}
};

typedef enum {
	T_ORACLE,
	T_MYSQL,
	T_BOTH
} DBMODE;

class DBManager {
public:
	static DBManager   *instance();

#ifndef _DB_TEST_
	void                Obtain(const char *pPrefix, OraSession *&pOraDB, MySession *&pMyDB, bool isSelectSQL = false, bool bRDOnlyUse = false);
#else
	void                Obtain(const char *pPrefix, MySession *&pMyDB, bool isSelectSQL = false, bool bRDOnlyUse = false);
#endif

#ifndef _DB_TEST_
	void                Release(const char *pPrefix, OraSession *pOraDB, MySession *pMyDB);
#else
	void                Release(const char *pPrefix, MySession *pMyDB);
#endif

	int                 GetPoolCount();
	const char         *GetPoolPrefix(int index);

	void                DisconnectAllSession(int nSleepTime, const char* pPrefix = NULL, bool bMysqlFailover = false);
	void                DisconnectRDSession(const char* pPrefix);
	bool                isNormalSessionStatus(const char* pfx, int rate, bool &RDStatus);
	void                SetSessHangupTimer(int pValue);

	void                SetThreadServicePtr(Ih::ThreadService* pIhService);

public:
	
	bool AddRange(const char *pStart, const char *pEnd,
		const char *pUser, const char *pPasswd,
		const char *pConnectString, const char *pMyDatabase,
		const char *pMyHost, const char *pMyRDHost,
		const char *pMyFailOverHost, const char *pMyFailOverRDHost,
		__u_short MyPort, __u_short MyRDPort,
		__u_short MyFailOverPort, __u_short MyFailOverRDPort,
		DBMODE DBMode, DBMODE ChoiceDB,
		int maxpool = 50,
#ifndef _DB_TEST_
		map<int32_t, StatementInfo*> *pSQLDB = NULL,
#endif
		int32_t pSessHangupTimer = 60);

	void                setDBMode(DBMODE DBMode, DBMODE ChoiceDB);
	DBMODE              getDBMode() { return m_DBMode; }
	DBMODE              getChoiceDB() { return m_ChoiceDB; }
	void                ChangeDBMode(DBMODE pDBMode, DBMODE pChoiceDB, const char* pPrefix = NULL);
	void                setMyRDUsed(bool bRdUsed, const char* pPrefix);
	bool                getMyRDUsed() { return m_bMyRDUse; }

	DBManager();
	~DBManager();

private:
#ifndef _DB_TEST_
	void                 FindDBServer(const char *pPrefix, OraSessionPool *&pOraPool, MySessionPool *&pMyPool);
#else
	void                 FindDBServer(const char *pPrefix, MySessionPool *&pMyPool);
#endif

	static DBManager   *m_pThis;
	list<stDBServer*>   m_DBServerTable;
	int                 m_index;
	AxMutex             m_mtx;
#ifndef _DB_TEST_
	friend class        OraSession;
#endif
	friend class        MySession;

	DBMODE              m_DBMode;
	DBMODE              m_ChoiceDB;
	bool                m_bOraUse;
	bool                m_bMyUse;
	bool                m_bMyRDUse;
};

#endif 
