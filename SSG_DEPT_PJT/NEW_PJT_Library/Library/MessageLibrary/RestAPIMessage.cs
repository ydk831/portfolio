using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;

namespace MessageLibrary
{
    public class RestAPIMessage
    {
        public string URLPath { get; set; }
        public HttpMethod Method { get; set; }
        public object ReqMsg { get; set; } = null;
        public object RspMsg { get; set; } = null;
        public Encoding MsgEnc { get; set; } = Encoding.Default;
        public HttpStatusCode HttpCode { get; set; }

        public RestAPIMessage(string path, HttpMethod method)
        {
            URLPath = path;
            Method = method;
        }
    }


    public class ForceCancelReq
    {
        public int apiId { get; set; } // api 아이디 (다건 등록 구분을 위한 임의의 ID)
        public string cmpType { get; set; } // 회사구분 (101 = 백화점)
        public string serDate { get; set; } // 승인일 (YYYYMMDD)
        public string jumpoNo { get; set; } // 점포번호 (4자리)
        public string posNo { get; set; } // 포스번호 (4자리)
        public string tranNo { get; set; } // 거래번호 (4자리)
        public string etcField { get; set; } // 기타 (취소사유 입력)

        public ForceCancelReq() { }

        public ForceCancelReq(int apiId, string cmpType, string serDate, string jumpoNo, string posNo, string tranNo, string etcField)
        {
            this.apiId = apiId;
            this.cmpType = cmpType;
            this.serDate = serDate;
            this.jumpoNo = jumpoNo;
            this.posNo = posNo;
            this.tranNo = tranNo;
            this.etcField = etcField;
        }

        public virtual void PrintMsg()
        {
            Console.WriteLine(
                "{\r\n" +
                $"\tapiId:{this.apiId}\r\n" +
                $"\tcmpType:{this.cmpType}\r\n" +
                $"\tserDate:{this.serDate}\r\n" +
                $"\tjumpoNo:{this.jumpoNo}\r\n" +
                $"\tposNo:{this.posNo}\r\n" +
                $"\ttranNo:{this.tranNo}\r\n" +
                $"\tetcField:{this.etcField}\r\n" +
                "}\r\n"
            );
        }
        public string GetMsg()
        {
            return "{\r\n" +
                    $"\tapiId:{this.apiId}\r\n" +
                    $"\tcmpType:{this.cmpType}\r\n" +
                    $"\tserDate:{this.serDate}\r\n" +
                    $"\tjumpoNo:{this.jumpoNo}\r\n" +
                    $"\tposNo:{this.posNo}\r\n" +
                    $"\ttranNo:{this.tranNo}\r\n" +
                    $"\tetcField:{this.etcField}\r\n" +
                    "}\r\n";
        }
    }

    public class ForceCancelRsp
    {
        public int apiId { get; set; } // Req에서 전달 된 apiId
        public string apiResult { get; set; } // 결과코드
        public string apiMsg { get; set; } // 결과메세지
        public string cmpType { get; set; } // 회사구분 (Req에서 전달된 CmpType)
        public string serDate { get; set; } // 승인일 (Req에서 전달된 SerDate)
        public string jumpoNo { get; set; } // 점포번호 (Req에서 전달된 점포번호)
        public string posNo { get; set; } // 포스번호 (Req에서 전달된 포스번호)
        public string tranNo { get; set; } // 거래번호 (Req에서 전달된 거래번호)
        public string approvalNo { get; set; } // 승인번호 (점포번호,포스번호,거래번호로 도출 된 현금영수증 승인번호)
        public string etcField { get; set; } // 기타 (성공인 경우 Req에서 전달된 사유, 실패의 경우 실패사유)
        public string cancelYN { get; set; } //
        public void PrintMsg()
        {
            Console.WriteLine(
                "{\r\n" +
                $"\tapiId:{this.apiId}\r\n" +
                $"\tapiResult:{this.apiResult}\r\n" +
                $"\tapiMsg:{this.apiMsg}\r\n" +
                $"\tcmpType:{this.cmpType}\r\n" +
                $"\tserDate:{this.serDate}\r\n" +
                $"\tjumpoNo:{this.jumpoNo}\r\n" +
                $"\tposNo:{this.posNo}\r\n" +
                $"\ttranNo:{this.tranNo}\r\n" +
                $"\tapprovalNo:{this.approvalNo}\r\n" +
                $"\tetcField:{this.etcField}\r\n" +
                $"\tCancelYN:{this.cancelYN}\r\n" +
                "}\r\n"
                );
        }
        public string GetMsg()
        {
            return "{\r\n" +
                $"\tapiId:{this.apiId}\r\n" +
                $"\tapiResult:{this.apiResult}\r\n" +
                $"\tapiMsg:{this.apiMsg}\r\n" +
                $"\tcmpType:{this.cmpType}\r\n" +
                $"\tserDate:{this.serDate}\r\n" +
                $"\tjumpoNo:{this.jumpoNo}\r\n" +
                $"\tposNo:{this.posNo}\r\n" +
                $"\ttranNo:{this.tranNo}\r\n" +
                $"\tapprovalNo:{this.approvalNo}\r\n" +
                $"\tetcField:{this.etcField}\r\n" +
                $"\tCancelYN:{this.cancelYN}\r\n" +
                "}\r\n";
        }
    }

    public class AddCashReceiptFranchiseeReq
    {
        public int apiId { get; set; }
        public string cmpType { get; set; }
        public string groupType { get; set; }
        public string calGroup { get; set; }
        public string jumpoNo { get; set; }
        public string businessNo { get; set; }
        public string name { get; set; }
        public string openDate { get; set; }
        public string address { get; set; }

        public virtual void PrintMsg()
        {
            Console.WriteLine(
                "{\r\n" +
                $"\tapiId:{this.apiId}\r\n" +
                $"\tcmpType:{this.cmpType}\r\n" +
                $"\tgroupType:{this.groupType}\r\n" +
                $"\tcalGroup:{this.calGroup}\r\n" +
                $"\tjumpoNo:{this.jumpoNo}\r\n" +
                $"\tbusnessNo:{this.businessNo}\r\n" +
                $"\tname:{this.name}\r\n" +
                $"\topenDate:{this.openDate}\r\n" +
                $"\taddress:{this.address}\r\n" +
                "}\r\n"
            );
        }

    }
    public class AddCashReceiptFranchiseeRsp
    {
        public int apiId { get; set; }
        public string cmpType { get; set; }
        public string groupType { get; set; }
        public string calGroup { get; set; }
        public string jumpoNo { get; set; }
        public string businessNo { get; set; }
        public string name { get; set; }
        public string openDate { get; set; }
        public string address { get; set; }
        public string apiResult { get; set; }
        public string apiMsg { get; set; }
        public void PrintMsg()
        {
            Console.WriteLine(
                "{\r\n" +
                $"\tapiId:{this.apiId}\r\n" +
                $"\tcmpType:{this.cmpType}\r\n" +
                $"\tgroupType:{this.groupType}\r\n" +
                $"\tcalGroup:{this.calGroup}\r\n" +
                $"\tjumpoNo:{this.jumpoNo}\r\n" +
                $"\tbusnessNo:{this.businessNo}\r\n" +
                $"\tname:{this.name}\r\n" +
                $"\topenDate:{this.openDate}\r\n" +
                $"\taddress:{this.address}\r\n" +
                $"\tapiResult:{this.apiResult}\r\n" +
                $"\tapiMsg:{this.apiMsg}\r\n" +
                "}\r\n"
                );
        }
    }
}
