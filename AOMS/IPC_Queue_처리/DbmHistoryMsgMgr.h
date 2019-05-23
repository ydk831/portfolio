#include <string>
#include <map>
#include <list>

#include "AxLib/AxLib.h"
#include "IpcQ/HistoryData.h"

class HistoryData;

using namespace std;

typedef list<HistoryData>           HDList;
typedef HDList::iterator            HDList_it;

typedef map<string, HistoryData>    HDMap;
typedef HDMap::iterator             HDMap_it;
typedef HDMap::value_type           HDMap_val;

class DbmHistoryMsgMgr
{
public:

	static DbmHistoryMsgMgr     *Instance();
	static DbmHistoryMsgMgr     *Instance(int polling, int margin);

	void                        RunHistoryGathering();
	void                        RunHistoryStatCreating();
	void                        RunHistoryStatRemoving();
	void                        RunSisHistoryRemoving();

	string                      GetStringPastTime(int pastHour, bool bUseTbl = false);

	void                        ProcessingHstatData(HDList oList);
	void                        InsertHstatData(string jobDate, int CsNum);

	void                        SetQueryTime(time_t &rTmStt, time_t &rTmEnd, bool isFail, int nCnt);
	void                        SetQueryTimeForSisRemoving(time_t &rTmStt, time_t &rTmEnd, bool isFail, int nCnt);

	void                        SetProcPollingTime(int tm) { m_PollingInterval = tm; }
	void                        SetProcMarginTime(int tm) { m_PollingMargin = tm; }

private:

	static DbmHistoryMsgMgr     *m_spInstance;
#ifdef _R440_DVT_
	DbmHistoryMsgMgr(int polling = 5, int margin = 1);
#endif
#ifndef _R440_DVT_
	DbmHistoryMsgMgr(int polling = 5, int margin = 720);
#endif
	virtual ~DbmHistoryMsgMgr();

	AxWorker                    *m_pHistoryGathering;
	AxWorker                    *m_pHistoryStatMaking;
	AxWorker                    *m_pHistoryStatRemoving;
	AxWorker                    *m_pSisHistoryRemoving;

	AxMutex                     m_HDataLock;

	HDMap                       m_HistoryDataMap;

	int                         m_PollingInterval;
	int                         m_PollingMargin;


};
