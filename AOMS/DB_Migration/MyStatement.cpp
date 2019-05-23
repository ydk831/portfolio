#include <errmsg.h>
#include <mysqld_error.h>

#include "CommIh/DebugLogger.h"
#include "AxLib/AxUtil.h"
#include "Mysql/MyStatement.h"


MyStatement::MyStatement(const string &rSQL, MySession *pSess)
{
	m_bAutocommit = false;
	m_pStmt = NULL;
	m_pSess = pSess;

	if (m_pSess) {
		char BUF[32];
		snprintf(BUF, sizeof(BUF), "[%02d(%s)]", m_pSess->GetID(), (m_pSess->isReadOnly() ? "ReadOnly" : "Master"));
		m_TAG = BUF;
		m_SQL = rSQL;
		m_pStmt = m_pSess->PrepareStmt(rSQL);
	}
	else {
		m_TAG = "[x:xx]";
	}

	m_bBindParam = false;
	memset(m_BindParam, 0x00, sizeof(m_BindParam));
	memset(m_BindParam_Length, 0x00, sizeof(m_BindParam_Length));


	m_bStoreResult = false;
	m_bBindResultSet = false;
	memset(m_BindResultSet, 0x00, sizeof(m_BindResultSet));
	memset(m_BindResultSet_Length, 0x00, sizeof(m_BindResultSet_Length));
	m_Prepare_meta_result = NULL;
}

MyStatement::~MyStatement()
{
	// MySQL prepare statement 의미가 미약하므로 
	// statement를 제거한다.

	if (m_Prepare_meta_result) {
		mysql_free_result(m_Prepare_meta_result);
	}

	ReleaseStmt();

	if (false == m_bAutocommit) {
		//SetAutocommit(true);
	}
}

int MyStatement::ExecuteQuery()
{
	static const char *FN = "[MyStatement::ExecuteQuery] ";

	if (NULL == m_pStmt) {
		WLOG(m_TAG << FN << "MyStatement(null)");
		return RET_DB_STATEMENT_FAIL;
	}

	if (m_bBindParam) {
		if (mysql_stmt_bind_param(m_pStmt, m_BindParam)) {
			CheckDBError(mysql_stmt_errno(m_pStmt));
			WLOG(m_TAG << FN << "Bind-Param FAIL. " << ERRORSTR());
			return RET_DB_STATEMENT_FAIL;
		}
	}

	if (m_bBindResultSet) {
		m_Prepare_meta_result = mysql_stmt_result_metadata(m_pStmt);
		if (NULL == m_Prepare_meta_result) {
			CheckDBError(mysql_stmt_errno(m_pStmt));
			WLOG(m_TAG << FN << "Result-Metadata FAIL. " << ERRORSTR());
			return RET_DB_STATEMENT_FAIL;
		}
	}

	if (m_pSess) {
		m_pSess->SetQrySttTime();
		m_pSess->SetLastUseTime(); // 20181213. YDK. 상세 설명은 Svn R2675 참조
	}

	/*
	time_t t;
	struct tm *lt;
	struct timeval tv;

	t = gettimeofday(&tv, NULL);
	lt = localtime(&tv.tv_sec);
	WLOG(m_TAG << FN << "Execute Query Start Time: " << lt->tm_hour << ":" << lt->tm_min << ":" << lt->tm_sec << "." << tv.tv_usec);
	*/

	if (mysql_stmt_execute(m_pStmt)) {
		unsigned int stmt_errno = mysql_stmt_errno(m_pStmt);
		switch (stmt_errno) {
		case ER_DUP_ENTRY: //1062
		{
			WLOG(m_TAG << FN << "Execute fail. " << ERRORSTR());
		} break;
		case ER_NO_SUCH_TABLE: //1142
		{
			WLOG(m_TAG << FN << "Execute fail. " << ERRORSTR());
		} break;

		case 9001: // user define, not-found
		case 9002: // user define, unique-constraint
		case 9006: // user define, Invalid Parameter
		case 9402: // user define, Payment Required
		case 9410: // user define, gone
		{
			IILOG(m_TAG << FN << "Execute fail. " << ERRORSTR());
		} break;

		default:
		{
			WLOG(m_TAG << FN << "Execute fail. " << ERRORSTR() << "\n" << m_SQL);
		} break;
		}

		return CheckDBError(stmt_errno);
	}

	/*
	t = gettimeofday(&tv, NULL);
	lt = localtime(&tv.tv_sec);
	WLOG(m_TAG << FN << "Execute Query Start Time: " << lt->tm_hour << ":" << lt->tm_min << ":" << lt->tm_sec << "." << tv.tv_usec);
	*/

	if (m_bBindResultSet) {
		if (mysql_stmt_bind_result(m_pStmt, m_BindResultSet)) {
			CheckDBError(mysql_stmt_errno(m_pStmt));
			WLOG(m_TAG << FN << "Bind-ResultSet FAIL. " << ERRORSTR());
			return RET_DB_STATEMENT_FAIL;
		}
	}

	return RET_DB_SUCCESS;
}

int MyStatement::Fetch(bool bStoreResultSet)
{
	static const char *FN = "[MyStatement::Fetch] ";

	if (NULL == m_pStmt) {
		WLOG(m_TAG << FN << "MyStatement(null)");
		//return eResStatementFail; 
		return RET_DB_STATEMENT_FAIL;
	}

	/*
	time_t t;
	struct tm *lt;
	struct timeval tv;
	*/

	if (bStoreResultSet && (false == m_bStoreResult)) {
		/*
		t = gettimeofday(&tv, NULL);
		lt = localtime(&tv.tv_sec);
		WLOG(m_TAG << FN << "Store Result Start Time: " << lt->tm_hour << ":" << lt->tm_min << ":" << lt->tm_sec << "." << tv.tv_usec);
		*/

		if (mysql_stmt_store_result(m_pStmt)) {
			WLOG(m_TAG << FN << "mysql_stmt_store_result FAIL. " << ERRORSTR());
			//return eResStatementFail;
			return RET_DB_STATEMENT_FAIL;
		}

		/*
		t = gettimeofday(&tv, NULL);
		lt = localtime(&tv.tv_sec);
		WLOG(m_TAG << FN << "Store Result End Time: " << lt->tm_hour << ":" << lt->tm_min << ":" << lt->tm_sec << "." << tv.tv_usec);
		WLOG(m_TAG << FN << "Store Num Rows : " << mysql_stmt_num_rows(m_pStmt));
		*/

		m_bStoreResult = true;
	}

	// http://dev.mysql.com/doc/refman/5.1/en/mysql-stmt-fetch.html
	switch (mysql_stmt_fetch(m_pStmt)) {
	case 0:
	{
		return RET_DB_SUCCESS;
	} break;

	case MYSQL_NO_DATA:
	{
		FLOG(m_TAG << FN << "MYSQL_DATA_NOT_FOUND. SQL: " << m_SQL);
		return RET_DB_NOT_FOUND;
	} break;

	case MYSQL_DATA_TRUNCATED:
	{
		ELOG(YELLOW(m_TAG << FN << "MYSQL_DATA_TRUNCATED. SQL: " << m_SQL));
		return RET_DB_TRUNCATED_FAIL;
	} break;

	default:
	{
		// if (ER_DUP_ENTRY == mysql_stmt_errno(m_pStmt)) {
		//    return eResUniqueConstraint;
		// }

		// 아래 error check로 흐른다.
	} break;
	}

	WLOG(m_TAG << FN << ERRORSTR());

	return CheckDBError(mysql_stmt_errno(m_pStmt));
}

int MyStatement::CheckDBError(int nErrorCode)
{
	switch (nErrorCode)
	{
	case ER_DUP_ENTRY:  // ER_DUP_ENTRY 1062, Duplicate entry
		return RET_DB_UNIQUE_ERR;
		break;

	case ER_DATA_TOO_LONG: // ER_DATA_TOO_LONG 1406, insert data too long than mysql column
		return RET_DB_ETC_FAIL;
		break;

	case ER_TABLE_EXISTS_ERROR:
		return RET_DB_EXIST_TBL;
		break;

	case ER_NO_SUCH_TABLE:
		return RET_DB_NOT_FOUND;
		break;

	case 9001: // user define, not-found
		return RET_DB_NOT_FOUND;
		break;

	case 9002: // user define, unique-constraint
		return RET_DB_UNIQUE_ERR;
		break;

	case 9006: // user define, Invalid Parameter
		return RET_DB_ETC_FAIL;
		break;

	default:
		if (m_pSess) {
			Rollback();
			m_pSess->SetAfterDown();
		}
		break;
	}

	return RET_DB_ETC_FAIL;
}