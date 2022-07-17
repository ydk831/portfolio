using System;
using System.Collections.Generic;
using System.Runtime.Serialization;
using System.ServiceModel;
using System.ServiceModel.Description;
using System.ServiceModel.Web;
using System.Text;

namespace RestAPIServerLibrary
{
    [DataContract]
    public class Response
    {
        [DataMember(Name = "Say")]
        public string Result { get; set; }
    }
    [DataContract]
    public class Somthing
    {
        [DataMember(Name = "Say")]
        public string Say { get; set; }
    }
}
