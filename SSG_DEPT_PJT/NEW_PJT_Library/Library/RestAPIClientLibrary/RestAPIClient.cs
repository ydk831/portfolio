using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Net.Sockets;
using System.Text;
using System.Threading.Tasks;
using MessageLibrary;

namespace RestAPIClientLibrary
{
    /// <summary>
    /// RestAPIClient는 실제 웹서버와 연동 될 Http 세션 정보를 갖는 클래스이며, RestApiClientManager 클래스를 통해서 데이터를 송수신한다.
    /// </summary>
    internal class RestAPIClient
    {
        /// <summary>
        /// Singleton 객체 폐기 Flag
        /// </summary>
        private bool m_isDispose = false;
        /// <summary>
        /// Http 세션 객체
        /// </summary>
        internal HttpClient Client { get; set; }
        /// <summary>
        /// 웹 URI 도메인 주소
        /// ex) http://www.naver.com/
        /// </summary>
        internal string BaseUri { get; set; }
        /// <summary>
        /// MediaTypeWithQualityHeaderValue
        /// </summary>
        internal string Metadata { get; set; }
        /// <summary>
        /// Connection, Send, Recv에 공통으로 적용되는 Timeout 값. (sec)
        /// </summary>
        internal int Timeout { get; set; }

        /// <summary>
        /// RestApiClient 생성자
        /// BaseUri, Metadata, Timeout 설정
        /// <param name="baseUri">접속할 ROOT URL</param>
        /// <param name="metadata">클라이언트가 허용할 수 있는 파일 형식(MIME TYPE)</param>
        /// <param name="timeout">Connect, Send, Recv 공통 Timeout (sec)</param>
        internal RestAPIClient(string baseUri, string metadata, int timeout) //HttpMethod method, )//, Encoding encoding, TcpClient tcpclient = null)
        {
            Client = new HttpClient();

            BaseUri = baseUri;
            Metadata = metadata;
            Timeout = timeout;

            try
            {
                Client.BaseAddress = new Uri(BaseUri); // Root URL
                Client.DefaultRequestHeaders.Accept.Clear();
                Client.DefaultRequestHeaders.Accept.Add(
                    new MediaTypeWithQualityHeaderValue(Metadata)); // METADATA FORMAT
                Client.Timeout = TimeSpan.FromSeconds(Timeout); // timeout second
            }
            catch (Exception e)
            {
                throw new Exception("RestApiClient Initialize Fail.\n" + e.Message);
            }
        }

        ~RestAPIClient()
        {
            if (!m_isDispose)
            {
                Dispose();
            }
        }

        /// <summary>
        /// Singleton 객체 폐기 함수
        /// </summary>
        internal void Dispose()
        {
            m_isDispose = true;

            // 메모리해제
            Client.Dispose();

            GC.SuppressFinalize(this); // GC가 이 객체에 대해서 Finalize를 호출하지 않도록 명시적으로 지정
        }

    }
}
