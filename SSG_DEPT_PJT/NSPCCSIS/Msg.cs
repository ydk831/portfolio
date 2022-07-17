using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace NSPCCSIS
{
    public class SPointAuthReq
    {
        public string clientId { get; set; } // 클라이언트ID
        public string apiKey { get; set; } // API KEY
        
        public SPointAuthReq() { }

        public SPointAuthReq(string clientId, string apiKey)
        {
            this.clientId = clientId;
            this.apiKey = apiKey;            
        }

        public virtual void PrintMsg()
        {
            Console.WriteLine(
                "{\r\n" +
                $"\tclientId:{this.clientId}\r\n" +
                $"\tapiKey:{this.apiKey}\r\n" +                
                "}\r\n"
            );
        }
        public string GetMsg()
        {
            return "{\r\n" +
                   $"\tclientId:{this.clientId}\r\n" +
                   $"\tapiKey:{this.apiKey}\r\n" +
                   "}\r\n";
        }
    }

    public class SPointAuthRsp
    {
        public string clientId { get; set; } // 클라이언트ID
        public string apiKey { get; set; } // API KEY
        public string responseCd { get; set; } // 응답코드 (성공:API0000, 그 외 오류)
        public string responseMsg { get; set; } // 응답메세지 (성공 :"정상", 그 외 오류메세지)
        public string tokenId { get; set; } // 토큰 값 (서비스 요청시 사용)
        public string avlbDt { get; set; } // 유효일자 (YYYYMMDD)
        public string avlbTs { get; set; } // 유효시간 (HH24MISS)
        
        public SPointAuthRsp() { }
        public SPointAuthRsp(string clientId, string apiKey, string responseCd, string responseMsg, string tokenId, string avlbdt, string avlbTs)
        {
            this.clientId = clientId;
            this.apiKey = apiKey;
            this.responseCd = responseCd;
            this.responseMsg = responseMsg;
            this.tokenId = tokenId;
            this.avlbDt = avlbdt;
            this.avlbTs = avlbTs;
        }

        public void PrintMsg()
        {
            Console.WriteLine(
                "{\r\n" +
                $"\tclientId:{this.clientId}\r\n" +
                $"\tapiKey:{this.apiKey}\r\n" +
                $"\tresponseCd:{this.responseCd}\r\n" +
                $"\tresponseMsg:{this.responseMsg}\r\n" +
                $"\ttokenId:{this.tokenId}\r\n" +
                $"\tavlbDt:{this.avlbDt}\r\n" +
                $"\tavlbTs:{this.avlbTs}\r\n" +                
                "}\r\n"
                );
        }
        public string GetMsg()
        {
            return "{\r\n" +
                $"\tclientId:{this.clientId}\r\n" +
                $"\tapiKey:{this.apiKey}\r\n" +
                $"\tresponseCd:{this.responseCd}\r\n" +
                $"\tresponseMsg:{this.responseMsg}\r\n" +
                $"\ttokenId:{this.tokenId}\r\n" +
                $"\tavlbDt:{this.avlbDt}\r\n" +
                $"\tavlbTs:{this.avlbTs}\r\n" +
                "}\r\n";
        }
    }

    public class SPointChicorPerformanceUpdateReq
    {
        public string clientId { get; set; } // 클라이언트 ID
        public string apiKey { get; set; } // API KEY
        public string tokenId { get; set; } // AUTH에서 받은 토큰 값
        public string reqTrcNo { get; set; } // 거래 고유번호 (유니크 한 키, YYYYMMDDHH24MISS+000000)
        public string siteCd { get; set; } // 사이트 코드 (40041 : 시코르)
        public string rowCnt { get; set; } // 변경 건수 (1회 호출 시 1천건 이내)

        public List<ChicorPerformanceUpdateData> inList;

        public SPointChicorPerformanceUpdateReq() { }
        public SPointChicorPerformanceUpdateReq(string clientId, string apiKey, string tokenId, string reqTrcNo, string siteCd, string rowCnt, List<ChicorPerformanceUpdateData> updateList)
        {
            this.clientId = clientId;
            this.apiKey = apiKey;
            this.tokenId = tokenId;
            this.reqTrcNo = reqTrcNo;
            this.siteCd = siteCd;
            this.rowCnt = rowCnt;
            this.inList = updateList.ToList();
        }

        public virtual void PrintMsg()
        {
            Console.WriteLine(
                "{\r\n" +
                $"\tclientId:{this.clientId}\r\n" +
                $"\tapiKey:{this.apiKey}\r\n" +
                $"\ttokenId:{this.tokenId}\r\n" +
                $"\treqTrcNo:{this.reqTrcNo}\r\n" +
                $"\tsiteCd:{this.siteCd}\r\n" +
                $"\trowCnt:{this.rowCnt}\r\n" +                                
                "}\r\n"
            );
            foreach (var data in inList)
            {
                data.PrintMsg();
            }
        }
        public string GetMsg()
        {
            string msg = "{\r\n" +
                $"\tclientId:{this.clientId}\r\n" +
                $"\tapiKey:{this.apiKey}\r\n" +
                $"\ttokenId:{this.tokenId}\r\n" +
                $"\treqTrcNo:{this.reqTrcNo}\r\n" +
                $"\tsiteCd:{this.siteCd}\r\n" +
                $"\trowCnt:{this.rowCnt}\r\n" +
                "}\r\n";
            foreach (var data in inList)
            {
                msg += data.GetMsg();
            }
            return msg;
        }
    }

    public class ChicorPerformanceUpdateData
    {
        public string cardNo { get; set; } // 신세계 포인트 카드 번호
        public string srvGb { get; set; } // 휴면 실적 구분 (01:적립, 02:적립취소, 03:사용, 04:사용취소, 05:추후적립, 06:기타)
        public string srvGbDesc { get; set; } // 휴면 실적 상세명
        public string srvDt { get; set; } // 실적 보유 일자 (YYYYMMDD)

        public ChicorPerformanceUpdateData() { }
        public ChicorPerformanceUpdateData(string cardNo, string srvGb, string srvGbDesc, string srvDt)
        {
            this.cardNo = cardNo;
            this.srvGb = srvGb;
            this.srvGbDesc = srvGbDesc;
            this.srvDt = srvDt;
        }

        public virtual void PrintMsg()
        {
            Console.WriteLine(
                "{\r\n" +
                $"\tcardNo:{this.cardNo}\r\n" +
                $"\tsrvGb:{this.srvGb}\r\n" +
                $"\tsrvGbDesc:{this.srvGbDesc}\r\n" +
                $"\tsrvDt:{this.srvDt}\r\n" +                
                "}\r\n"
            );
        }
        public string GetMsg()
        {
            return "{\r\n" +
                $"\tcardNo:{this.cardNo}\r\n" +
                $"\tsrvGb:{this.srvGb}\r\n" +
                $"\tsrvGbDesc:{this.srvGbDesc}\r\n" +
                $"\tsrvDt:{this.srvDt}\r\n" +
                "}\r\n";
        }
    }

    public class SPointChicorPerformanceUpdateRsp
    {
        public string clientId { get; set; } // 클라이언트ID
        public string apiKey { get; set; } // API KEY
        public string responseCd { get; set; } // 응답코드 (성공:API0000, 그 외 오류)
        public string responseMsg { get; set; } // 응답메세지 (성공 :"정상", 그 외 오류메세지)
        public string resTrcNo { get; set; } // 처리번호 (YYYYMMDD+SEQ(10))
        public string resRowCnt { get; set; } // 처리 건수
        public string failRowCnt { get; set; } // 실패 건수
        public List<ChicorPerformanceResultData> failList;

        public SPointChicorPerformanceUpdateRsp() { }
        public SPointChicorPerformanceUpdateRsp(string clientId, string apiKey, string responseCd, string responseMsg, string resTrcNo, string resRowCnt, string failRowCnt, List<ChicorPerformanceResultData> failList)
        {
            this.clientId = clientId;
            this.apiKey = apiKey;
            this.responseCd = responseCd;
            this.responseMsg = responseMsg;
            this.resTrcNo = resTrcNo;
            this.resRowCnt = resRowCnt;
            this.failList = failList.ToList();
        }
        
        public void PrintMsg()
        {
            Console.WriteLine(
                "{\r\n" +
                $"\tclientId:{this.clientId}\r\n" +
                $"\tapiKey:{this.apiKey}\r\n" +
                $"\tresponseCd:{this.responseCd}\r\n" +
                $"\tresponseMsg:{this.responseMsg}\r\n" +
                $"\tresTrcNo:{this.resTrcNo}\r\n" +
                $"\tresRowCnt:{this.resRowCnt}\r\n" +
                $"\tfailRowCnt:{this.failRowCnt}\r\n" +
                "}\r\n"
                );
            foreach (var data in failList)
            {
                data.PrintMsg();
            }
        }
        public string GetMsg()
        {
            string msg = "{\r\n" +
                $"\tclientId:{this.clientId}\r\n" +
                $"\tapiKey:{this.apiKey}\r\n" +
                $"\tresponseCd:{this.responseCd}\r\n" +
                $"\tresponseMsg:{this.responseMsg}\r\n" +
                $"\tresTrcNo:{this.resTrcNo}\r\n" +
                $"\tresRowCnt:{this.resRowCnt}\r\n" +
                $"\tfailRowCnt:{this.failRowCnt}\r\n" +
                "}\r\n";
            msg += "[\r\n";
            if (this.failList != null)
            {
                foreach (var data in failList)
                {
                    msg += data.GetMsg();
                }
            }
            msg += "]";
            return msg;
        }
    }

    public class ChicorPerformanceResultData
    {
        public string cardNo { get; set; } // 신세계포인트 카드번호
        public string srvGb { get; set; } // 휴면 실적 구분 (01:적립, 02:적립취소, 03:사용, 04:사용취소, 05:추후적립, 06:기타)

        public ChicorPerformanceResultData() { }
        public ChicorPerformanceResultData(string cardNo, string srvGb)
        {
            this.cardNo = cardNo;
            this.srvGb = srvGb;
        }
        public void PrintMsg()
        {
            Console.WriteLine(
                "{\r\n" +
                $"\tcardNo:{this.cardNo}\r\n" +
                $"\tsrvGb:{this.srvGb}\r\n" +
                "}\r\n"
                );
        }
        public string GetMsg()
        {
            return "{\r\n" +
                $"\tcardNo:{this.cardNo}\r\n" +
                $"\tsrvGb:{this.srvGb}\r\n" +
                "}\r\n";
        }
    }
}
