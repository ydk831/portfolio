#include "IpcQ.h"
#include "Hmsg.h"
#include "HpushMsg.h"
#include "HnotiMsg.h"
#include "Miscellaneous/TimeUtil.h"
#include "CommIh/IhUtil.h"

using namespace std;

extern int errno;

IpcQ*  IpcQ::m_pInstance = NULL;

void IpcQ::Init()
{
	if (m_pWorkerIpcQueProcessing) {
		delete m_pWorkerIpcQueProcessing;
		m_pWorkerIpcQueProcessing = NULL;
	}

	if (m_pWorkerJobQueProcessing) {
		delete m_pWorkerJobQueProcessing;
		m_pWorkerJobQueProcessing = NULL;
	}

	if (m_pWorkerHpushProcessing) {
		delete m_pWorkerHpushProcessing;
		m_pWorkerHpushProcessing = NULL;
	}

	{
		AxLock lock(m_JobQueLock);
		m_JobQue.clear();
	}

	{
		AxLock lock(m_HpushLock);
		m_HpushQue.clear();
	}

	m_bSender = true;
}

void IpcQ::RunJobQueProcessing()
{
	int ret = 0;

	while (true) {

		m_JobQueLock.Lock();
		if (m_JobQue.empty()) {
			m_JobQueLock.Unlock();
			sleep(1);
			continue;
		}

		Ih::BAOS *tmp = NULL;
		tmp = m_JobQue.front();
		if (tmp == NULL) {
			EELOG("[IpcQ::RunJobQueProcessing] JobQue Message is NULL. Why..?");
			m_JobQue.pop_front();
			m_JobQueLock.Unlock();
			continue;
		}
		else {
			m_JobQue.pop_front();
			m_JobQueLock.Unlock();
		}

		uint32_t tmp_type = 0;
		memcpy((void *)&tmp_type, tmp->getStream(), sizeof(uint32_t));

		if (tmp_type <= 0) {
			EELOG("[IpcQ::RunJobQueProcessing] JobQue Message is Wired. Message Type (" << tmp_type << ")");
			continue;
		}
		else {
			IILOG(YELLOW("[IpcQ::RunJobQueProcessing] JobQue Message Processing. Message Type (" << tmp_type << ")"));
		}

		if (true) {
			if ((ret = MsgQSnd(tmp, tmp_type)) == -1) {
				EELOG("[IpcQ::RunJobQueProcessing] Fail MsgQSnd()");
			}
		}
	}
}

void IpcQ::RunIpcQueProcessing()
{
	int ret = 0;

	while (true) {

		Ih::BAIS *Msg = new Ih::BAIS();
		long tmp_type = 0;

		if ((ret = MsgQRcv(*Msg, tmp_type)) == -1) {
			delete Msg;
			sleep(1);
			continue;
		}

		if (Msg->getStreamSize() == 0) {
			EELOG("[IpcQ::RunIpcQueProcessing] Ipc Queue Rev Success. But Data is NULL");
			delete Msg;
			continue;
		}

		if (tmp_type == 0) {
			EELOG("[IpcQ::RunIpcQueProcessing] Ipc Queue Rev Message Type is 0. So Drop This Message.");
			delete Msg;
			continue;
		}

		switch (tmp_type) {
		case IPC_HPUSH:
		case IPC_HPDD:
			IpcRecv4Hpush(*Msg);
			break;
		case IPC_HNOTI:
		case IPC_HTOPIC:
			IpcRecv4Hnoti(*Msg);
			break;
		default:
			EELOG("[IpcQ::RunIpcQueProcessing] tmp_type is Wired. (" << tmp_type << ")");
			break;
		}

		delete Msg;
	}
}

void IpcQ::RunHpushProcessing()
{
	int tps = 0;
	int mesc;
	struct timeval tm1, tm2, tmd;
	uint32_t nUpdateCnt = 0;
	int ret = 0;

	gettimeofday(&tm1, NULL);
	while (true) {
		gettimeofday(&tm2, NULL);
		TimeUtil::TimevalSub(&tm1, &tm2, &tmd);
		if (tmd.tv_sec > 0) {
			tps = 0;
			memcpy(&tm1, &tm2, sizeof(struct timeval));
		}

		Hmsg *tmp = NULL;

		m_HpushLock.Lock();
		if (m_HpushQue.empty()) {
			goto nextQueProc;
		}

		tmp = m_HpushQue.front();
		if (tmp == NULL) {
			EELOG("[IpcQ::RunHpushProcessing] Message is NULL. Why..?");
			m_HpushQue.pop_front();
			goto nextQueProc;
		}
		else {
			m_HpushQue.pop_front();
			m_HpushLock.Unlock();
		}

		if (true) {
			IpcQInfoDbApi rDB(m_UserID.c_str());

			switch (tmp->m_MsgType)
			{
			case IPC_HPUSH:
			case IPC_HPDD:
				ret = rDB.InsertHpushMsg(tmp, m_iCsNo);
				break;
			case IPC_HNOTI:
			case IPC_HTOPIC:
				ret = rDB.InsertHnotiMsg(tmp, m_iCsNo);
				break;
			default:
				EELOG("[IpcQ::RunHpushProcessing] NOT SUPPORT HISTORY PUSH MESSAGE TYPE (" << tmp->m_MsgType << ")");
				break;
			}

			delete tmp;

			if (ret != RET_DB_SUCCESS)
				EELOG("[IpcQ::RunHpushProcessing] DB FAIL");
			else
				nUpdateCnt++;
		}
		tps++;

		if (tps >= m_ProcCount) {
		nextQueProc:
			m_HpushLock.Unlock();
			if (nUpdateCnt)
				IILOG("[IpcQ::RunHpushProcessing] Insert Count:" << nUpdateCnt);
			nUpdateCnt = 0;

			mesc = 1000000 - tmd.tv_usec;
			uSleep(mesc);

			tps = 0;
			gettimeofday(&tm1, NULL);
		}
	}
}