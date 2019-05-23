#ifndef COMMLIB_MAPPINGTABLE_MAP_TBL_MANAGER_H_
#define COMMLIB_MAPPINGTABLE_MAP_TBL_MANAGER_H_

#include "AxLib/AxLib.h"
#include "MappingTable/MapTblInfoDbApi.h"

#define MAP_TBL_HASH_SIZE   50000

#define MAP_TBL_TYPE_6  6

class MapTblManager
{
public:
	static MapTblManager*       Instance();
	int                         Initialize(char *pUdbId, bool isCS, int interval = 5, int margin = 10, int count = 100);
	void                        Stop();
	void                        SetPollingInterval(int interval);
	void                        SetPollingMargin(int margin);
	string                      FindAomcId(char* mdn, int type = 0);
	string                      FindMdn(char* aomcid);
	void                        SetProcCountPerSec(int count);
	int                         RemainQueSize();

	int                         InsertMapTbl(string mdn, string aomcid, bool hdvSupp, uint32_t maid, const uint8_t *cisno, const uint8_t *imscisno, string pddSvcType, bool flag = false);
	int                         DeleteMapTbl(string mdn, string aomcid = "");
	/** R402. HOST/MA 정보 재구성 API
	* Secondary CS 및 CS 재기동시 수행됨. */
	//int                         UpdateMapTbl(string mdn, string aomcid, bool hdvSupp, uint32_t maid, const uint8_t *cisno, const uint8_t *imscisno, time_t tm=0);
	int                         UpdateMapTbl(string mdn, string aomcid, bool hdvSupp, uint32_t maid, const uint8_t *cisno, const uint8_t *imscisno, time_t tm = 0, string pddSvcType = "");
	//int                         UpdateMapTbl(string mdn, string aomcid, bool hdvSupp, uint16_t pddSvcType[], uint32_t maid, const uint8_t *cisno, const uint8_t *imscisno, time_t tm=0);

	void                        uSleep(int usec);

	/** R402. Multi Account Info API */
	uint32_t                    GetMultiAccountCount(string mdn);
	uint32_t                    GetMultiAccountCount(char* mdn) { return GetMultiAccountCount(string(mdn)); }
	uint32_t                    GetActiveAccountInfo(string mdn, AccountInfo &out);
	uint32_t                    GetActiveAccountInfo(char* mdn, AccountInfo &out) { return GetActiveAccountInfo(string(mdn), out); }
	uint32_t                    GetAccountInfoList(string mdn, list<AccountInfo> &out);
	uint32_t                    GetAccountInfoList(char* mdn, list<AccountInfo> &out) { return GetAccountInfoList(string(mdn), out); }

	bool                        ComparePddSvc(uint16_t org[], uint16_t comp[]);
	/** R420. PDD SVC 고도화 */
	bool                        FindPddSvc4Subs(const char *mdn, __uint16_t idx, string &pddsvc);
	string                      toStringPddIdx(uint16_t pddsvc[]);
	void                        GetPddSvcInfo(string aomcid, string &pddsvc);

private:
	MapTblManager();
	~MapTblManager();
	void                        Init(bool bMap = true);
	int                         Hash(string mdn);
	void                        RunMapTblInfoPolling();
	void                        RunMapTblInfoProcessing();
	void                        SetQueryTime(time_t &tmStt, time_t &tmEnd, bool isFail, int nCnt);

private:

	static MapTblManager       *m_pInstance;
	AxWorker                    *m_pWorkerPolling;
	int                         m_PollingInterval;
	int                         m_PollingMargin;
	AxString                    m_UdbId;


	AxMutex                     m_lock_map_mdn[MAP_TBL_HASH_SIZE];
	//AxMutex                     m_lock_map_aomcid[MAP_TBL_HASH_SIZE];
	mtiMap                      m_map_mdn[MAP_TBL_HASH_SIZE];   // MDN key
																//mtiMap                      m_map_aomcid[MAP_TBL_HASH_SIZE];  // AOMCID Key

	AxMutex                     m_QueLock;
	AxWorker                    *m_pWorkerProcessing;
	mtiQue                      m_ProcQue;
	int                         m_ProcCount;        // 초당 Insert/Delete 처리 Limit
	bool                        m_InitFlag;
	bool                        m_isCS;
};

#endif /* COMMLIB_MAPPINGTABLE_MAP_TBL_MANAGER_H_ */
