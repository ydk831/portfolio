using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Configuration;
using System.Runtime.InteropServices;
using LogLibrary;

namespace ConfigLibrary
{
    // 프로젝트 > 참조 우클릭 > 참조추가 > System.Configuration 검색 후 추가
    // 참고 : https://itbrain.tistory.com/entry/AppConfig-XML%ED%8C%8C%EC%9D%BC%EC%9D%84-%EC%9D%BD%EA%B3%A0-%EC%93%B0%EB%8A%94-%EA%B0%84%EB%8B%A8%ED%95%9C-%ED%95%A8%EC%88%98
    // 참고 : https://m.blog.naver.com/PostView.nhn?blogId=csaiur&logNo=220242402103&proxyReferer=https:%2F%2Fwww.google.com%2F


    public class ConfigManager
    {
        public const string DEF_DUMMY_KEY = "1001100110011001"; // 암/복호화용 더미키  : 20170407 WJ CHOI

        /// <summary>
        /// Singleton 객체 폐기 Flag
        /// </summary>
        private bool m_isDispose = false;

        /// <summary>
        /// ConfigManager Singleton 객체
        /// </summary>
        private static ConfigManager m_instance;
        /// <summary>
        /// Singleton 객체 접근을 위한 Lock
        /// </summary>
        private static readonly object m_instanceLock = new object();
        /// <summary>
        /// app.config 파일 접근 객체
        /// </summary>
        private static Configuration m_appConfig;


        //ini 파일 Read / Write를 위한 WinApi32 호출 설정
        [DllImport("kernel32")]
        private static extern long WritePrivateProfileString(string section, string key, string val, string filePath);
        [DllImport("kernel32")]
        private static extern int GetPrivateProfileString(string section, string key, string def, StringBuilder retVal, int size, string filePath);
        //암호화 ini 데이터 읽기 위한 w32dll 호출 설정
        [DllImport(@"C:\NSP\APL\W32DLL.dll")]
        //private static extern int dec_proc(string _pOrgMsg, string _pKey, string _pDecMsg);
        private static extern int AES_Base64_Decrypt(string key, string iv, string enc, StringBuilder data);

        /// <summary>
        /// 외부에서 생성자 접근을 못하게 private로 명시 
        /// </summary>
        private ConfigManager() { }
        ~ConfigManager()
        {
            if (!m_isDispose)
            {
                Dispose();
            }
        }

        /// <summary>
        /// Singleton 객체 폐기 함수
        /// </summary>
        private void Dispose()
        {
            m_isDispose = true;

            // 메모리해제

            GC.SuppressFinalize(this); // GC가 이 객체에 대해서 Finalize를 호출하지 않도록 명시적으로 지정
        }

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
            string strReturn = "";
            try
            {

                if (m_appConfig.AppSettings.Settings.AllKeys.Contains(key)) // key 가 있으면
                {
                    strReturn = m_appConfig.AppSettings.Settings[key].Value;
                }
                else // key 가 없으면
                {
                    LogManager.Instance.Log("Configuration Not Found " + key, Task.CurrentId.Value);
                    strReturn = "";
                }
                return strReturn;
            }
            catch
            {
                // exception 로그츨력
                return strReturn;
            }

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

        /// <summary>
        /// Ini 파일 read 함수
        /// </summary>
        /// <param name="iniPath">INI 파일 경로</param>
        /// <param name="section">INI 섹션</param>
        /// <param name="key">찾을 변수</param>
        /// <param name="defaultValue">찾을 변수의 값이 없을 경우 default 값 처리</param>
        /// <returns></returns>
        public string IniConfigRead(string iniPath, string section, string key, string defaultValue = "")
        {
            StringBuilder returnValue = new StringBuilder(255);
            GetPrivateProfileString(section, key, defaultValue, returnValue, returnValue.Capacity, iniPath);

            return returnValue.ToString();
        }
        /// <summary>
        /// Ini 파일 Write 함수
        /// </summary>
        /// <param name="iniPath">Ini 파일 경로</param>
        /// <param name="section">Ini 섹션</param>
        /// <param name="key">write 할 변수</param>
        /// <param name="writeValue"> write 할 값</param>
        public void IniConfigWrtie(string iniPath, string section, string key, string writeValue)
        {
            WritePrivateProfileString(section, key, writeValue, iniPath);
        }

        /// <summary>
        /// Ini 파일 암호화 데이터 read 함수
        /// </summary>
        /// <param name="iniPath">INI 파일 경로</param>
        /// <param name="section">INI 섹션</param>
        /// <param name="key">찾을 변수</param>
        /// <param name="defaultValue">찾을 변수의 값이 없을 경우 default 값 처리</param>
        /// <returns></returns>
        public string IniConfigReadDec(string iniPath, string section, string key, string defaultValue = "")
        {
            string encValue = IniConfigRead(iniPath, section, key, defaultValue);

            if (encValue == "")
            {
                return "";
            }

            StringBuilder decValue = new StringBuilder(255);
            int nLen = AES_Base64_Decrypt(DEF_DUMMY_KEY, DEF_DUMMY_KEY, encValue, decValue);
            if (nLen <= 0) // 복호화가 안되면 평문으로 간주하고 복호화 전 값 전달
            {
                return encValue;
            }
            else
            {
                return decValue.ToString();
            }
        }
    }
}
