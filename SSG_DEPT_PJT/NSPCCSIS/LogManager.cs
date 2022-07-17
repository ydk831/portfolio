using System;
using System.Collections.Concurrent;
using System.IO;
using System.Runtime.CompilerServices;
using System.Text;
using System.Threading.Tasks;
using System.Threading;
using System.Diagnostics;
using System.Runtime.InteropServices;

namespace NSPCCSIS
{
    class LogManager
    {
        private static LogManager m_instance;
        private static readonly object m_instanceLock = new object();

        private string m_logPath; // 로그 경로
        private string m_service; // 서비스명
        private string m_logFile; // 로그파일 (절대경로)
        private uint m_logSeq = 1; // 로그파일 시퀀스번호 (순차증가)
        private StreamWriter m_fileWriter; // 로그파일 writer
        private ConcurrentQueue<string> m_logQ; // 로그 메시지 유입 큐
        private bool m_isRunning; // LogManager 실행 체크 flag
        private FileInfo m_logFileCheck; // 로그파일 정보
        private uint m_logFileLimitSize; // 로그파일 최대 사이즈 (KB)

        #region AMS를 위한 DLL Import & extern 선언
        // DLL 경로
        private static readonly string AMSLOGDLL_PATH = @"C:\nsp\apl\AMSLogCtrlEx.dll";

        [DllImport(@"C:\nsp\apl\AMSLogCtrlEx.dll")]
        private static extern int AMSLogWriteEx(int _nPriority, string _sLogInfo, string _sProcessName, string _sModuleName, string _sErrorGB, int _nErrorCode, string _sErrorDetail, string _sErrorReason);
        #endregion


        /// <summary>
        /// 외부에서 생성자 접근을 못하게 private로 명시
        /// </summary>
        private LogManager() { }

        /// <summary>
        ///  Singleton 처리를 위한 객체 getter
        /// </summary>
        public static LogManager Instance
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
                            m_instance = new LogManager();
                        }
                    }
                }
                return m_instance;
            }
        }

        /// <summary>
        ///  LogManager 객체 초기화 
        ///   - LogPath 설정 (경로 디렉토리가 없으면 생성)
        ///   - Initialize 실패 시 윈도우 이벤트로그 작성
        /// </summary>
        /// <param name="logPath">Log파일 경로</param>
        /// <param name="service">프로그램 명</param>
        public void Initialize(string logPath, string service, uint logSize)
        {
            try
            {
                m_logPath = logPath;
                m_service = service;
                m_logFileLimitSize = logSize;
                m_logFile = m_logPath + m_service + "_" + Program.storeCode + ".log";

                if (!Directory.Exists(m_logPath)) // 디렉토리가 없다면 생성
                {
                    Directory.CreateDirectory(m_logPath);
                }

                m_fileWriter = new StreamWriter(m_logFile, true, Encoding.UTF8); // Log Append
                m_logQ = new ConcurrentQueue<string>();
                m_logFileCheck = new FileInfo(m_logFile);
                m_isRunning = true;

                WriteLog();
            }
            catch (Exception e)
            {
                throw new Exception("LogManager Initialize Fail.\n" + e.Message);
            }
        }
        public void DeInitialize()
        {
            LogQueueClear();
            m_fileWriter.Close();
        }

        public string LogPath
        {
            get { return m_logPath; }
            set { m_logPath = value; }
        }

        private void LogQueueClear()
        {
            string outMsg;
            m_isRunning = false;
            while (!m_logQ.IsEmpty)
            {
                if (m_logQ.TryDequeue(out outMsg))
                {
                    m_fileWriter.WriteLine(outMsg);
                    m_fileWriter.Flush();
                }
            }
        }

        /// <summary>
        /// Initialize 에서 호출 -> 쓰레드로 동작
        ///  - LogFile 사이즈가 최대치보다 크다면 시퀀스를 증가시켜 새로운 파일 생성
        ///  - m_LogQ 에서 로그 데이터를 Dequeue 하여 파일에 작성
        ///  - m_LogQ 에 데이터가 없다면 0.1초 sleep
        /// </summary>
        private void WriteLog()
        {
            var task = new Task(async () =>
            {
                string outMsg;
                while (m_isRunning)
                {
                    m_logFileCheck.Refresh();
                    if (m_logFileCheck.Length > m_logFileLimitSize)
                    {
                        m_fileWriter.Close();
                        string moveLog = m_logPath + m_service + "_" + Program.storeCode + "_" + (m_logSeq++).ToString() + ".log";
                        while (File.Exists(moveLog))
                        {
                            moveLog = m_logPath + m_service + "_" + Program.storeCode + "_" + m_logSeq.ToString() + ".log";
                        }
                        File.Move(m_logFile, moveLog);
                        m_fileWriter = new StreamWriter(m_logFile, true, Encoding.UTF8);
                    }
                    if (m_logQ.TryDequeue(out outMsg))
                    {
                        m_fileWriter.WriteLine(outMsg);
                        m_fileWriter.Flush();
                    }
                    else
                    {
                        await Task.Delay(10);
                    }
                }
            });
            task.Start();
        }

        /// <summary>
        /// 외부에서 호출하는 Log 작성 함수
        /// </summary>
        /// <param name="strMsg">로그 데이터</param>
        /// <param name="callingMethod">Log함수를 호출한 함수명</param>
        /// <param name="callingFilePath">Log함수를 호출한 위치의 소스파일</param>
        /// <param name="callingFileLineNumber">Log함수를 호출한 위치의 소스라인</param>
        public void Log(string strMsg,
            [CallerMemberName] string callingMethod = null,
            [CallerFilePath] string callingFilePath = null,
            [CallerLineNumber] int callingFileLineNumber = 0)
        {
            string[] filePath = callingFilePath.Split('\\');
            string fileName = filePath[filePath.Length - 2] + "/" + filePath[filePath.Length - 1];

            // Write out message
            DateTime tm = DateTime.Now;
            string log = "[" + DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss.fff") + "] " +
                         strMsg + "\t" +
                         "[" + callingMethod + "] " +
                         //callingFilePath +
                         fileName +
                         "(" + callingFileLineNumber + ")";

            if (m_isRunning)
            {
                m_logQ.Enqueue(log);
            }
            else // LogManager 가 동작하고 있지 않을 경우 윈도우 이벤트로그에 작성
            {
                LogManager.WriteEventLogEntry("NSPCRFCL", "Log is Not Running. So, Can not enqueuing. (" + log + ")");
                LogManager.AMSLogPrintStatic("NSPCRFCL", "ERR", "LogManager 미동작! 이벤트로그 확인 필요!");
            }
        }

        public void AMSLogPrint(string errType, string errMsg, [CallerMemberName] string callingMethod = null)
        {
            try
            {
                string szLogInf = "";
                int errCode = 2100;
                int nErrPriority = 0;

                szLogInf = "40" + "-" + errType + "-" + "APL-02";

                if (errType == "ERR")
                    nErrPriority = 4;
                else
                    nErrPriority = 0;

                AMSLogWriteEx(nErrPriority, szLogInf, ConfigManager.Instance.AppConfigRead("SERVICE"), callingMethod, "U", errCode, errMsg, "");
            }
            catch (Exception e)
            {
                throw new Exception("AMS Log Print Exception!!\n" + e.Message);
            }
        }

        // LogManager Instance 가 없을 때 사용하기 위한 함수
        public static void AMSLogPrintStatic(string service, string errType, string errMsg, [CallerMemberName] string callingMethod = null)
        {
            try
            {
                string szLogInf = "";
                int errCode = 2000;
                int nErrPriority = 0;

                szLogInf = "40" + "-" + errType + "-" + "APL-02";

                if (errType == "ERR")
                    nErrPriority = 4;
                else
                    nErrPriority = 0;

                AMSLogWriteEx(nErrPriority, szLogInf, service, callingMethod, "U", errCode, errMsg, "");
            }
            catch (Exception e)
            {
                throw new Exception("AMS Log Print Exception!!\n" + e.Message);
            }
        }

        /// <summary>
        /// LogManager 가 동작 못 할 경우 윈도우 이벤트로그에 작성하기 위한 함수.
        /// Logmanager 스레드와 관계없게 static으로 만들어 쓴다.
        /// </summary>
        /// <param name="msg">로그 메세지</param>
        public static void WriteEventLogEntry(string service, string msg)
        {
            System.Diagnostics.EventLog eventLog = new System.Diagnostics.EventLog();
            if (!System.Diagnostics.EventLog.SourceExists(service))
            {
                System.Diagnostics.EventLog.CreateEventSource(service, "Application");
            }
            eventLog.Source = service;
            int eventID = 8;
            eventLog.WriteEntry(msg, System.Diagnostics.EventLogEntryType.Error, eventID);
            eventLog.Close();
        }

    }
}
