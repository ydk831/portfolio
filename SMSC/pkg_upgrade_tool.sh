#!/usr/local/bin/bash
#####################################################
#Shell Information
#Shell Release : R0.0.1
#Shell Name : pkg_upgrade.sh
#Shell Code : 
#Shell SYS  : 
#Shell Used : CMD
#Create by  : Uangel System Dev
#Shell Discription :uangel smsc full pkg upgrade tool 
#####################################################
###################################################################################################################################
#environment setting zone

LOG_DATE=`date +%Y%m%d`
BASE_PATH="----"
who=`whoami`



####          Common Enviroment Call        ####

CFGname=$BASE_PATH/scripts/tool/PKG_INSTALL/pkg_up_env
USEname=$BASE_PATH/scripts/tool/PKG_INSTALL/.pkg_up_env

sed -e '/^\#/d' -e '/^$/d' $CFGname > "$USEname"

while read line
do
        declare -x $line

		done <$USEname

# owner check

if [ $who = $PKG_USR ]
then
	echo " Login user is $who "

else

	echo " This user is unpermitted ( $who ) !!"
	echo " plz relogin $PKG_USR user "
	exit 0

fi



SRC_BIN_PATH=`ls -al $BASE_PATH"/"bin | grep -e "->" | awk '{print $11}' | sed 's/^\///g' `
SRC_LIB_PATH=`ls -al $BASE_PATH"/"lib | grep -e "->" | awk '{print $11}' | sed 's/^\///g' `
SRC_BUILD_PATH=`ls -al $BASE_PATH"/"build | grep -e "->" | awk '{print $11}' | sed 's/^\///g' ` 
SRC_CONFIG_PATH=`ls -al $BASE_PATH"/"config | grep -e "->" | awk '{print $11}'| sed 's/^\///g'  `
SRC_DATA_PATH=`ls -al $BASE_PATH"/"data | grep -e "->" | awk '{print $11}' | sed 's/^\///g' `
OLD_VER=`grep VERSION $BASE_PATH"/"$SRC_CONFIG_PATH"/version.conf" |awk '{print $3}'`


###################################################################################################################################
#Function Block

function pkg_intro()
{
		clear
		echo "************************************************************"
    	echo " UANGEL SMSC Applocation pkg upgrade tool  "
        echo; echo  -n "Do you want to start pkg upgrade(n/y)? "
        read ans
        case "$ans" in
         [nN])   echo " Canceled CMD!"
                  exit 0 ;;
    	 [yY])   echo " Starting Menu" ;;
      		*)     echo " Typing error.... please retry.... "
                  exit 0 ;;
      esac
}

function chk()
{

    echo "===================================================================="
    echo; echo  -n "Do you want to proceed(n/y)? "
        read ans
        case "$ans" in
         [nN])   echo " Canceled CMD!"
                  exit 0 ;;

         [yY])   echo " Processing....." ;;


            *)   echo " Typing error.... please retry.... "
                  exit 0 ;;
      esac
}

function pkg_1st_menu()
{

until [ "$answer" = 0 ]
do
echo " "
echo " "
echo "             ================================================================="
echo "             Welcome to UANGEL SMSC APPLICATION PKG UPGRADE TOOL Main Menu !!!"
echo "             ================================================================="
echo " "
echo " ******************************************************************"
echo "  [1])   SHOW OLD PKG PATH INFO "
echo "  [2])   OLD PKG Backup Process  Menu"
echo "  [3])   NEW PKG Extract Process  Menu"
echo "  [4])   PKG lib / bin differ  Menu"
echo "  [5])   PKG relinking Menu"
echo "  [0])   EXIT" 
echo " ******************************************************************"
echo; echo  -n "Do you want to run CMD (0-5)? "
read answer

case "$answer" in
 [1])   echo " SHOW OLD PKG PATH INFO"
        echo "====================================================================================================================="
		pwd
        echo " ls -al  $BASE_PATH ......  bin lib data config build "
        echo "---------------------------------------------------------------------------------------------------------------------"
        ls -al  $BASE_PATH"/bin"
        ls -al  $BASE_PATH"/lib"
        ls -al  $BASE_PATH"/config"
        ls -al  $BASE_PATH"/data"
        ls -al  $BASE_PATH"/build"
        echo "====================================================================================================================="
        echo " "
        ;;

 [2])   echo " OLD PKG Backup Process  Menu"
	 	pkg_backup $*
        echo " "
        ;;

 [3])   echo " OLD PKG Backup Process  Menu"
	 	pkg_ext $*
        echo " "
        ;;
 [4])   echo " PKG lib/bin differ Menu"
	 	pkg_diff $*
        echo " "
        ;;
 [5])   echo " PKG link change Menu"
	 	pkg_chg_link $*
        echo " "
        ;;
 [0])   echo " EXIT"
        echo " Good Bye!!!" ;;
 
 *)     echo " Typing error.... please retry.... "
        echo " "
       echo " "
        ;;
  esac

done

}


function pkg_backup()
{

until [ "$answer" = 0 ]
do
echo " "
echo " "
echo "             ======================================================="
echo "             Welcome to UANGEL SMSC APPLICATION PKG UPGRADE TOOL !!!"
echo "             ======================================================="
echo " "
echo " ******************************************************************"
echo "  [1])   SHOW OLD PKG PATH INFO "
echo "  [2])   COPY OLD binary to New dir "
echo "  [3])   COPY OLD library to New dir  "
echo "  [4])   COPY OLD build to New dir  "
echo "  [5])   COPY OLD config to New dir  "
echo "  [6])   COPY OLD data to New dir  "
echo "  [7])   GOTO UPPER MENU" 
echo "  [0])   EXIT" 
echo " ******************************************************************"
echo; echo  -n "Do you want to run CMD (0-7)? "
read answer

# OLD PKG copy for each directory
case "$answer" in
 [1])   echo "====================================================================================================================="
 		echo " SHOW OLD PKG PATH INFO"
        echo "====================================================================================================================="
		pwd
        echo " ls -dl  $BASE_PATH ......  bin lib data config build "
        echo "---------------------------------------------------------------------------------------------------------------------"
        ls -dl  $BASE_PATH"/bin"
        ls -dl  $BASE_PATH"/lib"
        ls -dl  $BASE_PATH"/config"
        ls -dl  $BASE_PATH"/data"
        ls -dl  $BASE_PATH"/build"
        echo "====================================================================================================================="
        echo " "
        ;;

 [2])   echo "*******************************************************************************************************"
	 	echo " ATTENTION!! For Copy Setting orig value is ( $BASE_PATH"/"$SRC_BIN_PATH ) "
	 	echo "                      binary new version is ( $NEW_VER ) "
		echo "                        it will be make dir ( $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER ) "
		echo "*******************************************************************************************************"
		echo " If it has incorrect src values.  You can type new path. Next Step!! "
        echo; echo -n " if you use setting env value just press Enter! Otherwise input path info ex : /home/ktfsmc/config ?"
        read src
		echo "*******************************************************************************************************"
		echo " If it has incorrect version values.  You can type new version. Next Step!! "
        echo; echo -n " if you use setting env value just press Enter! Otherwise input NEW_VER name ex :  R4.6.0 ?"
        read new
		echo "*******************************************************************************************************"
        

		
		if [ -z $src ]
		then

			if [ -z $new ] 
			then
			
			echo "cp -rpf $BASE_PATH"/"$SRC_BIN_PATH $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"
			chk $*
			cp -rpf $BASE_PATH"/"$SRC_BIN_PATH $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER
			ll $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER

		else 
			
			echo "cp -rpf $BASE_PATH"/"$SRC_BIN_PATH $BASE_PATH"/"$SITE_INFO"_BIN_"$new"
			chk $*
			cp -rpf $BASE_PATH"/"$SRC_BIN_PATH $BASE_PATH"/"$SITE_INFO"_BIN_"$new
			ll $BASE_PATH"/"$SITE_INFO"_BIN_"$new

		fi

		else	
			if [ -z $new ] 
			then
			
				echo "cp -rpf $src $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"
				chk $*
				cp -rpf $src $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER
				ll $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER

			else 
			
				echo "cp -rpf $src $BASE_PATH"/"$SITE_INFO"_BIN_"$new"
				chk $*
				cp -rpf $src $BASE_PATH"/"$SITE_INFO"_BIN_"$new
				ll $BASE_PATH"/"$SITE_INFO"_BIN_"$new

			fi
		fi
        ;;
 
 
 [3])   echo "*******************************************************************************************************"
	 	echo " ATTENTION!! For Copy Setting orig value is ( $BASE_PATH"/"$SRC_LIB_PATH ) "
	 	echo "                      libary new version is ( $NEW_VER )"
		echo "                        it will be make dir ( $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER ) "
		echo "*******************************************************************************************************"
		echo " If it has incorrect src values.  You can type new path. Next Step!! "
        echo; echo -n " if you use setting env value just press Enter! Otherwise input path info ex : /home/ktfsmc/config ?"
        read src
		echo "*******************************************************************************************************"
		echo " If it has incorrect version values.  You can type new version. Next Step!! "
        echo; echo -n " if you use setting env value just press Enter! Otherwise input NEW_VER name ex : R4.6.0 ?"
        read new
		echo "*******************************************************************************************************"
        
		if [ -z $src ]
		then

			if [ -z $new ] 
			then
			
				echo "cp -rpf  $BASE_PATH"/"$SRC_LIB_PATH  $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"
				chk $*
				cp -rpf  $BASE_PATH"/"$SRC_LIB_PATH  $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER
				ll  $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER

			else 
			
				echo "cp -rpf  $BASE_PATH"/"$SRC_LIB_PATH  $BASE_PATH"/"$SITE_INFO"_LIB_"$new"
				chk $*
				cp -rpf  $BASE_PATH"/"$SRC_LIB_PATH $BASE_PATH"/"$SITE_INFO"_LIB_"$new
				ll $BASE_PATH"/"$SITE_INFO"_LIB_"$new

			fi
		else
			if [ -z $new ] 
			then
			
				echo "cp -rpf  $src  $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"
				chk $*
				cp -rpf  $src  $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER
				ll  $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER

			else 
			
				echo "cp -rpf  $src  $BASE_PATH"/"$SITE_INFO"_LIB_"$new"
				chk $*
				cp -rpf  $src $BASE_PATH"/"$SITE_INFO"_LIB_"$new
				ll $BASE_PATH"/"$SITE_INFO"_LIB_"$new

			fi
		fi
        ;;
 
 

 [4])   echo "*******************************************************************************************************"
	 	echo " ATTENTION!! For Copy Setting orig value is ( $BASE_PATH"/"$SRC_BUILD_PATH )"
	 	echo "                      build  new version is ( $NEW_VER ) "
		echo "                        it will be make dir ( $BASE_PATH"/"$SITE_INFO"_BUILD_"$NEW_VER ) "
		echo "*******************************************************************************************************"
		echo " If it has incorrect src values.  You can type new path. Next Step!! "
        echo; echo -n " if you use setting env value just press Enter! Otherwise input path info ex : /home/ktfsmc/config ?"
        read src 
		echo "*******************************************************************************************************"
		echo " If it has incorrect version values.  You can type new version. Next Step!! "
        echo; echo -n " if you use setting env value just press Enter! Otherwise input NEW_VER name ex : R4.6.0 ?"
        read new
		echo "*******************************************************************************************************"
        
		if [ -z $src ]
		then


			if [ -z $new ] 
			then
			
				echo "cp -rpf $BASE_PATH"/"$SRC_BUILD_PATH $BASE_PATH"/"$SITE_INFO"_BUILD_"$NEW_VER"
				chk $*
				cp -rpf $BASE_PATH"/"$SRC_BUILD_PATH  $BASE_PATH"/"$SITE_INFO"_BUILD_"$NEW_VER
				ll $BASE_PATH"/"$SITE_INFO"_BUILD_"$NEW_VER

			else 
			
				echo "cp -rpf $BASE_PATH"/"$SRC_BUILD_PATH   $BASE_PATH"/"$SITE_INFO"_BUILD_"$new"
				chk $*
				cp -rpf $BASE_PATH"/"$SRC_BUILD_PATH   $BASE_PATH"/"$SITE_INFO"_BUILD_"$new
				ll $BASE_PATH"/"$SITE_INFO"_BUILD_"$new

			fi

		else

			if [ -z $new ] 
			then
			
				echo "cp -rpf $src $BASE_PATH"/"$SITE_INFO"_BUILD_"$NEW_VER"
				chk $*
				cp -rpf $src  $BASE_PATH"/"$SITE_INFO"_BUILD_"$NEW_VER
				ll $BASE_PATH"/"$SITE_INFO"_BUILD_"$NEW_VER

			else 
			
				echo "cp -rpf $src   $BASE_PATH"/"$SITE_INFO"_BUILD_"$new"
				chk $*
				cp -rpf $src   $BASE_PATH"/"$SITE_INFO"_BUILD_"$new
				ll $BASE_PATH"/"$SITE_INFO"_BUILD_"$new

			fi
		fi
        ;;
 
 
 [5])   echo "*******************************************************************************************************"
	 	echo " ATTENTION!! For Copy Setting orig value is (  $BASE_PATH"/"$SRC_CONFIG_PATH  )"
	 	echo "                      config new version is (  $NEW_VER  ) "
		echo "                        it will be make dir (  $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER ) "
		echo "*******************************************************************************************************"
		echo " If it has incorrect src values.  You can type new path. Next Step!! "
        echo; echo -n " if you use setting env value just press Enter! Otherwise input path info ex : /home/ktfsmc/config ?"
        read src
		echo "*******************************************************************************************************"
		echo " If it has incorrect version values.  You can type new version. Next Step!! "
        echo; echo -n " if you use setting env value just press Enter! Otherwise input NEW_VER name ex : R4.6.0 ?"
        read new
		echo "*******************************************************************************************************"
        
		if [ -z $src ]
		then

			if [ -z $new ] 
			then
			
				echo "cp -rpf $BASE_PATH"/"$SRC_CONFIG_PATH $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER"
				chk $*
				cp -rpf $BASE_PATH"/"$SRC_CONFIG_PATH $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER
				echo " version info changing..."
				rm -f   $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER"/version.conf"
				echo " OLD PKG Version is $OLD_VER  New PKG Version is $NEW_VER "
				sed s/$OLD_VER/$NEW_VER/g $BASE_PATH"/"$SRC_CONFIG_PATH"/version.conf" > $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER"/version.conf"
				echo " changed version info."
				echo " RESULT PATH : $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER "
				ll $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER

			else 
			
				echo "cp -rpf $BASE_PATH"/"$SRC_CONFIG_PATH $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new"
				chk $*
				cp -rpf $BASE_PATH"/"$SRC_CONFIG_PATH $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new
				echo " version info changing..."
				rm -f   $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new"/version.conf"
				echo " OLD PKG Version is $OLD_VER New PKG Version is $new "
				sed s/$OLD_VER/$new/g  $BASE_PATH"/"$SRC_CONFIG_PATH"/version.conf" > $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new"/version.conf"
				echo " changed version info."
				echo " RESULT PATH : $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new "
				ll $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new

			fi
		else
			if [ -z $new ] 
			then
			
				echo "cp -rpf $src $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER"
				chk $*
				cp -rpf $src $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER
				echo " version info changing..."
				rm -f   $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER"/version.conf"
				OLD_VER=`grep VERSION $src"/version.conf" |awk '{print $3}'`		
				echo " OLD PKG Version is $OLD_VER  New PKG Version is $NEW_VER "
				sed s/$OLD_VER/$NEW_VER/g  $src"/version.conf" > $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER"/version.conf"
				echo " changed version info."
				echo " RESULT PATH : $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER "
				ll $BASE_PATH"/"$SITE_INFO"_CONFIG_"$NEW_VER

			else 
			
				echo "cp -rpf $src $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new"
				chk $*
				cp -rpf $src $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new
				echo " version info changing..."
				rm -f   $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new"/version.conf"
				OLD_VER=`grep VERSION $src"/version.conf" |awk '{print $3}'`		
				echo " OLD PKG Version is $OLD_VER New PKG Version is $new "
				sed s/$OLD_VER/$new/g $src"/version.conf" > $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new"/version.conf"
				echo " changed version info."
				echo " RESULT PATH : $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new "
				ll $BASE_PATH"/"$SITE_INFO"_CONFIG_"$new

			fi

		fi		
		
        ;;
 
 

 [6])   echo "*******************************************************************************************************" 
	 	echo " ATTENTION!! For Copy Setting orig value is ( $BASE_PATH"/"$SRC_DATA_PATH )"
	 	echo "                        data new version is ( $NEW_VER ) "
		echo "                        it will be make dir ( $BASE_PATH"/"$SITE_INFO"_DATA_"$NEW_VER ) "
		echo "*******************************************************************************************************"
		echo " If it has incorrect src values.  You can type new path. Next Step!! "
        echo; echo -n " if you use setting env value just press Enter! Otherwise input path info ex : /home/ktfsmc/config ?"
        read src
		echo "*******************************************************************************************************"
		echo " If it has incorrect version values.  You can type new version. Next Step!! "
        echo; echo -n " if you use setting env value just press Enter! Otherwise input NEW_VER name ex : R4.6.0 ?"
        read new
		echo "*******************************************************************************************************"
        
		if [ -z $src ]
		then


			if [ -z $new ] 
			then
			
				echo "cp -rpf $BASE_PATH"/"$SRC_DATA_PATH $BASE_PATH"/"$SITE_INFO"_DATA_"$NEW_VER"
				chk $*
				cp -rpf $BASE_PATH"/"$SRC_DATA_PATH $BASE_PATH"/"$SITE_INFO"_DATA_"$NEW_VER
				ll $BASE_PATH"/"$SITE_INFO"_DATA_"$NEW_VER

			else 
			
				echo "cp -rpf $BASE_PATH"/"$SRC_DATA_PATH  $BASE_PATH"/"$SITE_INFO"_DATA_"$new"
				chk $*
				cp -rpf $BASE_PATH"/"$SRC_DATA_PATH  $BASE_PATH"/"$SITE_INFO"_DATA_"$new
				ll $BASE_PATH"/"$SITE_INFO"_DATA_"$new

			fi
		else

			if [ -z $new ] 
			then
			
				echo "cp -rpf $src $BASE_PATH"/"$SITE_INFO"_DATA_"$NEW_VER"
				chk $*
				cp -rpf $src $BASE_PATH"/"$SITE_INFO"_DATA_"$NEW_VER
				ll $BASE_PATH"/"$SITE_INFO"_DATA_"$NEW_VER

			else 
			
				echo "cp -rpf $src  $BASE_PATH"/"$SITE_INFO"_DATA_"$new"
				chk $*
				cp -rpf $src  $BASE_PATH"/"$SITE_INFO"_DATA_"$new
				ll $BASE_PATH"/"$SITE_INFO"_DATA_"$new

			fi
		fi
        ;;
 

 [7])   echo " GOTO UPPER MENU"
	 	
		pkg_1st_menu $*
 
	 	;;

 [0])   echo " EXIT"
        echo " Good Bye!!!" ;;
 
 *)     echo " Typing error.... please retry.... "
        echo " "
       	echo " "
        ;;
  esac

 done

}
 



function pkg_ext()
{

until [ "$answer" = 0 ]
do
echo " "
echo " "
echo "             ==================================================================="
echo "             Welcome to UANGEL SMSC APPLICATION PKG UPGRADE TOOL for EXTRACT !!!"
echo "             ==================================================================="
echo " "
echo " ******************************************************************"
echo "  [1])   SHOW NEW(BACKUPED) PKG PATH INFO "
echo "  [2])   EXTRACT binary to New dir "
echo "  [3])   EXTRACT library to New dir  "
echo "  [4])   GOTO UPPER MENU " 
echo "  [0])   EXIT" 
echo " ******************************************************************"
echo; echo  -n "Do you want to run CMD (0-4)? "
read answer


#NEW PKG File extract for each directory
case "$answer" in
 [1])   echo " SHOW NEW(BACKUPED) PKG PATH INFO"
        echo "====================================================================================================================="
		echo " locate " 
		pwd
		echo ""
        echo " ls -al  $BASE_PATH ......  bin lib build config data "
        echo "---------------------------------------------------------------------------------------------------------------------"
        ls -al  $BASE_PATH"/$SITE_INFO"_BIN_"$NEW_VER"
        ls -al  $BASE_PATH"/$SITE_INFO"_LIB_"$NEW_VER"
        ls -al  $BASE_PATH"/$SITE_INFO"_BUILD_"$NEW_VER"
        ls -al  $BASE_PATH"/$SITE_INFO"_CONFIG_"$NEW_VER"
        ls -al  $BASE_PATH"/$SITE_INFO"_DATA_"$NEW_VER"
        echo "====================================================================================================================="
        echo " "
        ;;


 [2])   echo " Extract binary to $NEW_VER dir "
	 	
	 chk_new_bin_dir=`ls -d $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER  |wc -l`

		if [ $chk_new_bin_dir = "0" ]
		then
			echo "*************************************************************"
			echo " error!! unkown dir!! $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			echo "*************************************************************"
			pkg_1st_menu $*
		else
			echo "*************************************************************"
			echo " dir check ok!"
			echo "*************************************************************"

		echo ""
			echo "*************************************************************"
 			echo " Source file info is $SOURCE_PATH"/"$SOURCE_BIN_FILE "
			echo " Target Path Info is $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			echo "*************************************************************"
			echo ""
			echo "*************************************************************"
			echo " if info is incorrect and than next step you can typing"
			echo "*************************************************************"
			echo ""
			echo "--------------------------------------------------------------------------------------------------------"
        	echo; echo -n " if you use setting env value just press Enter! Otherwise input source path ex : /home/ktfsmc/home/ ?"
        read s_path
			echo ""
			echo "--------------------------------------------------------------------------------------------------------"
        	echo; echo -n " if you use setting env value just press Enter! Otherwise input source file name ex : kt_r460_bin.tar ?"
        read s_file

		fi

	if [ -z $s_path ] 
	then
		
		if [ -z $s_file ] 
		then
			
			echo "cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"
			cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER
			echo " locale info is" 
			pwd
			echo " extact $SOURCE_PATH"/"$SOURCE_BIN_FILE  to $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			chk $*
			rm -f $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/stopmc"
			echo "rm -f $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/stopmc" "
			tar xvf $SOURCE_PATH"/"$SOURCE_BIN_FILE  
			echo""
			echo "******************************************************"
			echo " change stopmc owner and permition. plz root login!!! "
			echo "******************************************************"
			su - root -c "chown root:$PKG_GRP $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc ; chmod +s  $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc"
			echo " [INFO] : chown root:$PKG_GRP $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc " 
			echo " [INFO] : chmod  $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc " 

		else

			echo "cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"
			cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER
			echo " locale info is" 
			pwd
			echo " extact $SOURCE_PATH"/"$s_file  to $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			chk $*
			rm -f $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/stopmc"
			echo "rm -f $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/stopmc" "
			tar xvf $SOURCE_PATH"/"$s_file  
			echo""
			echo "******************************************************"
			echo " change stopmc owner and permition. plz root login!!! "
			echo "******************************************************"
			su - root -c "chown root:$PKG_GRP $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc ; chmod +s  $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc"

			echo " [INFO] : chown root:$PKG_GRP $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc " 
			echo " [INFO] : chmod  $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc " 
		fi

	else
		if [ -z $s_file ] 
		then

			echo "cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"
			cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER
			echo " locale info is" 
			pwd
			echo " extact $s_path"/"$SOURCE_BIN_FILE  to $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			chk $*
			rm -f $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/stopmc"
			tar xvf $s_path"/"$SOURCE_BIN_FILE 
			echo""
			echo "******************************************************"
			echo " change stopmc owner and permition. plz root login!!! "
			echo "******************************************************"
			su - root -c "chown root:$PKG_GRP $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc ; chmod +s  $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc"

			echo " [INFO] : chown root:$PKG_GRP $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc " 
			echo " [INFO] : chmod  $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc " 

		else 
			
			echo "cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"
			cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER
			echo " locale info is" 
			pwd
			echo " extact $s_path"/"$s_file  to $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			chk $*
			rm -f $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/stopmc"
			tar xvf $s_path"/"$s_file
			echo""
			echo "******************************************************"
			echo " change stopmc owner and permition. plz root login!!! "
			echo "******************************************************"
			su - root -c "chown root:$PKG_GRP $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc ; chmod +s  $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc"

			echo " [INFO] : chown root:$PKG_GRP $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc " 
			echo " [INFO] : chmod  $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"/"stopmc " 

		fi
	fi
    ;;



 [3])   echo " Extract library $NEW_VER dir "
	 chk_new_lib_dir=`ls -d $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER  |wc -l`

		if [ $chk_new_lib_dir = "0" ]
		then
			echo "*************************************************************"
			echo " error!! unkown dir!! $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
			echo "*************************************************************"
			pkg_1st_menu $*
		else
			echo "*************************************************************"
			echo " dir check ok!"
			echo "*************************************************************"

			echo ""
			echo "*************************************************************"
 			echo " Source file info is $SOURCE_PATH"/"$SOURCE_LIB_FILE "
			echo " Target Path Info is $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
			echo "*************************************************************"
			echo ""
			echo "*************************************************************"
			echo " if it is incorrect and than next step you can typing"
			echo "*************************************************************"
			echo ""
			echo "--------------------------------------------------------------------------------------------------------"
    	    echo; echo -n " if you use setting env value just press Enter! Otherwise input source path ex : /home/ktfsmc/home/ ?"
        read s_path
			echo ""
			echo "--------------------------------------------------------------------------------------------------------"
        echo; echo -n " if you use setting env value just press Enter! Otherwise input source file name ex : kt_r460_lib.tar ?"
        read s_file

		fi

	if [ -z $s_path ] 
	then
			
			if [ -z $s_file ] 
			then
			
				echo "cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"
				cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER
				echo " locale info is" 
				pwd
				echo " extact $SOURCE_PATH"/"$SOURCE_LIB_FILE  to $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
				chk $*
				rm -rf  $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"/hpuxia64"
				tar xvf $SOURCE_PATH"/"$SOURCE_LIB_FILE  


			else

				echo "cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"
				cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER
				echo " locale info is" 
				pwd
				echo " extact $SOURCE_PATH"/"$s_file to $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
				chk $*
				rm -rf  $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"/hpuxia64"
				tar xvf $SOURCE_PATH"/"$s_file  

			fi

	else 
			
		if [ -z $s_file ] 
		then
			echo "cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"
			cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER
			echo " locale info is" 
			pwd
			echo " extact $s_path"/"$SOURCE_LIB_FILE  to $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
			chk $*
				rm -rf  $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"/hpuxia64"
			tar xvf $s_path"/"$SOURCE_LIB_FILE  

		else
			
			echo "cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"
			echo "cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"
			cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER
			echo " locale info is" 
			pwd
			echo " extact $s_path"/"$s_file  to $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
			chk $*
				rm -rf  $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"/hpuxia64"
			tar xvf $s_path"/"$s_file  

		fi
	fi

	;;



 [4])   echo " GOTO UPPER MENU"

         pkg_1st_menu $*

         ;;

 [0])   echo " EXIT"
        echo " Good Bye!!!" ;;

   *)     echo " Typing error.... please retry.... "
          echo " "
          echo " "
          ;;
     esac

 done

}


function pkg_diff()
{

until [ "$answer" = 0 ]
do
echo " "
echo " "
echo "             ==================================================================="
echo "             Welcome to UANGEL SMSC APPLICATION PKG UPGRADE TOOL for DIFFER !!!"
echo "             ==================================================================="
echo " "
echo " ******************************************************************"
echo "  [1])   SHOW NEW(BACKUPED) PKG PATH INFO "
echo "  [2])   DIFFER with source  binary VS New dir "
echo "  [3])   DIFFER with source library VS New dir  "
echo "  [4])   GOTO UPPER MENU " 
echo "  [0])   EXIT" 
echo " ******************************************************************"
echo; echo  -n "Do you want to run CMD (0-4)? "
read answer

case "$answer" in
 [1])   echo " SHOW NEW(BACKUPED) PKG PATH INFO"
        echo "====================================================================================================================="
		echo " locate " 
		pwd
		echo ""
        echo " ls -al  $BASE_PATH ......  bin lib " 
        echo "---------------------------------------------------------------------------------------------------------------------"
        ls -al  $BASE_PATH"/$SITE_INFO"_BIN_"$NEW_VER"
        echo "---------------------------------------------------------------------------------------------------------------------"
        ls -al  $BASE_PATH"/$SITE_INFO"_LIB_"$NEW_VER"
        echo "---------------------------------------------------------------------------------------------------------------------"
		ls -al  $BASE_PATH"/"*.tar
        echo "====================================================================================================================="
        echo " "
        ;;


 [2])   echo " DIFFER binary to $NEW_VER dir "
	 	
	 chk_new_bin_dir=`ls -d $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER  |wc -l`

		if [ $chk_new_bin_dir = "0" ]
		then
			echo "*************************************************************"
			echo " error!! unkown dir!! $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			echo "*************************************************************"
			pkg_1st_menu $*
		else
			echo "*************************************************************"
			echo " dir check ok!"
			echo "*************************************************************"

			echo ""
			echo "*************************************************************"
 			echo " Source tar file info is $SOURCE_PATH"/"$SOURCE_BIN_FILE "
			echo " Target Path Info is $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			echo "*************************************************************"
			echo ""
			echo "*************************************************************"
			echo " If it has incorrect values.  You can type Next Step "
			echo "*************************************************************"
			echo ""
			echo "--------------------------------------------------------------------------------------------------------"
        	echo; echo -n " if you use setting env value just press Enter! Otherwise input source path ex : /home/ktfsmc/home/ ? "
        read s_path
			echo ""
			echo "--------------------------------------------------------------------------------------------------------"
        	echo; echo -n " if you use setting env value just press Enter! Otherwise input source file name ex :  kt_r460_bin.tar ?"
        read s_file

		fi

	if [ -z $s_path ] 
	then
		
		if [ -z $s_file ] 
		then
			
			echo "cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"
			cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER
			echo " locale info is" 
			pwd
			echo " differ $SOURCE_PATH"/"$SOURCE_BIN_FILE  with $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			chk $*
		
			tar tvf $SOURCE_PATH"/"$SOURCE_BIN_FILE | awk '{print $3, $8}' > $LOG_PATH"/".pkg_src_info.log
			tar tvf $SOURCE_PATH"/"$SOURCE_BIN_FILE | awk '{print $8}' | xargs ls -al | awk '{print $5, $9}' > $LOG_PATH"/".pkg_tag_info.log
			echo " differ result....................."
			diff_result=`diff $LOG_PATH"/".pkg_src_info.log $LOG_PATH"/".pkg_tag_info.log`
			rm -f $LOG_PATH"/".pkg_src_info.log  $LOG_PATH"/".pkg_tag_info.log

			if [ -z $diff_result ]
			then

				echo ""
				echo " differ result is O.K "

			else
				
				echo ""
				echo " differ result error!!!"
				echo " $diff_result "

			fi

		else

			echo "cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"
			cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER
			echo " locale info is" 
			pwd
			echo " differ $SOURCE_PATH"/"$s_file with $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			chk $*
	
			tar tvf $SOURCE_PATH"/"$s_file | awk '{print $3, $8}' > $LOG_PATH"/".pkg_src_info.log
			tar tvf $SOURCE_PATH"/"$s_file | awk '{print $8}' | xargs ls -al | awk '{print $5, $9}' > $LOG_PATH"/".pkg_tag_info.log

			echo " differ result....................."
			diff_result=`diff $LOG_PATH"/".pkg_src_info.log $LOG_PATH"/".pkg_tag_info.log`
			di_re=`echo "$diff_result" |wc -l`
			rm -f $LOG_PATH"/".pkg_src_info.log  $LOG_PATH"/".pkg_tag_info.log

			if [ 0 -eq $di_re ]
			then

				echo ""
				echo " differ result is O.K "

			else
				
				echo ""
				echo " differ result error!!! -------------------------------"
				echo "-------------------------------------------------------"
				echo " $diff_result "
				echo "-------------------------------------------------------"
				echo "------------------------------- differ result error!!!"

			fi

		fi

	else
		if [ -z $s_file ] 
		then

			echo "cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"
			cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER
			echo " locale info is" 
			pwd
			echo " differ $s_path"/"$SOURCE_BIN_FILE  with $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			chk $*

			tar tvf $SOURCE_PATH"/"$SOURCE_BIN_FILE | awk '{print $3, $8}' > $LOG_PATH"/".pkg_src_info.log
			tar tvf $SOURCE_PATH"/"$SOURCE_BIN_FILE | awk '{print $8}' | xargs ls -al | awk '{print $5, $9}' > $LOG_PATH"/".pkg_tag_info.log
				
			echo " differ result....................."
			diff_result=`diff $LOG_PATH"/".pkg_src_info.log $LOG_PATH"/".pkg_tag_info.log`
			di_re=`echo "$diff_result" |wc -l`
			rm -f $LOG_PATH"/".pkg_src_info.log  $LOG_PATH"/".pkg_tag_info.log

			if [ 0 -eq $di_re ]
			then

				echo ""
				echo " differ result is O.K "

			else
				
				echo ""
				echo " differ result error!!! -------------------------------"
				echo "-------------------------------------------------------"
				echo " $diff_result "
				echo "-------------------------------------------------------"
				echo "------------------------------- differ result error!!!"

			fi

		else 
			
			echo "cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER"
			cd $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER
			echo " locale info is" 
			pwd
			echo " differ $s_path"/"$s_file  with $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			chk $*

			tar tvf $SOURCE_PATH"/"$s_file | awk '{print $3, $8}' > $LOG_PATH"/".pkg_src_info.log
			tar tvf $SOURCE_PATH"/"$s_file | awk '{print $8}' | xargs ls -al | awk '{print $5, $9}' > $LOG_PATH"/".pkg_tag_info.log

			echo " differ result....................."
			diff_result=`diff $LOG_PATH"/".pkg_src_info.log $LOG_PATH"/".pkg_tag_info.log`
			di_re=`echo "$diff_result" |wc -l`
			rm -f $LOG_PATH"/".pkg_src_info.log  $LOG_PATH"/".pkg_tag_info.log

			if [ 0 -eq $di_re ]
			then

				echo ""
				echo " differ result is O.K "

			else
				
				echo ""
				echo " differ result error!!! -------------------------------"
				echo "-------------------------------------------------------"
				echo " $diff_result "
				echo "-------------------------------------------------------"
				echo "------------------------------- differ result error!!!"


			fi
		fi
	fi
	;;



 [3])   echo " DIFFER library $NEW_VER dir "
	 chk_new_lib_dir=`ls -d $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER  |wc -l`

		if [ $chk_new_lib_dir = "0" ]
		then
			echo "*************************************************************"
			echo " error!! unkown dir!! $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
			echo "*************************************************************"
			pkg_1st_menu $*
		else
			echo "*************************************************************"
			echo " dir check ok!"
			echo "*************************************************************"

			echo ""
			echo "*************************************************************"
 			echo " Source tar file info is $SOURCE_PATH"/"$SOURCE_LIB_FILE "
			echo " Target Path Info is $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
			echo "*************************************************************"
			echo ""
			echo "*************************************************************"
			echo " If it has incorrect values.  You can type Next Step "
			echo "*************************************************************"
			echo ""
			echo "--------------------------------------------------------------------------------------------------------"
        	echo; echo -n " if you use setting env value just press Enter! Otherwise input source path ex : /home/ktfsmc/home/ ?"
        read s_path
			echo ""
			echo "--------------------------------------------------------------------------------------------------------"
       		echo; echo -n " if you use setting env value just press Enter! Otherwise input source tar file name ex : kt_r460_lib.tar ?"
        read s_file

		fi

		if [ -z $s_path ] 
		then
			
			if [ -z $s_file ] 
			then
			
				echo "cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"
				cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER
				echo " locale info is" 
				pwd
				echo " differ $SOURCE_PATH"/"$SOURCE_LIB_FILE with  $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
				chk $*

				tar tvf $SOURCE_PATH"/"$SOURCE_LIB_FILE | awk '{print $3, $8}' |grep sl > $LOG_PATH"/".pkg_src_info.log
				tar tvf $SOURCE_PATH"/"$SOURCE_LIB_FILE | awk '{print $8}' |grep sl | xargs ls -al | awk '{print $5, $9}' > $LOG_PATH"/".pkg_tag_info.log

				echo " differ result....................."
				diff_result=`diff $LOG_PATH"/".pkg_src_info.log $LOG_PATH"/".pkg_tag_info.log`
				di_re=`echo "$diff_result" |wc -l`
				rm -f $LOG_PATH"/".pkg_src_info.log  $LOG_PATH"/".pkg_tag_info.log

					if [ 0 -eq $di_re ]
					then

						echo ""
						echo " differ result is O.K "

					else
				
						echo ""
						echo " differ result error!!! -------------------------------"
						echo "-------------------------------------------------------"
						echo " $diff_result "
						echo "-------------------------------------------------------"
						echo "------------------------------- differ result error!!!"
					fi

				else

					echo "cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"
					cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER
					echo " locale info is" 
					pwd
					echo " differ $SOURCE_PATH"/"$s_file with $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
					chk $*

					tar tvf $SOURCE_PATH"/"$s_file | awk '{print $3, $8}' |grep sl > $LOG_PATH"/".pkg_src_info.log
					tar tvf $SOURCE_PATH"/"$s_file | awk '{print $8}' |grep sl | xargs ls -al | awk '{print $5, $9}' > $LOG_PATH"/".pkg_tag_info.log

					echo " differ result....................."
					diff_result=`diff $LOG_PATH"/".pkg_src_info.log $LOG_PATH"/".pkg_tag_info.log`
					di_re=`echo "$diff_result" |wc -l`
					rm -f $LOG_PATH"/".pkg_src_info.log  $LOG_PATH"/".pkg_tag_info.log

					if [ 0 -eq $di_re ]
					then

						echo ""
						echo " differ result is O.K "

					else
				
						echo ""
						echo " differ result error!!! -------------------------------"
						echo "-------------------------------------------------------"
						echo " $diff_result "
						echo "-------------------------------------------------------"
						echo "------------------------------- differ result error!!!"
					fi

				fi

		else 
			
			if [ -z $s_file ] 
			then
				echo "cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"
				cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER
				echo " locale info is" 
				pwd
				echo " differ $s_path"/"$SOURCE_LIB_FILE  with $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
				chk $*
	
				tar tvf $SOURCE_PATH"/"$SOURCE_LIB_FILE | awk '{print $3, $8}'  |grep sl > $LOG_PATH"/".pkg_src_info.log
				tar tvf $SOURCE_PATH"/"$SOURCE_LIB_FILE | awk '{print $8}' |grep sl | xargs ls -al | awk '{print $5, $9}' > $LOG_PATH"/".pkg_tag_info.log
				echo " differ result....................."
				diff_result=`diff $LOG_PATH"/".pkg_src_info.log $LOG_PATH"/".pkg_tag_info.log`
				di_re=`echo "$diff_result" |wc -l`
				rm -f $LOG_PATH"/".pkg_src_info.log  $LOG_PATH"/".pkg_tag_info.log

				if [ 0 -eq $di_re ]
				then

					echo ""
					echo " differ result is O.K "

				else
				
					echo ""
					echo " differ result error!!! -------------------------------"
					echo "-------------------------------------------------------"
					echo " $diff_result "
					echo "-------------------------------------------------------"
					echo "------------------------------- differ result error!!!"
				fi

			else
			
				echo "cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER"
				cd $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER
				echo " locale info is" 
				pwd
				echo " differ $s_path"/"$s_file  with $BASE_PATH"/"$SITE_INFO"_LIB_"$NEW_VER "
				chk $*

				tar tvf $SOURCE_PATH"/"$s_file | awk '{print $3, $8}' | grep sl > $LOG_PATH"/".pkg_src_info.log
				tar tvf $SOURCE_PATH"/"$s_file | awk '{print $8}' | grep sl | xargs ls -al | awk '{print $5, $9}' > $LOG_PATH"/".pkg_tag_info.log
	
				echo " differ result....................."
				diff_result=`diff $LOG_PATH"/".pkg_src_info.log $LOG_PATH"/".pkg_tag_info.log`
				di_re=`echo "$diff_result" |wc -l`
				rm -f $LOG_PATH"/".pkg_src_info.log  $LOG_PATH"/".pkg_tag_info.log

				if [ 0 -eq $di_re ]
				then

					echo ""
					echo " differ result is O.K "

				else
				
					echo ""
					echo " differ result error!!! -------------------------------"
					echo "-------------------------------------------------------"
					echo " $diff_result "
					echo "-------------------------------------------------------"
					echo "------------------------------- differ result error!!!"
				fi
			fi
		fi
	;;



 [4])   echo " GOTO UPPER MENU"

         pkg_1st_menu $*

         ;;

 [0])   echo " EXIT"
        echo " Good Bye!!!" 
		;;

   *)     echo " Typing error.... please retry.... "
          echo " "
          echo " "
          ;;
     esac


 done

}


function pkg_chg_link()
{


until [ "$answer" = 0 ]
do
echo " "
echo " "
echo "             ==================================================================="
echo "             Welcome to UANGEL SMSC APPLICATION PKG UPGRADE TOOL for change link !!!"
echo "             ==================================================================="
echo " "
echo " ******************************************************************"
echo "  [1])   SHOW PKG LINK INFO "
echo "  [2])   Change of each old link to new link "
echo "  [3])   Change of all old link to new link(not ready)  "
echo "  [4])   GOTO UPPER MENU " 
echo "  [0])   EXIT" 
echo " ******************************************************************"
echo; echo  -n "Do you want to run CMD (0-4)? "
read answer

case "$answer" in
 [1])   echo " SHOW NEW(BACKUPED) PKG PATH INFO"
        echo "====================================================================================================================="
		echo " locate " 
		pwd
		echo ""
        echo " ls -al  $BASE_PATH ......  bin lib " 
        echo "---------------------------------------------------------------------------------------------------------------------"
        ls -al  $BASE_PATH |grep -v data2 |grep -v data3 |grep "\->"
        echo "====================================================================================================================="
        echo " "
        ;;




 [2])   echo " Relink  to $NEW_VER  dir "
	 	echo " ex : xxx will be linked to xxx ->  $BASE_PATH"/"$SITE_INFO"_XXX_"$NEW_VER "
	 	
	 chk_new_bin_dir=`ls -d $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER  |wc -l`

		if [ $chk_new_bin_dir = "0" ]
		then
			echo "*************************************************************"
			echo " error!! unkown dir!! $BASE_PATH"/"$SITE_INFO"_BIN_"$NEW_VER "
			echo "*************************************************************"
			pkg_1st_menu $*
		else
			echo "*************************************************************"
			echo " dir check ok!"
			echo "*************************************************************"

			echo ""
			echo "*************************************************************"
 			echo " BASE link path  info is $BASE_PATH "
			echo " New PKG Path Info is $SITE_INFO"_xxx_"$NEW_VER "
			echo "*************************************************************"
			echo ""
			echo "*************************************************************"
			echo " If it has incorrect values.  You can type Next Step "
			echo "*************************************************************"
			echo ""
			echo "--------------------------------------------------------------------------------------------------------"
        	echo; echo -n " if you use setting env value just press Enter! Otherwise input link type (bin,lib,data,config,build) ex : bin ?"
        read s_info
			echo ""
			echo "--------------------------------------------------------------------------------------------------------"
        	echo; echo -n " if you use setting env value just press Enter! Otherwise input source path ex : /home/ktfsmc/home/ ? "
        read s_path
			echo ""
			echo "--------------------------------------------------------------------------------------------------------"
        	echo; echo -n " if you use setting env value just press Enter! Otherwise input source path version name ex :  KT_BIN_R4.6.0 ?"
        read s_file

		fi

if [ -z $s_info ]
then
	echo "no input data!!! please link type "
	pkg_chg_link $*

elif [ [ $s_info != bin -o  $s_info != lib ] -o [ $s_info != data -o $s_info != config ] -o  [ $s_info != build ] ]
then

	echo "input value is $s_info"
	echo "invalid input data!!! please link type (only type : bin lib data config build) "
	pkg_chg_link $*

else


	if [ -z $s_path ] 
	then
		
		if [ -z $s_file ] 
		then
			
			echo "cd $BASE_PATH"
			cd $BASE_PATH
			echo " old link info is" 
			ll $s_info |grep "\->"

			echo " this link will be remove"
			chk $*

			rm -f $s_info
			echo " removed $s_info link"

			ln -s $BASE_PATH"/"$SITE_INFO"_`echo "$s_info"|tr a-z A-Z `_"$NEW_VER $s_info
			echo "relink $BASE_PATH"/"$SITE_INFO"_`echo "$s_info"|tr a-z A-Z `_"$NEW_VER $s_info "

		

		else

			echo "cd $BASE_PATH"
			cd $BASE_PATH
			echo " old link info is" 
			ll $s_info |grep "\->"

			echo " this link will be remove"
			chk $*

			rm -f $s_info
			echo " removed $s_info link"

			ln -s $BASE_PATH"/"$s_file $s_info
			echo "relink $BASE_PATH"/"$s_file $s_info "

		


		fi

	else
		if [ -z $s_file ] 
		then

			
			echo "cd $s_path"
			cd $s_path
			echo " old link info is" 
			ll $s_info |grep "\->"

			echo " this link will be remove"
			chk $*

			rm -f $s_info
			echo " removed $s_info link"

			echo "cd $s_path"
			cd $s_path
			ln -s $s_path"/"$SITE_INFO"_`echo "$s_info"|tr a-z A-Z `_"$NEW_VER $s_info
			echo "relink $s_path"/"$SITE_INFO"_`echo "$s_info"|tr a-z A-Z `_"$NEW_VER $s_info "

		

		else 
			

			echo "cd $s_path"
			cd $s_path
			echo " old link info is" 
			ll bin |grep "\->"

			echo " this link will be remove"
			chk $*

			rm -f bin
			echo " removed bin link"

			echo "cd $s_path"
			cd $s_path
			ln -s $s_path"/"$s_file bin
			echo "relink $s_path"/"$s_file bin "

		

		fi
	fi
fi
	;;



     esac


 done

}



##########################################################################################################################
# CALL and RUN PART


pkg_intro $*
pkg_1st_menu $*



