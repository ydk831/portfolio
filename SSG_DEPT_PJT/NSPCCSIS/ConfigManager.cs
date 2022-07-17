using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Configuration;

namespace NSPCCSIS
{
    // 참고 : https://itbrain.tistory.com/entry/AppConfig-XML%ED%8C%8C%EC%9D%BC%EC%9D%84-%EC%9D%BD%EA%B3%A0-%EC%93%B0%EB%8A%94-%EA%B0%84%EB%8B%A8%ED%95%9C-%ED%95%A8%EC%88%98
    // 참고 : https://m.blog.naver.com/PostView.nhn?blogId=csaiur&logNo=220242402103&proxyReferer=https:%2F%2Fwww.google.com%2F

    class ConfigManager
    {
        private static ConfigManager m_instance;
        private static readonly object m_instanceLock = new object();

        private static Configuration m_appConfig;
        
        /// <summary>
        /// 외부에서 생성자 접근을 못하게 private로 명시 
        /// </summary>
        private ConfigManager() { }
        /// <summary>
        ///  Singleton 처리를 위한 객체 getter
        /// </summary>
        public static ConfigManager Instance
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
                            m_instance = new ConfigManager();
                        }
                    }
                }
                return m_instance;
            }
        }

        /// <summary>
        ///  ConfigManager 객체 초기화 
        ///   - app.config 접근권한 전체로 설정
        /// </summary>
        public void Initialize()
        {
            try
            {
                m_appConfig = ConfigurationManager.OpenExeConfiguration(ConfigurationUserLevel.None);
            }
            catch (Exception e)
            {
                throw new Exception("ConfigManager Initialize Fail.\n" + e.Message);
            }
        }

        /// <summary>
        /// app.config 파일 read 함수
        /// </summary>
        /// <param name="key"> config 파라미터 이름</param>
        /// <returns>파라미터 값</returns>
        public string AppConfigRead(string key)
        {
            string strReturn;

            if (m_appConfig.AppSettings.Settings.AllKeys.Contains(key)) // key 가 있으면
            {
                strReturn = m_appConfig.AppSettings.Settings[key].Value;
            }
            else // key 가 없으면
            {
                strReturn = "";
            }
            return strReturn;
        }

        /// <summary>
        /// app.config 파일 write 함수
        ///  - write 할 파라미터가 없다면 신규 추가 함
        ///  - ThreadSafe 하게 lock을 걸고 처리함
        /// </summary>
        /// <param name="key">write 할 파라미터 명</param>
        /// <param name="value">write 할 파라미터 값</param>
        /// <returns>성공/실패</returns>
        public bool AppConfigWrite(string key, string value)
        {
            try
            {
                lock (m_instanceLock)
                {
                    if (m_appConfig.AppSettings.Settings.AllKeys.Contains(key)) // key 가 있으면
                    {
                        m_appConfig.AppSettings.Settings[key].Value = value;
                    }
                    else // key 가 없으면
                    {
                        m_appConfig.AppSettings.Settings.Add(key, value);
                    }
                    m_appConfig.Save();
                    ConfigurationManager.RefreshSection("appSettings");
                }
                return true;
            }
            catch (Exception)
            {
                // Todo. Log Write
                return false;
            };
        }
    }
}
