using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Data.SqlClient;
using System.Linq;
using System.Reflection;
using System.Text;
using System.Threading.Tasks;

namespace NSPCCSIS
{
    class SQL
    {
        // Singleton
        private static SQL m_instance; // SQL
        private static object m_instanceLock = new object();

        // Member Value
        private string m_dbIp;
        private string m_dbName;
        private string m_dbUser;
        private string m_dbPassword;
        private string m_dbConnectString;
        private int m_dbConnectionCnt;

        public ConcurrentQueue<SqlConnection> m_sqlConnectionList;
        
        /// <summary>
        /// 외부에서 생성자 접근을 못하게 private로 명시 
        /// </summary>      
        private SQL() { }

        /// <summary>
        /// Singleton 처리를 위한 객체 getter
        /// </summary>
        public static SQL Instance
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
                            m_instance = new SQL();
                        }
                    }
                }
                return m_instance;
            }
        }

        /// <summary>
        /// SQL 객체 초기화
        /// </summary>
        /// <param name="dbIp">접속할 DB IP</param>
        /// <param name="dbName">접속할 Database 명</param>
        /// <param name="dbUser">접속할 DB 계정명</param>
        /// <param name="dbPassword">접속할 DB 계정의 비밀번호</param>
        /// <param name="connectionCnt">DB와 연결을 맺을 세션 수</param>
        public void Initialize(string dbIp, string dbName, string dbUser, string dbPassword, int connectionCnt)
        {
            m_dbIp = dbIp;
            m_dbName = dbName;
            m_dbUser = dbUser;
            m_dbPassword = dbPassword;
            m_dbConnectionCnt = connectionCnt;
            m_dbConnectString = $"Server={m_dbIp}; Database={m_dbName}; uid={m_dbUser}; pwd={m_dbPassword}";

            m_sqlConnectionList = new ConcurrentQueue<SqlConnection>();

            try
            {
                CreateConnection(m_dbConnectionCnt);
            }
            catch (Exception e)
            {
                throw new Exception("SQL Initialize Fail.\n" + e.Message);
            }
        }

        /// <summary>
        /// DB와 실제 연결을 맺는 함수
        /// </summary>
        /// <param name="connectionCount">연결을 맺을 세션 수</param>
        private void CreateConnection(int connectionCount)
        {
            for (var i = 0; i < connectionCount; i++)
            {
                SqlConnection connection = new SqlConnection(m_dbConnectString);
                connection.Open();
                m_sqlConnectionList.Enqueue(connection);
            }
        }

        /// <summary>
        /// 연결된 세션을 큐(세션보관용)에 넣는 함수
        /// </summary>
        /// <param name="connection"></param>
        public void EnqueuSqlConnection(SqlConnection connection)
        {
            if (connection != null)
            {
                if (connection.State == System.Data.ConnectionState.Open)
                {
                    m_sqlConnectionList.Enqueue(connection);
                }
                else if (connection.State == System.Data.ConnectionState.Closed)
                {
                    connection.Close();
                    connection.Open();
                    m_sqlConnectionList.Enqueue(connection);
                }
                else // Connecting, Executing, Featching, Broken
                {
                    // Do Something??
                }
            }
        }

        /// <summary>
        /// DB 연결을 사용하기 위한 함수
        /// </summary>
        /// <returns>DB와 연결된 세션</returns>
        public SqlConnection DequeuSqlConnection()
        {
            SqlConnection connection;
            if (m_sqlConnectionList.TryDequeue(out connection))
            {
                return connection;
            }
            else
            {
                return null;
            }
        }
    }
}
