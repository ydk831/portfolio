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

namespace NSPCCSIS
{
    /// <summary>
    /// HttpClient 를 사용하여 RestAPI Client 동작하는 Class
    /// - Send 할 때는 Newtonsoft.Json.JsonConvert.SerializeObject() 를 사용하여 Class를 HttpContent로 변환하여 전송
    /// - Recv 할 때는 Newtonsoft.Json.JsonConvert.DeserializeObject<T>() 를 사용하여 HttpContent를 Class로 변환하여 수신
    /// </summary>
    public class RestApiClient
    {
        private HttpClient m_client;
        private string m_url;
        private string m_metadata;

        /// <summary>
        /// 생성자
        /// </summary>
        public RestApiClient()
        {
            m_client = new HttpClient();
        }

        /// <summary>
        /// RestApiClient 초기화
        ///  - URL과 METADATA 설정
        /// </summary>
        /// <param name="url">접속할 ROOT URL</param>
        /// <param name="metadata">사용할 HTTP METADATA</param>
        /// <returns></returns>
        public void Initialize(string url, string metadata)
        {
            m_url = url;
            m_metadata = metadata;

            try
            {
                m_client.BaseAddress = new Uri(url); // Root URL
                m_client.DefaultRequestHeaders.Accept.Clear();
                m_client.DefaultRequestHeaders.Accept.Add(
                    new MediaTypeWithQualityHeaderValue(metadata)); // METADATA FORMAT
                m_client.Timeout = TimeSpan.FromSeconds(300); // timeout second
            }
            catch (Exception e)
            {
                throw new Exception("RestApiClient Initialize Fail.\n" + e.Message);
            }
        }

        /// <summary>
        /// GET METHOD 호출 함수...  현재 안씀
        /// </summary>
        /// <param name="getUri"></param>
        /// <returns></returns>
        //public async Task<object> GetAsync(string getUri)
        //{
        //    HttpResponseMessage response = await client.GetAsync(getUri);
        //    if (response.IsSuccessStatusCode)
        //    {
        //        return await response.Content.ReadAsStringAsync();
        //    }
        //    else
        //    {
        //        return "";
        //    }
        //}

        /// <summary>
        /// POST에 사용하는 SEND 함수
        /// </summary>
        /// <param name="method">HTTP METHOD</param>
        /// <param name="uri">HTTP URI. BaseURL 뒤에 부분만 입력</param>
        /// <param name="msg">HTTP BODY 데이터. class 사용</param>
        /// <returns>HttpResponse Body (string). 외부에서 class로 변환하여 사용</returns>
        public async Task<object> Send<T>(HttpMethod method, string uri, T msg)
        {
            string body = JsonConvert.SerializeObject(msg);
            HttpRequestMessage request = new HttpRequestMessage(method, uri)
            {
                Content = new StringContent(body, Encoding.UTF8, "application/json")
            };

            LogManager.Instance.Log($"Send Data :\r\n {body}");

            try
            {
                HttpResponseMessage response = await m_client.SendAsync(request);

                if (response.IsSuccessStatusCode)
                {
                    string resString = response.Content.ReadAsStringAsync().Result;
                    return resString;
                }
                else
                {
                    LogManager.Instance.Log($"Response Code Error : {response.StatusCode}");
                    LogManager.Instance.AMSLogPrint("ERR", "시코르 샵인샵 실적 업데이트 HTTP 응답 오류! (" + response.StatusCode.ToString() + ")");
                    return null;
                }
            }
            catch (Exception e)
            {
                LogManager.Instance.Log("RestApi Send Exception.\n" + e.Message);
                LogManager.Instance.AMSLogPrint("ERR", "시코르 샵인샵 실적 업데이트 SEND 오류! " + e.Message);
                return null;
            }
        }

        /// <summary>
        /// GET에 사용하는 SEND 함수
        /// </summary>
        /// <param name="method">HTTP METHOD</param>
        /// <param name="uri">HTTP URI. BaseURL 뒤에 부분만 입력</param>
        /// <returns>HttpResponse Body (string). 외부에서 class로 변환하여 사용</returns>
        public async Task<object> Send(HttpMethod method, string uri)
        {
            HttpRequestMessage request = new HttpRequestMessage(method, uri);

            HttpResponseMessage response = await m_client.SendAsync(request);
            if (response.IsSuccessStatusCode)
            {
                string resString = response.Content.ReadAsStringAsync().Result;
                return resString;
            }
            else
            {
                LogManager.Instance.Log($"Response Code Error : {response.StatusCode}");
                return null;

            }
        }

        /// <summary>
        /// Delete에 사용하는 함수
        /// </summary>
        /// <param name="deleteUri">HTTP URI. BaseURL 뒤에 부분만 입력</param>
        /// <returns></returns>
        public async Task<HttpStatusCode> DeleteAsync(string deleteUri)
        {
            HttpResponseMessage response = await m_client.DeleteAsync(
                deleteUri);
            return response.StatusCode;
        }

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
