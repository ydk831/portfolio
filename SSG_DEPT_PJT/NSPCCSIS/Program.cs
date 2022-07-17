using System;
using System.Collections;
using System.Collections.Generic;
using System.Data.SqlClient;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Threading.Tasks;
using System.Configuration;
using System.Threading;


namespace NSPCCSIS
{
    class Program
    {
        public static RestApiClient SPointClient = new RestApiClient(); // RestApiClient 객체 생성

        public static string date = ""; // 날짜 지정 처리 변수
        public static string storeCode = "";

        static void Main(string[] args)
        {
            int argc = args.Length;
            if (argc < 1)
            {
                //LogManager.Instance.Log("Command Error!! Usage : NSPCCSIS.exe {StoreCode} {Date}");
                LogManager.AMSLogPrintStatic("NSPCCSIS", "ERR", "NSPCCSIS 실행 실패! 점포코드 미입력!");
                Thread.Sleep(1000); // 로그 남기기 위해 임시조치로 1초 쉬고 종료..
                return;
            }
            else if (argc == 2)
            {
                date = args[1].ToString();
                if (date.Length != 8)
                {
                    //LogManager.Instance.Log("Command Error!! Usage : NSPCCSIS.exe {StoreCode} [YYYYMMDD]");
                    LogManager.AMSLogPrintStatic("NSPCCSIS", "ERR", "NSPCCSIS 실행 실패! 날짜 입력 오류!");
                    Thread.Sleep(1000); // 로그 남기기 위해 임시조치로 1초 쉬고 종료..
                    return;
                }
            }

            storeCode = args[0].ToString();
            if (storeCode.Length != 4)
            {
                //LogManager.Instance.Log("Command Error!! Usage : StoreCode Must 4 Length.");
                LogManager.AMSLogPrintStatic("NSPCCSIS", "ERR", "NSPCCSIS 실행 실패! 점포코드 입력 오류!");
                Thread.Sleep(1000); // 로그 남기기 위해 임시조치로 1초 쉬고 종료..
                return;
            }

            try
            {
                ConfigManager.Instance.Initialize();
                LogManager.Instance.Initialize(
                    ConfigManager.Instance.AppConfigRead("LOG_PATH"),
                    ConfigManager.Instance.AppConfigRead("SERVICE"),
                    UInt32.Parse(ConfigManager.Instance.AppConfigRead("LOG_FILE_SIZE"))
                    );

            }
            catch (Exception e)
            {
                LogManager.WriteEventLogEntry("NSPCCSIS", e.Message);       // 윈도우이벤트 로그
                LogManager.AMSLogPrintStatic("NSPCCSIS", "ERR", e.Message); // AMS 로그
                Thread.Sleep(1000); // 로그 남기기 위해 임시조치로 1초 쉬고 종료..
                return;
            }

            string storeType = ConfigManager.Instance.AppConfigRead("STORE_TYPE");
            string storeTypeNumber;
            if (storeType != "STORE" && storeType != "CENTER")
            {
                LogManager.Instance.Log("Configuration Error.(STORE_TYPE)");
                LogManager.Instance.AMSLogPrint("ERR", "NSPCCSIS 실행 실패! NSPCCSIS.exe.config 파일 내 STORE_TYPE 확인 필요!");
                Thread.Sleep(1000); // 로그 남기기 위해 임시조치로 1초 쉬고 종료..
                return;
            }
            else
            {
                storeTypeNumber = (storeType == "STORE") ? "3" : "5";
            }

            try
            {
                SQL.Instance.Initialize(
                        ConfigManager.Instance.AppConfigRead("DB_IP"),
                        ConfigManager.Instance.AppConfigRead("DB_NAME_PREFIX") + storeCode.Substring(2, 2) + storeTypeNumber,
                        ConfigManager.Instance.AppConfigRead("DB_USER_PREFIX") + storeCode,
                        ConfigManager.Instance.AppConfigRead("DB_PASSWORD_PREFIX") + storeCode,
                        Int32.Parse(ConfigManager.Instance.AppConfigRead("DB_CONNECTION"))
                        );

                SPointClient.Initialize(ConfigManager.Instance.AppConfigRead("SITE_URI"), "application/json"); // 접속URL & METATDATA 설정
            }
            catch (Exception e)
            {
                LogManager.Instance.Log(e.Message);
                LogManager.Instance.AMSLogPrint("ERR", "NSPCCSIS 실행 실패! MSSQL 초기화 오류!");
                Thread.Sleep(1000); // 로그 남기기 위해 임시조치로 1초 쉬고 종료..
                return;
            }

            LogManager.Instance.Log($"{ConfigManager.Instance.AppConfigRead("SERVICE")} PROGRAM START! ({storeCode} / {storeType})");
            RunAsync().GetAwaiter().GetResult(); // Main함수에서 await Test(httpsUrl) 사용못하므로, 이를 대신함
            LogManager.Instance.Log($"{ConfigManager.Instance.AppConfigRead("SERVICE")} PROGRAM END! ({storeCode} / {storeType})");

            Thread.Sleep(1000);
        }

        static async Task RunAsync()
        {
            // DB에서 UpdateList 조회(건수없으면 종료) -> { API 인증요청 -> TOKEN 수신 -> 서비스요청(TOKEN) -> 서비스응답 }
            // Config에 설정된 MAX_UPDATE_LIST_COUNT 만큼 (API 인증요청부터) 반복

            // 1.시코르 실적 업데이트를 위한 데이터 건수 취득
            List<ChicorPerformanceUpdateData> updateList = new List<ChicorPerformanceUpdateData>();
            using (DBApi db = new DBApi())
            {
                if (!db.SelectChicorShopInShopPointPerformance(updateList, date))
                {
                    LogManager.Instance.Log("Program Exit. Caused by SQL Error.");
                    LogManager.Instance.AMSLogPrint("ERR", "NSPCCSIS 비정상 종료!");
                    return;
                }
                if (updateList.Count == 0)
                {
                    LogManager.Instance.Log("시코르 샵인샵 실적 업데이트 건수 없음.");
                    LogManager.Instance.AMSLogPrint("INF", "시코르 샵인샵 실적 업데이트 건수 없음.");
                    return;
                }
            }

            // 2.시코르 실적 업데이트 데이터 건수만큼 반복하면서 REST API 처리
            while (updateList.Count != 0)
            {
                // 2-1. 인증 JSON 데이터 생성
                SPointAuthReq authReq = new SPointAuthReq(ConfigManager.Instance.AppConfigRead("CLIENT_ID"), ConfigManager.Instance.AppConfigRead("API_KEY"));
                LogManager.Instance.Log($"AuthReq Message \r\n{authReq.GetMsg()}");

                SPointAuthRsp authRsp = new SPointAuthRsp();

                var recvAuthData = await SPointClient.Send(HttpMethod.Post, ConfigManager.Instance.AppConfigRead("SITE_AUTH_PATH"), authReq);
                if (recvAuthData != null)
                {
                    authRsp = SPointClient.ConvertJsonString<SPointAuthRsp>(recvAuthData);
                    if (authRsp.responseCd != "API0000")
                    {
                        LogManager.Instance.Log($"Auth Response Code Error. ({authRsp.responseCd}) {authRsp.responseMsg}");
                        LogManager.Instance.AMSLogPrint("ERR", $"Auth Response Code Error. ({authRsp.responseCd}) {authRsp.responseMsg}");
                        return;
                    }
                    LogManager.Instance.Log($"AuthRsp Message \r\n{authRsp.GetMsg()}");
                }
                else
                {
                    LogManager.Instance.Log("authRsp is Null!!!");
                    LogManager.Instance.AMSLogPrint("ERR", "샵인샵 인증 응답 NULL!!");
                    return;
                }

                // 2-2. 실적 업데이트 데이터가 1000건 초과 일 경우를 대비해서 1000건씩 처리하도록 임시 list에 담아서 처리
                //      실적 업데이트 최대 건수는 config로 관리.
                List<ChicorPerformanceUpdateData> tmp = new List<ChicorPerformanceUpdateData>();
                int maxUpdateListCnt = Int32.Parse(ConfigManager.Instance.AppConfigRead("MAX_UPDATE_LIST_COUNT"));

                tmp.Clear();
                if (updateList.Count > maxUpdateListCnt)
                {
                    tmp = updateList.GetRange(0, maxUpdateListCnt);
                    updateList.RemoveRange(0, maxUpdateListCnt);
                }
                else
                {
                    tmp = updateList.GetRange(0, updateList.Count);
                    updateList.RemoveRange(0, updateList.Count);
                }

                SPointChicorPerformanceUpdateReq updateReq =
                    new SPointChicorPerformanceUpdateReq(ConfigManager.Instance.AppConfigRead("CLIENT_ID"),
                                                         ConfigManager.Instance.AppConfigRead("API_KEY"),
                                                         authRsp.tokenId,
                                                         GetTrcNo(),
                                                         "40041",
                                                         tmp.Count.ToString(),
                                                         tmp);

                LogManager.Instance.Log($"PerformanceUpdateReq Message \r\n {updateReq.GetMsg()}");

                var recvData = await SPointClient.Send(HttpMethod.Post, ConfigManager.Instance.AppConfigRead("SITE_SERVICE_PATH"), updateReq);
                if (recvData != null)
                {
                    try
                    {
                        SPointChicorPerformanceUpdateRsp result = SPointClient.ConvertJsonString<SPointChicorPerformanceUpdateRsp>(recvData);
                        // 응답이 정상이 아닐 경우 GetMsg() 함수에서 null 참조가 되어버려서 일단 이렇게 막았다.
                        // 근본적으로는 GetMsg()에서 null 처리 해주는게 맞다.
                        if (result.responseCd != "API0000")
                        {
                            LogManager.Instance.Log($"recvData is Not OK.\r\n responseCd : {result.responseCd}\r\n {result.responseMsg}");
                        }
                        else
                        {
                            LogManager.Instance.Log(result.GetMsg());
                        }
                    }
                    catch (Exception e)
                    {
                        LogManager.Instance.Log($"recvData is Error!!!\r\n{e.Message}");
                        LogManager.Instance.AMSLogPrint("ERR", "샵인샵 실적 업데이트 응답 오류!!");
                    }
                }
                else
                {
                    LogManager.Instance.Log("recvData is Null!!!");
                    LogManager.Instance.AMSLogPrint("ERR", "샵인샵 실적 업데이트 응답 NULL!!");
                }
            }
        }

        public static string GetTrcNo()
        {
            string trcStr = ConfigManager.Instance.AppConfigRead("TRANSACTION_DATE");
            int trcNo = Int32.Parse(ConfigManager.Instance.AppConfigRead("TRANSACTION_NO"));
            if (trcStr != DateTime.Now.ToString("yyyyMMdd"))
            {
                ConfigManager.Instance.AppConfigWrite("TRANSACTION_DATE", DateTime.Now.ToString("yyyyMMdd"));
                ConfigManager.Instance.AppConfigWrite("TRANSACTION_NO", "001");
            }
            // 20210728.YDK 같은 트랜잭션 번호로 보내면 포인트시스템에서 성공/실패 건수 계산처리가 중복되기 때문에, 점포 점/중앙코드를 넣어서 각각의 서버가 유니크하게 트랜잭션 번호를 갖게 수정.
            // YYYYMMDDHHMISS + [점코드2자리] + [점|중앙1자리 (3|5)] + 시퀀스번호 3자리
            trcStr = DateTime.Now.ToString("yyyyMMddHHmmss") + storeCode.Substring(2,2) + (ConfigManager.Instance.AppConfigRead("STORE_TYPE") == "STORE" ? "3" : "5") + trcNo++.ToString("D3");
            ConfigManager.Instance.AppConfigWrite("TRANSACTION_NO", trcNo.ToString("D3"));

            return trcStr;
        }
    }
}


