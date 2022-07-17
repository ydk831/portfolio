using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Sockets;
using System.Text;
using System.Threading.Tasks;

namespace MessageLibrary
{
    public class HostMessage
    {
        private byte[] reqMsg;
        private byte[] resMsg;

        private int reqSendSize;
        private int resRecvSize;

        private TcpClient client;

        public HostMessage() { reqSendSize = 0; }
        public HostMessage(int bufsize) : this()
        {
            reqMsg = new byte[bufsize];
        }
        public HostMessage(TcpClient tc) : this()
        {
            client = tc;
        }
        public HostMessage(TcpClient tc, int bufsize) : this()
        {
            client = tc;
            reqMsg = new byte[bufsize];
        }
        public TcpClient Client
        {
            get { return client; }
            set { client = value; }
        }
        public byte[] ReqMsg
        {
            get { return reqMsg; }
            set { reqMsg = value; }
        }
        public int ReqSendSize
        {
            get { return reqSendSize; }
            set { reqSendSize = value; }
        }
        public byte[] ResMsg
        {
            get { return resMsg; }
            set { resMsg = value; }
        }
        public int ResRecvSize
        {
            get { return resRecvSize; }
            set { resRecvSize = value; }
        }
    }
}
