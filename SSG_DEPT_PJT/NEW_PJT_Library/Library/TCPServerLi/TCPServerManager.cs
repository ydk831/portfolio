using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using LogLibrary;
using MessageLibrary;

namespace TcpServerLibrary
{
    public class TCPServerManager
    {
        /// <summary>
        /// TcpServerManager Singleton 객체
        /// </summary>
        private static TCPServerManager m_instance;
        /// <summary>
        /// Singleton 객체 접근을 위한 Lock
        /// </summary>
        private static object m_instanceLock = new object();
        /// <summary>
        /// Singleton 객체 폐기 Flag
        /// </summary>
        private bool m_isDispose = false;
        /// <summary>
        /// Server Socket
        /// </summary>
        private TcpListener m_socketServer;
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
        /// 연결된 Client 소켓을 관리하는 Queue
        /// </summary>
        private ConcurrentQueue<TcpClient> m_tcpClientQ = new ConcurrentQueue<TcpClient>();
        /// <summary>
        /// Client로부터 수신받은 메세지를 관리하는 Queue
        /// </summary>
        private ConcurrentQueue<PosMessage> m_recvRequestQ = new ConcurrentQueue<PosMessage>();
        /// <summary>
        /// Client로 전송 할 메세지를 관리하는 Queue
        /// </summary>
        private ConcurrentQueue<PosMessage> m_sendResponseQ = new ConcurrentQueue<PosMessage>();


        private CancellationTokenSource m_AcceptCancelToken = new CancellationTokenSource();

        /// <summary>
        /// 외부에서 생성자 접근을 못하게 private로 명시 
        /// </summary>      
        private TCPServerManager() { }

        /// <summary>
        /// Singleton 처리를 위한 객체 getter
        /// </summary>
        public static TCPServerManager Instance
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
                            m_instance = new TCPServerManager();
                        }
                    }
                }
                return m_instance;
            }
        }

        /// <summary>
        /// TCPServerManager 초기화 함수. TCP 연결에 필요한 값 설정.
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
        }

        /// <summary>
        /// TCP Server 시작 함수
        /// Accept, Request Receive, Response Send 를 비동기로 처리
        /// </summary>
        /// <param name="port">서버 Listen Port</param>
        public void TcpServerStart(int port)
        {
            try
            {
                m_socketServer = new TcpListener(IPAddress.Any, port);
                m_socketServer.Server.SetSocketOption(SocketOptionLevel.IP, SocketOptionName.ReuseAddress, true);
                m_socketServer.Server.SetSocketOption(SocketOptionLevel.Socket, SocketOptionName.DontLinger, true);
                m_socketServer.Start();
                Task Accept = AcceptAsync(m_AcceptCancelToken.Token);
                Task CheckRecvRequestQ = CheckTcpClientQ(m_AcceptCancelToken.Token);
                Task ChekcSendResponseQ = CheckPosResponseQ(m_AcceptCancelToken.Token);
                LogManager.Instance.Log("tcp server start success", Thread.CurrentThread.ManagedThreadId);
            }
            catch (Exception e)
            {
                //Console.WriteLine("Listen Fail");
                //Environment.Exit(0);
                LogManager.Instance.Log($"tcp server start fail {e.Message}", Thread.CurrentThread.ManagedThreadId);
            }
        }
        ~TCPServerManager()
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
        /// 비동기 Client Accept Task
        /// </summary>
        /// <param name="token">Task Cancellation Token</param>
        /// <returns>Task</returns>
        public async Task AcceptAsync(CancellationToken token)
        {
            while (true)
            {
                if (token.IsCancellationRequested)
                {
                    break;
                }
                try
                {
                    TcpClient client = await m_socketServer.AcceptTcpClientAsync();
                    //LogManager.Instance.Log("Accept Success");
                    client.ReceiveTimeout = 1000;
                    client.ReceiveBufferSize = 1024 * 10; // byte
                    client.SendTimeout = 1000;
                    m_tcpClientQ.Enqueue(client);
                }
                catch (Exception e)
                {
                    LogManager.Instance.Log($"Accept Fail. {e.Message}", Thread.CurrentThread.ManagedThreadId);
                }
            }
        }
        /// <summary>
        /// 연결된 Client에서 데이터 Read 하는 비동기 Task
        /// </summary>
        /// <returns>Task </returns>
        public async Task CheckTcpClientQ(CancellationToken token)
        {
            while (true)
            {
                try
                {
                    if (token.IsCancellationRequested && m_tcpClientQ.IsEmpty)
                    {
                        break;
                    }
                    if (m_tcpClientQ.TryDequeue(out TcpClient client))
                    {
                        //LogManager.Instance.Log("TCQueue TryDequeue.");
                        //Task _ = Task.Run(() => GetClientStreamAsync(tc));
                        //Task.Run(() => GetClientStreamAsync(client));
                        Task ReadSocket = GetClientStreamAsync(client);
                    }
                    else
                    {
                        //LogManager.Log("TCQueue is Empty. Wait 1Sec.");
                        await Task.Delay(1000);
                    }
                }
                catch (Exception e)
                {
                    LogManager.Instance.Log($"TryDequeue Fail. {e.Message}", Thread.CurrentThread.ManagedThreadId);
                }
            }
        }
        public async Task GetClientStreamAsync(TcpClient client)
        {
            try
            {
                using (CancellationTokenSource cts = new CancellationTokenSource(m_recvTimeout))
                {
                    PosMessage PosMsg = new PosMessage(client, client.ReceiveBufferSize);
                    NetworkStream stream = PosMsg.Client.GetStream();
                    //LogManager.Instance.Log("Client GetStream.");
                    PosMsg.ReqRecvSize = await stream.ReadAsync(PosMsg.ReqMsg, 0, PosMsg.ReqMsg.Length, cts.Token);

                    //LogManager.Instance.Log("ReadAsync.");
                    if (PosMsg.ReqRecvSize > 0)
                    {
                        //Console.WriteLine("PosReqQueue Enqueue.");
                        m_recvRequestQ.Enqueue(PosMsg);
                    }
                    else
                    {
                        //LogManager.Instance.Log("ReadAsync bytes < 0");
                    }
                }
            }
            catch (ObjectDisposedException e)
            {
                LogManager.Instance.Log($"TCPClient가 닫혔습니다. {e.Message}", Thread.CurrentThread.ManagedThreadId);
            }
            catch (InvalidOperationException e)
            {
                LogManager.Instance.Log($"TCPClient가 원격 호스트에 연결되어 있지 않음. {e.Message}", Thread.CurrentThread.ManagedThreadId);
            }
            catch (TimeoutException e)
            {
                // Read timeout wait for m_recvTimeout millisecond
                LogManager.Instance.Log($"TCPClient Read Timeout. {e.Message}", Thread.CurrentThread.ManagedThreadId);
            }
            catch (Exception e)
            {
                LogManager.Instance.Log($"GetclientStreamAsync Fail. {e.Message}", Thread.CurrentThread.ManagedThreadId);
            }
        }

        /// <summary>
        /// 서비스 스레드에서 이것처럼 posreqestQ를 풀링하자.
        /// </summary>
        /// <returns></returns>
        //public async Task<PosMessage> CheckPosRequestQ()
        //{
        //    int retry = 3;

        //    try
        //    {
        //        for (int i = 0; i < retry; i++)
        //        {
        //            if (m_recvRequestQ.TryDequeue(out PosMessage PosMsg))
        //            {
        //                //LogManager.Instance.Log("_posRequestQ TryDequeue.");
        //                //Task.Run(() => ServiceLogic1(PosMsg));
        //                //Task Service = ServiceLogic1(PosMsg);
        //                return PosMsg;
        //            }
        //            else
        //            {
        //                //Console.WriteLine("posReqQueue is Empty. Wait 1Sec.");
        //                await Task.Delay(100);
        //            }
        //        }
        //    }
        //    catch (Exception e)
        //    {
        //        //LogManager.Instance.Log("_posRequestQ Fail. {0}", e.Message);
        //    }
        //    return null;

        //}
        //public static async Task ServiceLogic1(PosMessage PosMsg)
        //{
        //    try
        //    {
        //        string msg = Encoding.UTF8.GetString(PosMsg.ReqMsg, 0, PosMsg.ReqRecvSize);
        //        LogManager.Instance.Log($"Recevie : {msg} at {DateTime.Now}");

        //        PosMsg.ResMsg = Encoding.UTF8.GetBytes("Response for " + msg);
        //        PosMsg.ResSendSize = PosMsg.ResMsg.Length;
        //        LogManager.Instance.Log("m_sendResponseQ Enqueue.");
        //        m_sendResponseQ.Enqueue(PosMsg);

        //    }
        //    catch (Exception e)
        //    {
        //        LogManager.Instance.Log("ServiceLogic1 Fail. {0}", e.Message);
        //    }
        //}

        public void SendResponse(PosMessage res)
        {
            m_sendResponseQ.Enqueue(res);
        }

        /// <summary>
        /// Client로 전송 할 응답 메시지 큐를 체크하는 비동기 Task
        /// </summary>
        /// <returns></returns>
        public async Task CheckPosResponseQ(CancellationToken token)
        {
            while (true)
            {
                try
                {
                    if (token.IsCancellationRequested && m_sendResponseQ.IsEmpty)
                    {
                        break;
                    }
                    using (CancellationTokenSource cts = new CancellationTokenSource(m_sendTimeout))
                    {
                        if (m_sendResponseQ.TryDequeue(out PosMessage PosMsg))
                        {
                            //LogManager.Instance.Log("PosRspQueue TryDequeue.");
                            NetworkStream stream = PosMsg.Client.GetStream();
                            //LogManager.Instance.Log("Client GetStream.");
                            await stream.WriteAsync(PosMsg.ResMsg, 0, PosMsg.ResSendSize, cts.Token);
                            //LogManager.Instance.Log("WriteAsync.");
                            stream.Close();
                            PosMsg.Client.Close();
                        }
                        else
                        {
                            //Console.WriteLine("posRspQueue is Empty. Wait 1Sec.");
                            await Task.Delay(1000);
                        }
                    }
                }
                catch (Exception e)
                {
                    LogManager.Instance.Log($"CheckPosResonseQ Fail. {e.Message}", Thread.CurrentThread.ManagedThreadId);
                }
            }
        }
    }
}
