#include <time.h>
#include <sys/time.h>
#include <stdlib.h>
#include <errno.h>

#include "CommIh/IhUtil.h"
#include "CommIh/DebugLogger.h"

#include "CsIpcTransceiver.h"
#include "CsIhData.h"
#include "CsIhGroupServer.h"
#include "GenCsVolumeNoti.h"
#include "GenSimpleSmsReq.h"
#include "GenIhKeepAliveReq.h"
#include "SisSession.h"
#include "CsAlarmFault.h"
#include "LgsThread.h"
#include "CsLLT.h"
#include "CsCommPool.h"
#include "CsIhSession.h"
#include "RecSession.h"
#include "GenInfraPushReq.h"

int CsIpcTransceiver::SendIhKeepAliveReq()
{
	int         nTotal = 0;
	CsIhData    *pData = CsIhData::instance();
	AxMutex     *pMtx;

	for (int i = 0; i < NMT_TOT; ++i)
	{
		if (i == NMT_CCHS && m_bCchsSupport == false)
			continue;

		list<iwsAddr_t>::iterator   iter;
		list<iwsAddr_t>             *pNL = NULL;
		int                         ntype = 0;
		LsgType                     ltype;
		string                      strNode;

		switch (i) {
		case NMT_CIS: pNL = pData->GetCisNodeInfo();   pMtx = pData->GetCisNodeInfoLock();   ntype = NMT_CIS;   ltype = CIS_LSG;   strNode = "CIS";   break;
		case NMT_SIS: pNL = pData->GetSisNodeInfo();   pMtx = pData->GetSisNodeInfoLock();   ntype = NMT_SIS;   ltype = SIS_LSG;   strNode = "SIS";   break;
		case NMT_LGS: pNL = pData->GetLgsNodeInfo();   pMtx = pData->GetLgsNodeInfoLock();   ntype = NMT_LGS;   ltype = LGS_LSG;   strNode = "LGS";   break;
		case NMT_UKIS: pNL = pData->GetUkisNodeInfo();  pMtx = pData->GetUkisNodeInfoLock();  ntype = NMT_UKIS;  ltype = UKIS_LSG;  strNode = "UKIS";  break;
		case NMT_SMSIS: pNL = pData->GetSmsisNodeInfo(); pMtx = pData->GetSmsisNodeInfoLock(); ntype = NMT_SMSIS; ltype = SMSIS_LSG; strNode = "SMSIS"; break;
		case NMT_CIMS: pNL = pData->GetCimsNodeInfo();  pMtx = pData->GetCimsNodeInfoLock();  ntype = NMT_CIMS;  ltype = CIMS_LSG;  strNode = "CIMS";  break;
		case NMT_QIIS: pNL = pData->GetQiisNodeInfo();  pMtx = pData->GetQiisNodeInfoLock();  ntype = NMT_QIIS;  ltype = QIIS_LSG;  strNode = "QIIS";  break;
		case NMT_CCHS: pNL = pData->GetCchsNodeInfo();  pMtx = pData->GetCchsNodeInfoLock();  ntype = NMT_CCHS;  ltype = CCHS_LSG;  strNode = "CCHS";  break;
		case NMT_GIS: pNL = pData->GetGisNodeInfo();   pMtx = pData->GetGisNodeInfoLock();   ntype = NMT_GIS;   ltype = GIS_LSG;   strNode = "GIS";   break;
		default: continue;
		}

		AxLock lock(*pMtx);

		for (iter = pNL->begin(); iter != pNL->end(); ++iter)
		{
			if ((iter->port == 0) || (iter->ip == ""))
			{
				// 2 ä�� ���� ��� �� Send ä���� �������� �ʾ��� ��� IhKeepAlive �����ۿ� ���� ����ó���� �Ʒ����� Recv ä�� ��ȿ�� �Ǻ���� ��
				// Recv ä�� �� ���� �ÿ��� �������.
				if ((iter->id < 100) && (ntype == NMT_CIS) || (ntype == NMT_SIS))
				{
					DDLOG("[CsIpcTranceiver::SendIhKeepAlive()] No Configured Send Channel For" << (ntype == NMT_CIS ? "CIS" : "SIS"));
				}
				else if ((iter->id == 0) && (ntype == NMT_CCHS))
				{
					DDLOG("[CsIpcTranceiver::SendIhKeepAlive()] No Configured Send Channel For CCHS");
				}
				else
					continue;
			}

			ManagedSocket *pSock = m_node_info[ntype].get(iter->id);
			if ((pSock == NULL) || (pSock->is_closed()))
			{
				// CIS,SIS Send ä���� ��ȿ���� ���� ��
				if (ntype == NMT_CIS || ntype == NMT_SIS)
				{
					if (iter->id < 100)
					{
						// Recv ä�� ��ȿ�� �˻�
						ManagedSocket *pSock = m_node_info[ntype].get(iter->id + 100);
						if ((pSock == NULL) || (pSock->is_closed()))
						{
							continue;
						}
						else
						{
							//2ä�� �������� ä�� ������ ��ȿ���� ���� �� �ش� ��忡 KeepAlive ������
							//����� Send ä�� ��ȿ���� ���� �� Recv�� ��ȿ���� �˻��Ͽ� �����ϸ� �׳� Sendä���� pbgid ����
							//��¥�� SendToIws() ���� Send ä�� ������ �� Recv�� ������ ������
							//���⼭�� �ܼ��� Ÿ ä���� ��ȿ���� Ȯ���Ͽ� �޽����� ����  �� �� �ֵ��� �Ѵ�.
						}
					}
				}
				else if (ntype == NMT_CCHS) // CCHS Send ä���� ��ȿ���� ���� ��
				{
					if (iter->id == 0)
					{
						// Recv ä�� ��ȿ�� �˻�
						ManagedSocket *pSock = m_node_info[ntype].get(iter->id + 1);
						if ((pSock == NULL) || (pSock->is_closed()))
						{
							continue;
						}
						else
						{
							//��
						}
					}
				}
				else // 1ä�� ����..
					continue;
			}

			CsIhSession*    p_ih_session;
			SessionIdType   ih_session_id = INVALID_ID;
			int             dstPbgID;
			RecSession      *p_rec_sess;
			MemberIdType    ih_member_id;
			bool            bSendFail;
			string          strComment;

			switch (i) {
				// ������ Send ä�� ���н� Recv ä���� ��ȿ���� Ȯ���Ͽ��� ������, 
				// ���⼱ Send ä�ο� ���ؼ��� �����ϰ� Recv�� ó������ �ʵ��� �Ѵ�.
			case NMT_CIS:
				if (iter->id >= 100)
					continue;
				else
					dstPbgID = get_cis_pbgid(iter->id);
				break;
			case NMT_SIS:
				if (iter->id >= 100)
					continue;
				else
					dstPbgID = get_sis_pbgid(iter->id);
				break;
			case NMT_CCHS:
				if (iter->id == 1)
					continue;
				else
					dstPbgID = get_cchs_pbgid(iter->id);
				break;
			case NMT_LGS: dstPbgID = get_lgs_pbgid(iter->id);   break;
			case NMT_UKIS: dstPbgID = get_ukis_pbgid(iter->id);  break;
			case NMT_SMSIS: dstPbgID = get_smsis_pbgid(iter->id); break;
			case NMT_CIMS: dstPbgID = get_cims_pbgid(iter->id);  break;
			case NMT_QIIS: dstPbgID = get_qiis_pbgid(iter->id);  break;
			case NMT_GIS: dstPbgID = get_gis_pbgid(iter->id); break;
			default: dstPbgID = INVALID_ID; break;
			}

			if ((p_ih_session = (CsIhSession*)CsCommPool::instance()->obtainSession(CS_IH_SESSION)) == NULL)
			{
				CELOG("OBTAIN SESSION FAIL IH SEND_IH_KEEPALIVE_REQ..");
				continue;
			}

			p_ih_session->m_Mutex.lock();

			if ((p_rec_sess = (RecSession*)CommPool::instance()->obtainSession(CS_CS_SESSION)) == NULL)
			{
				p_ih_session->detachIhSessMember(INVALID_ID);
				p_ih_session->m_Mutex.unlock();
				continue;
			}

			p_rec_sess->init();
			ih_session_id = p_ih_session->getIhSessId();
			ih_member_id = p_ih_session->attachIhSessMember(p_rec_sess);

			p_rec_sess->setIhSession(p_ih_session);
			p_rec_sess->setIhSessionId(p_ih_session->getIhSessId());
			p_rec_sess->setMemberId(ih_member_id);

			GenIhKeepAliveReq out;

			out.setSrcLsgType(Ih::CS_LSG);
			out.setSrcPbgId(pData->GetMySessAddr()->m_PbgId);
			out.setSrcIhSessId(ih_session_id);
			out.setSrcIhMemberId(ih_member_id);
			out.setDstLsgType(ltype);
			out.setDstPbgId(dstPbgID);
			out.setDstIhSessId(INVALID_ID);
			out.setDstIhMemberId(INVALID_ID);

			out.m_CenterID = pData->GetCenterNumber();
			out.m_bPri = pData->IsPrimary();
			out.m_CommTime = time(0);

			bSendFail = false;
			strComment = "";

			memcpy(&p_rec_sess->m_DstAddr, &out.getHeader().m_DstSessAddr, sizeof(SessionAddress));

			if (sendToIws(out) != 1)
			{
				bSendFail = true;
				strComment = "SEND FAIL";
				CELOG("[CsIpcTransceiver::SendIhKeepAliveReq] SEND FAIL TO " << strNode << "#" << Ih::IhUtil::itos(iter->id + 1) << "(PBG_ID:" << Ih::IhUtil::itos(dstPbgID) << ")");
				p_rec_sess->release();
				p_ih_session->m_Mutex.unlock();
				continue;
			}

			if ((p_rec_sess->m_TimerID = p_rec_sess->startTimer(IH_KEEPALIVE_TIMER, pData->getTimerValue(IH_KEEPALIVE_TIMER))) == INVALID_ID)
			{
				bSendFail = true;
				strComment = "START-TIMER FAIL";
				CELOG("[CsIpcTransceiver::SendIhKeepAliveReq] START-TIMER FAIL TO " << strNode << "#" << Ih::IhUtil::itos(iter->id + 1) << "(PBG_ID:" << Ih::IhUtil::itos(dstPbgID) << ")");
				p_rec_sess->release();
				p_ih_session->m_Mutex.unlock();
				continue;
			}

			++nTotal;
			CDLOG("SEND GEN_IH_KEEPALIVE_REQ MESSAGE TO " << strNode << " #" << Ih::IhUtil::itos(iter->id + 1) << "(PGB_ID:" << Ih::IhUtil::itos(iter->id) << ")");
			CSLlt::PrtGenIhKeepAliveReq(&out, true, bSendFail, strComment);

			p_ih_session->m_Mutex.unlock();
		}
	}

	return nTotal;
}