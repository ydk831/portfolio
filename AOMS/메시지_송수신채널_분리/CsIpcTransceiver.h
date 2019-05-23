#ifndef _CS_IPC_TRANSCEIVER_H
#define _CS_IPC_TRANSCEIVER_H

#include <pthread.h>

#include "CommIh/ConditionVariable.h"
#include "CommIh/IhPrimitive.h"
#include "CommIh/MutexLock.h"

#include "Connector.h"
#include "CsIhData.h"
#include "CsDef.h"

using namespace CsDef;

/*------------------------------------------------------------------------------
* definition
*--------------------------------------------------------------------------- */
#define IPC_THREAD_COUNT	1
#define CHECK_CONN_TIME		60 //sec
#define CS_MAX_BUF_SIZE     1024*8 // 2048 AOM_R100 2KB -> 4KB AOM_R440 4KB->8KB

typedef std::map<int, ManagedSocket*> SocketMap;

/*------------------------------------------------------------------------------
* class definition
*--------------------------------------------------------------------------- */
class CsIpcTransceiver : public Connector
{
public:

	CsIpcTransceiver();
	virtual ~CsIpcTransceiver();

	static CsIpcTransceiver* instance();

	void 		   load_peers();

	// Connector
	virtual void    error_proc(ManagedSocket* sock);
	virtual int     process(ManagedSocket* info); // recv

	int             sendToIws(Ih::IhPrimitive& ip); // send
	int             sendInitToIws(ManagedSocket* sock); // send-init
	int             SendCsVolumeNoti(__uint32_t volume);
	int             SendGenSimpleSmsReqToSmsis(string smstype, string aomcid, string mdn, string message, bool bCommonSms, string &strComment, string &strSmsisSendResult);
	int             SendIhKeepAliveReq();
	int             SendGenInfraPushReq(Ih::IhPrimitive &out);

	void            conn_check();
	//bool           reconnect();
	bool            start_conn_check();
	bool            stop_conn_check();

	// added by msryu
	NodeMapType                 get_node_map_type(int pbg_id);
	int                         GetIWSNodeNumber(IN int pbgID, IN NodeMapType nmt);
	inline static int           get_cis_pbgid(int cis_no);
	inline static int           get_sis_pbgid(int sis_no);
	inline static int           get_lgs_pbgid(int lgs_no);
	inline static int           get_ukis_pbgid(int ukis_no);
	inline static int           get_smsis_pbgid(int smsis_no);
	inline static int           get_cims_pbgid(int cims_no);
	inline static int           get_qiis_pbgid(int qiis_no);
	inline static int           get_cchs_pbgid(int cchs_no);
	inline static int           get_gis_pbgid(int gis_no);
	inline static iwsAddr_t*    GetIwsAddr(ManagedSocket *pSock);
	int                         closeLgsConn(std::string reason);
	int                         closeUkisConn(std::string reason);
	int                         closeCimsConn(std::string reason);
	int                         closeQiisConn(std::string reason);
	int                         closeCisConn(int nodeNumber, std::string reason);
	int                         closeSisConn(int nodeNumber, std::string reason);
	int                         closeSmsisConn(int nodeNumber, std::string reason);
	int                         closeCchsConn(std::string reason);
	int                         closeGisConn(int nodeNumber, std::string reason);
	int                         setTargetSmsisAddr(Ih::IhPrimitive &ip);
	int                         setTargetGisAddr(Ih::IhPrimitive &ip);

	ManagedSocket*              addNodeMapInfo(int lsgType, iwsAddr_t &rNodeInfo);  // R370. 2013.08.27
	bool                        delNodeMapInfo(int lsgType, int id);                // R370. 2013.08.27
	bool                        modNodeMapInfo(int lsgType, iwsAddr_t &rNodeInfo);  // R370. 2013.08.27

	void                        setCchsSupport(bool bOnOff);
	int                         getSendChnlFD(int pbg_id);
	int                         getRecvChnlFD(int pbg_id);
private:

	Ih::MutexLock  m_TransMutex;
	Ih::MutexLock  m_MutexLoadShareSmsis;
	Ih::MutexLock  m_MutexLoadShareGis;

	pthread_t      conn_thread;

	NodeMap 	   m_node_info[NMT_TOT];
	size_t		   m_index[NMT_TOT];

	Ih::ConditionVariable check_conn_cond;

	uint32_t        m_LoadShareSmsis;
	uint32_t        m_LoadShareGis;

	bool            m_bCchsSupport;
	//LengthFramer framer;
};

#endif // _CS_IPC_TRANSCEIVER_H
