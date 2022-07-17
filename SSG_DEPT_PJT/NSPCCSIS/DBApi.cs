using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Data.SqlClient;
using System.Reflection;

namespace NSPCCSIS
{
    public class DBApi : IDisposable
    {
        private bool isDispose = false;
        SqlConnection m_session;
        string m_query;


        public DBApi()
        {
            m_session = SQL.Instance.DequeuSqlConnection();

        }
        ~DBApi()
        {
            if (!isDispose)
            {
                Dispose();
            }
        }

        public void Dispose()
        {
            isDispose = true;

            SQL.Instance.EnqueuSqlConnection(m_session);
            m_session = null;
            GC.SuppressFinalize(this); // GC가 이 객체에 대해서 Finalize를 호출하지 않도록 명시적으로 지정
        }


        public bool SelectChicorShopInShopPointPerformance(List<ChicorPerformanceUpdateData> performanceDataList, string date = "")
        {
            if (m_session == null)
            {
                LogManager.Instance.Log("Ther is not have able DB Session!");
                LogManager.Instance.AMSLogPrint("ERR", "DB세션 없음! 확인 필요!");
                return false;
            }

            try
            {
                /// <summary> 
                /// SCMMPK_BACK 과 SCMMPK2_BACK에 중복 데이터가 존재 할 수 있고, 포인트 조회 시 같은 거래번호에서 다른 포인트번호가 존재 할 수 있기 때문에, 방지하기 위해 아래와 같이 쿼리 설계 함
                /// 1. SCMMPK_BACK 과 SCMMPK2_BACK 을 동일날짜와 동일 HOST_KIND로 조회하여 UNION ALL (UM_ALL)
                /// 2. SCMMPK_BACK 과 SCMMPK2_BACK 을 동일날짜와 동일 HOST_KIND로 조회 후 POS_NO,TRAN_NO로 GROUP 하여 INS_DATE(insert 시간)이 MAX은 ROW 추출 (UM_INS)
                ///    UM_INS 추출 이유는, 가장 마지막에 저장된 매핑키 기준 카드번호로만 실적 업데이트 처리하기 위함.
                ///    예를들어서, 자사 카드 결제 시, 매핑키에 저장된 포인트 카드번호가 자사카드가 되는데, 포스 OP상으로 다른 포인트 카드로 처리 할 수 있기 때문.
                /// 3. 전체 매핑키 데이터 (UM_ALL)에서 가장 최근에 저장 된 매핑키 데이터 (UM_INS)의 카드번호 추출 (A+B=M) // A는 UM_ALL B는 UM_INS
                /// 4. Tran_PointItem_Back 테이블의 샵인샵 거래 데이터 (GiftShopFlag='2')의 포인트 카드번호(M)를 join 하여 SELECT
                /// </summary>
                if (date == "")
                {
                    m_query = "SELECT P.POS_NO posNo, P.TRAN_NO tranNo, " +
                                " M.KEY_DATA cardNo, " +
                                " CASE P.TranType " +
                                " WHEN '20' THEN '01' " +
                                " WHEN '21' THEN '02' " +
                                " WHEN '30' THEN '03' " +
                                " WHEN '31' THEN '04' " +
                                " WHEN '40' THEN '05' " +
                                " WHEN '41' THEN '06' " +
                                " ELSE '07' " +
                                " END srvGB, " +
                                " CASE P.TranType " +
                                " WHEN '20' THEN '적립' " +
                                " WHEN '21' THEN '적립 취소' " +
                                " WHEN '30' THEN '사용' " +
                                " WHEN '31' THEN '사용 취소' " +
                                " WHEN '40' THEN '적립+사용' " +
                                " WHEN '41' THEN '적립+사용 취소' " +
                                " ELSE '기타' " +
                                " END srvGbDesc, " +
                                " P.SALEDATE srvDt " +
                                "FROM Tran_PointItem_Back P with(nolock) " +
                                "inner join " +
                                "( " +
                                "    select B.STORE_CODE, B.SALE_DATE, B.POS_NO, B.TRAN_NO, B.HOST_KIND, master.dbo.dec_char_sel(0,'CARD',A.KEY_DATA) KEY_DATA, B.INS_DATE from " +
                                "    ( " +
                                "        SELECT UM_ALL.STORE_CODE, UM_ALL.SALE_DATE, UM_ALL.POS_NO, UM_ALL.TRAN_NO, UM_ALL.HOST_KIND, UM_ALL.KEY_DATA, UM_ALL.INS_DATE " +
                                "        FROM " +
                                "        ( " +
                                "            SELECT STORE_CODE, SALE_DATE, POS_NO, TRAN_NO, HOST_KIND, KEY_DATA, INS_DATE " +
                                "            FROM SCMMPK_BACK M1 with(nolock) where M1.SALE_DATE= CONVERT(VARCHAR, GETDATE() - 1, 112) AND M1.HOST_KIND = 'SPOINT' " +
                                "            UNION ALL " +
                                "            SELECT STORE_CODE, SALE_DATE, POS_NO, TRAN_NO, HOST_KIND, KEY_DATA, INS_DATE " +
                                "            FROM SCMMPK2_BACK M2 with(nolock) where M2.SALE_DATE= CONVERT(VARCHAR, GETDATE() - 1, 112) AND M2.HOST_KIND = 'SPOINT' " +
                                "        ) UM_ALL " +
                                "	) A inner join " +
                                "    ( " +
                                "        SELECT UM_INS.STORE_CODE, UM_INS.SALE_DATE, UM_INS.POS_NO, UM_INS.TRAN_NO, UM_INS.HOST_KIND, MAX(UM_INS.INS_DATE) INS_DATE " +
                                "        FROM " +
                                "        ( " +
                                "            SELECT STORE_CODE, SALE_DATE, POS_NO, TRAN_NO, HOST_KIND, INS_DATE " +
                                "            FROM SCMMPK_BACK M1 with(nolock) where M1.SALE_DATE = CONVERT(VARCHAR, GETDATE() - 1, 112) AND M1.HOST_KIND = 'SPOINT' " +
                                "            UNION ALL " +
                                "            SELECT STORE_CODE, SALE_DATE, POS_NO, TRAN_NO, HOST_KIND, INS_DATE " +
                                "            FROM SCMMPK2_BACK M2 with(nolock) where M2.SALE_DATE = CONVERT(VARCHAR, GETDATE() - 1, 112) AND M2.HOST_KIND = 'SPOINT' " +
                                "        ) UM_INS GROUP BY STORE_CODE, SALE_DATE, POS_NO, TRAN_NO, HOST_KIND " +
                                "    ) B on " +
                                "            A.STORE_CODE = B.STORE_CODE " +
                                "        AND A.SALE_DATE = B.SALE_DATE " +
                                "        AND A.POS_NO = B.POS_NO " +
                                "        AND A.TRAN_NO = B.TRAN_NO " +
                                "        AND A.HOST_KIND = B.HOST_KIND " +
                                "        AND A.INS_DATE = B.INS_DATE " +
                                ") M " +
                                "ON P.STORE_NO = M.STORE_CODE " +
                                "    AND P.SALEDATE = M.SALE_DATE " +
                                "    AND P.POS_NO = M.POS_NO " +
                                "    AND P.TRAN_NO = M.TRAN_NO " +
                                "WHERE P.GiftShopFlag = '2'"
                                ;
                }
                else
                {
                    m_query = "SELECT P.POS_NO posNo, P.TRAN_NO tranNo, " +
                                " M.KEY_DATA cardNo, " +
                                " CASE P.TranType " +
                                " WHEN '20' THEN '01' " +
                                " WHEN '21' THEN '02' " +
                                " WHEN '30' THEN '03' " +
                                " WHEN '31' THEN '04' " +
                                " WHEN '40' THEN '05' " +
                                " WHEN '41' THEN '06' " +
                                " ELSE '07' " +
                                " END srvGB, " +
                                " CASE P.TranType " +
                                " WHEN '20' THEN '적립' " +
                                " WHEN '21' THEN '적립 취소' " +
                                " WHEN '30' THEN '사용' " +
                                " WHEN '31' THEN '사용 취소' " +
                                " WHEN '40' THEN '적립+사용' " +
                                " WHEN '41' THEN '적립+사용 취소' " +
                                " ELSE '기타' " +
                                " END srvGbDesc, " +
                                " P.SALEDATE srvDt " +
                                "FROM Tran_PointItem_Back P with(nolock) " +
                                "inner join " +
                                "( " +
                                "    select B.STORE_CODE, B.SALE_DATE, B.POS_NO, B.TRAN_NO, B.HOST_KIND, master.dbo.dec_char_sel(0,'CARD',A.KEY_DATA) KEY_DATA, B.INS_DATE from " +
                                "    ( " +
                                "        SELECT UM_ALL.STORE_CODE, UM_ALL.SALE_DATE, UM_ALL.POS_NO, UM_ALL.TRAN_NO, UM_ALL.HOST_KIND, UM_ALL.KEY_DATA, UM_ALL.INS_DATE " +
                                "        FROM " +
                                "        ( " +
                                "            SELECT STORE_CODE, SALE_DATE, POS_NO, TRAN_NO, HOST_KIND, KEY_DATA, INS_DATE " +
                                "            FROM SCMMPK_BACK M1 with(nolock) where M1.SALE_DATE='" + date + "' AND M1.HOST_KIND = 'SPOINT' " +
                                "            UNION ALL " +
                                "            SELECT STORE_CODE, SALE_DATE, POS_NO, TRAN_NO, HOST_KIND, KEY_DATA, INS_DATE " +
                                "            FROM SCMMPK2_BACK M2 with(nolock) where M2.SALE_DATE='" + date + "' AND M2.HOST_KIND = 'SPOINT' " +
                                "        ) UM_ALL " +
                                "	) A inner join " +
                                "    ( " +
                                "        SELECT UM_INS.STORE_CODE, UM_INS.SALE_DATE, UM_INS.POS_NO, UM_INS.TRAN_NO, UM_INS.HOST_KIND, MAX(UM_INS.INS_DATE) INS_DATE " +
                                "        FROM " +
                                "        ( " +
                                "            SELECT STORE_CODE, SALE_DATE, POS_NO, TRAN_NO, HOST_KIND, INS_DATE " +
                                "            FROM SCMMPK_BACK M1 with(nolock) where M1.SALE_DATE='" + date + "' AND M1.HOST_KIND = 'SPOINT' " +
                                "            UNION ALL " +
                                "            SELECT STORE_CODE, SALE_DATE, POS_NO, TRAN_NO, HOST_KIND, INS_DATE " +
                                "            FROM SCMMPK2_BACK M2 with(nolock) where M2.SALE_DATE='" + date + "' AND M2.HOST_KIND = 'SPOINT' " +
                                "        ) UM_INS GROUP BY STORE_CODE, SALE_DATE, POS_NO, TRAN_NO, HOST_KIND " +
                                "    ) B on " +
                                "            A.STORE_CODE = B.STORE_CODE " +
                                "        AND A.SALE_DATE = B.SALE_DATE " +
                                "        AND A.POS_NO = B.POS_NO " +
                                "        AND A.TRAN_NO = B.TRAN_NO " +
                                "        AND A.HOST_KIND = B.HOST_KIND " +
                                "        AND A.INS_DATE = B.INS_DATE " +
                                ") M " +
                                "ON P.STORE_NO = M.STORE_CODE " +
                                "    AND P.SALEDATE = M.SALE_DATE " +
                                "    AND P.POS_NO = M.POS_NO " +
                                "    AND P.TRAN_NO = M.TRAN_NO " +
                                "WHERE P.GiftShopFlag = '2'"
                                ;
                }

                using (SqlCommand sqlCommand = new SqlCommand(m_query, m_session))
                {
                    // 센텀점 같은 대형 점포는 쿼리 실행 시간이 좀 오래 걸려서 timeout 값을 더 준다. default는 30초임.)
                    if (Int32.Parse(ConfigManager.Instance.AppConfigRead("DB_TIMEOUT")) != 0)
                        sqlCommand.CommandTimeout = Int32.Parse(ConfigManager.Instance.AppConfigRead("DB_TIMEOUT"));

                    SqlDataReader sqlReader = sqlCommand.ExecuteReader();
                    if (sqlReader.HasRows)
                    {
                        while (sqlReader.Read())
                        {
                            performanceDataList.Add(new ChicorPerformanceUpdateData(
                                                            AES.Encrypt256(sqlReader["cardNo"].ToString().Trim(), ConfigManager.Instance.AppConfigRead("AES_KEY"), ConfigManager.Instance.AppConfigRead("AES_IV")),
                                                            sqlReader["srvGb"].ToString().Trim(),
                                                            sqlReader["srvGbDesc"].ToString().Trim(),
                                                            sqlReader["srvDt"].ToString().Trim()
                                                            )
                                        );

                            LogManager.Instance.Log($"{AES.Encrypt256(sqlReader["cardNo"].ToString().Trim(), ConfigManager.Instance.AppConfigRead("AES_KEY"), ConfigManager.Instance.AppConfigRead("AES_IV"))}," +
                                                    $"{sqlReader["srvGb"].ToString().Trim()}," +
                                                    $"{sqlReader["srvGbDesc"].ToString().Trim()}," +
                                                    $"{sqlReader["srvDt"].ToString().Trim()},"
                                                    );

                            //클래스 필드 파싱해서 자동으로 넣어주게끔 (나중에 사용해보자, 전문하고 DB컬럼하고 완전 매칭되야 한다.)
                            //Type tp = typeof(T);
                            //FieldInfo[] fields = tp.GetFields(BindingFlags.Instance |
                            //                                  BindingFlags.NonPublic |
                            //                                  BindingFlags.Public |
                            //                                  BindingFlags.NonPublic);
                            //
                            //foreach (var f in fields)
                            //{
                            //    string s = sqlReader[f.Name] as string;
                            //    Console.WriteLine($"필드 자동화 테스트 ({s})");
                            //    Console.WriteLine();
                            //}

                            //string encoding = AES.Encrypt256("9350130267740005", aesKey, aesIv);
                            //Console.WriteLine($"{encoding}");
                        }
                    }
                    else
                    {
                        // 건수없음. 밖에서 로그찍자.
                        // LogManager.Instance.Log("No Rows for Cancel CashReceipt Data!");
                    }
                }

                return true;
            }
            catch (Exception e)
            {
                LogManager.Instance.Log("SelectChicorShopInShopPointPerformance() Exception Error.\n" + e.Message);
                LogManager.Instance.Log(m_query);
                LogManager.Instance.AMSLogPrint("ERR", "시코르 샵인샵 포인트 실적 데이터 select 오류! " + e.Message);
                return false;
            }
        }
    } // DBApi END
}
