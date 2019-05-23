#ifndef __SWG_FILE_SERVICE_H__ 
#define __SWG_FILE_SERVICE_H__ 

#include <string>
#include <list>


#include "AxLib/AxLib.h"
#include "SwgDef.h"
#include "SwgMgr.h"

using namespace std;
using namespace SwgDef;

class SVCFilesInfo {
public:
	~SVCFilesInfo() {};
	SVCFilesInfo() {};
	SVCFilesInfo(string& pFilePath) : m_LocalFile(pFilePath) {};

	string& GetLocalFile() { return m_LocalFile; }
	void SetLocalFile(string& pFilePath) { m_LocalFile = pFilePath; }

private:
	string m_LocalFile;
};

class DevInfo {
public:
	DevInfo() {};
	~DevInfo() {};

	void SetDevInfo(string code, string model, string time, string nick)
	{
		m_DevCode = code;
		m_DevModel = model;
		m_DevNick = nick;
		m_DevModTime = time;
	}

	string GetDevCode() { return  m_DevCode; }
	string GetDevModel() { return  m_DevModel; }
	string GetDevNick() { return  m_DevNick; }
	string GetDevModTime() { return  m_DevModTime; }

private:
	string  m_DevCode;
	string  m_DevModel;
	string  m_DevNick;
	string  m_DevModTime;
};

class SwgFileService {
public:
	SwgFileService();
	~SwgFileService();

	string toString() { return string("SwgFileService"); }

	bool Create();
	void Destroy();
	static SwgFileService* Instance();

	bool StartSwgFileService();
	bool StopSwgFileService();

	void AddSVCFile(string& pPath);
	list<DevInfo*> GetDevInfo() { return m_DevInfo; }
	int32_t FtpFileProc(char* pFilePath);
	int32_t FtpInsertDevInfo(DevInfo *Info);

protected:
	enum
	{
		DEFAULT_STACK_SIZE = 1024 * 1024
	};
	enum
	{
		DEFAULT_POLL_PERIOD = 1000 * 50
	}; // in usec

	static SwgFileService* m_spInstance;
	static void* RunSwgFileService(void* data);



private:
	pthread_t   m_PthreadId;
	bool        m_fTerm;
	bool        m_bWorkDoneStt;
	int         m_FtpFileLineCnt;

	list<SVCFilesInfo*>         m_SvcFilesInfo;
	AxMutex                     m_SvcMutex;

	list<DevInfo*>              m_DevInfo;
};



#endif /* __SWG_FILE_SERVICE_H__ */
