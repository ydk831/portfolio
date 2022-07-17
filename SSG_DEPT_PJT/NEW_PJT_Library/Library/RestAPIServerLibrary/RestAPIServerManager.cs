using System;
using System.Collections.Generic;
using System.Runtime.Serialization;
using System.ServiceModel;
using System.ServiceModel.Description;
using System.ServiceModel.Web;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using LogLibrary;

// C# 콘솔 프로젝트 > 참조 > ServiceModel, ServiceModel.Web 추가
// App.Config 수정
// C# 콘솔 프로젝트 > 참조 >system.runtime.serialization 추가
// newtonsoft.json 누겟 추가 필요
namespace RestAPIServerLibrary
{
    /// <summary>
    /// REST API Server를 시작하는 Class
    /// 미완성
    /// </summary>
    public class RestAPIServerManager
    {
        /// <summary>
        /// Singleton 객체 폐기 Flag
        /// </summary>
        private bool m_isDispose = false;
        /// <summary>
        /// MSSQL Singleton 객체
        /// </summary>
        private static RestAPIServerManager m_instance;
        /// <summary>
        /// Singleton 객체 접근을 위한 Lock
        /// </summary>
        private static object m_instanceLock = new object();
        /// <summary>
        /// App.config baseaddress 자동 입력
        /// </summary>
        private static WebServiceHost host = new WebServiceHost(typeof(Service));

        private RestAPIServerManager() { }

        ~RestAPIServerManager()
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

            GC.SuppressFinalize(this); // GC가 이 객체에 대해서 Finalize를 호출하지 않도록 명시적으로 지정
        }

        /// <summary>
        /// Singleton 처리를 위한 객체 getter
        /// </summary>
        public static RestAPIServerManager Instance
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
                            m_instance = new RestAPIServerManager();
                        }
                    }
                }
                return m_instance;
            }
        }

        public void RestAPIServerStart()
        {
            try
            {
                host.Open();
                LogManager.Instance.Log("RestApiServerStart!", Thread.CurrentThread.ManagedThreadId);

                //host.Close();
            }
            catch (Exception e)
            {
                LogManager.Instance.Log($"RestApiServerStart Exception! {e.Message}!", Thread.CurrentThread.ManagedThreadId);
                host.Abort();
            }
        }
        public void RestAPIServerStop()
        {

            try
            {
                host.Close();
            }
            catch (CommunicationException cex)
            {
                LogManager.Instance.Log($"RestApiServerStop Exception {cex.Message}!", Task.CurrentId.Value);
                host.Abort();
            }
        }
    }
}
