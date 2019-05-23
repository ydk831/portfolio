#ifndef __SisTlsStack_H__
#define __SisTlsStack_H__ 

#include <queue>
#include <map>
#include <sys/epoll.h>
#include "cyassl/openssl/ssl.h"

#include "AxLib/AxWorker.h"
#include "AxLib/AxWorkerPool.h"
#include "AxLib/AxTransport.h"
#include "SisDef.h"
#include "SisTlsSession.h"

struct _FDIndex2SessionEl {
	AXINDEX_T       m_SocketID;
	SisTlsSession  *m_pSocket;
#ifdef _SUPP_IPV6_
	bool            m_IsIpv6;
#endif
	_FDIndex2SessionEl() { Clear(); }
	void Clear() {
		m_SocketID = 0;
		m_pSocket = NULL;
#ifdef _SUPP_IPV6_
		m_IsIpv6 = false;
#endif
	}
};

#define INADDR_ANY_IP_STR   ("0.0.0.0")
#define INADDR_ANY_IP6_STR  ("::")

//multimap<string, SisTlsSession *>  g_AidIndex2Socket;
//void                               DeleteAllConnSP();

using namespace SisDef;

class SisTlsStack {
public:
	static const __u_int    MAX_FDSET = SisDef::MAX_FDSET;
	static const __u_int    CPUs = 6;

	virtual void            NotifySocketStatus(AxTransport::enNotifySocket status, AxSocketAddr &rAddr, string &rAID) = 0;
	virtual bool            Recv(SisTlsSession *pSocket) = 0;
	bool                    Send(AxPacketBuf *pSendMsg, AxFdAferAct optAfertAct = AFTER_KEEP);
	bool                    Close(AxSocketAddr &addr);

	__u_int                 GetSessionCount() { return m_SessionCount; }
#ifdef _SUPP_IPV6_
	__u_int                 GetSessionCountIpv6() { return m_SessionCountIpv6; }
#endif
	__u_int                 GetMaxFD() { return m_MaxFD; }

	time_t                  GetLastUsedTime(AxSocketAddr &rAddr);
	void                    SetSessionKeepAliveExpiry(AxSocketAddr &rAddr, time_t iExpiry);
	time_t                  GetSessionKeepAliveExpiry(AxSocketAddr &rAddr);

	bool					IsExistSocket(int fd, string &rAID, AxSocketAddr *pAddr = NULL);
	bool                    UpdateSessionQoS(int fd, string &rAID, int iQoSLevel, __uint32_t iTPS);

protected:
	AXINDEX_T               GetNextSeqNo();
	SisTlsSession          *AllocSocket();
	void                    ReleaseSocket(SisTlsSession *pSocket);

	bool					IsExistSocket(AxSocketAddr &addr);
	SisTlsSession          *FindSocket(AxSocketAddr &addr);
	bool                    AddNewSocket(SisTlsSession *pSocket);

	void                    PollingNewSocketPipe(int pipeFD);
	void                    PollingCloseFdPipe(int pipeFD);
	void                    PollingSendMsgPipe(int pipeFD);

	bool                    PreSend(SisTlsSession *pSocket, AxPacketBuf *pSendMsg, AxFdAferAct optAfertAct);
	bool                    DeferredSend(SisTlsSession *pSocket);

	bool                    AddPollingSet(SisTlsSession *pSocket);
	bool                    RemovePollingSet(AxSocketAddr &addr);
	void                    EpollCtlMod(SisTlsSession *pSocket, uint32_t events);

	void                    RunTlsVerifySubject(SisTlsSession *pSocket);
	bool                    EraseTlsRegiEl(AxSocketAddr &rAddr);

	void                    InitFdSet(int TASKID);

public:
	SisTlsStack();
	virtual ~SisTlsStack();
	bool                    Initialize(const char *pCA_Cert, const char *pServer_Cert, const char *pServer_Key
		, AxSocketAddr &listenAddr, int iSNDBUFSize = 1024 * 1024, int iRCVBUFSize = 1024 * 1024);
	//, AxSocketAddr &listenAddr, int iSNDBUFSize = 1024*16, int iRCVBUFSize = 1024*16);
	void                    RunPollingTask(int *arg);
	void                    RunTlsListener();
#ifdef _SUPP_IPV6_
	void                    RunTlsListenerIpv6();
#endif
	void                    RunCloser(SisTlsSession *pSocket);
	void                    RunTryConnectChecker();
	void                    RunTlsRegiChecker();
	void                    RunSessionChecker();
	void                    RunSpConnChecker();
	void                    DeleteAllConnSP();
	void                    Clear();
	void                    SetName(string name) { m_name = name; }
	void                    SetEnable() { m_bEnable = true; }
	void                    SetDisable() { m_bEnable = false; }
	void                    SetPacketBufferSize(__u_int iPacketBufferSize) { m_PacketBufferSize = iPacketBufferSize; }

	/** R391. 둘다 안쓰임. */
	void                    SetSpConnectionLimit(__uint32_t iLimit) { m_SpConnLimit = iLimit; }
	__uint32_t              GetSpConnectionLimit() { return m_SpConnLimit; }

	/* R420. TLS Verify와 Close 처리 담당 Thread. */
	AxWorkerPool*           GetSessionWorkerPool() { return m_pWorkerPool; }

private:
	string                  m_name;
	AxWorker               *m_pPollingTask[CPUs];
	int                     m_efd[CPUs];
	struct epoll_event     *m_pEventLists[CPUs];

	AxWorker               *m_pRunTlsListener;
	AxSocketAddr            m_listenAddr;
#ifdef _SUPP_IPV6_
	AxWorker               *m_pRunTlsListenerIpv6;
	AxSocketAddr            m_listenTlsAddrIpv6;
#endif

	bool                    m_bEnable;
	__u_int                 m_PacketBufferSize;
	int                     m_SocketSNDBUFSize;
	int                     m_SocketRCVBUFSize;

	AxWorkerPool           *m_pWorkerPool;

	AxMutex                 m_NewSockLock[CPUs];
	int                     m_NewSocketPipe[CPUs][2]; // 0:Read, 1:Write
	AxMutex                 m_CloseFdLock[CPUs];
	int                     m_CloseFdPipe[CPUs][2];   // 0:Read, 1:Write
	AxMutex                 m_SendMsgLock[CPUs];
	int                     m_SendMsgPipe[CPUs][2];   // 0:Read, 1:Write

	_FDIndex2SessionEl     *m_FDIndex2Socket;       // FD to SisTlsSession
	int                     m_MaxFD;

	AxMutex                 m_socketCacheLock;
	list<SisTlsSession *>   m_socketCacheList;
	__u_int                 m_socketPoolCount;

	// SocketID Sequence Number
	AxMutex                 m_lockSeqNo;
	AXINDEX_T               m_iSeqNo;

	// ##### SessionCount Checker #####
	AxWorker               *m_pSessChecker;
	__u_int                 m_SessionCount;
#ifdef _SUPP_IPV6_
	__u_int                 m_SessionCountIpv6;
#endif

	__uint32_t              m_SpConnLimit;

	// ##### TLS #####
	SSL_CTX                *m_pServerSslCtx;

	// ##### TLS Regi Check #####
	AxMutex                 m_LockTlsRegi;
	list<AxTransport::_AddrTimeEl> m_TlsRegiCheckList;
	AxWorker               *m_pTlsRegiChecker;

	//R420
	AxMutex                 m_Aid2SockLock;
	multimap<string, SisTlsSession *>  m_AidIndex2Socket;
	AxWorker               *m_pSpConnChecker;
};

#endif
