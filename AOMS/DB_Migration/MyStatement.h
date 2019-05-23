#ifndef DB_MYSQLAPI_STATEMENT_H_
#define DB_MYSQLAPI_STATEMENT_H_

#include <string>
#include <mysql.h>

#include "AxLib/AxDefConst.h"
#include "Mysql/MySession.h"

#define RET_DB_SUCCESS              0
#define RET_DB_UNIQUE_ERR           -1
#define RET_DB_NOT_FOUND            -1403
#define RET_DB_DISCONNECTED         -3114
#define RET_DB_EXIST_TBL            -1050
#define RET_DB_ETC_FAIL             -9990
#define RET_DB_STATEMENT_FAIL       -9991
#define RET_DB_OBTAIN_FAIL          -9992
#define RET_DB_TRUNCATED_FAIL       -9993
#define RET_DB_INSERT_FAIL          -11111          /* 에러코드는 수정 */
#define RET_DB_UPDATE_FAIL          -11112          /* 에러코드는 수정 */

#define MAX_RECORD_CNT              (D_ORA_ARRAY*10)


class MyStatement {
public:
	static const __uint32_t MAX_BIND_SIZE = 50;

public:
	MyStatement(const string &rSQL, MySession *pSess);
	~MyStatement();

	void            BindParamString(unsigned int index, const char *pValuePtr, unsigned int nLen);
	void            BindParamInt8(unsigned int index, __int8_t *pValuePtr);
	void            BindParamUInt8(unsigned int index, uint8_t *pValuePtr);
	void            BindParamInt16(unsigned int index, __int16_t *pValuePtr);
	void            BindParamUInt16(unsigned int index, uint16_t *pValuePtr);
	void            BindParamInt32(unsigned int index, __int32_t *pValuePtr);
	void            BindParamUInt32(unsigned int index, uint32_t *pValuePtr);
	void            BindParamInt64(unsigned int index, __int64_t *pValuePtr);
	void            BindParamUInt64(unsigned int index, uint64_t *pValuePtr);
	void            BindParamBlob(unsigned int index, char *pValuePtr, unsigned int nLen);

	void            BindResultSetString(unsigned int index, const char *pValuePtr, unsigned int nBufSize);
	void            BindResultSetChar(unsigned int index, const char *pValuePtr, unsigned int nBufSize);
	void            BindResultSetInt8(unsigned int index, __int8_t *pValuePtr);
	void            BindResultSetInt16(unsigned int index, __int16_t *pValuePtr);
	void            BindResultSetInt32(unsigned int index, __int32_t *pValuePtr);
	void            BindResultSetInt64(unsigned int index, __int64_t *pValuePtr);
	void            BindResultSetUInt8(unsigned int index, uint8_t *pValuePtr);
	void            BindResultSetUInt16(unsigned int index, uint16_t *pValuePtr);
	void            BindResultSetUInt32(unsigned int index, uint32_t *pValuePtr);
	void            BindResultSetUInt64(unsigned int index, uint64_t *pValuePtr);
	void            BindResultSetBlob(unsigned int index, char *pValuePtr, unsigned int nBufSize);

	int               ExecuteQuery();
	int               Fetch(bool bStoreResultSet = true);
	unsigned long   AffectedRows();
	void            Clear();

	void            Commit();
	void            Rollback();
	void            Rollback(int dbRet); // 20130204-ISSUE:148 Don't call rollback() on FatalError [altibase]
	bool            SetAutocommit(bool flag);

	void            ReleaseStmt();

	char           *ERRORSTR();

private:
	void            ClearBindParam();
	void            ClearBindResult();

	int             CheckDBError(int nErrorCode);

private:
	std::string     m_TAG;
	MySession      *m_pSess;
	MYSQL_STMT     *m_pStmt;
	bool            m_bAutocommit;

	bool            m_bBindParam;
	MYSQL_BIND      m_BindParam[MAX_BIND_SIZE];
	unsigned long   m_BindParam_Length[MAX_BIND_SIZE];

	bool            m_bBindResultSet;
	bool            m_bStoreResult;
	MYSQL_BIND      m_BindResultSet[MAX_BIND_SIZE];
	unsigned long   m_BindResultSet_Length[MAX_BIND_SIZE];
	MYSQL_RES      *m_Prepare_meta_result;

	char            ERRSTR[256];  // tmp
	string          m_SQL;
};

#endif
