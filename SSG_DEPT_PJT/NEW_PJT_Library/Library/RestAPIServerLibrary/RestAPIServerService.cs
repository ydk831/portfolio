using System;
using System.Collections.Generic;
using System.Runtime.Serialization;
using System.ServiceModel;
using System.ServiceModel.Description;
using System.ServiceModel.Web;
using System.Text;

namespace RestAPIServerLibrary
{
    /// <summary>
    /// URL 별로 처리되는 함수가 구현 된 Class
    /// </summary>
    // 서비스 다중 처리 모드 설정
    [ServiceBehavior(InstanceContextMode = InstanceContextMode.PerCall, ConcurrencyMode = ConcurrencyMode.Multiple)]
    public class Service : IService
    {
        // /data/{value}의 형식으로 접속되면 호출되어 처리한다.
        public Response GetResponse(string value, string n1, int n2)
        {
            // Response 클래스 타입으로 리턴하면 자동으로 Json형식으로 변환한다.
            string s = OperationContext.Current.SessionId;
            return new Response() { Result = "get - " + value };
        }
        public Response GetResponse2(string value)
        {
            // Response 클래스 타입으로 리턴하면 자동으로 Json형식으로 변환한다.
            return new Response() { Result = "get2 - " + value };
        }
        public Somthing PostResponse(Somthing value)
        {
            // Response 클래스 타입으로 리턴하면 자동으로 Json형식으로 변환한다.            
            Console.WriteLine("들어는왔냐");
            Console.WriteLine(value.Say);
            return new Somthing() { Say = "Post" };
        }
    }
}
