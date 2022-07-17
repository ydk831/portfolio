using System;
using System.Collections.Generic;
using Newtonsoft.Json;
using Newtonsoft.Json.Converters;
using Newtonsoft.Json.Serialization;
using Newtonsoft.Json.Linq;
using System.Linq;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Threading.Tasks;
using System.Net;
using LogLibrary;
using MessageLibrary;
using System.Collections.Concurrent;
using System.Threading;

namespace RestAPIClientLibrary
{
    /// <summary>
    /// HttpClient 를 사용하여 RestAPI Client 동작하는 Class
    /// - Send 할 때는 Newtonsoft.Json.JsonConvert.SerializeObject() 를 사용하여 Class를 HttpContent로 변환하여 전송
    /// - Recv 할 때는 Newtonsoft.Json.JsonConvert.DeserializeObject<T>() 를 사용하여 HttpContent를 Class로 변환하여 수신
    /// </summary>
    /// 프로젝트 > 참조 우클릭 > 참조추가 > 프로젝트에서 LogLibrary 추가
    public class RestAPIClientManager
    {
        /// <summary>
        /// RestAPIClientManager Singleton 객체
        /// </summary>
        private static RestAPIClientManager m_instance;
        /// <summary>
        /// Singleton 객체 접근을 위한 Lock
        /// </summary>
        private static readonly object m_instanceLock = new object();
        /// <summary>
        /// Singleton 객체 폐기 Flag
        /// </summary>
        private bool m_isDispose = false;

        /// <summary>
        /// BaseUri 별로 세션을 갖는 Dictionary
        /// </summary>
        private ConcurrentDictionary<string, RestAPIClient> m_restAPISessionDictionary = new ConcurrentDictionary<string, RestAPIClient>();
        /// <summary>
        /// BaseUri 별로 Request 메세지를 처리하는 Queue를 갖는 Dictionary
        /// </summary>
        private ConcurrentDictionary<string, ConcurrentQueue<RestAPIMessage>> m_restAPISendDictionary = new ConcurrentDictionary<string, ConcurrentQueue<RestAPIMessage>>();
        /// <summary>
        /// /// BaseUri 별로 Response 메세지를 처리하는 Queue를 갖는 Dictionary
        /// </summary>
        private ConcurrentDictionary<string, ConcurrentQueue<RestAPIMessage>> m_restAPIRecvDictionary = new ConcurrentDictionary<string, ConcurrentQueue<RestAPIMessage>>();


        /// <summary>
        /// 외부에서 접근 할 수 없도록 생성자를 private로 선언
        /// </summary>
        private RestAPIClientManager() { }

        ~RestAPIClientManager()
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
        ///  Singleton 처리를 위한 객체 getter
        /// </summary>
        public static RestAPIClientManager Instance
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
                            m_instance = new RestAPIClientManager();
                        }
                    }
                }
                return m_instance;
            }
        }

        /// <summary>
        /// BaseUri 별로 세션을 만들어서 저장
        /// BaseUri 별로 Request / Response 메세지를 처리하는 Dictionary를 생성
        /// </summary>
        /// <param name="baseUri">Http Base Uri (ex. www.naver.com/)</param>
        /// <param name="metadata">클라이언트가 허용할 수 있는 파일 형식(MIME TYPE)</param>
        /// <param name="timeout">Connect, Send, Recv 공통 Timeout (sec)</param>
        /// <returns></returns>
        public bool Initialize(string baseUri, string metadata, int timeout) //HttpMethod method, )
        {
            try
            {
                if (m_restAPISessionDictionary.TryAdd(baseUri, new RestAPIClient(baseUri, metadata, timeout)))
                {
                    if (m_restAPISendDictionary.TryAdd(baseUri, new ConcurrentQueue<RestAPIMessage>()))
                    {
                        if (m_restAPIRecvDictionary.TryAdd(baseUri, new ConcurrentQueue<RestAPIMessage>()))
                        {
                            Task task = CheckSendRestAPI(baseUri);
                            LogManager.Instance.Log("rest client success", Thread.CurrentThread.ManagedThreadId);
                            return true;
                        }
                        else
                        {
                            m_restAPISendDictionary.TryRemove(baseUri, out ConcurrentQueue<RestAPIMessage> queue);
                            m_restAPISessionDictionary.TryRemove(baseUri, out RestAPIClient client);
                            client.Dispose();
                            LogManager.Instance.Log("rest client fail because can not adding recv dictionary", Thread.CurrentThread.ManagedThreadId);
                        }
                    }
                    else
                    {
                        m_restAPISessionDictionary.TryRemove(baseUri, out RestAPIClient client);
                        client.Dispose();
                        LogManager.Instance.Log("rest client fail because can not adding send dictionnary", Thread.CurrentThread.ManagedThreadId);
                    }

                }
                else
                {
                    // what can i do?                    
                    LogManager.Instance.Log("rest client fail becuase can not adding session dictionnary", Thread.CurrentThread.ManagedThreadId);
                }

                return false;
            }
            catch (Exception e)
            {
                LogManager.Instance.Log($"rest client fail {e.Message}", Thread.CurrentThread.ManagedThreadId);
                return false;
            }
        }

        private RestAPIClient GetSession(string baseUri)
        {
            return m_restAPISessionDictionary[baseUri];
        }

        public void SendAsync(string baseUri, RestAPIMessage msg)
        {
            m_restAPISendDictionary[baseUri].Enqueue(msg);
        }


        public async Task CheckSendRestAPI(string baseUri)
        {
            RestAPIClient client = GetSession(baseUri);
            while (true)
            {
                if (m_restAPISendDictionary[baseUri].TryDequeue(out RestAPIMessage msg))
                {
                    Send(client, msg);
                }
                else
                {
                    await Task.Delay(100);
                }
            }
        }
        private async void Send(RestAPIClient client, RestAPIMessage msg)
        {
            string body = JsonConvert.SerializeObject(msg.ReqMsg);
            HttpRequestMessage request;
            if (msg.Method == HttpMethod.Get)
            {
                request = new HttpRequestMessage(msg.Method, msg.URLPath);
            }
            else
            {
                request = new HttpRequestMessage(msg.Method, msg.URLPath)
                {
                    Content = new StringContent(body, msg.MsgEnc, "application/json")
                };
            }

            LogManager.Instance.Log($"Send Data :\r\n {body}", Thread.CurrentThread.ManagedThreadId);

            try
            {
                HttpResponseMessage response = await client.Client.SendAsync(request);

                if (response.IsSuccessStatusCode)
                {
                    msg.ReqMsg = response.Content.ReadAsStringAsync().Result;
                    m_restAPIRecvDictionary[client.BaseUri].Enqueue(msg);
                }
                else
                {
                    LogManager.Instance.Log($"Response Code Error : {response.StatusCode}", Thread.CurrentThread.ManagedThreadId);
                    LogManager.Instance.AMSLogPrint("ERR", "HTTP 응답 오류! (" + response.StatusCode.ToString() + ")");
                }
            }
            catch (Exception e)
            {
                LogManager.Instance.Log("RestApi Send Exception.\n" + e.Message, Thread.CurrentThread.ManagedThreadId);
                LogManager.Instance.AMSLogPrint("ERR", "RestApi Send Exception! " + e.Message);
            }
        }

        /// <summary>
        /// POST에 사용하는 SEND 함수
        /// </summary>
        /// <param name="method">HTTP METHOD</param>
        /// <param name="uri">HTTP URI. BaseURL 뒤에 부분만 입력</param>
        /// <param name="msg">HTTP BODY 데이터. class 사용</param>
        /// <returns>HttpResponse Body (string). 외부에서 class로 변환하여 사용</returns>
        public async Task<bool> Send(string baseUri, RestAPIMessage msg)
        {
            RestAPIClient client = GetSession(baseUri);

            string body = JsonConvert.SerializeObject(msg.ReqMsg);
            HttpRequestMessage request;
            if (msg.Method == HttpMethod.Get)
            {
                request = new HttpRequestMessage(msg.Method, msg.URLPath);
            }
            else
            {
                request = new HttpRequestMessage(msg.Method, msg.URLPath)
                {
                    Content = new StringContent(body, msg.MsgEnc, "application/json")
                };
            }

            LogManager.Instance.Log($"Send Data :\r\n {body}", Thread.CurrentThread.ManagedThreadId);

            try
            {
                HttpResponseMessage response = await client.Client.SendAsync(request);

                if (response.IsSuccessStatusCode)
                {
                    msg.RspMsg = response.Content.ReadAsStringAsync().Result;
                    return true;
                }
                else
                {
                    LogManager.Instance.Log($"Response Code Error : {response.StatusCode}", Thread.CurrentThread.ManagedThreadId);
                    LogManager.Instance.AMSLogPrint("ERR", "HTTP 응답 오류! (" + response.StatusCode.ToString() + ")");
                    return false;
                }
            }
            catch (Exception e)
            {
                LogManager.Instance.Log("RestApi Send Exception.\n" + e.Message, Thread.CurrentThread.ManagedThreadId);
                LogManager.Instance.AMSLogPrint("ERR", "RestApi Send Exception! " + e.Message);
                return false;
            }
        }

        /// <summary>
        /// 서비스 스레드에서 이것처럼 posreqestQ를 풀링하자. 이건 삭제
        /// </summary>
        /// <returns></returns>
        public async void RecvAsync(string baseUri)
        {
            while (m_restAPIRecvDictionary[baseUri].TryDequeue(out RestAPIMessage msg))
            {
                try
                {
                    LogManager.Instance.Log($"{msg.ToString()}", Thread.CurrentThread.ManagedThreadId);
                }
                catch (Exception e)
                {
                    LogManager.Instance.Log($"_posRequestQ Fail. {e.Message}", Thread.CurrentThread.ManagedThreadId);
                }
            }
        }

        public ConcurrentQueue<RestAPIMessage> GetRecvQueue (string baseUri)
        {
            return m_restAPIRecvDictionary[baseUri];
        }

        /// <summary>
        /// Delete에 사용하는 함수
        /// </summary>
        /// <param name="deleteUri">HTTP URI. BaseURL 뒤에 부분만 입력</param>
        /// <returns></returns>
        //public async Task<HttpStatusCode> DeleteAsync(string deleteUri)
        //{
        //    HttpResponseMessage response = await m_client.DeleteAsync(
        //        deleteUri);

        //    return response.StatusCode;
        //}

        /// <summary>
        /// SEND()후 HTTP RESPONSE Body를 class로 변환해주는 함수
        ///  ex) var response = SEND(HttpMethod.POST, URI, MSG);
        ///      T resClass = ConvertJsonString<T>(response);
        /// </summary>
        /// <typeparam name="T">변환할 class 객체</typeparam>
        /// <param name="obj">HTTP RESPONSE CONTENT STRING (BODY)</param>
        /// <returns>JsonToClass</returns>
        public T ConvertJsonString<T>(object obj)
        {
            return JsonConvert.DeserializeObject<T>(obj.ToString());
        }
    }
}
