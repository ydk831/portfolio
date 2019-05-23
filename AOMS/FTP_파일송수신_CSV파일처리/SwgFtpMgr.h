#ifndef __SWG_FTP_MGR_H__ 
#define __SWG_FTP_MGR_H__ 

using namespace std;

#include <string>

#include "AxLib/AxLib.h"

#include "SwgDef.h"
#include "SwgMgr.h"
#include "SwgMessage.h"

typedef enum {
	FTP_PROC_READY,
	FTP_PROC_START,
	FTP_PROC_END,
	FTP_PROC_FAIL
} ftpProcStatus_t;

class FTPFilesInfo {
public:
	~FTPFilesInfo() {};
	FTPFilesInfo() {};
	FTPFilesInfo(ftpSrvType_t pFtpType, string& pFilePath) :
		m_FtpType(pFtpType), m_FilePath(pFilePath) {};

	ftpSrvType_t GetFtpType() { return m_FtpType; }
	string& GetFilePath() { return m_FilePath; }
	string& GetFileName() { return m_FileName; }

	void SetFtpType(ftpSrvType_t pType) { m_FtpType = pType; }
	void SetFilePath(string& pPath) { m_FilePath = pPath; }
	void SetFileName(char* pName) { m_FileName = pName; }

	ftpProcStatus_t GetStatus() { return m_Status; }
	void SetStatus(ftpProcStatus_t pStatus) { m_Status = pStatus; }


private:
	ftpSrvType_t m_FtpType;
	string m_FilePath;
	string m_FileName;

	string m_FinPath;
	string m_FinName;

	ftpProcStatus_t   m_Status;
};


class SwgFtpMgr {
public:
	SwgFtpMgr();
	~SwgFtpMgr();

	string toString() { return string("SwgFtpMgr"); }

	bool Create();
	void Destroy();
	static SwgFtpMgr* Instance();

	bool StartSwgFtpMgr();
	bool StopSwgFtpMgr();

	void AddFTPFile(ftpSrvType_t pType, const string& pPath);
	int TimeParseProc(char cmpH[2], char cmpM[2]);

protected:
	enum
	{
		DEFAULT_STACK_SIZE = 1024 * 1024
	};
	enum
	{
		DEFAULT_POLL_PERIOD = 1000 * 50
	}; // in usec

	static SwgFtpMgr* m_spInstance;
	static void* RunSwgFtpMgr(void* data);

private:
	pthread_t   m_PthreadId;
	bool        m_fTerm;
	bool        m_bWorkDoneStt;

	list<FTPFilesInfo*>         m_FtpFilesInfo;
	AxMutex                     m_FtpMutex;

};

#endif /* __SWG_FTP_MGR_H__ */
