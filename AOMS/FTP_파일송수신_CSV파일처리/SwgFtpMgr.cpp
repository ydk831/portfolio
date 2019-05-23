#include <iostream>
#include <string>
#include <string.h>
#include <time.h>
#include <dirent.h>
#include <unistd.h>
#include <sys/stat.h>
#include <sys/types.h>
#include "error.h"

#include "AxLib/AxLock.h"
#include "ftp.h"
#include "CommIh/IhUtil.h"

#include "SwgIhData.h"
#include "SwgFtpMgr.h"
#include "SwgFileService.h"

SwgFtpMgr* SwgFtpMgr::m_spInstance = NULL;

void* SwgFtpMgr::RunSwgFtpMgr(void *data)
{
	CDLOG("[SwgFtpMgr::RunSwgFtpMgr]  Starting Thread.");

	SwgFtpMgr* ftpMgr = reinterpret_cast<SwgFtpMgr*>(data);
	bool    bRetry = false;
	bool    bComplete = false;
	string  TargetFileName;
	int     mode;

	while (true)
	{
		if (ftpMgr->m_fTerm == true)
		{
			usleep(1000);
			continue;
		}

		//sleep(10);
		usleep(1000 * 1000);


		TargetFileName.clear();
		TargetFileName = "AOM_DCDM_";

		char confHour[2] = { '\0', };
		char confMin[2] = { '\0', };
		char RetryConfHour[2] = { '\0' };

		if (ftpMgr->TimeParseProc(confHour, confMin) < 0)
		{
			CELOG("TimeParseProc() Error... Please Check Configuration at FTP Time format!!");
			continue;
		}
		else
		{
			//CDLOG("Config Timer Parsing : Hour(" << confHour << "), Min(" << confMin << ")");
			int H = atoi(confHour) + 1;
			if (H == 24)
				strcpy(RetryConfHour, "00");
			else
				sprintf(RetryConfHour, "%02d", H);
			//CDLOG("Retry Timer Set : Hour(" << RetryConfHour << "), Min(" << confMin << ")");
		}

		time_t     now = time(0); //현재 시간을 time_t 타입으로 저장
		struct tm  tstruct;
		tstruct = *localtime(&now);
		char       curHour[3] = { '\0', };
		char       curMin[3] = { '\0', };
		char       Today[9] = { '\0', };
		bool       startFtp = false;

		strftime(curHour, sizeof(curHour), "%H", &tstruct);
		strftime(curMin, sizeof(curMin), "%M", &tstruct);
		strftime(Today, sizeof(Today), "%Y%m%d", &tstruct);

		if (bRetry == false)
		{
			CDLOG("Conf Time(" << confHour << ":" << confMin << "), Current Time(" << curHour << ":" << curMin << ")");
			if (!(strcmp(confHour, curHour) || strcmp(confMin, curMin)))
			{
				if (bComplete == true)
				{
					continue;
				}
				else
				{
					TargetFileName.append(Today);
					TargetFileName.append(".dat");
					startFtp = true;
				}
			}
			else
			{
				if (bComplete == true)
				{
					bComplete = false;
					continue;
				}
				else
				{
					continue;
				}
			}
		}
		else
		{
			CDLOG("Retry Time(" << RetryConfHour << ":" << confMin << "), Current Time(" << curHour << ":" << curMin << ")");
			if (!(strcmp(RetryConfHour, curHour) || strcmp(confMin, curMin)))
			{
				TargetFileName.append(Today);
				TargetFileName.append(".dat");
				startFtp = true;
			}
			else
				startFtp = false;
		}

		if (startFtp)
		{
			mode = SwgIhData::instance()->GetProcMode();

			// Target Dir
			char szTargetFileDir[_LEN_FTP_FILE];
			memset(szTargetFileDir, 0x00, _LEN_FTP_FILE);
			strcpy(szTargetFileDir, SwgIhData::instance()->GetSWGFtpTargetFileDir());

			// Target File
			char szTargetFile[_LEN_FTP_FILE];
			memset(szTargetFile, 0x00, _LEN_FTP_FILE);
			strcpy(szTargetFile, TargetFileName.c_str());

			WLOG("[SwgFtpMgr::RunSwgFtpMgr] TargetFileDir(" << szTargetFileDir << ") TargetFile(" << szTargetFile << ")");

			// My LocalFile
			char szMyFile[_LEN_FTP_FILE];
			memset(szMyFile, 0x00, _LEN_FTP_FILE);

			strcpy(szMyFile, SwgIhData::instance()->GetSWGFtpLocalFileDir().c_str());    /* Local File Path/Name */
			strcat(szMyFile, "/");
			strcat(szMyFile, szTargetFile);

			WLOG("[SwgFtpMgr::RunSwgFtpMgr] MyFile(" << szMyFile << ")");


			int iCrtDirRst = mkdir(SwgIhData::instance()->GetSWGFtpLocalFileDir().c_str(), 0755);
			if (iCrtDirRst < 0)
			{
				if (errno == EEXIST)
					CDLOG("[SwgFtpMgr::RunSwgFtpMgr] FTP Local Directory Exist");
				else
				{
					CELOG("[SwgFtpMgr::RunSwgFtpMgr] Can not Create FTP Local Directory. Please check Local Directory Nameing or Permission");
				}
			}

			int iRet = -1;
			if (mode == FTP_MODE)
			{
				WLOG("[SwgFtpMgr::RunSwgFtpMgr] Start FTP MODE Procedure");
				// FTP-GET-PUT
				ftpsession_t *fObj = ftp_connect(SwgIhData::instance()->GetSWGFtpIp(),
					SwgIhData::instance()->GetSWGFtpPort());

				if (fObj)
				{
					if (ftp_login(fObj, SwgIhData::instance()->GetSWGFtpUser(), SwgIhData::instance()->GetSWGFtpPassword()) == 0)
					{
						iRet = 0;

						/* Success is return 0. */
						iRet += ftp_cwd(fObj, szTargetFileDir);
						iRet += ftp_get(fObj, szTargetFile, szMyFile, TRAN_BINARY);
						//iRet += ftp_put(fObj, szFinFile, szFinFile+iFinFilePos, TRAN_BINARY);
					}
					else
					{
						CELOG("[SwgFtpMgr::RunSwgFtpMgr] ERROR_FTP.Login(" << SwgIhData::instance()->GetSWGFtpUser() << ":" << SwgIhData::instance()->GetSWGFtpPassword() << ")");
						ftp_disconnect(fObj);
						if (bRetry == false)
							bRetry = true;
						else
							bRetry = false;
						continue;
					}
					ftp_disconnect(fObj);
				}
				else
				{
					CELOG("[SwgFtpMgr::RunSwgFtpMgr] ERROR_FTP.Connection(" << SwgIhData::instance()->GetSWGFtpIp() << ":" << SwgIhData::instance()->GetSWGFtpPort() << ")");
					startFtp = false;
					if (bRetry == false)
					{
						CELOG("[SwgFtpMgr::RunSwgFtpMgr] Try to Retry Procedure");
						bRetry = true;
					}
					else
						bRetry = false;
					continue;
				}
			}
			else if (mode == EAI_MODE)
			{
				WLOG("[SwgFtpMgr::RunSwgFtpMgr] Start EAI MODE Procedure");
				char *check_file = szMyFile;
				int scriptResult;
				string EAIcmd = SwgIhData::instance()->GetEAIScript();
				EAIcmd.append(" >> /log/AOM/swgis.exe/log/EAI.log");

				if (0 == access(check_file, F_OK))
				{
					iRet = 0;
				}
				else
				{
					WLOG("[SwgFtpMgr::RunSwgFtpMgr] File Not Exist: " << szMyFile << ", Try EAI Script: " << EAIcmd);
					scriptResult = system(EAIcmd.c_str());
					if (scriptResult == 0)
					{
						if (0 == access(check_file, F_OK))
							iRet = 0;
						else
							WLOG("[SwgFtpMgr::RunSwgFtpMgr] Sucsses Script Excute, But File Not Exist: ");
					}
					else
					{
						EELOG("EAI Script Execute Fail!!!");
						startFtp = false;
						if (bRetry == false)
						{
							EELOG("[SwgFtpMgr::RunSwgFtpMgr] Try to Retry Procedure");
							bRetry = true;
						}
						else
							bRetry = false;
						continue;
					}
				}
			}
			else
			{
				EELOG("SWGIS Processing Mode is Not Available!!! Please Check Configuration!!!");
				continue;
			}

			// DataPare&Save
			if (iRet == 0)
			{
				string svcFile(szMyFile);

				WLOG("[SwgFtpMgr::RunSwgFtpMgr] Success Processing svcFile : " << svcFile);
				SwgFileService::Instance()->AddSVCFile(svcFile);
				bComplete = true;
				bRetry = false; // 성공 시 retry 초기화
			}
			else
			{
				if (bRetry == false)
				{
					EELOG("[SwgFtpMgr::RunSwgFtpMgr] Try to Retry Procedure");
					bRetry = true;
				}
				else
					bRetry = false;
			}
			startFtp = false;
		}
		else
		{
			CDLOG("[SwgFtpMgr::RunSwgFtpMgr] Do Not SWGIS Proceesing Time!");
		}
	}
}