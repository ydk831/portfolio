using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using LogLibrary;
using System.Collections.Concurrent;
using System.Net.Sockets;
using MessageLibrary;
using System.Net;
using System.Threading;

namespace TcpClientLibrary
{
    public class TCPClientManager
    {
        /// <summary>
        /// 점포코드-Host 별 비연결 소켓 관리 딕셔너리 
        /// </summary>
        private ConcurrentDictionary<string, ConcurrentDictionary<bool, ConcurrentQueue<TcpClient>>> m_clientDisconnectDictionary = new ConcurrentDictionary<string, ConcurrentDictionary<bool, ConcurrentQueue<TcpClient>>>();
        /// <summary>
        /// 점포코드-Host 별 연결 소켓 관리 딕셔너리 
        /// </summary>
        private ConcurrentDictionary<string, ConcurrentDictionary<bool, ConcurrentQueue<TcpClient>>> m_clientConnectDictionary = new ConcurrentDictionary<string, ConcurrentDictionary<bool, ConcurrentQueue<TcpClient>>>();
        /// <summary>
        /// 점포코드-Host 별 전송 메세지 관리 딕셔너리
        /// </summary>
        private ConcurrentDictionary<string, ConcurrentDictionary<bool, ConcurrentQueue<HostMessage>>> m_sendToHostMsgDictionary = new ConcurrentDictionary<string, ConcurrentDictionary<bool, ConcurrentQueue<HostMessage>>>();
        /// <summary>
        /// 점포코드-Host 별 수신 메세지 관리 딕셔너리 
        /// </summary>        
        private ConcurrentDictionary<string, ConcurrentDictionary<bool, ConcurrentQueue<HostMessage>>> m_recvFromHostMsgDictionary = new ConcurrentDictionary<string, ConcurrentDictionary<bool, ConcurrentQueue<HostMessage>>>();

        /// <summary>
        /// 점포코드-Host 별 메세지 전송 Task 관리 딕셔너리 
        /// </summary>
        private ConcurrentDictionary<string, Task> m_sendToHostDictionary = new ConcurrentDictionary<string, Task>();

        /// <summary>
        /// millisecond (Tcpclient.SendTimeout == NetworkStream.WriteTimeout)
        /// </summary>
        private int m_sendTimeout = 3000;
        /// <summary>
        /// millisecond (Tcpclient.RecvTimeout == NetworkStream.ReadTimeout)
        /// </summary>
        private int m_recvTimeout = 3000;
        /// <summary>
        /// millisecond 
        /// </summary>
        private int m_connectTimeout = 3000;
        /// <summary>
        /// 수신 할 최대 Byte
        /// </summary>
        private int m_recvBufferSize = 1024 * 10;
        /// <summary>
        /// 송신 할 최대 Byte
        /// </summary>
        private int m_sendBufferSize = 1024 * 10;

        /// <summary>
        /// Initialize, CreateSession, TcpClientConnect 호출 판단 Flag 
        /// </summary>
        bool[] m_initCheckFlag = new bool[3] { false, false, false };

        /// <summary>
        /// TCPCLientManager Singleton 객체
        /// </summary>
        private static TCPClientManager m_instance;
        /// <summary>
        /// Singleton 객체 접근을 위한 Lock
        /// </summary>
        private static object m_instanceLock = new object();

        /// <summary>
        /// Singleton 객체 폐기 Flag
        /// </summary>
        private bool m_isDispose = false;

        /// <summary>
        /// 외부에서 생성자 접근을 못하게 private로 명시 
        /// </summary>      
        private TCPClientManager() { }

        /// <summary>
        /// Singleton 처리를 위한 객체 getter
        /// </summary>
        public static TCPClientManager Instance
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
                            m_instance = new TCPClientManager();
                        }
                    }
                }
                return m_instance;
            }
        }
        ~TCPClientManager()
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

            GC.SuppressFinalize(this); // GC가 이 객체에 대해서 Finalize를 호출하지 않도록 명시적으로 지정
        }

        /// <summary>
        /// TCPClientManager 초기화 함수. TCP 연결에 필요한 값 설정.
        /// </summary>
        /// <param name="sendTimeout">TCP 데이터 Send Timeout 설정 값 (ms)</param>
        /// <param name="recvTimeout">TCP 데이터 Recv Timeout 설정 값 (ms)</param>
        /// <param name="connectTimeout">TCP Connect Timeout 설정 값 (ms)</param>
        /// <param name="recvBufferSize">TCP 데이터 Recv 버퍼 최대 값 (byte)</param>
        /// <param name="sendBufferSize">TCP 데이터 Send 버퍼 최대 값 (byte)</param>
        public void Initialize(int sendTimeout, int recvTimeout, int connectTimeout, int recvBufferSize, int sendBufferSize)
        {
            m_sendTimeout = sendTimeout;
            m_recvTimeout = recvTimeout;
            m_connectTimeout = connectTimeout;
            m_recvBufferSize = recvBufferSize;
            m_sendBufferSize = sendBufferSize;

            // m_initCheckFlag[0] = true;
        }

        /// <summary>
        /// 점포코드-HOST Type 연결 객체 생성 함수.
        /// 점포코드 별 호스트의 Main/Sub, 호스트와 연결 할 세션 수를 설정.
        /// 서비스 함수에서 필요한 만큼 반복 호출한다.
        /// </summary>
        /// <param name="storeCode">점포코드 4자리</param>
        /// <param name="isHost">Host Type (True == Main / False == Sub)</param>
        /// <param name="sessionCnt">HOST와 연결 할 세션 수</param>
        public bool CreateSession(string storeCode, bool isHost, uint sessionCnt)
        {
            if (storeCode.Length != 4)
            {
                // 점포코드 정합성 확인 에러로그 출력
                return false;
            }
            if (sessionCnt <= 0)
            {
                // 세션카운트 정합성 확인 에러로그 출력
                return false;
            }

            m_clientDisconnectDictionary.TryAdd(storeCode, new ConcurrentDictionary<bool, ConcurrentQueue<TcpClient>>());
            m_clientDisconnectDictionary[storeCode].TryAdd(isHost,new ConcurrentQueue<TcpClient>());
            m_clientConnectDictionary.TryAdd(storeCode, new ConcurrentDictionary<bool, ConcurrentQueue<TcpClient>>());
            m_clientConnectDictionary[storeCode].TryAdd(isHost, new ConcurrentQueue<TcpClient>());

            for (int i = 0; i < sessionCnt; i++)
            {
                m_clientDisconnectDictionary[storeCode][isHost].Enqueue(new TcpClient());
            }

            // m_initCheckFlag[1] = true;
            LogManager.Instance.Log("Create tcp client sucecss", Thread.CurrentThread.ManagedThreadId);
            return true;
        }

        /// <summary>
        /// 점포코드 별 초기화한 연결 정보에 실제 호스트의 ip/port 정보를 입력하여 연결을 맺는다.
        /// 서비스 함수에서 필요한 만큼 반복 호출한다.        
        /// </summary>
        /// <param name="storeCode">점포코드 4자리</param>
        /// <param name="ip">연결 할 HOST IP</param>
        /// <param name="port">연결 할 HOST PORT</param>
        /// <param name="isHost">Host Type (True == Main / False == Sub)</param>
        public bool TcpClientConnect(string storeCode, string ip, int port, bool isHost = true)
        {
            if (storeCode.Length != 4)
            {
                // 점포코드 정합성 확인 에러로그 출력
                return false;
            }
            if (IPAddress.TryParse(ip, out IPAddress ipCheck) == false)
            {
                // ip 정합성 확인 에러로그 출력
                return false;
            }
            if (port < 0)
            {
                // port 정합성 확인 에러로그 출력
                return false;
            }

            Task.Run(() => TcpClientConnectTask(storeCode, ip, port, isHost));

            // 한번이라도 호출 됐다면 성공처리.. 얼마나 많은 연결이 있을지 모르므로..
            // 좀더 근본적인 관리는 시간관계상 하지 않았다..
            // m_initCheckFlag[1] = true; 
            LogManager.Instance.Log("connect tcp client sucecss", Thread.CurrentThread.ManagedThreadId);
            return true;
        }

        /// <summary>
        /// TcpClientConnect() 에서 Task로 실행하는 함수.
        /// 점포코드-Host Type 별 m_clientDisconnectDictionary 를 체크하여 HOST 연결 및 m_clientConnectDictionary에 저장한다.
        /// </summary>
        /// <param name="storeCode">점포코드 4자리</param>
        /// <param name="ip">연결 할 HOST IP</param>
        /// <param name="port">연결 할 HOST PORT</param>
        /// <param name="isHost">Host Type (True == Main / False == Sub)</param>
        public async void TcpClientConnectTask(string storeCode, string ip, int port, bool isHost = true)
        {
            while (true)
            {
                if (m_clientDisconnectDictionary[storeCode][isHost].TryDequeue(out TcpClient client))
                {
                    try
                    {
                        client.ReceiveTimeout = m_recvTimeout; // ms
                        client.ReceiveBufferSize = m_recvBufferSize; // byte

                        client.SendTimeout = m_sendTimeout; // ms
                        client.SendBufferSize = m_sendBufferSize; // byte

                        if (client.ConnectAsync(ip, port).Wait(m_connectTimeout))
                        {
                            m_clientConnectDictionary[storeCode][isHost].Enqueue(client);
                        }
                        else
                        {
                            // 연결 Timeout 로깅, 1초후 재연결 로깅
                            m_clientDisconnectDictionary[storeCode][isHost].Enqueue(client);
                        }
                    }
                    catch (Exception e)
                    {
                        // Exception 로깅, 1초후 재연결 로깅
                        client.Close();
                        m_clientDisconnectDictionary[storeCode][isHost].Enqueue(client);
                        LogManager.Instance.Log($"TCPCleitnConnectTask Fail. {e.Message}", Thread.CurrentThread.ManagedThreadId);
                    }
                    finally
                    {
                        await Task.Delay(1000); //Todo. 재연결 대기시간 Config로 정리필요
                    }
                }
                else
                {
                    await Task.Delay(1000);
                }
            }
        }


        /// <summary>
        /// 점포코드-Host Type 별 메세지 송수신 처리 Task 실행 함수. SendToHost()를 Task로 실행.        
        /// 서비스에서 이 함수를 호출해야 함. 
        /// </summary>
        public bool StartTcpClient()
        {
            //foreach (bool check in m_initCheckFlag)
            //{
            //    if (!check)
            //    {
            //        // TcpClientManager 동작에 필요한 필수 함수 실행 안됐다는 에러 출력.
            //        return false;
            //    }
            //}

            foreach (string storeCode in m_clientConnectDictionary.Keys)
            {
                foreach (bool isHost in m_clientConnectDictionary[storeCode].Keys)
                {
                    m_sendToHostDictionary.TryAdd(storeCode + isHost.ToString(), Task.Run(() => SendToHost(storeCode, isHost)));
                }
            }
            LogManager.Instance.Log("start tcp client sucecss", Thread.CurrentThread.ManagedThreadId);
            return true;
        }



        /// <summary>
        /// 점포코드 별 호스트에 연결 된 세션객체를 얻어온다. (필요할까??)         
        /// </summary>
        /// <param name="storeCode">점포코드 4자리</param>
        /// <param name="isHost">Host Type (True == Main / False == Sub)</param>
        /// <returns></returns>
        public TcpClient GetConnect(string storeCode, bool isHost = true)
        {
            if (storeCode.Length != 4)
            {
                // 점포코드 정합성 오류 로그
                return null;
            }
            if (m_clientConnectDictionary[storeCode][isHost].TryDequeue(out TcpClient client))
            {
                return client;
            }
            else
            {
                // 세션객체 없음 에러 로깅
                return null;
            }
        }

        /// <summary>
        /// HostMessage 클래스 안에는 점포코드, 전송할 호스트 세션의 Main/Sub 여부, POS의 TCPClient 객체, 전송 메세지 (Byte) 
        /// 큐에 넣기 전에 이미 byte 버퍼는 encoding이 되어있어야 한다.
        /// </summary>
        /// <param name="storeCode">점포코드 4자리</param>
        /// <param name="isHost">Host Type (True == Main / False == Sub)</param>
        /// <param name="msg">Host에 전송 할 데이터 Class</param>
        public void Send(string storeCode, bool isHost, HostMessage msg)
        {
            m_sendToHostMsgDictionary[storeCode][isHost].Enqueue(msg);
        }

        /// <summary>
        /// Host와 데이터 송수신 하는 Work Function.
        /// 점포코드-Host Type 별 Task가 생성되어 동작한다.
        /// </summary>
        /// <param name="storeCode">점포코드 4자리</param>
        /// <param name="isHost">Host Type (True == Main / False == Sub)</param>
        public void SendToHost(string storeCode, bool isHost)
        {
            // 여기서는 파라미터에 대한 정합성 체크를 별도로 하지 않는다. 왜냐면 SendToHost가 호출되기까지 넘어오는 인자들은 이미 정합성 체크가 완료되었기 때문이다.
            LogManager.Instance.Log($"Start SendToHost Task (ID : {Task.CurrentId.Value}) (Storecode : {storeCode}) (isHost : {isHost})", Task.CurrentId.Value);
            TcpClient hostSession;
            string ip;
            int port;

            while (true)
            {
                // Todo. 현재는 멀티세션이여도 1개 세션에 대해서만 처리되도록 구현하였다. 멀티세션에 대해서 처리 방법 고려 필요.
                if (m_clientConnectDictionary[storeCode][isHost].TryDequeue(out hostSession))
                {
                    ip = GetClientIP(hostSession);
                    port = GetClientPort(hostSession);
                    while (true)
                    {
                        if (m_sendToHostMsgDictionary[storeCode][isHost].TryDequeue(out HostMessage msg))
                        {
                            // 비동기 함수에서는 timeout 멤버변수를 사용하지 않는다. CancellationToekn을 사용해야 한다.
                            // 현재까지는 동기함수 사용.
                            NetworkStream stream = hostSession.GetStream();
                            stream.Write(msg.ReqMsg, 0, msg.ReqSendSize);
                            msg.ResRecvSize = stream.Read(msg.ResMsg, 0, msg.ResMsg.Length);

                            if (msg.ResRecvSize > 0)
                            {
                                m_recvFromHostMsgDictionary[storeCode][isHost].Enqueue(msg);
                            }
                            else
                            {
                                LogManager.Instance.Log("ReadAsync bytes < 0", Task.CurrentId.Value);
                                break;
                            }

                        }
                        else // 메세지 큐에 메세지가 없다면?
                        {
                            // 할일없지..                            
                        }
                    }
                    // Recv Length <= 0 이라서 연결을 끊고 재연결
                    TcpClientReconnect(storeCode, isHost, hostSession);
                }
                else // 연결 큐에 호스트 연결정보가 없다면?
                {
                    // 연결이 없습니다. 오류 발생
                }
            }
        }

        // 서비스함수가 이 함수를 사용하여 계속 응답 메세지를 polling 한다.
        // 
        public HostMessage RecvFromHost(string storeCode, bool isHost)
        {
            while (true)
            {
                if (m_recvFromHostMsgDictionary[storeCode][isHost].TryDequeue(out HostMessage recvMsg))
                {
                    return recvMsg;
                }
                else
                {
                    // 큐에 메세지가 없을경우 여기서 계속 지연이 발생해도 될까??
                }
            }
        }

        public void TcpClientDisconnect()
        {
            // 서비스가 종료될 때 가지고있는 모든 연결을 끊어야 함. 구현 필요
        }

        public void TcpClientReconnect(string storeCode, bool isHost, TcpClient client)
        {

        }

        public static string GetClientIP(TcpClient client)
        {
            return ((IPEndPoint)client.Client.RemoteEndPoint).Address.ToString();
        }
        public static int GetClientPort(TcpClient client)
        {
            return ((IPEndPoint)client.Client.RemoteEndPoint).Port;
        }
    }
}
