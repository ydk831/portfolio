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
    /// URL과 Service 함수를 매칭하는 Interface
    /// </summary>
    //[ServiceContract(SessionMode = SessionMode.Required)]
    [ServiceContract]
    public interface IService
    {
        [OperationContract]
        // 메소드 지정과 리턴 포멧 지정(리턴 포멧은 RestAPI를 만들 것이기 때문에 Json으로 한다.)
        // UriTemplate에 보간법으로 중괄호로 값을 지정할 수 있다.        
        [WebGet(UriTemplate = "/data/{value}?n1={n1}&n2={n2}", ResponseFormat = WebMessageFormat.Json, BodyStyle = WebMessageBodyStyle.Wrapped)]
        //[return: MessageParameter(Name = "value")] // BodyStyle이 WrappedResponse 또는 Wrapped (이건 Req,Rsp 둘다 선택하는거임)일 경우 MessageParameter로 응답을 감쌈
        // 리턴을 String으로 해도 상관없는데.. 여기서는 클래스 형식한다.
        // 클래스 형식으로 하면 자동으로 Newtonsoft.Json 라이브러리를 통해 자동으로 Json 형식으로 변환한다.
        Response GetResponse(string value, string n1, int n2);

        [OperationContract]
        [WebInvoke(Method = "GET", UriTemplate = "/data2/{value}", ResponseFormat = WebMessageFormat.Json, BodyStyle = WebMessageBodyStyle.Wrapped)]
        Response GetResponse2(string value);

        [OperationContract]
        [WebInvoke(Method = "POST", UriTemplate = "/Post", RequestFormat = WebMessageFormat.Json, ResponseFormat = WebMessageFormat.Json, BodyStyle = WebMessageBodyStyle.Bare)]
        [return: MessageParameter(Name = "test")] // BodyStyle이 WrappedResponse 또는 Wrapped (이건 Req,Rsp 둘다 선택하는거임)일 경우 MessageParameter로 응답을 감쌈
        // 리턴을 String으로 해도 상관없는데.. 여기서는 클래스 형식한다.
        // 클래스 형식으로 하면 자동으로 Newtonsoft.Json 라이브러리를 통해 자동으로 Json 형식으로 변환한다.
        Somthing PostResponse(Somthing value);
    }
}
