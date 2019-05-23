#!/usr/local/bin/bash
#####################################################
#Shell Information
#Shell Release : R0.1.0
#Shell Name : ck_sys.sh
#Shell Code :
#Shell SYS  : OM only
#Shell Used : CMD
#Create by  : Uangel System Dev
#Create date: 2014.08.08
#Shell Discription : uangel smsc system check tool
#####################################################


##### USER CONFIG #####

# resource useage limit percent
MEM_CHK=75
DISK_CHK=70


# HOSTS_CNT is offset of smsom server. this is same thing that HOSTS list index
# you must write hosts list like ( "SMSC" "SMSC" SMSCOM" ) pattern

# TB environment
#HOSTS_CNT="2"
#HOSTS=( "SMS1A" "SMS1B" "SMSOM1" )


# YI enviroment
HOSTS_CNT="2 5 8"
HOSTS=( "SMS6A" "SMS6B" "SMSOM6" "SMS8A" "SMS8B" "SMSOM8" "SMS9A" "SMS9B" "SMSOM9" )


# mysql connection information
MYSQL_ID=---
MYSQL_PASS=---

# YI ODBC
MYSQL_DB=---
# GR ODBC
#MYSQL_DB=--- 



# using run alarm check shell
ALM_CHK_LOCATION='---'

# ignore alarm when occured less equal then ALM_CHK_TIME
ALM_CHK_TIME=8




##### NO USER CONFIG ######


NO_ARGS=0 
OPTERROR=65


###########################











function head_line() {

	echo -e "|---------------------------------------------------------------------------------------------------------------|"

}










# CPU #

function CPU_CHK() {
	CPU_MAX_A=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select round(used,1) from ST_resource_cpu_5M where used = (select max(used) from ST_resource_cpu_5M where system_id=\"${HOSTS[$omnum-2]}\")" | tail -1`
	CPU_AVG_A=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select round(avg(used),1) from ST_resource_cpu_5M where system_id=\"${HOSTS[$omnum-2]}\"" | tail -1`
	CPU_TIM_A=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select hour,min from ST_resource_cpu_5M where used = (select max(used) from ST_resource_cpu_5M where system_id=\"${HOSTS[$omnum-2]}\")" | tail -1 | awk '{print $1":"$2}'`
	
	CPU_MAX_B=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select round(used,1) from ST_resource_cpu_5M where used = (select max(used) from ST_resource_cpu_5M where system_id=\"${HOSTS[$omnum-1]}\")" | tail -1`
	CPU_AVG_B=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select round(avg(used),1) from ST_resource_cpu_5M where system_id=\"${HOSTS[$omnum-1]}\"" | tail -1`
	CPU_TIM_B=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select hour,min from ST_resource_cpu_5M where used = (select max(used) from ST_resource_cpu_5M where system_id=\"${HOSTS[$omnum-1]}\")" | tail -1 | awk '{print $1":"$2}'`
	
	CPU_MAX_OM=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select round(used,1) from ST_resource_cpu_5M where used = (select max(used) from ST_resource_cpu_5M where system_id=\"${HOSTS[$omnum]}\")" | tail -1`
	CPU_AVG_OM=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select round(avg(used),1) from ST_resource_cpu_5M where system_id=\"${HOSTS[$omnum]}\"" | tail -1`
	CPU_TIM_OM=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select hour,min from ST_resource_cpu_5M where used = (select max(used) from ST_resource_cpu_5M where system_id=\"${HOSTS[$omnum]}\")" | tail -1 | awk '{print $1":"$2}'`
	
	echo -e "| CPU  (MAX / AVG)   |\t$CPU_MAX_A%\t/\t$CPU_AVG_A%\t|\t$CPU_MAX_B%\t/\t$CPU_AVG_B%\t|\t$CPU_MAX_OM%\t/\t$CPU_AVG_OM%\t|"
	echo -e "| MAX_TIME           |\t$CPU_TIM_A\t \t\t|\t$CPU_TIM_B\t \t\t|\t$CPU_TIM_OM\t \t\t|"
}

























# MEMORY #

function MEM_CHK() {

	MEM_AVG_A=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select round(avg(used),0) from ST_resource_mem_5M where system_id=\"${HOSTS[$omnum-2]}\"" | tail -1`
	MEM_AVG_B=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select round(avg(used),0) from ST_resource_mem_5M where system_id=\"${HOSTS[$omnum-1]}\"" | tail -1`
	MEM_AVG_OM=`mysql -h${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "select round(avg(used),0) from ST_resource_mem_5M where system_id=\"${HOSTS[$omnum]}\"" | tail -1`
	
	function MEM_CMP() {
	    if [ $MEM_AVG_A -ge $MEM_CHK ]; then
	        MEM_CHK_A="NG"
	    else
	        MEM_CHK_A="OK"
	    fi
	
	    if [ $MEM_AVG_B -ge $MEM_CHK ]; then
	        MEM_CHK_B="NG"
	    else
	        MEM_CHK_B="OK"
	    fi
	
	    if [ $MEM_AVG_OM -ge $MEM_CHK ]; then
	        MEM_CHK_OM="NG"
	    else
	        MEM_CHK_OM="OK"
	    fi
	}

	MEM_CMP

	echo -e "| MEM  CHECK_RESULT  |\t$MEM_AVG_A%\t/\t$MEM_CHK_A\t|\t$MEM_AVG_B%\t/\t$MEM_CHK_B\t|\t$MEM_AVG_OM%\t/\t$MEM_CHK_OM\t|"

}



























# DISK #

function DISK_CHK_FUNC() {

	DISK_CHK_A=`remsh ${HOSTS[$omnum-2]} bdf | sed "s/%//" | awk -v CHK="$DISK_CHK" '{if ($5 >= CHK) print $5, $6}' | sed "s/ /% /g" | xargs echo`
	DISK_CHK_B=`remsh ${HOSTS[$omnum-1]} bdf | sed "s/%//" | awk -v CHK="$DISK_CHK" '{if ($5 >= CHK) print $5, $6}' | sed "s/ /% /g" | xargs echo`
	DISK_CHK_OM=`remsh ${HOSTS[$omnum]} bdf | sed "s/%//" | awk -v CHK="$DISK_CHK" '{if ($5 >= CHK) print $5, $6}' | sed "s/ /% /g" | xargs echo`
	
	
	DISK_A_LIST=( $DISK_CHK_A )
	DISK_B_LIST=( $DISK_CHK_B )
	DISK_OM_LIST=( $DISK_CHK_OM )
	
	DISK_A_CNT=`echo "${#DISK_A_LIST[@]}"`
	DISK_B_CNT=`echo "${#DISK_B_LIST[@]}"`
	DISK_OM_CNT=`echo "${#DISK_OM_LIST[@]}"`
	
		if [ $DISK_A_CNT -ge $DISK_B_CNT ]; then
			DISK_CNT=$DISK_A_CNT
		else
			DISK_CNT=$DISK_B_CNT
		fi
	
		if [ $DISK_CNT -ge $DISK_OM_CNT ]; then
			DISK_CNT=$DISK_CNT
		else
			DISK_CNT=$DISK_OM_CNT
		fi


	
	function DISK_CMP() {

		DISK_CHK_A_CNT=`remsh ${HOSTS[$omnum-2]} bdf | awk -v CHK="$DISK_CHK" '{if (substr($5,1,2) >= CHK) print $5, $6}' | wc -l`
		DISK_CHK_B_CNT=`remsh ${HOSTS[$omnum-1]} bdf | awk -v CHK="$DISK_CHK" '{if (substr($5,1,2) >= CHK) print $5, $6}' | wc -l`
		DISK_CHK_OM_CNT=`remsh ${HOSTS[$omnum]} bdf | awk -v CHK="$DISK_CHK" '{if (substr($5,1,2) >= CHK) print $5, $6}' | wc -l`
		
		    if [ $DISK_CHK_A_CNT -eq 0 ]; then
		        DISK_CHK_A_RESULT="OK"
		    else
		        DISK_CHK_A_RESULT="NG"
		    fi
		
		    if [ $DISK_CHK_B_CNT -eq 0 ]; then
		        DISK_CHK_B_RESULT="OK"
		    else
		        DISK_CHK_B_RESULT="NG"
		    fi
		
		    if [ $DISK_CHK_OM_CNT -eq 0 ]; then
		        DISK_CHK_OM_RESULT="OK"
		    else
		        DISK_CHK_OM_RESULT="NG"
		    fi
	}
	#excute function DISK_CMP
	DISK_CMP
	
	
	function DISK_PRINT() {
	
		IDX=1
		echo -e "| DISK CHECK_RESULT  |\t$DISK_CHK_A_RESULT\t\t\t|\t$DISK_CHK_A_RESULT\t\t\t|\t$DISK_CHK_OM_RESULT\t\t\t|"
		
		while [ $DISK_CNT -ge $IDX ]
		do
			echo -e "| PARTITION / USEAGE |\t${DISK_A_LIST[$IDX]}\t\t${DISK_A_LIST[$IDX-1]}\t|\t${DISK_B_LIST[$IDX]}\t\t${DISK_B_LIST[$IDX-1]}\t|\t${DISK_OM_LIST[$IDX]}\t\t${DISK_OM_LIST[$IDX-1]}\t|"
			IDX=`expr $IDX + 2`
		done
	}
	#excute function DISK_PRINT
	DISK_PRINT

}















function PROC_CHK() {

	P_LINE_A=`remsh ${HOSTS[$omnum-2]} dismc -a | awk '{if(substr($10,1,1) != 0) print $1}' | sed "1,3d" | wc -l`
	P_LINE_A=`expr $P_LINE_A - 3`
	
	
	P_LINE_B=`remsh ${HOSTS[$omnum-1]} dismc -a | awk '{if(substr($10,1,1) != 0) print $1}' | sed "1,3d" | wc -l`
	P_LINE_B=`expr $P_LINE_B - 3`
	
	
	P_LINE_OM=`remsh ${HOSTS[$omnum]} dismc -a | awk '{if(substr($10,1,1) != 0) print $1}' | sed "1,3d" | wc -l`
	P_LINE_OM=`expr $P_LINE_OM - 3`
	
	
	
	DEAD_P_CNT_A=`remsh ${HOSTS[$omnum-2]} dismc -a | awk '{if(substr($10,1,1) != 0) print $1}' | sed "1,3d" | sed -n "1,$P_LINE_A p" | xargs echo | wc -w `
	DEAD_P_CNT_B=`remsh ${HOSTS[$omnum-1]} dismc -a | awk '{if(substr($10,1,1) != 0) print $1}' | sed "1,3d" | sed -n "1,$P_LINE_B p" | xargs echo | wc -w `
	DEAD_P_CNT_OM=`remsh ${HOSTS[$omnum]} dismc -a | awk '{if(substr($10,1,1) != 0) print $1}' | sed "1,3d" | sed -n "1,$P_LINE_OM p" | xargs echo | wc -w `
	
	
	DEAD_P_A=`remsh ${HOSTS[$omnum-2]} dismc -a | awk '{if(substr($10,1,1) != 0) print $1}' | sed "1,3d" | sed -n "1,$P_LINE_A p" | xargs echo `
	DEAD_P_B=`remsh ${HOSTS[$omnum-1]} dismc -a | awk '{if(substr($10,1,1) != 0) print $1}' | sed "1,3d" | sed -n "1,$P_LINE_B p" | xargs echo `
	DEAD_P_OM=`remsh ${HOSTS[$omnum]} dismc -a | awk '{if(substr($10,1,1) != 0) print $1}' | sed "1,3d" | sed -n "1,$P_LINE_OM p" | xargs echo `


		if [ $DEAD_P_CNT_A -eq 1 ] && [ $DEAD_P_A = "=====================================================================================" ]; then
			DEAD_P_A=( "" )
			DEAD_P_CNT_A=0
		fi
		if [ $DEAD_P_CNT_B -eq 1 ] && [ $DEAD_P_B = "=====================================================================================" ]; then
			DEAD_P_B=( "" )
			DEAD_P_CNT_B=0
		fi
		if [ $DEAD_P_CNT_OM -eq 1 ] && [ $DEAD_P_OM = "=====================================================================================" ]; then
			DEAD_P_OM=( "" )
			DEAD_P_CNT_OM=0
		fi



	PROC_LIST_A=( $DEAD_P_A )
	PROC_LIST_B=( $DEAD_P_B )
	PROC_LIST_OM=( $DEAD_P_OM )



		if [ $DEAD_P_CNT_A -ge 1 ]; then
			PROC_STATE_A="NG"
		else
			PROC_STATE_A="OK"
		fi
		if [ $DEAD_P_CNT_B -ge 1 ]; then
			PROC_STATE_B="NG"
		else
			PROC_STATE_B="OK"
		fi
		if [ $DEAD_P_CNT_OM -ge 1 ]; then
			PROC_STATE_OM="NG"
		else
			PROC_STATE_OM="OK"
		fi





		if [ $DEAD_P_CNT_A -ge $DEAD_P_CNT_B ]; then
			PROC_LIST_CNT=$DEAD_P_CNT_A
		    if [ $DEAD_P_CNT_A -ge $DEAD_P_CNT_OM ]; then
	    	    PROC_LIST_CNT=$DEAD_P_CNT_A
	   		 else
	      		PROC_LIST_CNT=$DEAD_P_CNT_OM
	    	fi
		else
			PROC_LIST_CNT=$DEAD_P_CNT_B
		    if [ $DEAD_P_CNT_B -ge $DEAD_P_CNT_OM ]; then
	   		    PROC_LIST_CNT=$DEAD_P_CNT_B
	    	else
	        	PROC_LIST_CNT=$DEAD_P_CNT_OM
	    	fi
		fi



	echo -e "| PROC  CHECK_RESULT |\t\t$PROC_STATE_A\t\t|\t\t$PROC_STATE_B\t\t|\t\t$PROC_STATE_OM\t\t|"
	
	
	IDX=0
		while [ $PROC_LIST_CNT -gt $IDX ]
		do
			echo -e "| INIT_PROC_LIST     |\t${PROC_LIST_A[$IDX]}\t\t\t|\t${PROC_LIST_B[$IDX]}\t\t\t|\t${PROC_LIST_OM[$IDX]}\t\t\t|"
			IDX=`expr $IDX + 1`
		done
	
}	






































function TPS() {
	#mysql -h${HOSTS[3]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "Select ROUND(SUM(Summary),0) as 1H_TPS FROM ST_smsc_tps_1H WHERE  ( DATE = curdate()+0 ) ORDER BY DATE, HOUR, MIN ;"
	
	#mysql -h${HOSTS[3]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "Select ROUND(SUM(VALUEB0),0) as RT_TPS FROM ST_realtime_stat_1M WHERE ( DATE = curdate()+0) group by min order by date,HOUR,min desc limit 1;"



	
	TPS_QUERY=`mysql -h ${HOSTS[$omnum]} -u$MYSQL_ID -p$MYSQL_PASS $MYSQL_DB -A -e "Select hour,min,ROUND(SUM(Summary),0) as 5M_TPS FROM ST_smsc_tps_5M WHERE ( DATE = curdate()+0 ) group by hour,min order by 5M_TPS desc limit 1" | tail -1`

	SMS_TPS=( $TPS_QUERY )

	
	Echo -e "|     TPS(TIME)      |\t\t\t${SMS_TPS[2]} (${SMS_TPS[0]}:${SMS_TPS[1]})\t\t\t\t|\t\tNONE\t\t|"
	
	
	
}




function SYSTEM_CHK() {

	echo ""
	echo ""
	echo ""
	echo -e "[ CHECK USEAGE LEVEL          //     MEMORY : $MEM_CHK% ,    DISK : $DISK_CHK% ]"

	head_line
	echo -e "|                    |\t\t${HOSTS[$omnum-2]}\t\t|\t\t${HOSTS[$omnum-1]}\t\t|\t\t${HOSTS[$omnum]}\t\t|"
	head_line
	CPU_CHK
	head_line
	MEM_CHK
	head_line
	DISK_CHK_FUNC
	head_line
	PROC_CHK
	head_line
	TPS
	head_line

}










############################### MAIN FLOW OF SHELL SCRIPT ############################


clear


if [ $# -eq "$NO_ARGS" ]  # if called shell without option
then
  echo -e "useage: `basename $0` options (-asr)"
  echo -e "-a : system and alarm check\n-s : only system check\n-r : only alarm check" 
  exit $OPTERROR          # script exit
fi  


# option check
while getopts "asr" opt
do
	case $opt in
	a)
		for omnum in ${HOSTS_CNT[*]}
		do
		    SYSTEM_CHK
		    sleep 3
		done
		
		echo -e "\n\n\n\n"		
        echo -e " ################################# ALARM INFORMATION OF SMSC  #################################"
		echo -e "==============================================================================================="
		echo -e " HOST_NAME\t| ALARM_TIME\t| ALARM_LEVEL\t| ALARM_INFORMATION"
		echo -e "==============================================================================================="
		for hosts in ${HOSTS[*]}
		do
		    remsh $hosts "$ALM_CHK_LOCATION $ALM_CHK_TIME"
		done
		;;

	s) 
        for omnum in ${HOSTS_CNT[*]}
        do
            SYSTEM_CHK
            sleep 3
        done
		;;

	r)
        echo -e " ################################# ALARM INFORMATION OF SMSC  #################################"
        echo -e "==============================================================================================="
        echo -e " HOST_NAME\t| ALARM_TIME\t| ALARM_LEVEL\t| ALARM_INFORMATION"
        echo -e "==============================================================================================="
        for hosts in ${HOSTS[*]}
        do
            remsh $hosts $ALM_CHK_LOCATION $ALM_CHK_TIME
        done
        ;;

		
	*)
		echo -e "useage: `basename $0` options (-asr)"
  		echo -e "-a : system and alarm check\n-s : only system check\n-r : only alarm check" 
		exit 0
		;;
	esac
done
