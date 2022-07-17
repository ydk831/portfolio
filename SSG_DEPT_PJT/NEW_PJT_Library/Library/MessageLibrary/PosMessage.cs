using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Sockets;
using System.Text;
using System.Threading.Tasks;

namespace MessageLibrary
{
    public class PosMessage
    {
        private byte[] reqMsg;
        private byte[] resMsg;

        private int reqRecvSize;
        private int resSendSize;

        private TcpClient client;

        public PosMessage() { reqRecvSize = 0; }
        public PosMessage(int bufsize) : this()
        {
            reqMsg = new byte[bufsize];
        }
        public PosMessage(TcpClient tc) : this()
        {
            client = tc;
        }
        public PosMessage(TcpClient tc, int bufsize) : this()
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
        public int ReqRecvSize
        {
            get { return reqRecvSize; }
            set { reqRecvSize = value; }
        }
        public byte[] ResMsg
        {
            get { return resMsg; }
            set { resMsg = value; }
        }
        public int ResSendSize
        {
            get { return resSendSize; }
            set { resSendSize = value; }
        }
    }
}
