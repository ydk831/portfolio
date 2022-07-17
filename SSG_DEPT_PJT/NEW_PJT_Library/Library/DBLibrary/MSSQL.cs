using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Data.SqlClient;
using System.Linq;
using System.Net;
using System.Reflection;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using LogLibrary;


namespace DBLibrary
{
    public enum DB_LOCATION { LOCAL, REMOTE };
    public class MSSQL : IDisposable
    {
        /// <summary>
        /// Singleton 객체 폐기 Flag
        /// </summary>
        private bool m_isDispose = false;
        /// <summary>
        /// MSSQL Singleton 객체
        /// </summary>
        private static MSSQL m_instance;
        /// <summary>
        /// Singleton 객체 접근을 위한 Lock
        /// </summary>
        private static object m_instanceLock = new object();
        /// <summary>
        /// 점포코드 별 SqlConnectionManger 객체를 갖는 Dictionary (로컬서버)
        /// </summary>
        private static ConcurrentDictionary<string, SqlConnectinoManager> m_localSQLDictionary = new ConcurrentDictionary<string, SqlConnectinoManager>();
        /// <summary>
        /// 점포코드 별 SqlConnectionManger 객체를 갖는 Dictionary (원격서버)
        /// </summary>
        private static ConcurrentDictionary<string, SqlConnectinoManager> m_remoteSQLDictionary = new ConcurrentDictionary<string, SqlConnectinoManager>();
        /// <summary>
        /// 외부에서 생성자 접근을 못하게 private로 명시 
        /// </summary>      
        private MSSQL() { }
        ~MSSQL()
        {
            if (!m_isDispose)
            {
                Dispose();
            }
        }

        /// <summary>
        /// Singleton 객체 폐기 함수
        /// </summary>
        public void Dispose()
        {
            m_isDispose = true;

            // 메모리 해제
            foreach (string key in m_localSQLDictionary.Keys)
            {
                m_localSQLDictionary[key].Dispose();
            }
            m_localSQLDictionary.Clear();

            GC.SuppressFinalize(this); // GC가 이 객체에 대해서 Finalize를 호출하지 않도록 명시적으로 지정
        }

        /// <summary>
        /// Singleton 처리를 위한 객체 getter
        /// </summary>
        public static MSSQL Instance
        {
            get
            {
                // lock은 비싼 연산에 속하므로 Double-checked locking을 사용
                // 자세한 설명 : https://stackoverflow.com/questions/12316406/thread-safe-c-sharp-singleton-pattern
                if (m_instance == null)
                {
                    lock (m_instanceLock)
                    {
                        if (m_instance == null)
                        {
                            m_instance = new MSSQL();
                        }
                    }
                }
                return m_instance;
            }
        }

        /// <summary>
        /// MSSQL 객체 초기화
        /// 서비스에서 점포코드별 연동 할 DB정보를 입력하면 해당 정보를 토대로 SqlonnectinoManager 클래스를 생성하여 m_localSQLDictionary에 저장한다.
        /// </summary>
        /// <param name="storeCode">점포코드 4자리</param>
        /// <param name="dbIp">접속할 DB IP</param>
        /// <param name="dbName">접속할 Database 명</param>
        /// <param name="dbUser">접속할 DB 계정명</param>
        /// <param name="dbPassword">접속할 DB 계정의 비밀번호</param>
        /// <param name="connectionCnt">DB와 연결을 맺을 세션 수</param>
        /// <param name="dbConnTimeout">DB연결 TImeout 시간 (ms)</param>
        /// <returns>초기화 성공여부</returns>
        public bool Initialize(DB_LOCATION dbLocation, string storeCode, string dbIp, string dbName, string dbUser, string dbPassword, uint connectionCnt, int dbConnTimeout)
        {
            try
            {
                if (storeCode.Length != 4)
                {
                    // 점포코드 정합성 에러
                    return false;
                }
                else if (dbIp == null || IPAddress.TryParse(dbIp, out IPAddress addr) == false)
                {
                    // dbIp형식 에러
                    return false;
                }
                else if (connectionCnt <= 0)
                {
                    // db연결 세션 수 에러, 1로 초기화
                    connectionCnt = 1;
                }
                else if (dbConnTimeout < 1000)
                {
                    // 1초보다 적게 설정했을 경우 1초로 초기화
                    dbConnTimeout = 1000;
                }

                // 로컬DB 연결
                if (dbLocation == DB_LOCATION.LOCAL)
                {
                    if (m_localSQLDictionary.TryAdd(storeCode, new SqlConnectinoManager(storeCode, dbIp, dbName, dbUser, dbPassword, connectionCnt, dbConnTimeout)))
                    {
                        // 연결로그 작성
                        LogManager.Instance.Log("success local db", Thread.CurrentThread.ManagedThreadId);
                        return true;
                    }
                    else // 동일 Key가 존재하는 경우 Fail
                    {
                        // 에러로그 작성
                        return false;
                    }
                }
                else // 원격DB 연결
                {
                    if (m_remoteSQLDictionary.TryAdd(storeCode, new SqlConnectinoManager(storeCode, dbIp, dbName, dbUser, dbPassword, connectionCnt, dbConnTimeout)))
                    {
                        // 연결로그 작성
                        LogManager.Instance.Log("success remote db", Thread.CurrentThread.ManagedThreadId);
                        return true;
                    }
                    else // 동일 Key가 존재하는 경우 Fail
                    {
                        // 에러로그 작성
                        return false;
                    }
                }
            }
            catch (Exception e)
            {
                // exception 에러로그 발생
                LogManager.Instance.Log($"MSSQL Exception {e.Message}", Thread.CurrentThread.ManagedThreadId);
                return false;
            }
        }

        /// <summary>
        /// 점포코드에 매칭되는 SqlConnectinoManager 객체를 가져오는 함수
        /// GetConnectino(string storeCode)에서만 호출한다.
        /// </summary>
        /// <param name="connectionCount">점포코드 4자리</param>
        /// <returns>해당 점포코드에 매칭되는 SqlConnectionManager 객체</returns>
        internal SqlConnectinoManager GetSqlConnectinoManager(string storeCode, DB_LOCATION dbLocation = DB_LOCATION.LOCAL)
        {
            try
            {
                if (dbLocation == DB_LOCATION.LOCAL)
                {
                    if (m_localSQLDictionary.TryGetValue(storeCode, out SqlConnectinoManager sqlConnMgr))
                    {
                        return sqlConnMgr;
                    }
                    else // SqlConnectionManager가 Dictionary에 없는 경우.
                    {
                        // 에러로그
                        return null;
                    }
                }
                else
                {
                    if (m_remoteSQLDictionary.TryGetValue(storeCode, out SqlConnectinoManager sqlConnMgr))
                    {
                        return sqlConnMgr;
                    }
                    else // SqlConnectionManager가 Dictionary에 없는 경우.
                    {
                        // 에러로그
                        return null;
                    }
                }
            }
            catch (Exception)
            {
                // exception 에러로그                
                return null;
            }
        }

        /// <summary>
        /// 서비스에서 사용하는 실제 DB세션 객체를 얻어오는 함수.
        /// </summary>
        /// <param name="storeCode">점포코드 4자리</param>
        /// <returns>DB세션객체</returns>
        public SqlConnection GetConnection(string storeCode, DB_LOCATION dbLocation = DB_LOCATION.LOCAL)
        {
            if (storeCode.Length != 4)
            {
                // 점포코드 정합성 에러
                return null;
            }

            SqlConnectinoManager sqlConnMgr = GetSqlConnectinoManager(storeCode, dbLocation);
            if (sqlConnMgr == null)
            {
                // SqlConnectionManager 오류. 없는 점포코드로 조회할 경우 발생할 수 있음. 에러로그 작성
                return null;
            }

            // return 되는게 null인것도 사용하는 측에서 체크
            return sqlConnMgr.GetConnection();
        }

        /// <summary>
        /// 서비스에서 사용한 DB세션 객체를 반환하는 함수
        /// </summary>
        /// <param name="storeCode">점포코드 4자리</param>
        /// <param name="dbConnection">사용한 DB세션객체</param>
        /// <returns>반환 성공여부</returns>
        public bool ReturnConnection(string storeCode, SqlConnection dbConnection, DB_LOCATION dbLocation = DB_LOCATION.LOCAL)
        {
            if (storeCode.Length != 4)
            {
                // 점포코드 정합성 에러
                return false;
            }
            SqlConnectinoManager sqlConnMgr = GetSqlConnectinoManager(storeCode, dbLocation);
            if (sqlConnMgr == null)
            {
                // SqlConnectionManager 오류. 없는 점포코드로 조회할 경우 발생할 수 있음. 에러로그 작성
                return false;
            }

            sqlConnMgr.ReturnConnection(dbConnection);
            return true;
        }
    }
}
