using System;
using System.Collections.Concurrent;
using System.IO;
using System.Runtime.CompilerServices;
using System.Text;
using System.Threading.Tasks;
using System.Threading;
using System.Diagnostics;
using System.Runtime.InteropServices;

namespace LogLibrary
{
    // 프로젝트 > 참조 우클릭 > 참조 추가 > 프로젝트 ConfigureLibrary 추가
    public class LogManager
    {
        /// <summary>
        /// LogManager Singleton 객체
        /// </summary>
        private static LogManager m_instance;
        /// <summary>
        /// Singleton 객체 접근을 위한 Lock
        /// </summary>
        private static readonly object m_instanceLock = new object();
        /// <summary>
        /// Singleton 객체 폐기 Flag
        /// </summary>
        private bool m_isDispose = false;

        /// <summary>
        /// Log 폴더 경로
        /// </summary>
        private string m_logPath;
        /// <summary>
        /// Log를 사용하는 서비스 프로그램 명
        /// </summary>
        private string m_service;
        /// <summary>
        /// Log 파일 절대경로
        /// </summary>
        private string m_logFile;
        /// <summary>
        /// Log File 시퀀스번호 (제한 용량에 다다르면 순차증가하여 Log File 재생성)
        /// </summary>
        private uint m_logSeq = 1;
        /// <summary>
        /// Log File Writer 객체
        /// </summary>
        private StreamWriter m_fileWriter;

        /// <summary>
        /// Log FileStream 객체, FileStream을 쓰는게 더 좋을 것 같은데 분석을 다 못해서 주석으로 막아놨다.
        /// </summary>
        //private FileStream m_fileStream;

        /// <summary>
        /// Log Message 유입 Queue
        /// </summary>
        private ConcurrentQueue<string> m_logQ;
        /// <summary>
        /// LogManager 실행 체크 Flag
        /// </summary>
        private bool m_isRunning;
        /// <summary>
        /// Log File 정보 객체
        /// </summary>
        private FileInfo m_logFileCheck;
        /// <summary>
        /// Log File 최대 사이즈 (Byte)
        /// </summary>
        private uint m_logFileLimitSize;


        //AMS를 위한 DLL Import & extern 선언
        /// <summary>
        /// AMS Log 작성 함수
        /// </summary>
        /// <param name="_nPriority">0:???, 1:STA, 2:INF, 3:WAN, 4:ERR</param>
        /// <param name="_sLogInfo">로그상태</param>
        /// <param name="_sProcessName">프로세스명</param>
        /// <param name="_sModuleName">모듈명</param>
        /// <param name="_sErrorGB">에러구분(의미없음)</param>
        /// <param name="_nErrorCode">에러코드</param>
        /// <param name="_sErrorDetail">에러내용</param>
        /// <param name="_sErrorReason">에러원인</param>
        /// <returns></returns>
        [DllImport(@"C:\NSP\APL\AMSLogCtrlEx.dll")]
        private static extern int AMSLogWriteEx(int _nPriority, string _sLogInfo, string _sProcessName, string _sModuleName, string _sErrorGB, int _nErrorCode, string _sErrorDetail, string _sErrorReason);


        /// <summary>
        /// 외부에서 생성자 접근을 못하게 private로 명시
        /// </summary>
        private LogManager() { }

        ~LogManager()
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
            LogQueueClear();
            m_fileWriter.Close();

            GC.SuppressFinalize(this); // GC가 이 객체에 대해서 Finalize를 호출하지 않도록 명시적으로 지정
        }

        /// <summary>
        /// m_isRunning = false 처리하여 더이상 로그를 받지 못하게 한 뒤, 로그 Q에 남은 로그들은 모두 남기고 종료.
        /// </summary>
        private void LogQueueClear()
        {
            string outMsg;
            //m_isRunning = false;
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
        /// LogManager 객체 초기화 
        ///  - LogPath 설정 (경로 디렉토리가 없으면 생성)
        ///  - Initialize 실패 시 윈도우 이벤트로그 작성
        /// </summary>
        /// <param name="logPath">Log파일 경로</param>
        /// <param name="service">프로그램 명</param>
        /// <param name="logSize">로그파일 최대 사이즈(Byte)</param>
        public void Initialize(string logPath, string service, uint logSize)
        {
            try
            {
                m_logPath = logPath;
                m_service = service;
                m_logFileLimitSize = logSize;
                m_logFile = m_logPath + m_service + "_" + m_logSeq.ToString() + ".log";

                if (!Directory.Exists(m_logPath)) // 디렉토리가 없다면 생성
                {
                    Directory.CreateDirectory(m_logPath);
                }

                m_fileWriter = new StreamWriter(m_logFile, true, Encoding.Default); // Log Append, ANSI Encoding
                //m_fileStream = new FileStream(m_logFile, FileMode.OpenOrCreate, FileAccess.ReadWrite, FileShare.ReadWrite, 524288); // TEST CODE
                m_logQ = new ConcurrentQueue<string>();
                m_logFileCheck = new FileInfo(m_logFile);
                m_isRunning = true;

                Task WriteLogTask = this.WriteLog();

                AMSLogPrint("STS", "LogManager Initialize Success");
            }
            catch (Exception e)
            {
                throw new Exception("LogManager Initialize Fail.\n" + e.Message);
            }
        }

        /// <summary>
        /// Initialize 에서 호출 -> 쓰레드로 동작
        ///  - LogFile 사이즈가 최대치보다 크다면 시퀀스를 증가시켜 새로운 파일 생성
        ///  - m_LogQ 에서 로그 데이터를 Dequeue 하여 파일에 작성
        ///  - m_LogQ 에 데이터가 없다면 0.1초 sleep
        /// </summary>
        private async Task WriteLog()
        {
            while (m_isRunning)
            {
                m_logFileCheck.Refresh();
                if (m_logFileCheck.Length > m_logFileLimitSize)
                {
                    m_fileWriter.Close();
                    //m_fileStream.Close();
                    string moveLog = m_logPath + m_service + "_" + (m_logSeq++).ToString() + ".log";
                    while (File.Exists(moveLog))
                    {
                        moveLog = m_logPath + m_service + "_" + m_logSeq.ToString() + ".log";
                    }
                    File.Move(m_logFile, moveLog);
                    m_fileWriter = new StreamWriter(m_logFile, true, Encoding.UTF8);
                    //m_fileStream = new FileStream(m_logFile, FileMode.OpenOrCreate, FileAccess.ReadWrite, FileShare.ReadWrite, 524288); // TEST CODE
                }
                if (m_logQ.TryDequeue(out string outMsg))
                {
                    m_fileWriter.WriteLine(outMsg);
                    //await m_fileStream.WriteAsync(Encoding.Default.GetBytes(outMsg), 0, outMsg.Length);
                    m_fileWriter.Flush();
                }
                else
                {
                    await Task.Delay(10);
                }
            }
        }

        /// <summary>
        /// 외부에서 호출하는 Log 작성 함수
        /// </summary>
        /// <param name="strMsg">로그 데이터</param>
        /// <param name="taskId">Log 함수를 호출하는 서비스 스레드 ID</param>
        /// <param name="callingMethod">Log함수를 호출한 함수명</param>
        /// <param name="callingFilePath">Log함수를 호출한 위치의 소스파일</param>
        /// <param name="callingFileLineNumber">Log함수를 호출한 위치의 소스라인</param>
        public void Log(string strMsg,
            int taskId,
            [CallerMemberName] string callingMethod = null,
            [CallerFilePath] string callingFilePath = null,
            [CallerLineNumber] int callingFileLineNumber = 0)
        {
            string[] filePath = callingFilePath.Split('\\');
            string fileName = filePath[filePath.Length - 2] + "/" + filePath[filePath.Length - 1];

            // Write out message
            DateTime tm = DateTime.Now;
            string log = "(" + taskId.ToString() + ") " +
                         "[" + DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss.fff") + "] " +
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
                LogManager.WriteEventLogEntry(m_service, "Log is Not Running. So, Can not enqueuing. (" + log + ")");
                LogManager.AMSLogPrintStatic(m_service, "ERR", "LogManager 미동작! 이벤트로그 확인 필요!");
            }
        }

        /// <summary>
        /// 서비스에서 사용하는 AMS Log 작성 함수
        /// </summary>
        /// <param name="errType">에러타입 ("ERR" 외에는 다 똑같음)</param>
        /// <param name="errMsg">로그 메세지</param>
        /// <param name="callingMethod">AMSLogPrint를 호출한 함수</param>
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

                AMSLogWriteEx(nErrPriority, szLogInf, m_service, callingMethod, "U", errCode, errMsg, "");
            }
            catch (Exception e)
            {
                throw new Exception("AMS Log Print Exception!!\n" + e.Message);
            }
        }

        /// <summary>
        /// LogManager Instance 가 없을 때 사용하기 위한 함수 
        /// </summary>
        /// <param name="service">서비스 프로그램 명</param>
        /// <param name="errType">에러타입 ("ERR" 외에는 다 똑같음)</param>
        /// <param name="errMsg">로그 메세지</param>
        /// <param name="callingMethod">AMSLogPrint를 호출한 함수</param>
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
