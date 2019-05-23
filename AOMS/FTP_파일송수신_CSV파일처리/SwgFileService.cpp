#include <iostream>
#include <fstream>
#include <dirent.h>
#include <sys/stat.h>
#include <sys/types.h>

#include "OraDBApi.h"
#include "SwgIhData.h"
#include "SwgFileService.h"

SwgFileService* SwgFileService::m_spInstance = NULL;

void* SwgFileService::RunSwgFileService(void *data)
{
	CDLOG(" [SwgFileService::RunSwgFileService] Starting Thread");

	SwgFileService* fileSvc = reinterpret_cast<SwgFileService*>(data);

	while (true)
	{
		if (fileSvc->m_fTerm == true)
		{
			usleep(1000);
			continue;
		}


		SVCFilesInfo* lFileService = NULL;
		{
			AxLock lock(fileSvc->m_SvcMutex);
			if (fileSvc->m_SvcFilesInfo.empty())
			{
				usleep(1000);
				continue;
			}

			lFileService = fileSvc->m_SvcFilesInfo.front();
			fileSvc->m_SvcFilesInfo.pop_front();
			WLOG("[SwgFileService::RunSwgFileService()] Found FTP File : " << lFileService->GetLocalFile());
		}

		if (fileSvc->FtpFileProc((char*)lFileService->GetLocalFile().c_str()) < 0)
		{
			EELOG("fileSvc->FtpFileProc is Fail, Skip Processing");
			continue;
		}
		else
		{

			string FinDirPath = SwgIhData::instance()->GetSWGFtpLocalFileDir();
			FinDirPath.append("/FIN/");

			string FinFile = lFileService->GetLocalFile().substr((SwgIhData::instance()->GetSWGFtpLocalFileDir().size() + 1), _LEN_FTP_FILE);

			FinFile = FinDirPath + FinFile;

			int iCrtDirRst = mkdir(FinDirPath.c_str(), 0755);
			if (iCrtDirRst < 0)
			{
				if (errno == EEXIST)
					CDLOG("[SwgFileService::RunSwgFileService] FTP FIN File Directory Exist");
				else
				{
					CELOG("[SwgFileService::RunSwgFileService] Can not Create FTP FIN File Directory. Please check Local Directory Nameing or Permission...");
					CELOG("[SwgFileService::RunSwgFileService] Skip Processing");
					continue;
				}
			}

			int iMoveFinRst = rename(lFileService->GetLocalFile().c_str(), FinFile.c_str());
			if (iMoveFinRst < 0)
			{
				CELOG("[SwgFileService::RunSwgFileService] Move Ftp File to Fin Directory Fail(" << lFileService->GetLocalFile().c_str() << ")");
			}

			if (fileSvc->m_DevInfo.empty())
			{
				CELOG("[SwgFileService::RunSwgFileService] FTP File have no Data.(" << lFileService->GetLocalFile().c_str() << ")");
				continue;
			}

			{
				list<DevInfo*> lDevInfo = fileSvc->m_DevInfo;
				list<DevInfo*>::iterator iter;

				for (iter = lDevInfo.begin(); iter != lDevInfo.end(); ++iter)
				{
					DevInfo *info = *iter;
					if (fileSvc->FtpInsertDevInfo(info) < 0)
					{
						//DB Insert Fail Procedure????
					}
				}
			}

			fileSvc->m_DevInfo.clear();
		}

		delete lFileService;
		usleep(1000);
	}
}