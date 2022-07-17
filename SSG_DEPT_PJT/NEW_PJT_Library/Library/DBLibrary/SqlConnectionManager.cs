using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Data.SqlClient;
using System.Linq;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using LogLibrary;

namespace DBLibrary
{
    /// <summary>
    /// DB연결 및 관리, 연결된 세션 객체를 보관, 반환, 회수하는 Class
    /// 점포코드별로 생성된다.
    /// </summary>
    internal class SqlConnectinoManager : IDisposable
    {
        /// <summary>
        /// Singleton 객체 폐기 Flag
        /// </summary>
        private bool m_isDispose = false;
        /// <summary>
        /// 연결되지 않은 SqlConnection 객체를 보관하는 큐
        /// </summary>
        internal ConcurrentQueue<SqlConnection> m_sqlDisConnectionQ = new ConcurrentQueue<SqlConnection>();
        /// <summary>
        /// 연결된 SqlConnectino 세션 객체를 보관하는 큐
        /// </summary>
        internal ConcurrentQueue<SqlConnection> m_sqlConnectionQ = new ConcurrentQueue<SqlConnection>();

        /// <summary>
        /// 점포코드 4자리
        /// </summary>
        internal string StoreCode { get; set; }
        /// <summary>
        /// 연결할 DB서버 IP
        /// </summary>
        internal string DbIp { get; set; }
        /// <summary>
        /// 연결할 DB서버 명칭
        /// </summary>
        internal string DbName { get; set; }
        /// <summary>
        /// DB연결에 사용 될 계정
        /// </summary>
        internal string DbUser { get; set; }
        /// <summary>
        /// DB연결에 사용 될 계정 비밀번호
        /// </summary>
        internal string DbPassword { get; set; }
        /// <summary>
        /// DB연결에 사용 될 DB연결문자열
        /// </summary>
        internal string DbConnectString { get; set; }
        /// <summary>
        /// DB와 연결 할 세션 수
        /// </summary>
        internal uint DbConnectionCnt { get; set; }
        /// <summary>
        /// DB 연결 Timeout sec
        /// </summary>
        internal int DbConnectionTimeout { get; set; }

        CancellationTokenSource m_CheckDisConnSessionTaskCT = new CancellationTokenSource();

        Task m_CheckDisconnSessionTask;

        /// <summary>
        /// DBMS 클래스에서 생성자를 호출한다.
        /// 파라미터로 전달 된 정보를 토대로 DB세션 객체를 생성하고 DisConnectionQ에 저장한다.
        /// 그리고 CheckDisconnSession()을 Task로 수행하여 연결되지 않은 DB객체들을 연결한다.
        /// </summary>
        /// <param name="storeCode">점포코드 4자리</param>
        /// <param name="dbIp">DB서버 IP</param>
        /// <param name="dbName">DB서버 명</param>
        /// <param name="dbUser">DB 접속 계정</param>
        /// <param name="dbPassword">DB 접속 계정의 비밀번호</param>
        /// <param name="m_dbConnectionCnt">DB에 연결 할 세션 수</param>
        /// <param name="dbConnectionTimeout">DB 연결 Timeout 시간</param>
        internal SqlConnectinoManager(string storeCode, string dbIp, string dbName, string dbUser, string dbPassword, uint m_dbConnectionCnt, int dbConnectionTimeout)
        {
            StoreCode = storeCode;
            DbIp = dbIp;
            DbName = dbName;
            DbUser = dbUser;
            DbPassword = dbPassword;
            DbConnectionTimeout = dbConnectionTimeout;
            DbConnectString = $"Server={DbIp}; Database={DbName}; uid={DbUser}; pwd={DbPassword}; Connection Timeout={dbConnectionTimeout}";

            for (int i = 0; i < m_dbConnectionCnt; i++)
            {
                m_sqlDisConnectionQ.Enqueue(new SqlConnection(DbConnectString));
            }

            m_CheckDisconnSessionTask = CheckDisconnSession();
            LogManager.Instance.Log("SqlConnectionManager Create Success", Thread.CurrentThread.ManagedThreadId);
        }

        ~SqlConnectinoManager()
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

            // 메모리 정리
            m_CheckDisConnSessionTaskCT.Cancel();

            GC.SuppressFinalize(this); // GC가 이 객체에 대해서 Finalize를 호출하지 않도록 명시적으로 지정
        }

        /// <summary>
        /// 생성자에서 TASK로 호출되며, 연결되지 않은 세션들을 연결한다.
        /// CancellationToken을 이용해 연결 Timeout을 지정한다.
        /// </summary>
        /// <returns></returns>
        internal async Task CheckDisconnSession()
        {
            LogManager.Instance.Log("SqlConnectinoManager Start CheckDisConnSession", Thread.CurrentThread.ManagedThreadId);
            while (true)
            {
                if (m_CheckDisConnSessionTaskCT.Token.IsCancellationRequested)
                {
                    break;
                }
                try
                {
                    if (m_sqlDisConnectionQ.TryDequeue(out SqlConnection sqlConn))
                    {
                        if (sqlConn == null)
                        {
                            LogManager.Instance.Log("SqlConnectino is null", Thread.CurrentThread.ManagedThreadId);
                            sqlConn = new SqlConnection(DbConnectString);
                            //using (CancellationTokenSource dbOpenTC = new CancellationTokenSource((int)DbConnectionTimeout))
                            //{
                            //    //dbOpenTC.CancelAfter((int)DbConnectionTimeout);
                            //    await sqlConn.OpenAsync(dbOpenTC.Token);
                            //}

                            sqlConn.Open();
                            m_sqlConnectionQ.Enqueue(sqlConn);
                            LogManager.Instance.Log($"SqlConnect Enqueue with null", Thread.CurrentThread.ManagedThreadId);
                        }
                        else if (sqlConn.State != System.Data.ConnectionState.Open)
                        {
                            LogManager.Instance.Log($"SqlConnectino state is {sqlConn.State}", Thread.CurrentThread.ManagedThreadId);
                            //sqlConn.ConnectionString = DbConnectString;
                            await sqlConn.OpenAsync();
                            LogManager.Instance.Log($"SqlConnect Enqueue with {sqlConn.State}", Thread.CurrentThread.ManagedThreadId);
                            m_sqlConnectionQ.Enqueue(sqlConn);
                        }
                        else // 그 외 상태들에 대해서...
                        {
                            // 뭐할까...
                        }
                    }
                    else
                    {
                        // Dequeue Error.. or no more exist dequeue session
                        //LogManager.Instance.Log("CheckDisconnsession can't deque", Thread.CurrentThread.ManagedThreadId);
                    }
                }
                catch (Exception e)
                {
                    // 에러 출력
                    LogManager.Instance.Log($"CheckDisconnsession {e.Message}", Thread.CurrentThread.ManagedThreadId);
                }

                await Task.Delay(100);
            }
            LogManager.Instance.Log($"SqlConnectionManager Ended", Thread.CurrentThread.ManagedThreadId);
        }

        /// <summary>
        /// DBMS 싱글톤 객체에서 커넥션을 가져가기 위한 함수
        /// </summary>
        /// <returns>DB세션 객체</returns>
        internal SqlConnection GetConnection()
        {
            try
            {
                // 연결 큐에서 연결이 맺어진 객체를 꺼낼 때 까지
                while (m_sqlConnectionQ.TryDequeue(out SqlConnection sqlConn))
                {
                    // Open 상태가 아니면 끊고 비연결큐에 넣음 (자동 재접속되게)
                    if (sqlConn.State != System.Data.ConnectionState.Open)
                    {
                        sqlConn.Close();
                        m_sqlDisConnectionQ.Enqueue(sqlConn);
                    }
                    // Open 상태인건 리턴
                    else
                    {
                        return sqlConn;
                    }

                }

                // 연결큐에 연결된게 하나도 없다면 에러로그 남기고 null 리턴
                return null;

            }
            catch
            {
                // 익셉션
                return null;
            }
        }

        internal void ReturnConnection(SqlConnection sqlConn)
        {
            m_sqlConnectionQ.Enqueue(sqlConn);
        }

    }
}
