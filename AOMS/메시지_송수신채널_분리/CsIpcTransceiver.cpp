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
				// 2 채널 지원 노드 중 Send 채널을 설정하지 않았을 경우 IhKeepAlive 미전송에 대한 예외처리로 아래에서 Recv 채널 유효성 판별토록 함
				// Recv 채널 미 설정 시에는 상관없다.
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
				// CIS,SIS Send 채널이 유효하지 않을 때
				if (ntype == NMT_CIS || ntype == NMT_SIS)
				{
					if (iter->id < 100)
					{
						// Recv 채널 유효성 검사
						ManagedSocket *pSock = m_node_info[ntype].get(iter->id + 100);
						if ((pSock == NULL) || (pSock->is_closed()))
						{
							continue;
						}
						else
						{
							//2채널 이전에는 채널 소켓이 유효하지 않을 때 해당 노드에 KeepAlive 미전송
							//현재는 Send 채널 유효하지 않을 때 Recv의 유효성을 검사하여 존재하면 그냥 Send채널의 pbgid 설정
							//어짜피 SendToIws() 에서 Send 채널 끊겼을 때 Recv로 보내기 때문에
							//여기서는 단순히 타 채널의 유효성만 확인하여 메시지가 생성  될 수 있도록 한다.
						}
					}
				}
				else if (ntype == NMT_CCHS) // CCHS Send 채널이 유효하지 않을 때
				{
					if (iter->id == 0)
					{
						// Recv 채널 유효성 검사
						ManagedSocket *pSock = m_node_info[ntype].get(iter->id + 1);
						if ((pSock == NULL) || (pSock->is_closed()))
						{
							continue;
						}
						else
						{
							//상동
						}
					}
				}
				else // 1채널 노드들..
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
				// 위에서 Send 채널 실패시 Recv 채널의 유효성을 확인하였기 때문에, 
				// 여기선 Send 채널에 대해서만 설정하고 Recv는 처리하지 않도록 한다.
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