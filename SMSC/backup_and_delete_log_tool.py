#!/usr/bin/env /usr/local/bin/python

import os
import sys
import time
import glob
import re
from datetime import date

gw_oam_list = "---"
gw_dn_list = "---"
gw_up_list = "---"




# LOG Directory List Reading Function
def readConfig(filename):
    fconf = open(filename)

    listOpts = []

    for list in fconf.readlines():
        if (list.startswith('#',0) or list.startswith('\n',0)):
            continue
        list = list[:list.find('\n')]
        #list = list.split(',')  # This step can supported ',' separate indicator
        listOpts.append(list)
    fconf.close()

    return listOpts





# Check the input option is number Function
def isNumber(s):
    return s.isdigit()


# Calculation for delete OAM log files
def dayCheck(t_val):
    if isNumber(t_val) == True :
        t_val = int(t_val)*60*60*24
        dateinfo = date.fromtimestamp(time.time() - t_val)
        chk_time = dateinfo.strftime("%Y%m%d")
#        print chk_time
    else:
        sys.exit("please don't input character....")

    return chk_time





# Calculation for compress GWFEP log files
def dayCheck_separate(t_val):
	chk_time = []
	if isNumber(t_val) == True:
		t_val = int(t_val)*60*60*24
		dateinfo = date.fromtimestamp(time.time() - t_val)
		chk_year = dateinfo.strftime("%Y")
		chk_month = dateinfo.strftime("%m")
		chk_day = dateinfo.strftime("%d")
		chk_time.append(chk_year)
		chk_time.append(chk_month)
		chk_time.append(chk_day)
#		print chk_time
	else:
		sys.exit("please don't input character....")

	return chk_time







# OAM log delete before than bk_date
def Deletelog(listfile):
	bk_date = raw_input("input backup date : ")
	bk_date = dayCheck(bk_date)
	for loglist in listfile:
		cmd = "/home/gwfep/HOME/log/%s" % loglist

		print cmd
		list = os.listdir(cmd)
		print list

		for logs in list :
			s=re.search(r'\d', logs)
			idx=[]

			if type(s) == type(None) :
				continue
			else :	
				idx =  s.start()
				if idx :
					if logs[idx:idx+8] <= bk_date :
						#print ' delete file is %s ' % logs
						delcmd = "rm -f %s" % cmd + "/" + "%s" % logs
						print delcmd
						#os.system(delcmd)
					else :
						continue
				else :
					print ' check date error... please check file name format '
	return ''





############### search low directory or files ###############
#def search(dirname):
#    flist = os.listdir(dirname)
#    for f in flist:
#        next = os.path.join(dirname, f)
#        if os.path.isdir(next):
#            search(next)
#        else:
#			ext = os.path.splitext(next)[-2]
#			print ext[-10:]
#			print ext
#			doFileWork(next)
#
#
#def doFileWork(filename):
#	#print os.getcwd()
#	#print os.chdir('../../')
#	ext = os.path.splitext(filename)[-1]
#    #print(filename)
#
##############################################################





# GWFEP log compress before than bk_date
def Compress_Download_log(listfile):
	bk_date = raw_input("input backup date : ")
	bk_date = dayCheck_separate(bk_date)
	for loglist in listfile:
		logs = "/home/gwfep/HOME/log/%s/DOWNLOAD" % (loglist)
		for path,dir,files in os.walk(logs):
			for file in files:
				log_date = path[-7:].replace("/","")+file
				if isNumber(log_date):
					if int(log_date) <= int(bk_date[0]+bk_date[1]+bk_date[2]):
						#os.system("gzip %s" % (path+'/'+file))
						print path+'/'+file
					else:
						#print "not yet"
						continue
				else:
					#print "type error"
					continue
	return ''





# GWFEP log compress before than bk_date
def Compress_Upload_log(listfile):
	bk_date = raw_input("input backup date : ")
	bk_date = dayCheck_separate(bk_date)
	for loglist in listfile:
		logs = "/HOME/log/%s/UPLOAD" % (loglist)
		ftplogs = "/HOME/log/FTP"
		for path,dir,files in os.walk(logs):
			#print path
			for file in files:
				#print os.path.join(path,file)
				log_date = path[-7:].replace("/","")+file[-2:]
				if isNumber(log_date):
				    if int(log_date) <= int(bk_date[0]+bk_date[1]+bk_date[2]):
				        #os.system("gzip %s" % os.path.join(path,file))
				        print path+'/'+file
				    else:
				        #print "not yet"
				        continue
				else:
				    #print "type error"
				    continue

        for path,dir,files in os.walk(ftplogs):
            #print path
            for file in files:
                #print os.path.join(path,file)
                ftplog_date = path[-7:].replace("/","")+file[-2:]
                if isNumber(ftplog_date):
                    if int(ftplog_date) <= int(bk_date[0]+bk_date[1]+bk_date[2]):
                        #os.system("gzip %s" % os.path.join(path,file))
                        print path+'/'+file
                    else:
                        #print "not yet"
                        continue
                else:
                    #print "type error"
                    continue
	return ''













def main(argv):
	if os.path.exists(gw_dn_list):
		dn_list = readConfig(gw_dn_list)
	else:
		print 'Error : %s is not found.' % gw_dn_list
		exit()


	if os.path.exists(gw_oam_list):
		oam_list = readConfig(gw_oam_list)
	else:
		print 'Error : %s is not found.' % gw_oam_list
		exit()


	if os.path.exists(gw_up_list):
		up_list = readConfig(gw_up_list)
	else:
		print 'Error : %s is not found.' % gw_up_list
		exit()


	Compress_Upload_log(up_list)
	Deletelog(oam_list)
	Compress_Download_log(dn_list)
	return ''



if __name__ == "__main__":
    main(sys.argv)

