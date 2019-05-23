#ifndef __AOM_IPC_QUEUE_H__
#define __AOM_IPC_QUEUE_H__

#include "CommIh/IhDef.h"   
#include "CommIh/BaseTime.h"
#include "CommIh/BAIS.h"
#include "CommIh/BAOS.h"
#include "AxLib/AxLib.h"
#include "CommIh/DebugLogger.h"
#include "CommIh/IhData.h"
#include "CommIh/IhUtil.h"
#include "IpcQInfoDbApi.h"

#include <sys/types.h>
#include <sys/ipc.h>
#include <sys/msg.h>
#include <stdlib.h>
#include <unistd.h>
#include <signal.h>


class Hmsg;
class HpushMsg;
class HnotiMsg;

// ******************************************************************************
// * Ipc Queue Message Structure Define
// ******************************************************************************
#define MAX_DATA_SIZE 4096

typedef deque<Ih::BAOS*>     IpcJobQue_t;
typedef deque<Hmsg*>     HpushJobQue_t;
typedef struct {
	long type;
	int  len; // data의 실제 길이
	char data[MAX_DATA_SIZE]; // 어떤 데이터든 간에 binary type으로 저장하자

	void Clear() {
		type = 0;
		len = 0;
		memset(data, NULL, sizeof(data));
	}
} Qdata;

// ******************************************************************************
// * Service Message Type Define
// * When You Adding Message Type, MUST ADD Definitiona Sequencial
// ******************************************************************************
#define IPC_HPUSH 1
#define IPC_HNOTI 2
#define IPC_HPDD 3
#define IPC_HTOPIC 4

class IpcQ
{
public:
	IpcQ();
	IpcQ(const IpcQ &src);
	~IpcQ();

	static IpcQ*    Instance(); // single ton으로 하자
	void            Init();     // 초기화
	int             Initialize(char* pUdbId, int iCsNo, bool bSender = true, int interval = 5, int count = 100);
	void            Stop();

	void            SetQKey(int key) { m_Qkey = key; }
	int             GetQKey() { return m_Qkey; }

	int             GetQid() { return m_Qid; }

	int             MsgQSnd(Ih::BAOS *Msg, long type); // msg send
	int             MsgQRcv(Ih::BAIS &Msg, long &type); // msg recv
	int             MsgQCtl(); // queue status

	void            RunIpcQueProcessing();
	void            RunJobQueProcessing();
	void            RunHpushProcessing();


	int             IpcSend4Hpush(HpushMsg &pMsg);
	int             IpcRecv4Hpush(Ih::BAIS &Msg);
	int             IpcSend4Hnoti(HnotiMsg &pMsg);
	int             IpcRecv4Hnoti(Ih::BAIS &Msg);

	void            uSleep(int usec);

private:
	static IpcQ     *m_pInstance;
	int             m_Qkey; // ipc queie key.. use create and queue get
	int             m_Qid; // ipc queie id.. use snd/rcv

						   //Qdata           m_Qdata; // real data
	AxMutex         m_JobQueLock;
	AxMutex         m_HpushLock;

	AxWorker        *m_pWorkerIpcQueProcessing;
	AxWorker        *m_pWorkerJobQueProcessing;
	AxWorker        *m_pWorkerHpushProcessing;

	IpcJobQue_t     m_JobQue;
	HpushJobQue_t   m_HpushQue;

	int             m_ProcCount;
	int             m_PollingInterval;

	bool            m_bSender;
	string          m_UserID;

	int             m_iCsNo;

};


#endif // __AOM_IPC_QUEUE_H__
