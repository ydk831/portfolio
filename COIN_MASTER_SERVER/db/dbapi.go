package db

import (
	"common"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"message"
	"sort"
	"strconv"
	"strings"
	"time"

	// mysql driver

	_ "github.com/go-sql-driver/mysql"
	// IOS IAP
	_ "github.com/awa/go-iap/appstore"

	// xsql mysql driver
	"github.com/jmoiron/sqlx"
)

func MyErrCode(err error) int {
	if strings.Contains(err.Error(), "no rows in result set") {
		return 4040
	} else {
		x := strings.Fields(err.Error())
		res, e := strconv.Atoi(strings.TrimSuffix(x[1], ":"))
		if e != nil {
			log.Println("MyErrCode Parse Error : ", e)
			return -1
		}
		return res
	}
}

func GetUpdateError(err error) string {
	if err == nil {
		return "No Rows Affected"
	} else {
		return err.Error()
	}
}

func RechargeSP(uid int, mydb *sqlx.DB) (error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	type RechargeSPData struct {
		Spin                 int16     `db:"Spin"`
		MaxSpin              int16     `db:"MaxSpin"`
		LastSpinRechargeTime time.Time `db:"LastSpinRechargeTime"`
		Now                  time.Time `db:"Now"`
	}

	userinfo := RechargeSPData{}
	err := tx.Get(&userinfo, "select Spin, MaxSpin, LastSpinRechargeTime, Now() Now FROM User WHERE AccountUID = ?;", uid)
	if err != nil {
		log.Println("Select User Fail : ", err.Error())
		log.Println("Query : select Spin, MaxSpin, LastSpinRechargeTime, Now() Now FROM User WHERE AccountUID = ", uid)
		err = fmt.Errorf("RechargeSP : Find user info fail.")
		return err, 500
	}

	if userinfo.Spin < userinfo.MaxSpin {
		userinfo.Spin, userinfo.LastSpinRechargeTime = common.CalcRechargeSP(userinfo.Spin, userinfo.MaxSpin, userinfo.Now, userinfo.LastSpinRechargeTime)
		result, err := tx.Exec("UPDATE User SET Spin = ?, UpdateDate = NOW(6), LastSpinRechargeTime = ?  WHERE AccountUID=?",
			userinfo.Spin, userinfo.LastSpinRechargeTime, uid)
		if err != nil {
			log.Println("[DBAPI:RechargeSP] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET Spin = ", userinfo.Spin, ", UpdateDate = NOW(6)",
				", LastSpinRechargeTime = ", userinfo.LastSpinRechargeTime, "  WHERE AccountUID=", uid)
			err = fmt.Errorf("Update user info fail.")
			return err, 500
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:RechargeSP] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET Spin = ", userinfo.Spin, ", UpdateDate = NOW(6) ",
				", LastSpinRechargeTime = ", userinfo.LastSpinRechargeTime, "  WHERE AccountUID=", uid)
			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("Update user info fail.")
				return err, 400
			} else {
				err = fmt.Errorf("Update user info fail.")
				return err, 500
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:FindUserForAttack] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return err, 500
	}

	return nil, 200
}

func FindUserForAttack(req message.AttackRequest, mydb *sqlx.DB) (message.AttackResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	rsp := message.AttackResponse{}

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	// 타겟 선정 : 친구도 건물이 없으면 안됨
	// Todo. 불타는 건물이 n개 이하여야 함.

	BuildingCount := 0
	err := tx.Get(&BuildingCount, "SELECT count(A.AccountUID) FROM UserTile A INNER JOIN User B ON A.AccountUID = B.AccountUID WHERE A.TileStatus = 2 AND A.AccountUID = ? ", req.TargetID)
	if err != nil {
		log.Println("[DBAPI:FindUserForAttack] Select " + err.Error())
		log.Println("Query : SELECT count(A.AccountUID) FROM UserTile A INNER JOIN User B ON A.AccountUID = B.AccountUID WHERE A.TileStatus = 2 AND A.AccountUID = ", req.TargetID)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Exist Data in UserTile & User Table.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Find tile info fail.")
			return rsp, err, 500
		}
	}

	if BuildingCount <= 0 {
		log.Println("[DBAPI:FindUserForAttack] Target has no Building.(", strconv.Itoa(BuildingCount), ")")
		err = fmt.Errorf("Target has No Building..")
		return rsp, err, 400
	}

	// 타겟 선정 : 부서진 건물 수가 ChapterInfo 의 기준보다 적으면 안됨
	// 먼저 유저의 부서진 건물 수를 찾고
	DestroyedBuildingCount := 0
	err = tx.Get(&DestroyedBuildingCount, "SELECT count(TileStatus) FROM UserTile WHERE TileStatus = 3 AND AccountUID = ? ", req.TargetID)
	if err != nil {
		log.Println("[DBAPI:FindUserForAttack] Select " + err.Error())
		log.Println("Query : SELECT count(TileStatus) FROM UserTile WHERE TileStatus = 3 AND AccountUID =", req.TargetID)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Exist Data in UserTile Table.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Find tile info fail.")
			return rsp, err, 500
		}
	}
	// 챕터 별 최대 부숴진 건물 기준을 구한다음
	BreakBuildingMax := 0
	err = tx.Get(&BreakBuildingMax, "select BreakBuildingMax from DataDB.ChapterInfo where ChapterNumber = (select CurrentChapterIDX from User where AccountUID = ?)", req.TargetID)
	if err != nil {
		log.Println("[DBAPI:FindUserForAttack] Select " + err.Error())
		log.Println("Query : select BreakBuildingMax from DataDB.ChapterInfo where ChapterNumber = (select CurrentChapterIDX from User where AccountUID = ", req.TargetID, ")")

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Exist Data in DataDB.ChapterInfo Table.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Find tile info fail.")
			return rsp, err, 500
		}
	}
	// 비교
	if DestroyedBuildingCount >= BreakBuildingMax {
		log.Println("[DBAPI:FindUserForAttack] Target Building are Max Destroyed.(Des ", strconv.Itoa(DestroyedBuildingCount), ", Max ", strconv.Itoa(BreakBuildingMax), ")")
		err = fmt.Errorf("Target Building were Max Destroyed..")
		return rsp, err, 400
	}

	userinfo := message.AttackRspUserInfo{}

	err = tx.Get(&userinfo, "SELECT User.AccountUID, User.Name, Gold, Spin, Shield, CurrentChapterIDX, Keyid FROM User INNER JOIN Account ON User.AccountUID = Account.AccountUID WHERE User.AccountUID = ?", req.TargetID)
	if err != nil {
		log.Println("[DBAPI:FindUserForAttack] Select User Fail : " + err.Error())
		log.Println("Query : SELECT User.AccountUID, User.Name, Gold, Spin, Shield, CurrentChapterIDX, Keyid FROM User INNER JOIN Account ON User.AccountUID = Account.AccountUID WHERE User.AccountUID = ", req.TargetID)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Exist Data in User & Account Table.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Find user info fail.")
			return rsp, err, 500
		}
	}

	tileinfo := []message.AttackRspTileInfo{}

	err = tx.Select(&tileinfo, "SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ?", req.TargetID)
	if err != nil {
		log.Println("[DBAPI:FindUserForAttack] Select UserTile Fail : " + err.Error())
		log.Println("Query : SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ", req.TargetID)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Exist Data in UserTile Table.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Find tile info fail")
			return rsp, err, 500
		}
	}

	rsp.MultipleValue = req.MultipleValue
	rsp.UserInfo = userinfo
	rsp.TileInfo = tileinfo

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:FindUserForAttack] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, nil, 200
}

func FindRandUserForAttack(req message.AttackRequest, mydb *sqlx.DB) (message.AttackResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	rsp := message.AttackResponse{}

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	// 타겟선정 : 건물이 하나라도 없으면 안됨
	// ToDo. 불타는 건물이 n개 이하여야 함.

	var AccountUID uint

	// 대상 리스트를 뽑고, 리스트 순회하면서 조건에 부합하는지 검사
	type chk struct {
		UID     int `db:"AccountUID"`
		Chapter int `db:"CurrentChapterIDX"`
	}

	var targetlist []chk
	// 대상 리스트 추출 여기서 에러면 그냥 끝
	// 대상 리스트는 건물이 하나 이상 존재하고 로그인이 24시간 이내이며 나 자신은 제외
	err := tx.Select(&targetlist, "SELECT distinct A.AccountUID, B.CurrentChapterIDX FROM UserTile A "+
		"INNER JOIN User B ON A.AccountUID = B.AccountUID "+
		"WHERE (select count(TileStatus) from UserTile where AccountUID=A.AccountUID AND TileStatus=2) >= 1 "+
		"AND B.LoginDate >= SUBDATE(NOW(), INTERVAL 1 DAY) "+
		"AND A.AccountUID != ? "+
		"ORDER BY RAND()", req.ID)
	if err != nil {
		log.Println("[DBAPI:FindRandUserForAttack] Select " + err.Error())
		log.Println("Query : SELECT distinct A.AccountUID, B.CurrentChapterIDX FROM UserTile A INNER JOIN User B ",
			"ON A.AccountUID = B.AccountUID ",
			"WHERE (select count(TileStatus) from UserTile where AccountUID=A.AccountUID",
			"AND TileStatus=2) >= 1 ",
			"AND B.LoginDate >= SUBDATE(NOW(), INTERVAL 1 DAY) ",
			"AND A.AccountUID != ", req.ID,
			"ORDER BY RAND()")
		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Exist Data in UserTile & User Table.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Find tile info fail.")
			return rsp, err, 500
		}
	}

	// 대상 리스트를 순회 하면서 조건 검사
	for _, target := range targetlist {
		// 부서진 건물 수 추출
		BreakBuildingCnt := -1
		err := tx.Get(&BreakBuildingCnt, "select count(TileStatus) from UserTile Where AccountUID=? AND TileStatus=3", target.UID)
		if err != nil { // 무슨 에러든 간에 다음 리스트로 간다
			log.Println("[DBAPI:FindRandUserForAttack] Select " + err.Error())
			log.Println("Query : select count(TileStatus) from UserTile Where AccountUID=", target.UID, " AND TileStatus=3")
			log.Println("[DBAPI:FindRandUserForAttack] Select Next Target. Cause Select error.")
		} else { // 부서진 건물수가 추출되면
			// 현 챕터에서 부서진 건물수가 최대인지 판단
			istarget := false
			err := tx.Get(&istarget, "select if(BreakBuildingMax>?,true,false) from DataDB.ChapterInfo WHERE ChapterNumber = ?", BreakBuildingCnt, target.Chapter)
			if err != nil { // 무슨 에러든 간에 다음 리스트로 간다
				log.Println("[DBAPI:FindRandUserForAttack] Select " + err.Error())
				log.Println("Query : select if(BreakBuildingMax<=", BreakBuildingCnt, ",true,false) from DataDB.ChapterInfo WHERE ChapterNumber = ", target.Chapter)
				log.Println("[DBAPI:FindRandUserForAttack] Select Next Target. Cause Select error.")
			} else { // 비교결과 판단
				if istarget == true { // true 면 더 부셔도 됨
					AccountUID = uint(target.UID)
					log.Println("[DBAPI:FindRandUserForAttack] Find Target(", AccountUID, "). BreakBuilding(", BreakBuildingCnt, ")")
					break
				} else { // false 면 다음 타겟 확인
					log.Println("[DBAPI:FindRandUserForAttack] Select Next Target. Cause Target condition is not ready.")
				}
			}
		}
	}

	if AccountUID == 0 {
		err = fmt.Errorf("There is no target for attack random match.")
		return rsp, err, 204
	}

	userinfo := message.AttackRspUserInfo{}
	err = tx.Get(&userinfo, "SELECT User.AccountUID, User.Name, Gold, Spin, Shield, CurrentChapterIDX, Keyid FROM User INNER JOIN Account ON User.AccountUID = Account.AccountUID WHERE User.AccountUID = ?", AccountUID)
	if err != nil {
		log.Println("[DBAPI:FindRandUserForAttack] Select User Fail : " + err.Error())
		log.Println("Query : SELECT User.AccountUID, User.Name, Gold, Spin, Shield, CurrentChapterIDX, Keyid FROM User INNER JOIN Account ON User.AccountUID = Account.AccountUID WHERE User.AccountUID = ", AccountUID)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Exist Data in User & Account Table.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Find user info fail.")
			return rsp, err, 500
		}
	}

	tileinfo := []message.AttackRspTileInfo{}
	err = tx.Select(&tileinfo, "SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ?", AccountUID)
	if err != nil {
		log.Println("[DBAPI:FindRandUserForAttack] Select UserTile Fail : " + err.Error())
		log.Println("Query : SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ", AccountUID)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Exist Data in UserTile Table.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Find tile info fail.")
			return rsp, err, 500
		}
	}

	rsp.MultipleValue = req.MultipleValue
	rsp.UserInfo = userinfo
	rsp.TileInfo = tileinfo

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:FindRandUserForAttack] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, nil, 200
}

//////
func FindUserForRaid(req message.AttackRequest, mydb *sqlx.DB) (message.AttackResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	rsp := message.AttackResponse{}

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	// 타겟 선정  DataDB.ChapterInfo 테이블에서 유저 챕터 별 레이드 제한 골드 확인 후 진행
	var LimitGold int
	err := tx.Get(&LimitGold, "SELECT RaidTargetGoldLimit FROM DataDB.ChapterInfo "+
		"WHERE (SELECT CurrentChapterIDX FROM User WHERE AccountUID = ?)", req.TargetID)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:FindUserForRaid] Select User Fail : " + err.Error())
			log.Println("Query : SELECT RaidTargetGoldLimit FROM DataDB.ChapterInfo "+
				"WHERE (SELECT CurrentChapterIDX FROM User WHERE AccountUID = ?)", req.TargetID)
			err = fmt.Errorf("Can't Find Client Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:FindUserForRaid] Select User Fail : " + err.Error())
			log.Println("Query : SELECT RaidTargetGoldLimit FROM DataDB.ChapterInfo "+
				"WHERE (SELECT CurrentChapterIDX FROM User WHERE AccountUID = ?)", req.TargetID)
			err = fmt.Errorf("Find user info fail.")
			return rsp, err, 500
		}
	}

	userinfo := message.AttackRspUserInfo{}

	err = tx.Get(&userinfo, "SELECT User.AccountUID, User.Name, Gold, Spin, Shield, CurrentChapterIDX, Keyid FROM User INNER JOIN Account ON User.AccountUID = Account.AccountUID WHERE User.AccountUID = ? AND Gold >= ?", req.TargetID, LimitGold)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:FindUserForRaid] Select User Fail : " + err.Error())
			log.Println("Query : SELECT User.AccountUID, User.Name, Gold, Spin, Shield, CurrentChapterIDX, Keyid FROM User INNER JOIN Account ON User.AccountUID = Account.AccountUID WHERE User.AccountUID = ", req.TargetID, "AND Gold >= ", LimitGold)
			err = fmt.Errorf("Can't Find Client Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:FindUserForRaid] Select User Fail : " + err.Error())
			log.Println("Query : SELECT User.AccountUID, User.Name, Gold, Spin, Shield, CurrentChapterIDX, Keyid FROM User INNER JOIN Account ON User.AccountUID = Account.AccountUID WHERE User.AccountUID = ", req.TargetID, "AND Gold >= ", LimitGold)
			err = fmt.Errorf("Find user info fail.")
			return rsp, err, 500
		}
	}

	tileinfo := []message.AttackRspTileInfo{}

	err = tx.Select(&tileinfo, "SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ?", req.TargetID)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:FindUserForRaid] Select UserTile Fail : " + err.Error())
			log.Println("Query : SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ", req.TargetID)
			err = fmt.Errorf("Can't Find Client Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:FindUserForRaid] Select UserTile Fail : " + err.Error())
			log.Println("Query : SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ", req.TargetID)
			err = fmt.Errorf("Find tile info fail")
			return rsp, err, 500
		}
	}

	rsp.MultipleValue = req.MultipleValue
	rsp.UserInfo = userinfo
	rsp.TileInfo = tileinfo

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:FindUserForRaid] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, nil, 200
}

func FindRandUserForRaid(req message.AttackRequest, mydb *sqlx.DB) (message.AttackResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	rsp := message.AttackResponse{}

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	var AccountUID int
	/*
		err := tx.Get(&AccountUID, "SELECT AccountUID FROM User WHERE LoginDate >= SUBDATE(NOW(), INTERVAL 1 DAY) AND AccountUID != ? AND Gold >= 30000 ORDER BY RAND() LIMIT 1", req.ID)
		if err != nil {
			if MyErrCode(err) == 4040 {
				log.Println("[DBAPI:FindRandUserForRaid] Select " + err.Error())
				log.Println("Query : SELECT AccountUID FROM User WHERE LoginDate >= SUBDATE(NOW(), INTERVAL 1 DAY) AND AccountUID != ", req.ID, " AND Gold >= 30000 ORDER BY RAND() LIMIT 1")
				err = fmt.Errorf("Can't Find Client Information in Server.")
				return rsp, err, 400
			} else {
				log.Println("[DBAPI:FindRandUserForRaid] Select " + err.Error())
				log.Println("Query : SELECT AccountUID FROM User WHERE LoginDate >= SUBDATE(NOW(), INTERVAL 1 DAY) AND AccountUID != ", req.ID, " AND Gold >= 30000 ORDER BY RAND() LIMIT 1")
				err = fmt.Errorf("Find tile info fail.")
				return rsp, err, 500
			}
		}
	*/

	type chk struct {
		UID     int `db:"AccountUID"`
		Chapter int `db:"CurrentChapterIDX"`
		Gold    int `db:"Gold"`
	}

	var targetlist []chk
	// 타겟 선정 : 유저 리스트를 추출하고 해당 유저의 골드보유량과 챕터별 제한 골드 비교하여 최종 선발
	err := tx.Select(&targetlist, "SELECT AccountUID, CurrentChapterIDX, Gold FROM User WHERE LoginDate >= SUBDATE(NOW(), INTERVAL 1 DAY) AND AccountUID != ? ORDER BY RAND()", req.ID)
	if err != nil { // 리스트 추출 실패는 에러처리
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:FindRandUserForRaid] Select " + err.Error())
			log.Println("Query : SELECT AccountUID, CurrentChapterIDX, Gold FROM User WHERE LoginDate >= SUBDATE(NOW(), INTERVAL 1 DAY) AND AccountUID != ", req.ID, " ORDER BY RAND()")
			err = fmt.Errorf("Can't Find Client Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:FindRandUserForRaid] Select " + err.Error())
			log.Println("Query : SELECT AccountUID, CurrentChapterIDX, Gold FROM User WHERE LoginDate >= SUBDATE(NOW(), INTERVAL 1 DAY) AND AccountUID != ", req.ID, " ORDER BY RAND()")
			err = fmt.Errorf("Find User info fail.")
			return rsp, err, 500
		}
	}

	for _, target := range targetlist {
		istarget := false
		err := tx.Get(&istarget, "select "+
			"if((select RaidTargetGoldLimit from DataDB.ChapterInfo "+
			" where ChapterNumber = (select CurrentChapterIDX From User where AccountUID=1)) "+
			"<= (select Gold From User Where AccountUID=?),true,false)", target.UID)
		if err != nil { // 무슨 에러든 간에 다음 타겟으로 넘어감
			log.Println("[DBAPI:FindRandUserForRaid] Select " + err.Error())
			log.Println("Query : if((select RaidTargetGoldLimit from DataDB.ChapterInfo "+
				" where ChapterNumber = (select CurrentChapterIDX From User where AccountUID=1)) "+
				"<= (select Gold From User Where AccountUID=", target.UID, "),true,false)")
			log.Println("[DBAPI:FindRandUserForRaid] Select Next Target. Cause Select error.")
		} else {
			if istarget == true { // 추출된 리스트에서 보유골드가 제한 골드보다 많다면 타겟선정
				AccountUID = target.UID
				log.Println("[DBAPI:FindRandUserForRaid] Find Target(", AccountUID, "). Gold(", target.Gold, ")")
				break
			} else { // 아니면 다른 타겟 검색
				log.Println("[DBAPI:FindRandUserForRaid] Select Next Target. Cause Target condition is not ready.")
			}
		}
	}

	if AccountUID == 0 {
		err = fmt.Errorf("There is no target for raid random match.")
		return rsp, err, 400
	}

	userinfo := message.AttackRspUserInfo{}
	err = tx.Get(&userinfo, "SELECT User.AccountUID, User.Name, Gold, Spin, Shield, CurrentChapterIDX, Keyid FROM User INNER JOIN Account ON User.AccountUID = Account.AccountUID WHERE User.AccountUID = ?", AccountUID)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:FindRandUserForRaid] Select User Fail : " + err.Error())
			log.Println("Query : SELECT User.AccountUID, User.Name, Gold, Spin, Shield, CurrentChapterIDX, Keyid FROM User INNER JOIN Account ON User.AccountUID = Account.AccountUID WHERE User.AccountUID = ", AccountUID)
			err = fmt.Errorf("Can't Find Client Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:FindRandUserForRaid] Select User Fail : " + err.Error())
			log.Println("Query : SELECT User.AccountUID, User.Name, Gold, Spin, Shield, CurrentChapterIDX, Keyid FROM User INNER JOIN Account ON User.AccountUID = Account.AccountUID WHERE User.AccountUID = ", AccountUID)
			err = fmt.Errorf("Find user info fail.")
			return rsp, err, 500
		}
	}

	tileinfo := []message.AttackRspTileInfo{}
	err = tx.Select(&tileinfo, "SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ?", AccountUID)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:FindRandUserForRaid] Select UserTile Fail : " + err.Error())
			log.Println("Query : SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ", AccountUID)
			err = fmt.Errorf("Can't Find Client Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:FindRandUserForRaid] Select UserTile Fail : " + err.Error())
			log.Println("Query : SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ", AccountUID)
			err = fmt.Errorf("Find tile info fail.")
			return rsp, err, 500
		}
	}

	rsp.MultipleValue = req.MultipleValue
	rsp.UserInfo = userinfo
	rsp.TileInfo = tileinfo

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:FindRandUserForRaid] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, nil, 200
}

//////

func UpdateAttackResult(req message.AttackResultRequest, mydb *sqlx.DB) (message.AttackResultResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()
	rsp := message.AttackResultResponse{}
	rsp.IsUpdate = false

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	// dummy data update
	if req.TargetID == 0 {
		result, err := tx.Exec("UPDATE User SET gold=gold+?, UpdateDate=NOW() WHERE AccountUID=?",
			req.StealGold, req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateAttackResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET gold=gold+", req.StealGold, ", UpdateDate=NOW() WHERE AccountUID=", req.ID)
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 500
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateAttackResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET gold=gold+", req.StealGold, ", UpdateDate=NOW() WHERE AccountUID=", req.ID)

			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("No Rows Update User Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 400
			}
		}
		rsp.IsUpdate = true
		return rsp, err, 200
	}

	if req.IsSuccess == true {
		// 공격 성공 시, 타겟 골드 미차감, 공격 골드 증가, 타겟 타일 파괴, 뉴스
		// 공격 골드 증가
		result, err := tx.Exec("UPDATE User SET gold=gold+?, UpdateDate=NOW() WHERE AccountUID=?",
			req.StealGold, req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateAttackResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET gold=gold+", req.StealGold, ", UpdateDate=NOW() WHERE AccountUID=", req.ID)
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 500
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateAttackResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET gold=gold+", req.StealGold, ", UpdateDate=NOW() WHERE AccountUID=", req.ID)

			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("No Rows Update User Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 500
			}
		}

		// 타겟 타일 파괴
		result, err = tx.Exec("UPDATE UserTile SET TileStatus=3, UpdateDate=NOW() WHERE AccountUID=? AND TileIdx=?",
			req.TargetID, req.AttackTileIdx)
		if err != nil {
			log.Println("[DBAPI:UpdateAttackResult] Update UserTile Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE UserTile SET TileStatus=3, UpdateDate=NOW() WHERE AccountUID=", req.TargetID, " AND TileIdx=", req.AttackTileIdx)
			err = fmt.Errorf("Update tile info fail.")
			return rsp, err, 500
		}
		rowsAffected, err = result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateAttackResult] Update UserTile Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE UserTile SET TileStatus=3, UpdateDate=NOW() WHERE AccountUID=", req.TargetID, " AND TileIdx=", req.AttackTileIdx)
			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("No Rows Update UserTile Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Update tile info fail.")
				return rsp, err, 500
			}
		}

		// 뉴스
		//var name string
		var attacker message.AttackerInfo
		err = tx.Get(&attacker, "SELECT Name, Gender, Type, Keyid FROM Account WHERE AccountUID = ?", req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateAttackResult] Select Account Fail : " + err.Error())
			log.Println("Query : SELECT Name, Gender, Type, Keyid FROM Account WHERE AccountUID = ", req.ID)

			if MyErrCode(err) == 4040 {
				err = fmt.Errorf("Not Exist Data in Account Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Find account info fail.")
				return rsp, err, 500
			}
		}

		result, err = tx.Exec("Insert Into News(TargetUser, AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid) Values(?,?,?,?,?,NOW(),?,?,?)",
			req.TargetID, req.ID, attacker.Name, req.IsSuccess, req.StealGold, attacker.Gender, attacker.AppType, attacker.Keyid)
		if err != nil {
			log.Println("[DBAPI:UpdateAttackResult] Insert News Fail : " + GetUpdateError(err))
			log.Println("Query : Insert Into News(TargetUser, AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid) Values(", req.TargetID, ",", req.ID, ",", attacker.Name, ",", req.IsSuccess, ",", req.StealGold, ",NOW(),", attacker.Gender, ",", attacker.AppType, ",", attacker.Keyid, ")")
			err = fmt.Errorf("Update news fail.")
			return rsp, err, 500
		}
		rowsAffected, err = result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateAttackResult] Insert News Fail : " + GetUpdateError(err))
			log.Println("Query : Insert Into News(TargetUser, AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid) Values(", req.TargetID, ",", req.ID, ",", attacker.Name, ",", req.IsSuccess, ",", req.StealGold, ",NOW(),", attacker.Gender, ",", attacker.AppType, ",", attacker.Keyid, ")")
			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("No Rows Update News Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Update news fail.")
				return rsp, err, 500
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Println("[DBAPI:UpdateAttackResult] Commit Fail : " + err.Error())
			err = fmt.Errorf("DB Commit Error.")
			return rsp, err, 500
		}

		rsp.IsUpdate = true
		return rsp, err, 200
	} else {
		// 공격 방어 시, 타겟 실드 차감, 타겟 골드 미차감, 공격 골드 증가, 뉴스
		// 타겟 실드 차감
		result, err := tx.Exec("UPDATE User SET shield=shield-1, UpdateDate=NOW() WHERE AccountUID=?", req.TargetID)
		if err != nil {
			log.Println("[DBAPI:UpdateAttackResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET shield=shield-1, UpdateDate=NOW() WHERE AccountUID=", req.TargetID)
			if MyErrCode(err) == 1690 {
				err = fmt.Errorf("Target dosen't have shield. but attack is fail..")
				return rsp, err, 400
			} else {
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 500
			}
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateAttackResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET shield=shield-1, UpdateDate=NOW() WHERE AccountUID=", req.TargetID)
			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("No Rows Update User Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 500
			}
		}

		// 공격 골드 증가
		result, err = tx.Exec("UPDATE User SET gold=gold+?, UpdateDate=NOW() WHERE AccountUID=?",
			req.StealGold, req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateAttackResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET gold=gold+", req.StealGold, ", UpdateDate=NOW() WHERE AccountUID=", req.ID)
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 500
		}
		rowsAffected, err = result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateAttackResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET gold=gold+", req.StealGold, ", UpdateDate=NOW() WHERE AccountUID=", req.ID)
			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("No Rows Update User Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 500
			}
		}

		// 뉴스
		//var name string
		var attacker message.AttackerInfo
		err = tx.Get(&attacker, "SELECT Name, Gender, Type, Keyid FROM Account WHERE AccountUID = ?", req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateAttackResult] Select Account Fail : " + err.Error())
			log.Println("Query : SELECT Name, Gender, Type, Keyid FROM Account WHERE AccountUID = ", req.ID)

			if MyErrCode(err) == 4040 {
				err = fmt.Errorf("Not Exist Data in Account Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Find account info fail.")
				return rsp, err, 500
			}
		}

		result, err = tx.Exec("Insert Into News(TargetUser, AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid) Values(?,?,?,?,?,NOW(),?,?,?)",
			req.TargetID, req.ID, attacker.Name, req.IsSuccess, req.StealGold, attacker.Gender, attacker.AppType, attacker.Keyid)
		if err != nil {
			log.Println("[DBAPI:UpdateAttackResult] Insert News Fail : " + GetUpdateError(err))
			log.Println("Query : Insert Into News(TargetUser, AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid) Values(", req.TargetID, ",", req.ID, ",", attacker.Name, ",", req.IsSuccess, ",", req.StealGold, ",NOW(),", attacker.Gender, ",", attacker.AppType, ",", attacker.Keyid, ")")
			err = fmt.Errorf("Update news fail.")
			return rsp, err, 500
		}
		rowsAffected, err = result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateAttackResult] Insert News Fail : " + GetUpdateError(err))
			log.Println("Query : Insert Into News(TargetUser, AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid) Values(", req.TargetID, ",", req.ID, ",", attacker.Name, ",", req.IsSuccess, ",", req.StealGold, ",NOW(),", attacker.Gender, ",", attacker.AppType, ",", attacker.Keyid, ")")
			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("No Rows Update News Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Update news fail.")
				return rsp, err, 500
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Println("[DBAPI:UpdateAttackResult] Commit Fail : " + err.Error())
			err = fmt.Errorf("DB Commit Error.")
			return rsp, err, 500
		}

		rsp.IsUpdate = true
		return rsp, err, 200
	}
}

func UpdateRaidResult(req message.RaidResultRequest, mydb *sqlx.DB) (message.AttackResultResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()
	rsp := message.AttackResultResponse{}
	rsp.IsUpdate = false

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	// MultipleValue 가 추가되어서 정합성 처리 해줌
	if req.MultipleValue <= 0 {
		req.MultipleValue = 1
	}

	// dummy data update
	if req.TargetID == 0 {
		result, err := tx.Exec("UPDATE User SET gold=gold+?, UpdateDate=NOW() WHERE AccountUID=?",
			req.StealGold*uint32(req.MultipleValue), req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateRaidResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET gold=gold+", req.StealGold*uint32(req.MultipleValue), ", UpdateDate=NOW() WHERE AccountUID=?", req.ID)
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 500
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateRaidResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET gold=gold+", req.StealGold*uint32(req.MultipleValue), ", UpdateDate=NOW() WHERE AccountUID=?", req.ID)
			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 400
			} else {
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 500
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Println("[DBAPI:UpdateRaidResult] Commit Fail : " + err.Error())
			err = fmt.Errorf("DB Commit Error.")
			return rsp, err, 500
		}

		rsp.IsUpdate = true
		return rsp, err, 200
	}

	// 공격 성공 시, 타겟 골드 차감, 공격 골드 증가, 뉴스
	// 타겟 골드 차감
	result, err := tx.Exec("UPDATE User SET gold=gold-?, UpdateDate=NOW() WHERE AccountUID=?",
		req.StealGold, req.TargetID)
	if err != nil {
		if MyErrCode(err) == 1690 { // 차감후 음수가 된다면 0으로 세팅
			result, err = tx.Exec("UPDATE User SET gold=0, UpdateDate=NOW() WHERE AccountUID=?",
				req.TargetID)
			if err != nil {
				log.Println("[DBAPI:UpdateRaidResult] Update User Fail : " + GetUpdateError(err))
				log.Println("Query : UPDATE User SET gold=0, UpdateDate=NOW() WHERE AccountUID=", req.TargetID)
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 500
			}
			rowsAffected, err := result.RowsAffected()
			if err != nil || rowsAffected == 0 {
				log.Println("[DBAPI:UpdateRaidResult] Update User Fail : " + GetUpdateError(err))
				log.Println("Query : UPDATE User SET gold=0, UpdateDate=NOW() WHERE AccountUID=", req.TargetID)
				if rowsAffected == 0 {
					tx.Rollback()
					err = fmt.Errorf("Update user info fail.")
					return rsp, err, 400
				} else {
					err = fmt.Errorf("Update user info fail.")
					return rsp, err, 500
				}
			}
		} else {
			log.Println("[DBAPI:UpdateRaidResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET gold=gold-", req.StealGold, ", UpdateDate=NOW() WHERE AccountUID=", req.TargetID)
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 500
		}
	} else {
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateRaidResult] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET gold=gold-", req.StealGold, ", UpdateDate=NOW() WHERE AccountUID=", req.TargetID)
			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 400
			} else {
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 500
			}
		}
	}

	// 공격 골드 증가
	result, err = tx.Exec("UPDATE User SET gold=gold+?, UpdateDate=NOW() WHERE AccountUID=?",
		req.StealGold*uint32(req.MultipleValue), req.ID)
	if err != nil {
		log.Println("[DBAPI:UpdateRaidResult] Update User Fail : " + err.Error())
		log.Println("Query : UPDATE User SET gold=gold+", req.StealGold*uint32(req.MultipleValue), ", UpdateDate=NOW() WHERE AccountUID=", req.TargetID)
		err = fmt.Errorf("Update user info fail.")
		return rsp, err, 500
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateRaidResult] Update User Fail : " + GetUpdateError(err))
		log.Println("Query : UPDATE User SET gold=gold+", req.StealGold*uint32(req.MultipleValue), ", UpdateDate=NOW() WHERE AccountUID=", req.TargetID)
		if rowsAffected == 0 {
			tx.Rollback()
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 400
		}
		err = fmt.Errorf("Update user info fail.")
		return rsp, err, 500
	}

	// 뉴스
	//var name string
	var attacker message.AttackerInfo
	err = tx.Get(&attacker, "SELECT Name, Gender, Type, Keyid FROM Account WHERE AccountUID = ?", req.ID)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:UpdateRaidResult] Select Account Fail : " + err.Error())
			log.Println("Query : SELECT Name, Gender, Type, Keyid FROM Account WHERE AccountUID = ", req.ID)
			err = fmt.Errorf("Can't Find Client Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:UpdateRaidResult] Select Account Fail : " + err.Error())
			log.Println("Query : SELECT Name, Gender, Type, Keyid FROM Account WHERE AccountUID = ", req.ID)
			err = fmt.Errorf("Find account info fail.")
			return rsp, err, 500
		}
	}
	result, err = tx.Exec("Insert Into News(TargetUser, AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid) Values(?,?,?,2,?,NOW(),?,?,?)",
		req.TargetID, req.ID, attacker.Name, req.StealGold, attacker.Gender, attacker.AppType, attacker.Keyid)
	if err != nil {
		log.Println("[DBAPI:UpdateRaidResult] Insert News Fail : " + GetUpdateError(err))
		log.Println("Query : Insert Into News(TargetUser, AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid) Values(", req.TargetID, ",", req.ID, ",", attacker.Name, ",2,", req.StealGold, ",NOW(),", attacker.Gender, ",", attacker.AppType, ",", attacker.Keyid, ")")
		err = fmt.Errorf("Update news fail.")
		return rsp, err, 500
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateRaidResult] Insert News Fail : " + GetUpdateError(err))
		log.Println("Query : Insert Into News(TargetUser, AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid) Values(", req.TargetID, ",", req.ID, ",", attacker.Name, ",2,", req.StealGold, ",NOW(),", attacker.Gender, ",", attacker.AppType, ",", attacker.Keyid, ")")
		if rowsAffected == 0 {
			tx.Rollback()
			err = fmt.Errorf("Update news fail.")
			return rsp, err, 400
		} else {
			err = fmt.Errorf("Update news fail.")
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:UpdateRaidResult] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	rsp.IsUpdate = true
	return rsp, err, 200
}

func GetNews(id interface{}, mydb *sqlx.DB) (message.News, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	rsp := message.News{}

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	newsinfo := []message.NewsInfo{}

	err := tx.Select(&newsinfo, "SELECT AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid FROM News WHERE TargetUser = ? ORDER BY AttackTime DESC LIMIT 10", id)
	//err := tx.Select(&newsinfo, "SELECT AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType FROM News WHERE TargetUser = ? ORDER BY AttackTime DESC LIMIT 10", id)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:GetNews] Select News Fail : " + err.Error())
			log.Println("Query : SELECT AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid FROM News WHERE TargetUser = ", id, " ORDER BY AttackTime DESC LIMIT 10")
			err = fmt.Errorf("Can't Find Client Information in Server.")
			return rsp, err, 200
		} else {
			log.Println("[DBAPI:GetNews] Select News Fail : " + err.Error())
			log.Println("Query : SELECT AttackUser, AttackName, Result, StolenGold, AttackTime, AttackGender, AttackAppType, AttackKeyid FROM News WHERE TargetUser = ", id, " ORDER BY AttackTime DESC LIMIT 10")
			err = fmt.Errorf("Find news fail.")
			return rsp, err, 500
		}
	}

	rsp.MyNews = newsinfo

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:GetNews] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, nil, 200
}

func UpdateAdvReward(req message.AdvRewardRequest, mydb *sqlx.DB) (message.AdvRewardResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()
	rsp := message.AdvRewardResponse{}
	rsp.IsUpdate = false

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	type RewardTime struct {
		Update time.Time `db:"AdvUpdateDate"`
		Now    time.Time `db:"Now()"`
	}

	rewardtime := RewardTime{}
	err := tx.Get(&rewardtime, "SELECT AdvUpdateDate, Now() FROM User WHERE AccountUID = ?", req.ID)
	if err != nil {
		log.Println("[DBAPI:UpdateAdvReward] Select User Fail : " + err.Error())
		log.Println("Query : SELECT AdvUpdateDate, Now() FROM User WHERE AccountUID =", req.ID)
		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Exist Data in User Table.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Find user info fail.")
			return rsp, err, 500
		}
	}

	if rewardtime.Update.Day() < rewardtime.Now.Day() {
		result, err := tx.Exec("UPDATE User SET AdvGoldRewardCnt=0, AdvSpinRewardCnt=0, UpdateDate=NOW(), AdvUpdateDate=NOW() WHERE AccountUID=?", req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateAdvReward] Update User Fail : " + err.Error())
			log.Println("Query : UPDATE User SET AdvGoldRewardCnt=0, AdvSpinRewardCnt=0, UpdateDate=NOW(), AdvUpdateDate=NOW() WHERE AccountUID=", req.ID)
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 500
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateAdvReward] Update User Fail(RowsAffected) : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET AdvGoldRewardCnt=0, AdvSpinRewardCnt=0, UpdateDate=NOW(), AdvUpdateDate=NOW() WHERE AccountUID=", req.ID)
			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("No Rows Update User Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 500
			}
		}
	}

	if req.RewardType == "gold" {
		var advCnt int
		err := tx.Get(&advCnt, "SELECT AdvGoldRewardCnt FROM User WHERE AccountUID = ?", req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateAdvReward] Select User Fail : " + err.Error())
			log.Println("Query : SELECT AdvGoldRewardCnt FROM User WHERE AccountUID =", req.ID)

			if MyErrCode(err) == 4040 {
				err = fmt.Errorf("Not Exist Data in User Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Find user info fail.")
				return rsp, err, 500
			}
		}

		if advCnt >= 15 {
			log.Println("Query : SELECT AdvGoldRewardCnt FROM User WHERE AccountUID = ", req.ID)
			err = fmt.Errorf("Daily Gold AdvReward Count Exceeded")
			log.Println("[DBAPI:UpdateAdvReward] " + err.Error() + ". " + strconv.Itoa(advCnt))
			return rsp, err, 400
		}

		result, err := tx.Exec("UPDATE User SET Gold=Gold+?, AdvGoldRewardCnt=AdvGoldRewardCnt+1, UpdateDate=NOW() WHERE AccountUID=?",
			req.Reward, req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateAdvReward] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET Gold=Gold+", req.Reward, ", AdvGoldRewardCnt=AdvGoldRewardCnt+1, UpdateDate=NOW() WHERE AccountUID=", req.ID)
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 500
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateAdvReward] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET Gold=Gold+", req.Reward, ", AdvGoldRewardCnt=AdvGoldRewardCnt+1, UpdateDate=NOW() WHERE AccountUID=", req.ID)

			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("No Rows Update User Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 500
			}
		}
	} else if req.RewardType == "spin" {
		var advCnt int
		err := tx.Get(&advCnt, "SELECT AdvSpinRewardCnt FROM User WHERE AccountUID = ? ", req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateAdvReward] Select User Fail : " + err.Error())
			log.Println("Query : SELECT AdvSpinRewardCnt FROM User WHERE AccountUID = ", req.ID)

			if MyErrCode(err) == 4040 {
				err = fmt.Errorf("Not Exist Data in User Table.")
				return rsp, err, 400
			} else {
				err = fmt.Errorf("Find user info fail.")
				return rsp, err, 500
			}
		}

		if advCnt >= 15 {
			log.Println("Query : SELECT AdvSpinRewardCnt FROM User WHERE AccountUID = ", req.ID)
			err = fmt.Errorf("Daily Spin AdvReward Count Exceeded")
			log.Println("[DBAPI:UpdateAdvReward] " + err.Error() + ". " + strconv.Itoa(advCnt))
			return rsp, err, 400
		}

		result, err := tx.Exec("UPDATE User SET Spin=Spin+?, AdvSpinRewardCnt=AdvSpinRewardCnt+1, UpdateDate=NOW() WHERE AccountUID=?",
			req.Reward, req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateAdvReward] Update User Fail : " + err.Error())
			log.Println("Query : UPDATE User SET Spin=Spin+", req.Reward, ", AdvSpinRewardCnt=AdvSpinRewardCnt+1, UpdateDate=NOW() WHERE AccountUID=", req.ID)
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 500
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateAdvReward] Update User Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE User SET Spin=Spin+", req.Reward, ", AdvSpinRewardCnt=AdvSpinRewardCnt+1, UpdateDate=NOW() WHERE AccountUID=", req.ID)

			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("No Rows Update User Table.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Update user info fail.")
				return rsp, err, 500
			}
		}
	} else {
		log.Println("[DBAPI:UpdateAdvReward] Not Allow RewardType Parameter Value")
		err := fmt.Errorf("Invalid RewardType Parameter value.")
		return rsp, err, 400
	}

	err = tx.Get(&rsp, "SELECT AdvGoldRewardCnt, AdvSpinRewardCnt FROM User WHERE AccountUID = ? ", req.ID)
	if err != nil {
		log.Println("[DBAPI:UpdateAdvReward] Select User Fail : " + err.Error())
		log.Println("Query : SELECT AdvGoldRewardCnt, AdvSpinRewardCnt FROM User WHERE AccountUID = ", req.ID)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Exist Data in User Table.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Find user info fail.")
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:UpdateAdvReward] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	rsp.IsUpdate = true
	return rsp, err, 200
}

func FindTarget(keyidlist []string, mydb *sqlx.DB) (message.TargetResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	rsp := message.TargetResponse{}

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	for {
		var ranidx int
		if len(keyidlist) == 0 {
			log.Println("[DBAPI:FindTarget] keyidlist size is 0")
			err := fmt.Errorf("There is no exist target")
			return rsp, err, 400
		} else {
			randsrc := rand.NewSource(time.Now().UnixNano())
			r := rand.New(randsrc)
			ranidx = r.Intn(len(keyidlist))
		}

		keyid := keyidlist[ranidx]
		log.Println("[DBAPI:FindTarget] Target keyid = ", keyidlist[ranidx])

		// 타겟 선정 : DataDB.ChapterInfo 테이블에서 유저 챕터 별 레이드 제한 골드 확인 후 진행
		var LimitGold int
		err := tx.Get(&LimitGold, "SELECT RaidTargetGoldLimit FROM DataDB.ChapterInfo "+
			"WHERE ChapTerNumber = "+
			"(SELECT CurrentChapterIDX FROM User INNER JOIN Account "+
			"ON User.AccountUID = Account.AccountUID "+
			"WHERE Keyid=?)", keyid)
		if err != nil {
			log.Println("[DBAPI:FindTarget] Select User Fail : " + err.Error())
			log.Println("Query : SELECT RaidTargetGoldLimit FROM DataDB.ChapterInfo "+
				"WHERE ChapTerNumber = "+
				"(SELECT CurrentChapterIDX FROM User INNER JOIN Account "+
				"ON User.AccountUID = Account.AccountUID "+
				"WHERE Keyid=", keyid, ")")
			if len(keyidlist) == 1 {
				err = fmt.Errorf("Can't Find Client Information in Server.")
				return rsp, err, 400
			}
			keyidlist = keyidlist[:ranidx+copy(keyidlist[ranidx:], keyidlist[ranidx+1:])]
			log.Println("[DBAPI:FindTarget] Select Next Target. Cause Target condition is not ready.")
		}

		/*
			err := tx.Get(&rsp, "SELECT Account.AccountUID, Gender, Type, Keyid, Gold, User.Name FROM Account INNER JOIN User ON Account.AccountUID = User.AccountUID WHERE Keyid = ? AND Gold >= 30000", keyid)
			if err != nil {
				if MyErrCode(err) == 4040 {
					log.Println("[DBAPI:FindTarget] Select " + err.Error())
					log.Println("Query : SELECT Account.AccountUID, Gender, Type, Keyid, Gold, User.Name FROM Account INNER JOIN User ON Account.AccountUID = User.AccountUID WHERE Keyid = ", keyid, " AND Gold >= 30000")
					if len(keyidlist) == 1 {
						err = fmt.Errorf("Can't Find Client Information in Server.")
						return rsp, err, 400
					}
					keyidlist = keyidlist[:ranidx+copy(keyidlist[ranidx:], keyidlist[ranidx+1:])]
					continue
				} else {
					log.Println("[DBAPI:FindTarget] Select " + err.Error())
					log.Println("Query : SELECT Account.AccountUID, Gender, Type, Keyid, Gold, User.Name FROM Account INNER JOIN User ON Account.AccountUID = User.AccountUID WHERE Keyid = ", keyid, " AND Gold >= 30000")
					err = fmt.Errorf("Find account info fail.")
					return rsp, err, 500
				}
			}
		*/
		err = tx.Get(&rsp, "SELECT Account.AccountUID, Gender, Type, Keyid, Gold, User.Name FROM Account INNER JOIN User ON Account.AccountUID = User.AccountUID WHERE Keyid = ? AND Gold >= ?", keyid, LimitGold)
		if err != nil {
			if MyErrCode(err) == 4040 {
				log.Println("[DBAPI:FindTarget] Select " + err.Error())
				log.Println("Query : SELECT Account.AccountUID, Gender, Type, Keyid, Gold, User.Name FROM Account INNER JOIN User ON Account.AccountUID = User.AccountUID WHERE Keyid = ", keyid, " AND Gold >= ", LimitGold)
				if len(keyidlist) == 1 {
					err = fmt.Errorf("Can't Find Client Information in Server.")
					return rsp, err, 400
				}
				keyidlist = keyidlist[:ranidx+copy(keyidlist[ranidx:], keyidlist[ranidx+1:])]
				continue
			} else {
				log.Println("[DBAPI:FindTarget] Select " + err.Error())
				log.Println("Query : SELECT Account.AccountUID, Gender, Type, Keyid, Gold, User.Name FROM Account INNER JOIN User ON Account.AccountUID = User.AccountUID WHERE Keyid = ", keyid, " AND Gold >= ", LimitGold)
				err = fmt.Errorf("Find account info fail.")
				return rsp, err, 500
			}
		}
		return rsp, nil, 200

		// 타겟이 건물이 없으면 안됨
		// 2019.09.04 건물제한 없는 거로 변경
		/*
			tileinfo := message.AttackRspTileInfo{}
			err = tx.Get(&tileinfo, "SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ? AND TileStatus = 2", rsp.TargetID)
			if err != nil {
				if MyErrCode(err) == 4040 {
					log.Println("[DBAPI:FindUserForAttack] Not found user tile info. select retry")
					log.Println("Query : SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ", rsp.TargetID, "AND TileStatus = 2")
					if len(keyidlist) == 1 {
						err = fmt.Errorf("There is no Available Target.")
						return rsp, err, 400
					}
					keyidlist = keyidlist[:ranidx+copy(keyidlist[ranidx:], keyidlist[ranidx+1:])]
					continue
				} else {
					log.Println("[DBAPI:FindUserForAttack] Select UserTile Fail : " + err.Error())
					log.Println("Query : SELECT TileIdx, TileStatus, TileChargeTime FROM UserTile WHERE AccountUID = ", rsp.TargetID, "AND TileStatus = 2")
					err = fmt.Errorf("Find tile info fail")
					return rsp, err, 500
				}
			} else {
				return rsp, nil, 200
			}
		*/
	}
}

func FindRandTarget(id int, mydb *sqlx.DB) (message.TargetResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	rsp := message.TargetResponse{}
	rsplist := []message.TargetResponse{}

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	// 2019.09.04 건물제한 없는 거로 변경
	err := tx.Select(&rsplist, "Select A.AccountUID, B.Name, A.Gender, A.Type, B.Gold, A.Keyid From Account A, User B where A.AccountUID != ? AND (A.AccountUID = B.AccountUID) AND B.LoginDate >= subdate(now(), INTERVAL 1 DAY) order by rand()", id)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:FindRandTarget] Select " + err.Error())
			log.Println("Query : select A.AccountUID, B.Name, A.Gender, A.Type From Account A, User B, UserTile C where A.AccountUID != ", id, "(A.AccountUID = B.AccountUID) AND B.LoginDate >= subdate(now(), INTERVAL 1 DAY) order by rand()")
			err = fmt.Errorf("Can't Find Client Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:FindRandTarget] Select " + err.Error())
			log.Println("Query : select A.AccountUID, B.Name, A.Gender, A.Type From Account A, User B, UserTile C where A.AccountUID != ", id, "(A.AccountUID = B.AccountUID) AND B.LoginDate >= subdate(now(), INTERVAL 1 DAY) order by rand()")
			err = fmt.Errorf("Find account info fail.")
			return rsp, err, 500
		}
	}

	for _, rsp := range rsplist {
		istarget := false
		err := tx.Get(&istarget, "select "+
			"if((select RaidTargetGoldLimit from DataDB.ChapterInfo "+
			"where ChapterNumber = (select CurrentChapterIDX From User where AccountUID=?)) "+
			"<= (select Gold From User Where AccountUID=?),true,false)", rsp.TargetID, rsp.TargetID)
		if err != nil {
			log.Println("[DBAPI:FindRandTarget] Select User Fail : " + err.Error())
			log.Println("Query : select "+
				"if((select RaidTargetGoldLimit from DataDB.ChapterInfo "+
				"where ChapterNumber = (select CurrentChapterIDX From User where AccountUID=", rsp.TargetID, ")) "+
				"<= (select Gold From User Where AccountUID=", rsp.TargetID, "),true,false)")
			log.Println("[DBAPI:FindRandTarget] Select Next Target. Cause Select error.")
		} else {
			if istarget == true {
				log.Println("[DBAPI:FindRandTarget] Find Target(", rsp, ")")
				return rsp, err, 200
			} else {
				log.Println("[DBAPI:FindRandTarget] Select Next Target. Cause Target condition is not ready.")
			}
		}
	}

	err = fmt.Errorf("There is no target for random match.")
	return rsp, err, 400
}

func FindFriend(keyidlist []string, mydb *sqlx.DB) (message.FriendResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	rsp := message.FriendResponse{}
	friendinfo := message.FriendInfo{}
	//friendinfolist := []message.FriendInfo{}
	var err error

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	for _, keyid := range keyidlist {
		err = tx.Get(&friendinfo, "SELECT AccountUID, Name, Gender, Type, Keyid FROM Account WHERE KeyID=?", keyid)
		if err != nil {
			if MyErrCode(err) == 4040 {
				// 없는 친구 ID에 대해서는 skip 함
				/*
					log.Println("[DBAPI:FindFriend] Select no rows...")
					log.Println("Query : SELECT AccountUID, Name, Gender, Type, Keyid FROM Account WHERE KeyID=", keyid)
					err = fmt.Errorf("Can't Find Client Information in Server.")
					return rsp, err, 400
				*/
				log.Println("There is no exist KeyID(", keyid, "). Skip!")
				continue
			} else {
				log.Println("[DBAPI:FindFriend] Select " + err.Error())
				log.Println("Query : SELECT AccountUID, Name, Gender, Type, Keyid FROM Account WHERE KeyID=", keyid)
				err = fmt.Errorf("Find account info fail.")
				return rsp, err, 500
			}
		} else {
			rsp.MyFriends = append(rsp.MyFriends, friendinfo)
		}
	}

	return rsp, err, 200
}

func SaveReceiptGoogle(id int, receipt message.Receipt, mydb *sqlx.DB) error {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	var err error
	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	payload, err := json.Marshal(receipt.Payload)
	if err != nil {
		log.Println("[DBApi:SaveReceiptGoogle] payload convert error. ", err)
	}

	price := receipt.Payload.SkuDetails.PriceAmountMicros / 1000000
	result, err := tx.Exec("Insert Into Receipt(AccountUID, Store, TransactionID, ProductID, Price, PriceCurrencyCode, Description, Payload, UpdateDate) Values (?,?,?,?,?,?,?,?,NOW())",
		id, receipt.Store, receipt.TransactionID, receipt.Payload.SkuDetails.ProductId, price, receipt.Payload.SkuDetails.PriceCurrencyCode, receipt.Payload.SkuDetails.Description, string(payload))
	if err != nil {
		log.Println("[DBAPI:SaveReceiptGoogle] Insert Receipt Fail : " + err.Error())
		log.Println("Query : Insert Into Receipt(AccountUID, Store, TransactionID, ProductID, Price, PriceCurrencyCode, Description, Payload, UpdateDate) Values (", id, ",", receipt.Store, ",", receipt.TransactionID, ",", receipt.Payload.SkuDetails.ProductId, ",", price, ",", receipt.Payload.SkuDetails.PriceCurrencyCode, ",", receipt.Payload.SkuDetails.Description, ",", string(payload), ",NOW())")
		err = fmt.Errorf("Update Receipt fail.")
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:SaveReceiptGoogle] Insert Receipt Fail : " + GetUpdateError(err))
		log.Println("Query : Insert Into Receipt(AccountUID, Store, TransactionID, ProductID, Price, PriceCurrencyCode, Description, Payload, UpdateDate) Values(", id, ",", receipt.Store, ",", receipt.TransactionID, ",", receipt.Payload.SkuDetails.ProductId, ",", price, ",", receipt.Payload.SkuDetails.PriceCurrencyCode, ",", receipt.Payload.SkuDetails.Description, ",", string(payload), ",NOW())")
		if rowsAffected == 0 {
			tx.Rollback()
			err = fmt.Errorf("Update Google Receipt fail.")
			return err
		} else {
			err = fmt.Errorf("Update Google Receipt fail.")
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:SaveReceiptGoogle] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return err
	}

	return err
}

func SaveReceiptApple(info *message.ReceiptIOS, mydb *sqlx.DB) error {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	var err error
	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	payload, err := json.Marshal(info)
	if err != nil {
		log.Println("[DBApi:SaveReceiptApple] receipt convert error. ", err)
	}

	result, err := tx.Exec("Insert Into Receipt(AccountUID, Store, TransactionID, ProductID, Price, PriceCurrencyCode, Description, Payload, UpdateDate) Values (?,?,?,?,?,?,?,?,NOW())",
		info.AccountUID, info.Store, info.TransactionID, info.ProductID, info.Price, info.PriceCurrencyCode, info.Description, string(payload))
	if err != nil {
		log.Println("[DBAPI:SaveReceiptApple] Insert Receipt Fail : " + err.Error())
		log.Println("Query : Insert Into Receipt(AccountUID, Store, TransactionID, ProductID, Price, PriceCurrencyCode, Description, Payload, UpdateDate) Values (", info.AccountUID, ",", info.Store, ",", info.TransactionID, ",", info.ProductID, ",", info.Price, ",", info.PriceCurrencyCode, ",", info.Description, ",", string(payload), ",NOW())")
		err = fmt.Errorf("Update Receipt fail.")
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:SaveReceiptApple] Insert Receipt Fail : " + GetUpdateError(err))
		log.Println("Query : Insert Into Receipt(AccountUID, Store, TransactionID, ProductID, Price, PriceCurrencyCode, Description, Payload, UpdateDate) Values (", info.AccountUID, ",", info.Store, ",", info.TransactionID, ",", info.ProductID, ",", info.Price, ",", info.PriceCurrencyCode, ",", info.Description, ",", string(payload), ",NOW())")
		if rowsAffected == 0 {
			tx.Rollback()
			err = fmt.Errorf("Update Apple Receipt fail.")
			return err
		} else {
			err = fmt.Errorf("Update Apple Receipt fail.")
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:SaveReceiptApple] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return err
	}

	return err
}

//func RewardShop(id int, receipt message.Receipt, mydb *sqlx.DB) (message.Item, error, int) {
func RewardShop(id int, receipt interface{}, mydb *sqlx.DB) (message.Item, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	var ProductId string
	switch receipt.(type) {
	case message.Receipt:
		ProductId = receipt.(message.Receipt).Payload.SkuDetails.ProductId
	case message.ReceiptIOS:
		ProductId = receipt.(message.ReceiptIOS).ProductID
	}

	var ItemInfo message.Item
	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	var UserChapterInfo int
	err := tx.Get(&UserChapterInfo, "select CurrentChapterIDX from User WHERE AccountUID = ?", id)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:RewardShop] Select no rows...")
			log.Println("Query : select CurrentChapterIDX from User WHERE AccountUID =", id)
			err = fmt.Errorf("Can't Find User Chapter Information in Server.")
			return ItemInfo, err, 400
		} else {
			log.Println("[DBAPI:RewardShop] Select " + err.Error())
			log.Println("Query : select CurrentChapterIDX from User WHERE AccountUID =", id)
			err = fmt.Errorf("Find User info fail.")
			return ItemInfo, err, 500
		}
	}

	//err = tx.Get(&ItemInfo, "select ShopRewardType, ShopRewardValue, PurchaseLimit from DataDB.ShopInfoDetail INNER JOIN DataDB.ShopInfo ON DataDB.ShopInfoDetail.GroupNumber = DataDB.ShopInfo.ShopDetailGroupNumber WHERE AOSInAppCode = ? AND ? BETWEEN ChapterMin AND ChapterMax", receipt.Payload.SkuDetails.ProductId, UserChapterInfo)
	err = tx.Get(&ItemInfo, "select ShopRewardType, ShopRewardValue, PurchaseLimit from DataDB.ShopInfoDetail INNER JOIN DataDB.ShopInfo ON DataDB.ShopInfoDetail.GroupNumber = DataDB.ShopInfo.ShopDetailGroupNumber WHERE AOSInAppCode = ? AND ? BETWEEN ChapterMin AND ChapterMax", ProductId, UserChapterInfo)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:RewardShop] Select no rows...")
			//log.Println("Query : select ShopRewardType, ShopRewardValue, PurchaseLimit from DataDB.ShopInfoDetail INNER JOIN DataDB.ShopInfo ON DataDB.ShopInfoDetail.GroupNumber = DataDB.ShopInfo.ShopDetailGroupNumber WHERE AOSInAppCode = ", receipt.Payload.SkuDetails.ProductId, " AND ", UserChapterInfo, "BETWEEN ChapterMin AND ChapterMax")
			log.Println("Query : select ShopRewardType, ShopRewardValue, PurchaseLimit from DataDB.ShopInfoDetail INNER JOIN DataDB.ShopInfo ON DataDB.ShopInfoDetail.GroupNumber = DataDB.ShopInfo.ShopDetailGroupNumber WHERE AOSInAppCode = ", ProductId, " AND ", UserChapterInfo, "BETWEEN ChapterMin AND ChapterMax")
			err = fmt.Errorf("Can't Find Shop Information in Server.")
			return ItemInfo, err, 400
		} else {
			log.Println("[DBAPI:RewardShop] Select " + err.Error())
			//log.Println("Query : select ShopRewardType, ShopRewardValue, PurchaseLimit from DataDB.ShopInfoDetail INNER JOIN DataDB.ShopInfo ON DataDB.ShopInfoDetail.GroupNumber = DataDB.ShopInfo.ShopDetailGroupNumber WHERE AOSInAppCode = ", receipt.Payload.SkuDetails.ProductId, " AND ", UserChapterInfo, "BETWEEN ChapterMin AND ChapterMax")
			log.Println("Query : select ShopRewardType, ShopRewardValue, PurchaseLimit from DataDB.ShopInfoDetail INNER JOIN DataDB.ShopInfo ON DataDB.ShopInfoDetail.GroupNumber = DataDB.ShopInfo.ShopDetailGroupNumber WHERE AOSInAppCode = ", ProductId, " AND ", UserChapterInfo, "BETWEEN ChapterMin AND ChapterMax")
			err = fmt.Errorf("Find account info fail.")
			return ItemInfo, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:SaveReceipt] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return ItemInfo, err, 500
	}

	return ItemInfo, err, 200
}

func UpdateSpinItem(id int, rewardvalue int, mydb *sqlx.DB) (message.ShopUpdateResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()
	rsp := message.ShopUpdateResponse{}
	rsp.IsUpdate = false

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	result, err := tx.Exec("UPDATE User SET Spin=Spin+?, UpdateDate=NOW() WHERE AccountUID=?", rewardvalue, id)
	if err != nil {
		log.Println("[DBAPI:UpdateSpinItem] Update User Fail : " + err.Error())
		log.Println("Query : UPDATE User SET Spin=Spin+", rewardvalue, ", UpdateDate=NOW() WHERE AccountUID=", id)
		err = fmt.Errorf("Update user info fail.")
		return rsp, err, 500
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateSpinItem] Update User Fail : " + GetUpdateError(err))
		log.Println("Query : UPDATE User SET Spin=Spin+", rewardvalue, ", UpdateDate=NOW() WHERE AccountUID=", id)
		if rowsAffected == 0 {
			tx.Rollback()
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 400
		} else {
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 500
		}
	}

	err = tx.Get(&rsp, "SELECT Gold, Spin FROM User WHERE AccountUID = ?", id)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:UpdateSpinItem] Select User Fail : " + err.Error())
			log.Println("Query : SELECT Gold, Spin FROM User WHERE AccountUID = ?", id)
			err = fmt.Errorf("Can't Find Your Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:UpdateSpinItem] Select User Fail : " + err.Error())
			log.Println("Query : SELECT Gold, Spin FROM User WHERE AccountUID = ?", id)
			err = fmt.Errorf("Find user info fail.")
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:UpdateSpinItem] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	rsp.IsUpdate = true
	return rsp, err, 200
}

func UpdateGoldItem(id int, rewardvalue int, mydb *sqlx.DB) (message.ShopUpdateResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()
	rsp := message.ShopUpdateResponse{}
	rsp.IsUpdate = false

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	result, err := tx.Exec("UPDATE User SET Gold=Gold+?, UpdateDate=NOW() WHERE AccountUID=?", rewardvalue, id)
	if err != nil {
		log.Println("[DBAPI:UpdateGoldItem] Update User Fail : " + err.Error())
		log.Println("Query : UPDATE User SET Gold=Gold+", rewardvalue, ", UpdateDate=NOW() WHERE AccountUID=", id)
		err = fmt.Errorf("Update user info fail.")
		return rsp, err, 500
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateGoldItem] Update User Fail : " + GetUpdateError(err))
		log.Println("Query : UPDATE User SET Gold=Gold+", rewardvalue, ", UpdateDate=NOW() WHERE AccountUID=", id)
		if rowsAffected == 0 {
			tx.Rollback()
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 400
		} else {
			err = fmt.Errorf("Update user info fail.")
			return rsp, err, 500
		}
	}

	err = tx.Get(&rsp, "SELECT Gold, Spin FROM User WHERE AccountUID = ?", id)
	if err != nil {
		log.Println("[DBAPI:UpdateGoldItem] Select User Fail : " + err.Error())
		log.Println("Query : SELECT Gold, Spin FROM User WHERE AccountUID = ?", id)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Can't Find Your Information in Server.")
			return rsp, err, 400
		} else {
			err = fmt.Errorf("Find user info fail.")
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:UpdateGoldItem] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	rsp.IsUpdate = true
	return rsp, err, 200
}
func CheckMonitorNotice(mydb *sqlx.DB) (int, string, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	type MonitorData struct {
		MonitorCheck int    `db:"NotiType"`
		MonitorMsg   string `db:"NotiMsg"`
	}

	Monitor := MonitorData{}

	err := tx.Get(&Monitor, "Select ifnull(NotiType,0) as NotiType, ifnull(NotiMsg,\"Not Exist\") as NotiMsg "+
		"From DataDB.Notice "+
		"where NotiType = 1 and NotiFlag = 1 and EndDate >= NOW()")
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:CheckMonitorNotice] Select no rows...")
			log.Println("Query : Select ifnull(NotiType,0) as NotiType, ifnull(NotiMsg,\"Not Exist\") as NotiMsg " +
				"From DataDB.Notice " +
				"where NotiType = 1 and NotiFlag = 1 and EndDate >= NOW()")
			return 0, Monitor.MonitorMsg, 200
		} else {
			log.Println("[DBAPI:CheckMonitorNotice] Select " + err.Error())
			log.Println("Query : Select ifnull(NotiType,0) as NotiType, ifnull(NotiMsg,\"Not Exist\") as NotiMsg " +
				"From DataDB.Notice " +
				"where NotiType = 1 and NotiFlag = 1 and EndDate >= NOW()")
			return 1, "Internal Server Error", 500
		}
	}
	return Monitor.MonitorCheck, Monitor.MonitorMsg, 503

	/* 점검 공지가 과연 복수개가 있을까?? 있으면 이거 쓰자
	isMonitoring := make([]int, 1)
	err := tx.Select(&isMonitoring, "select NotiType, NotiMsg From DataDB.Notice"+
		"where NotiType = 1 and NotiFlag = 1 and EndDate >= NOW()")
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:CheckMonitorNotice] Select no rows...")
			log.Println("Query : select NotiType From DataDB.Notice" +
				"where NotiType = 1 and NotiFlag = 1 and EndDate >= NOW()")
			return false, 200
		} else {
			log.Println("[DBAPI:CheckMonitorNotice] Select " + err.Error())
			log.Println("Query : select NotiType From DataDB.Notice" +
				"where NotiType = 1 and NotiFlag = 1 and EndDate >= NOW()")
			return true, 500
		}
	}

	for _, tmp := range isMonitoring {
		if tmp != 0 {
			return true, 204
		} else {
			return false, 200
		}
	}
	*/

}

func GetNotice(id interface{}, mydb *sqlx.DB) (message.NotiInfoResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	rsp := message.NotiInfoResponse{}
	rsplist := []message.NotiInfo{}

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	// NotiType이 점검공지가 아니고, NotiFlag가 0이 아니고 EndDate가 오늘보다 과거가 아니고
	// AccountUID가 NoticeRewardManage 테이블에 존재하지 않는 NotiID만 검색
	err := tx.Select(&rsplist, "Select "+
		"NotiID, NotiType, StartDate, EndDate, NotiMsg "+
		"From DataDB.Notice "+
		"Where NotiType != 1 And NotiFlag != 0 And EndDate >= Now() "+
		"And NotiID != ifnull("+
		"(Select NotiID From DataDB.NoticeRewardManage where AccountUID = ?),0"+
		")", id)
	if err != nil {
		log.Println("[DBAPI:GetNotice] Select Notice Fail : " + err.Error())
		if MyErrCode(err) == 4040 {
			log.Println("Query : Select "+
				"NotiID, NotiType, StartDate, EndDate, NotiMsg "+
				"From DataDB.Notice "+
				"Where NotiType != 1 And NotiFlag != 0 And EndDate >= Now() "+
				"And NotiID != ifnull("+
				"(Select NotiID From DataDB.NoticeRewardManage where AccountUID = ", id, "),0)")
			err = fmt.Errorf("Can't Find Notice Information in Server.")
			return rsp, err, 200
		} else {
			log.Println("Query : Select "+
				"NotiID, NotiType, StartDate, EndDate, NotiMsg "+
				"From DataDB.Notice "+
				"Where NotiType != 1 And NotiFlag != 0 And EndDate >= Now() "+
				"And NotiID != ifnull("+
				"(Select NotiID From DataDB.NoticeRewardManage where AccountUID = ", id, "),0)")
			err = fmt.Errorf("Find Notice fail.")
			return rsp, err, 500
		}
	}

	for _, tmp := range rsplist {
		notireward := message.NotiReward{}
		switch tmp.NotiType {
		case common.NOTI_MORNITORING: // 점검공지
			// 점검공지가 있으면 안됨..
			log.Println("Monitoring Notice Detected When Check Notices")
			err = fmt.Errorf("Find Notice fail.")
			return rsp, err, 500
		case common.NOTI_REWARD: // 보상공지
			// 1. 패킷에 담기 위한 보상정보 select
			err := tx.Get(&notireward, "Select RewardGold, RewardSpin "+
				"From DataDB.Notice Where NotiID = ?", tmp.NotiID)
			if err != nil {
				// 위에서 찾은게 여기 없는게 말이 안되니 로그는 남기되
				// 보상 못주는게 게임에 영향 가는건 아닌거 같아서 그냥 넘어감
				log.Println("[DBAPI:GetNotice] Select Notice Fail : " + err.Error())
				log.Println("Query : Select RewardGold, RewardSpin "+
					"From DataDB.Notice Where NotiID = ?", tmp.NotiID)
			}
			tmp.NotiRewardInfo = notireward

			// 2. 보상을 담고 나선 유저 계정정보 업데이트
			result, err := tx.Exec("UPDATE User SET Gold=Gold+?, Spin=Spin+?, UpdateDate=NOW() "+
				"WHERE AccountUID=?",
				notireward.RewardGold, notireward.RewardSpin, id)
			if err != nil {
				// 유저 정보 갱신 실패했어도 로그만 남기자
				log.Println("[DBAPI:GetNotice] Update User Fail : " + err.Error())
				log.Println("Query : UPDATE User SET Gold=Gold+", notireward.RewardGold,
					", Spin=Spin+", notireward.RewardSpin,
					", UpdateDate=NOW() WHERE AccountUID=", id)
			}
			rowsAffected, err := result.RowsAffected()
			if err != nil || rowsAffected == 0 {
				tx.Rollback()
				log.Println("[DBAPI:GetNotice] Update User Fail : " + GetUpdateError(err))
				log.Println("Query : UPDATE User SET Gold=Gold+", notireward.RewardGold,
					", Spin=Spin+", notireward.RewardSpin,
					", UpdateDate=NOW() WHERE AccountUID=", id)
			}

			// 3. 중복 보상 방지를 위해 보상내역 테이블에 업데이트 처리
			result, err = tx.Exec("Insert Into DataDB.NoticeRewardManage Values (?,?,NOW())", tmp.NotiID, id)
			if err != nil {
				// 실패했어도 로그만 남기자
				log.Println("[DBAPI:GetNotice] Insert DataDB.NoticeRewardManage Fail : " + err.Error())
				log.Println("Query : Insert Into NoticeRewardManage Values (", tmp.NotiID, ",", id, ",NOW())")
			}
			rowsAffected, err = result.RowsAffected()
			if err != nil || rowsAffected == 0 {
				tx.Rollback()
				log.Println("[DBAPI:GetNotice] Insert DataDB.NoticeRewardManage Fail : " + GetUpdateError(err))
				log.Println("Query : Insert Into NoticeRewardManage Values (", tmp.NotiID, ",", id, ",NOW())")
			}

		case common.NOTI_EVENT: // 이벤트공지
			// 이벤트 공지는 따로 서버가 할거 없이 그냥 내려주면 된다.
		default:
			// 정의되지 않은 이벤트 타입에 대해서는 로그만 남긴다.
			log.Println("[DBAPI:GetNotice] Undefined NotiType (", tmp.NotiType, ")")
		}
		rsp.Notice = append(rsp.Notice, tmp)
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:GetNotice] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, nil, 200
}

func UpdateEventResult(req message.EventResultRequest, mydb *sqlx.DB) (message.EventResultResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	rsp := message.EventResultResponse{}

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	// 유저 정보 찾기
	type UserInfo struct {
		Chapter              int       `db:"CurrentChapterIdx"`
		Gold                 int       `db:"Gold"`
		Spin                 int       `db:"Spin"`
		MaxSpin              int       `db:"MaxSpin"`
		LastSpinRechargeTime time.Time `db:"LastSpinRechargeTime"`
		Coin                 int       `db:"Coin"`
	}

	userinfo := UserInfo{}
	err := tx.Get(&userinfo, "SELECT CurrentChapterIdx, Gold, Spin, MaxSpin, LastSpinRechargeTime, Coin From User WHERE AccountUID = ?", req.ID)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:UpdateEventResult] Select User Fail : " + err.Error())
			log.Println("Query : SELECT CurrentChapterIdx, Gold, Spin, MaxSpin, LastSpinRechargeTime, Coin From User WHERE AccountUID = ?", req.ID)
			err = fmt.Errorf("Can't Find Your Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:UpdateEventResult] Select User Fail : " + err.Error())
			log.Println("Query : SELECT CurrentChapterIdx, Gold, Spin, MaxSpin, LastSpinRechargeTime, Coin From User WHERE AccountUID = ?", req.ID)
			err = fmt.Errorf("Find user info fail.")
			return rsp, err, 500
		}
	}

	// 이벤트 보상 정보 찾기
	err = tx.Get(&rsp, "SELECT RewardGold, RewardSpin, RewardEventCoin From DataDB.SuperAttackResult "+
		"WHERE CurrentChapterMin <= ? and currentChapterMax >= ? and "+
		"ScoreMin <= ? and ScoreMax >= ? ",
		userinfo.Chapter, userinfo.Chapter,
		req.Score, req.Score)
	if err != nil {
		log.Println("[DBAPI:UpdateEventResult] Select User Fail : " + err.Error())
		log.Println("Query : SELECT RewardGold, RewardSpin, RewardEventCoin From DataDB.SuperAttackResult ",
			"WHERE CurrentChapterMin <= ", userinfo.Chapter, " and currentChapterMax >= ", userinfo.Chapter, " and "+
				"ScoreMin <= ", req.Score, " and ScoreMax >= ", req.Score)
		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Can't Find SuperAttackResult Information in Server.")
			return rsp, err, 400
		} else {
			err = fmt.Errorf("Find SuperAttackResult info fail.")
			return rsp, err, 500
		}
	}

	// 응답 패킷 값 넣음
	rsp.ID = req.ID
	rsp.EventID = req.EventID
	rsp.LastSpinRechargeTime = userinfo.LastSpinRechargeTime
	// 이벤트 보상정보를 응답 패킷에 넣음
	rsp.UserGold = (userinfo.Gold + rsp.PlusGold) * req.MultipleValue
	// Spin 은 최대치가 있으므로 계산해줌
	if userinfo.MaxSpin < (userinfo.Spin+rsp.PlusSpin)*req.MultipleValue {
		rsp.UserSpin = userinfo.MaxSpin
	} else {
		rsp.UserSpin = (userinfo.Spin + rsp.PlusSpin) * req.MultipleValue
	}
	rsp.UserCoin = (userinfo.Coin + rsp.PlusCoin) * req.MultipleValue

	// DB 에 업데이트
	result, err := tx.Exec("UPDATE User SET Gold=?, Spin=?, Coin=?, UpdateDate=NOW() WHERE AccountUID=?",
		rsp.UserGold, rsp.UserSpin, rsp.UserCoin, rsp.ID)
	if err != nil {
		log.Println("[DBAPI:SuperAttackResult] Update User Fail : " + err.Error())
		log.Println("Query : UPDATE User SET Gold=", rsp.UserGold, ", Spin=", rsp.UserSpin, ", Coin=", rsp.UserCoin, ", UpdateDate=NOW() WHERE AccountUID=", rsp.ID)
		err = fmt.Errorf("Update User info fail.")
		return rsp, err, 500
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		tx.Rollback()
		log.Println("[DBAPI:SuperAttackResult] Update User Fail : " + GetUpdateError(err))
		log.Println("Query : UPDATE User SET Gold=", rsp.UserGold, ", Spin=", rsp.UserSpin, ", Coin=", rsp.UserCoin, ", UpdateDate=NOW() WHERE AccountUID=", rsp.ID)
		err = fmt.Errorf("Update User info fail.")
		if rowsAffected == 0 {
			tx.Rollback()
			return rsp, err, 400
		} else {
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:UpdateEventResult] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func InsertMansionItem(id int, item int, mydb *sqlx.DB) (int64, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	var gold int64 // 중복 보상 골드

	result, err := tx.Exec("INSERT into UserMansionItem values(?,?,0)", id, item)
	if err != nil {
		if MyErrCode(err) == 1062 { // Duplicate
			err := tx.Get(&gold, "Select DuplicationGold From DataDB.MansionItemInfo where MansionItemIndex = ?", item)
			if err != nil {
				log.Println("[DBAPI:InsertMansionItem] Select MansionItemInfo Fail : " + err.Error())
				log.Println("Query : Select DuplicationGold From DataDB.MansionItemInfo where MansionItemIndex = ", item)
				err = fmt.Errorf("Select DuplicationMansionItem For User fail.")
				return 0, err, 500
			}

			result, err := tx.Exec("Update User Set Gold = Gold + ? where AccountUID=?", gold, id)
			if err != nil {
				log.Println("[DBAPI:InsertMansionItem] Update User Fail : " + err.Error())
				log.Println("Query : Update User Set Gold = Gold + ", gold, " where AccountUID=", id)
				err = fmt.Errorf("Update DuplicationMansionItem For User fail.")
				return 0, err, 500
			} else {
				rowsAffected, err := result.RowsAffected()
				if err != nil || rowsAffected == 0 {
					log.Println("[DBAPI:InsertMansionItem] Update DuplicationMansionItem For User Fail : " + GetUpdateError(err))
					log.Println("Query : Update User Set Gold = Gold + ", gold, " where AccountUID=", id)
					err = fmt.Errorf("Update DuplicationMansionItem For User fail.")
					if rowsAffected == 0 {
						tx.Rollback()
						return 0, err, 400
					} else {
						return 0, err, 500
					}
				}
			}

		} else { // Not Duplicate, Other Error
			log.Println("[DBAPI:InsertMansionItem] ", MyErrCode(err))
			log.Println("[DBAPI:InsertMansionItem] Insert UserMansionItem Fail : " + err.Error())
			log.Println("Query : INSERT into UserMansionItem values (", id, ",", item, ",0)")
			err = fmt.Errorf("Insert MansionItem info fail.")
			return 0, err, 500
		}

	} else { // Insert Success
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:InsertMansionItem] Insert UserMansionItem Fail : " + GetUpdateError(err))
			log.Println("Query : INSERT into UserMansionItem values (", id, ",", item, ",0)")
			err = fmt.Errorf("Insert MansionItem info fail.")
			if rowsAffected == 0 {
				tx.Rollback()
				return 0, err, 400
			} else {
				return 0, err, 500
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:InsertMansionItem] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return 0, err, 500
	}

	return gold, err, 200
}

func SelectUserMansionInfo(id int, rand int, mydb *sqlx.DB) (message.MansionInfoResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.MansionInfoResponse{}
	rspitem := []message.MansionItem{}

	// rand 가 0이 아니면 랜덤 유저 정보를 전달해준다.
	// id 에다가 랜덤한 유저의 uid를 넣어서 아래 유저멘션 정보 플로우를 태운다.
	if rand > 0 {
		tmp := id
		err := tx.Get(&id, "select AccountUID from UserMansion where AccountUID != ? and Mansion > 0 order by rand() limit 1", tmp)
		if err != nil {
			log.Println("[DBAPI:SelectUserMantinoInfo] Select User Fail : " + err.Error())
			log.Println("Query : select AccountUID from UserMansion where AccountUID != ", tmp, " and Mansion > 0 order by rand() limit 1")
			if MyErrCode(err) == 4040 {
				err = fmt.Errorf("Not Exist Random User Mansion Info.")
				return rsp, err, 400
			} else {
				err = fmt.Errorf("Find UserMansion info fail.")
				return rsp, err, 500
			}
		}
	}

	// 아이템 리스트 select
	err := tx.Select(&rspitem, "SELECT ItemIndex, ItemPosition From UserMansionItem WHERE AccountUID = ?", id)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:SelectUserMantinoInfo] This User(", id, ") has no item.")
			log.Println("Query : SELECT ItemIndex, ItemPosition From UserMansionItem WHERE AccountUID = ", id)
		} else {
			log.Println("[DBAPI:SelectUserMantinoInfo] Select User Fail : " + err.Error())
			log.Println("Query : SELECT ItemIndex, ItemPosition From UserMansionItem WHERE AccountUID = ", id)
			err = fmt.Errorf("Find UserMansionItem info fail.")
			return rsp, err, 500
		}
	}

	// 유저 멘션 정보 select
	err = tx.Get(&rsp, "SELECT AccountUID, Mansion, MansionRoom, KeyID, Name, Gender, AppType, HaveLike, RemainLike "+
		" From UserMansion WHERE AccountUID = ?", id)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:SelectUserMantinoInfo] Select User Fail : " + err.Error())
			log.Println("Query : SELECT AccountUID, Mansion, MansionRoom, KeyID, Name, Gender, AppType, HaveLike, RemainLike From UserMansion WHERE AccountUID = ", id)
			err = fmt.Errorf("Can't Find UserMansion Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:SelectUserMantinoInfo] Select User Fail : " + err.Error())
			log.Println("Query : SELECT AccountUID, Mansion, MansionRoom, KeyID, Name, Gender, AppType, HaveLike, RemainLike From UserMansion WHERE AccountUID = ", id)
			err = fmt.Errorf("Find UserMansion info fail.")
			return rsp, err, 500
		}
	}

	rsp.MyMansionItem = rspitem

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:SelectUserMantinoInfo] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func SelectKeyIDMansionInfo(id int, keyid string, mydb *sqlx.DB) (message.MansionInfoResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.MansionInfoResponse{}
	rspitem := []message.MansionItem{}

	// 유저 멘션 정보 select
	err := tx.Get(&rsp, "SELECT AccountUID, Mansion, MansionRoom, KeyID, Name, Gender, AppType, HaveLike, RemainLike "+
		" From UserMansion WHERE KeyID = ?", keyid)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:SelectKeyIDMansionInfo] Select User Fail : " + err.Error())
			log.Println("Query : SELECT AccountUID, Mansion, MansionRoom, KeyID, Name, Gender, AppType, HaveLike, RemainLike From UserMansion WHERE KeyID = ", keyid)
			err = fmt.Errorf("Can't Find UserMansion Information in Server.")
			return rsp, err, 400
		} else {
			log.Println("[DBAPI:SelectKeyIDMansionInfo] Select User Fail : " + err.Error())
			log.Println("Query : SELECT AccountUID, Mansion, MansionRoom, KeyID, Name, Gender, AppType, HaveLike, RemainLike From UserMansion WHERE KeyID = ", keyid)
			err = fmt.Errorf("Find UserMansion info fail.")
			return rsp, err, 500
		}
	}

	// 아이템 리스트 select
	err = tx.Select(&rspitem, "SELECT ItemIndex, ItemPosition From UserMansionItem WHERE AccountUID = ?", rsp.ID)
	if err != nil {
		if MyErrCode(err) == 4040 {
			log.Println("[DBAPI:SelectKeyIDMansionInfo] This User(", id, ") has no item.")
			log.Println("Query : SELECT ItemIndex, ItemPosition From UserMansionItem WHERE AccountUID = ", rsp.ID)
		} else {
			log.Println("[DBAPI:SelectKeyIDMansionInfo] Select User Fail : " + err.Error())
			log.Println("Query : SELECT ItemIndex, ItemPosition From UserMansionItem WHERE AccountUID = ", rsp.ID)
			err = fmt.Errorf("Find UserMansionItem info fail.")
			return rsp, err, 500
		}
	}

	rsp.MyMansionItem = rspitem

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:SelectKeyIDMansionInfo] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

//func CreateUserMansionInfo(id uint32, mydb *sqlx.DB) (error, int) {
func CreateUserMansionInfo(user *message.CrtMansion, mydb *sqlx.DB) (error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	// 유저 멘션 정보 select
	var tmp int
	err := tx.Get(&tmp, "SELECT AccountUID From UserMansion WHERE AccountUID = ?", user.AccountUID)
	if err != nil {
		if MyErrCode(err) == 4040 {

			result, err := tx.Exec("Insert Into UserMansion values(?,0,0,?,?,?,?,?,?,NOW())",
				user.AccountUID, user.KeyID, user.Name, user.Gender, user.Type, user.HaveLike, user.RemainLike)
			if err != nil {
				log.Println("[DBAPI:CreateUserMansionInfo] Insert UserMansion Fail : " + err.Error())
				log.Println("Query : Insert Into UserMansion values(", user.AccountUID, ",0,0,", user.KeyID, ",", user.Name, ",", user.Gender, ",", user.Type, ",", user.HaveLike, ",", user.RemainLike, ",NOW())")
				err = fmt.Errorf("Insert UserMansionInfo fail.")
				return err, 500
			}
			rowsAffected, err := result.RowsAffected()
			if err != nil || rowsAffected == 0 {
				log.Println("[DBAPI:CreateUserMansionInfo] Insert UserMansion Fail : " + GetUpdateError(err))
				log.Println("Query : Insert Into UserMansion values(", user.AccountUID, ",0,0,", user.KeyID, ",", user.Name, ",", user.Gender, ",", user.Type, ",", user.HaveLike, ",", user.RemainLike, ",NOW())")
				err = fmt.Errorf("Insert UserMansionInfo fail.")
				if rowsAffected == 0 {
					tx.Rollback()
					return err, 400
				} else {
					return err, 500
				}
			}

		} else {
			log.Println("[DBAPI:CreateUserMansionInfo] Select User Fail : " + err.Error())
			log.Println("Query : SELECT AccountUID From UserMansion WHERE AccountUID = ", user.AccountUID)
			err = fmt.Errorf("Find UserMansion info fail.")
			return err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:CreateUserMansionInfo] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return err, 500
	}

	return err, 200
}

func UpdateMoveMansionItem(req message.MansionItemMoveRequest, mydb *sqlx.DB) (message.MansionItemMoveResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.MansionItemMoveResponse{}
	rsp.ID = req.ID

	for _, item := range req.MoveMansionItem {
		result, err := tx.Exec("Update UserMansionItem Set ItemPosition = ? Where AccountUID=? AND ItemIndex = ?",
			item.ItemPosition, req.ID, item.ItemIndex)
		if err != nil {
			log.Println("[DBAPI:UpdateMoveMansionItem] Update UserMansionItem Fail : " + err.Error())
			log.Println("Query : Update UserMansionItem Set ItemPosition = ", item.ItemPosition, " Where AccountUID=", req.ID, " AND ItemIndex = ", item.ItemIndex)
			err = fmt.Errorf("Update UserMansionItem For User fail.")
			return rsp, err, 500
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateMoveMansionItem] Update UserMansionItem For User Fail : " + GetUpdateError(err))
			log.Println("Query : Update UserMansionItem Set ItemPosition = ", item.ItemPosition, " Where AccountUID=", req.ID, " AND ItemIndex = ", item.ItemIndex)
			err = fmt.Errorf("Update UserMansionItem For User fail.")
			if rowsAffected == 0 {
				tx.Rollback()
				return rsp, err, 400
			} else {
				return rsp, err, 500
			}
		}

		moveitem := message.MansionItem{}
		err = tx.Get(&moveitem, "SELECT ItemIndex, ItemPosition FROM UserMansionItem WHERE AccountUID=? AND ItemIndex=?", req.ID, item.ItemIndex)
		if err != nil { // 무슨 에러든지 에러면 안됨. 왜냐! 위에서 다 한거니까
			log.Println("[DBAPI:UpdateMoveMansionItem] Select UserMansionItem Fail : " + err.Error())
			log.Println("Query : SELECT ItemIndex, ItemPosition FROM UserMansionItem WHERE AccountUID=", req.ID, " AND ItemIndex=", item.ItemIndex)
			err = fmt.Errorf("Select UserMansionItem fail.")
			return rsp, err, 500
		} else {
			rsp.MoveMansionItem = append(rsp.MoveMansionItem, moveitem)
		}
	}

	err := tx.Commit()
	if err != nil {
		log.Println("[DBAPI:UpdateMoveMansionItem] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func SelectMansionFriend(id int, mydb *sqlx.DB) (message.MansionFirendResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.MansionFirendResponse{}
	rsplist := []message.MansionFirend{}

	err := tx.Select(&rsplist, "Select FriendUID as AccountUID, "+
		"FriendKeyID as KeyID, "+
		"FriendName as Name, "+
		"FriendGender as Gender, "+
		"FriendAppType as AppType, "+
		"HaveLike "+
		"From MansionFriend F Left Join UserMansion U On F.FriendUID = U.AccountUID "+
		"Where F.AccountUID=?", id)
	if err != nil {
		log.Println("[DBAPI:SelectMansionFriend] Select MansionFriend Fail : " + err.Error())
		log.Println("Query : Select FriendUID as AccountUID, "+
			"FriendKeyID as KeyID, "+
			"FriendName as Name, "+
			"FriendGender as Gender, "+
			"FriendAppType as AppType "+
			"HaveLike "+
			"From MansionFriend F Left Join UserMansion U On F.FriendUID = U.AccountUID "+
			"Where F.AccountUID=", id)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Found Mansion Friend Info.")
			return rsp, err, 400 // 200?
		} else {
			err = fmt.Errorf("Select MansionFriend Fail.")
			return rsp, err, 500
		}
	}

	rsp.MyMansionFirend = rsplist

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:SelectMansionFriend] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func InsertMansionFriend(id int, friend int, mydb *sqlx.DB) (message.MansionFirend, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.MansionFirend{}

	// 유저 정보를 꺼내서 존재하면 친구정보로 등록
	err := tx.Get(&rsp, "Select AccountUID, KeyID, Name, Gender, AppType, HaveLike from UserMansion Where AccountUID=?", friend)
	if err != nil {
		log.Println("[DBAPI:InsertMansionFriend] Select Account Info Fail : " + err.Error())
		log.Println("Query : Select AccountUID, KeyID, Name, Gender, AppType, HaveLike from UserMansion Where AccountUID =", friend)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Found Friend Info.")
			return rsp, err, 400 // 200?
		} else {
			err = fmt.Errorf("Select User Info Fail.")
			return rsp, err, 500
		}
	}

	// 친구추가 최대 초과 시 에러처리
	var friendcnt int
	err = tx.Get(&friendcnt, "Select count(*) From MansionFriend Where AccountUID=?", id)
	if err != nil {
		log.Println("[DBAPI:InsertMansionFriend] Select MansionFriend Info Fail : " + err.Error())
		log.Println("Query : Select count(*) Where AccountUID =", id)

		if MyErrCode(err) == 4040 {
			// 없는건 상관 없다.
		} else {
			err = fmt.Errorf("Select User Info Fail.")
			return rsp, err, 500
		}
	}
	if friendcnt > common.MAX_FRIEND {
		err = fmt.Errorf("Mansion Friend Count Exceed.")
		return rsp, err, 400
	}

	result, err := tx.Exec("Insert Into MansionFriend (AccountUID, FriendUID, FriendKeyID, FriendName, FriendGender, FriendAppType) "+
		"Values (?, ?, ?, ?, ?, ?)",
		id, rsp.ID, rsp.KeyID, rsp.Name, rsp.Gender, rsp.AppType)
	if err != nil {
		log.Println("[DBAPI:InsertMansionFriend] Insert MansionFriend Fail : " + err.Error())
		log.Println("Query : Insert Into MansionFriend (AccountUID, FriendUID, FriendKeyID, FriendName, FriendGender, FriendAppType) "+
			"Values (", id, ", ", rsp.ID, ", ", rsp.KeyID, ", ", rsp.Name, ", ", rsp.Gender, ", ", rsp.AppType, ")")
		err = fmt.Errorf("Insert MansionFriend Fail.")
		return rsp, err, 500
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:InsertMansionFriend] Insert MansionFriend For User Fail : " + GetUpdateError(err))
		log.Println("Query : Insert Into MansionFriend (AccountUID, FriendUID, FriendKeyID, FriendName, FriendGender, FriendAppType) "+
			"Values (", id, ", ", rsp.ID, ", ", rsp.KeyID, ", ", rsp.Name, ", ", rsp.Gender, ", ", rsp.AppType, ")")
		err = fmt.Errorf("Insert MansionFriend fail.")
		if rowsAffected == 0 {
			tx.Rollback()
			return rsp, err, 400
		} else {
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:InsertMansionFriend] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func DeleteMansionFriend(id int, friend int, mydb *sqlx.DB) (message.MansionFirend, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.MansionFirend{}

	result, err := tx.Exec("Delete From MansionFriend Where AccountUID = ? AND FriendUID = ?", id, friend)
	if err != nil {
		log.Println("[DBAPI:DeleteMansionFriend] Delete MansionFriend Fail : " + err.Error())
		log.Println("Query : Delete From MansionFriend Where AccountUID = ", id, " AND FriendUID = ", friend)
		err = fmt.Errorf("Delete MansionFriend Fail.")
		return rsp, err, 500
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:DeleteMansionFriend] Delete MansionFriend For User Fail : " + GetUpdateError(err))
		log.Println("Query : Delete From MansionFriend Where AccountUID = ", id, " AND FriendUID = ", friend)
		err = fmt.Errorf("Delete MansionFriend fail.")
		if rowsAffected == 0 {
			tx.Rollback()
			return rsp, err, 400
		} else {
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:DeleteMansionFriend] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func UpdateMansionLike(id int, target int, mydb *sqlx.DB) (int, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	// RemainLike 초기화 확인
	type RemainLikeTime struct {
		Update time.Time `db:"RemainLikeUpdateTime"`
		Now    time.Time `db:"Now()"`
	}

	liketime := RemainLikeTime{}
	err := tx.Get(&liketime, "SELECT RemainLikeUpdateTime, Now() FROM UserMansion WHERE AccountUID = ?", id)
	if err != nil {
		log.Println("[DBAPI:UpdateMansionLike] Select User Fail : " + err.Error())
		log.Println("Query : RemainLikeUpdateTime, Now() FROM UserMansion WHERE AccountUID = ?", id)
		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Exist Data in UserMansion Table.")
			return -1, err, 204
		} else {
			err = fmt.Errorf("Find UserMansion info fail.")
			return -1, err, 500
		}
	}

	if liketime.Update.Day() < liketime.Now.Day() {
		result, err := tx.Exec("UPDATE UserMansion SET RemainLike=?, RemainLikeUpdateTime=NOW() WHERE AccountUID=?", common.MANSION_LIKE, id)
		if err != nil {
			log.Println("[DBAPI:UpdateMansionLike] Update UserMansion Fail : " + err.Error())
			log.Println("Query : UPDATE UserMansion SET RemainLike=", common.MANSION_LIKE, ", RemainLikeUpdateTime=NOW() WHERE AccountUID=", id)
			err = fmt.Errorf("Update UserMansion info fail.")
			return -1, err, 500
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateMansionLike] Update UserMansion Fail : " + GetUpdateError(err))
			log.Println("Query : UPDATE UserMansion SET RemainLike=", common.MANSION_LIKE, ", RemainLikeUpdateTime=NOW() WHERE AccountUID=", id)
			if rowsAffected == 0 {
				tx.Rollback()
				err = fmt.Errorf("No Rows Update UserMansion Table.")
				return -1, err, 204
			} else {
				err = fmt.Errorf("Update UserMansion info fail.")
				return -1, err, 500
			}
		}
	}

	// RemainLike 확인
	var remainLike int
	err = tx.Get(&remainLike, "Select if(RemainLike > 0, RemainLike, 0) from UserMansion Where AccountUID=?", id)
	if err != nil {
		log.Println("[DBAPI:UpdateMansionLike] Select UserMansion Info Fail : " + err.Error())
		log.Println("Query : Select if(RemainLike > 0, RemainLike, 0) from UserMansion Where AccountUID=", id)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Found Friend Info.")
			return -1, err, 400 // 200?
		} else {
			err = fmt.Errorf("Select User Info Fail.")
			return -1, err, 500
		}
	}

	if remainLike <= 0 {
		err = fmt.Errorf("Not Enough Like.")
		return remainLike, err, 400
	}

	// 먼저 내 LIke 수를 차감하고
	result, err := tx.Exec("Update UserMansion Set RemainLike = RemainLike - 1, RemainLikeUpdateTime = Now() Where AccountUID = ? ", id)
	if err != nil {
		log.Println("[DBAPI:UpdateMansionLike] Update UserMansion Fail : " + err.Error())
		log.Println("Query : Update UserMansion Set RemainLike = RemainLike - 1, RemainLikeUpdateTime = Now() Where AccountUID =", id)
		err = fmt.Errorf("Update UserMansion Fail.")
		return remainLike, err, 500
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateMansionLike] Update UpdateMansion Fail : " + GetUpdateError(err))
		log.Println("Query : Update UserMansion Set RemainLike = RemainLike - 1, RemainLikeUpdateTime = Now() Where AccountUID =", id)
		err = fmt.Errorf("Update UserMansion fail.")
		if rowsAffected == 0 {
			tx.Rollback()
			return remainLike, err, 400
		} else {
			return remainLike, err, 500
		}
	}

	// 상대 Like를 올려줌
	result, err = tx.Exec("Update UserMansion Set HaveLike = HaveLike + 1 Where AccountUID = ? ", target)
	if err != nil {
		log.Println("[DBAPI:UpdateMansionLike] Update UserMansion Fail : " + err.Error())
		log.Println("Query : Update UserMansion Set HaveLike = HaveLike + 1 Where AccountUID = ", target)
		err = fmt.Errorf("Update UserMansion Fail.")
		return remainLike, err, 500
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateMansionLike] Update UpdateMansion Fail : " + GetUpdateError(err))
		log.Println("Query : Update UserMansion Set HaveLike = HaveLike + 1 Where AccountUID = ", target)
		err = fmt.Errorf("Update UserMansion fail.")
		if rowsAffected == 0 {
			tx.Rollback()
			return remainLike, err, 400
		} else {
			return remainLike, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:DeleteMansionFriend] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return remainLike, err, 500
	}

	// 최종 차감된 Like 수를 여기서 계산해서 돌려줌.
	remainLike -= 1
	return remainLike, err, 200
}

func SelectMainsionRanking(req message.MansionRankRequest, mydb *sqlx.DB) (message.MansionRankResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.MansionRankResponse{}
	rsplist := []message.MansionRank{}

	err := tx.Select(&rsplist, "SELECT AccountUID, KeyID, Name, Gender, AppType, Rank, HaveLike "+
		"FROM ( "+
		" SELECT AccountUID, KeyID, Name, Gender, AppType, Mansion, HaveLike, "+
		" ( @real_rank := IF ( @last > HaveLike, @real_rank:=@real_rank+1, @real_rank ) ) AS Rank, "+
		" ( @last := HaveLike )  as TMP "+
		" FROM UserMansion A, "+
		" (SELECT @last := 0, @real_rank := 1 ) AS B "+
		" ORDER BY  HaveLike desc"+
		") C "+
		" WHERE (Rank >= ? and Rank <= ?) and Mansion >= 1",
		req.RankMin, req.RankMax)

	if err != nil {
		log.Println("[DBAPI:SelectMainsionRanking] Select MansionRank Info Fail : " + err.Error())
		log.Println("Query : SELECT AccountUID, KeyID, Name, Gender, AppType, Rank, HaveLike "+
			"FROM ( "+
			" SELECT AccountUID, KeyID, Name, Gender, AppType, Mansion, HaveLike, "+
			" ( @real_rank := IF ( @last > HaveLike, @real_rank:=@real_rank+1, @real_rank ) ) AS Rank, "+
			" ( @last := HaveLike )  as TMP "+
			" FROM UserMansion A, "+
			" (SELECT @last := 0, @real_rank := 1 ) AS B "+
			" ORDER BY  HaveLike desc"+
			") C "+
			" WHERE (Rank >= ", req.RankMin, " and Rank <= ", req.RankMax, ") and Mansion >= 1")

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Found Ranking Info.")
			return rsp, err, 400
		} else {
			err = fmt.Errorf("Select Mansion Rank Fail.")
			return rsp, err, 500
		}
	}

	// 찾은 범위 내에서 나의 정보가 있는지 확인
	// 복사
	tmp := rsplist
	// 정렬
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].ID <= tmp[j].ID
	})
	// 검색
	idx := sort.Search(len(tmp), func(i int) bool {
		return tmp[i].ID >= req.ID
	})
	// 존재한다면 내 순위를 패킷에 넣음
	if tmp[idx].ID == req.ID {
		rsp.MyRank = tmp[idx].Rank
	} else { // 존재하지 않는다면 별도로 Select
		err := tx.Get(&rsp.MyRank, "SELECT Rank "+
			"FROM ( "+
			" SELECT AccountUID, "+
			" ( @real_rank := IF ( @last > HaveLike, @real_rank:=@real_rank+1, @real_rank ) ) AS Rank, "+
			" ( @last := HaveLike )  as TMP "+
			" FROM UserMansion A, "+
			" (SELECT @last := 0, @real_rank := 1 ) AS B "+
			" ORDER BY  HaveLike desc"+
			") C "+
			" Where AccountUID = ?", req.ID)

		if err != nil {
			log.Println("[DBAPI:SelectMainsionRanking] Select MansionRank Info Fail : " + err.Error())
			log.Println("Query : SELECT Rank "+
				"FROM ( "+
				" SELECT AccountUID, "+
				" ( @real_rank := IF ( @last > HaveLike, @real_rank:=@real_rank+1, @real_rank ) ) AS Rank, "+
				" ( @last := HaveLike )  as TMP "+
				" FROM UserMansion A, "+
				" (SELECT @last := 0, @real_rank := 1 ) AS B "+
				" ORDER BY  HaveLike desc"+
				") C "+
				" Where AccountUID = ", req.ID)

			if MyErrCode(err) == 4040 {
				err = fmt.Errorf("Not Found Ranking Info.")
				return rsp, err, 400
			} else {
				err = fmt.Errorf("Select Mansion Rank Fail.")
				return rsp, err, 500
			}
		}
	}

	// 전체 유저 수 가져옴
	err = tx.Get(&rsp.TotalUser, "SELECT Count(*) From UserMansion")
	if err != nil {
		log.Println("[DBAPI:SelectMainsionRanking] Select MansionRank Info Fail : " + err.Error())
		log.Println("Query : SELECT Count(*) From UserMansion")

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Found Ranking Info.")
			return rsp, err, 400
		} else {
			err = fmt.Errorf("Select Mansion Rank Fail.")
			return rsp, err, 500
		}
	}

	rsp.MansionRankInfo = rsplist

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:SelectMainsionRanking] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func SelectMainsionRankReward(id int, mydb *sqlx.DB) (message.MansionRankRewardResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.MansionRankRewardResponse{}
	// 랭크 보상 테이블 가져옴
	// 클라이언트에서 RankID를 통해 보상정보를 가져가도록 변경함
	//err := tx.Get(&rsp, "Select AccountUID, KeyID, Name, Gender, AppType, HaveLike, Rank, RewardGold, RewardSpin, RewardRandomBox From WeeklyRankReward Where AccountUID=?", id)
	err := tx.Get(&rsp, "Select AccountUID, KeyID, Name, Gender, AppType, HaveLike, Rank, RankID From WeeklyRankReward Where AccountUID=?", id)
	if err != nil {
		log.Println("[DBAPI:SelectMainsionRankReward] Select MansionRankReward Info Fail : " + err.Error())
		//log.Println("Query : Select AccountUID, KeyID, Name, Gender, AppType, HaveLike, Rank, RewardGold, RewardSpin, RewardRandomBox From WeeklyRankReward Where AccountUID=", id)
		log.Println("Query : Select AccountUID, KeyID, Name, Gender, AppType, HaveLike, Rank, RankID From WeeklyRankReward Where AccountUID=", id)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Found Ranking Reward Info.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Select Ranking Reward Fail.")
			return rsp, err, 500
		}
	}

	// 보상 테이블에서 삭제
	result, err := tx.Exec("Delete From WeeklyRankReward Where AccountUID = ? ", id)
	if err != nil {
		log.Println("[DBAPI:SelectMainsionRankReward] Delete WeeklyRankReward Fail : " + err.Error())
		log.Println("Query : Delete From WeeklyRankReward Where AccountUID =", id)
		err = fmt.Errorf("Delete WeeklyRankReward Fail.")
		return rsp, err, 500
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:SelectMainsionRankReward] Delete WeeklyRankReward Fail : " + GetUpdateError(err))
		log.Println("Query : Delete From WeeklyRankReward Where AccountUID =", id)
		err = fmt.Errorf("Delete WeeklyRankReward Fail.")
		if rowsAffected == 0 {
			tx.Rollback()
			return rsp, err, 204
		} else {
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:SelectMainsionRankReward] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func SelectQuestInfo(id int, mydb *sqlx.DB) (message.QuestInfoResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.QuestInfoResponse{}
	err := tx.Get(&rsp, "SELECT AccountUID, GainGold, UseGold, UseSpin, "+
		"TryRaid, TryAttack, TryEventGame, CompleteChapter, "+
		"StructureAction, MoveFurniture, GainFurniture, CompleteQuest FROM UserQuestData WHERE AccountUID=?", id)
	if err != nil {
		log.Println("[DBAPI:SelectQuestInfo] Select UserQuestData Info Fail : " + err.Error())
		log.Println("Query : SELECT AccountUID, GainGold, UseGold, UseSpin, TryRaid, TryAttack, TryEventGame, CompleteChapter, StructureAction, MoveFurniture, GainFurniture, CompleteQuest FROM UserQuestData WHERE AccountUID=", id)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Found UserQuest Info.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Select UserQuest Fail.")
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:SelectMainsionRankReward] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func UpdateUserQuest(req message.QuestUpdateRequest, mydb *sqlx.DB) (message.QuestInfoResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.QuestInfoResponse{}

	result, err := tx.Exec("Insert Into UserQuestData "+
		"(AccountUID, GainGold, UseGold, UseSpin, TryRaid, TryAttack, TryEventGame, CompleteChapter, StructureAction, MoveFurniture, GainFurniture) "+
		"Values (?,?,?,?,?,?,?,?,?,?,?) "+
		"ON DUPLICATE KEY UPDATE "+
		"GainGold = GainGold + ?, UseGold = UseGold + ?, UseSpin = UseSpin + ?, "+
		"TryRaid = TryRaid + ?, TryAttack = TryAttack + ?, TryEventGame = TryEventGame + ?, "+
		"CompleteChapter = ?, StructureAction = StructureAction + ?, MoveFurniture = MoveFurniture + ?, "+
		"GainFurniture = GainFurniture + ?",
		req.ID, req.GainGold, req.UseGold, req.UseSpin,
		req.TryRaid, req.TryAttack, req.TryEventGame, req.CompleteChapter,
		req.StructureAction, req.MoveFurniture, req.GainFurniture,
		req.GainGold, req.UseGold, req.UseSpin,
		req.TryRaid, req.TryAttack, req.TryEventGame,
		req.CompleteChapter, req.StructureAction, req.MoveFurniture,
		req.GainFurniture)
	if err != nil {
		log.Println("[DBAPI:UpdateUserQuest] Update UserQuestData Fail : " + err.Error())
		log.Println("Query : Insert Into UserQuestData "+
			"(AccountUID, GainGold, UseGold, UseSpin, TryRaid, TryAttack, TryEventGame, CompleteChapter, StructureAction, MoveFurniture, GainFurniture) "+
			"Values (", req.ID, ",", req.GainGold, ",", req.UseGold, ",", req.UseSpin, ",",
			req.TryRaid, ",", req.TryAttack, ",", req.TryEventGame, ",", req.CompleteChapter, ",",
			req.StructureAction, ",", req.MoveFurniture, ",", req.GainFurniture, ") "+
				"ON DUPLICATE KEY UPDATE "+
				"GainGold = GainGold + ", req.GainGold, ", UseGold = UseGold + ", req.UseGold, ", UseSpin = UseSpin + ", req.UseSpin, ", "+
				"TryRaid = TryRaid + ", req.TryRaid, ", TryAttack = TryAttack + ", req.TryAttack, ", TryEventGame = TryEventGame + ", req.TryEventGame, ", "+
				"CompleteChapter = ", req.CompleteChapter, ", StructureAction = StructureAction + ", req.StructureAction, ", MoveFurniture = MoveFurniture + ", req.MoveFurniture, ", "+
				"GainFurniture = GainFurniture + ", req.GainFurniture)
		err = fmt.Errorf("Update UserQuestData Fail.")
		return rsp, err, 500
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateUserQuest] Update UserQuestData Fail : " + GetUpdateError(err))
		log.Println("Query : Insert Into UserQuestData "+
			"(AccountUID, GainGold, UseGold, UseSpin, TryRaid, TryAttack, TryEventGame, CompleteChapter, StructureAction, MoveFurniture, GainFurniture) "+
			"Values (", req.ID, ",", req.GainGold, ",", req.UseGold, ",", req.UseSpin, ",",
			req.TryRaid, ",", req.TryAttack, ",", req.TryEventGame, ",", req.CompleteChapter, ",",
			req.StructureAction, ",", req.MoveFurniture, ",", req.GainFurniture, ") "+
				"ON DUPLICATE KEY UPDATE "+
				"GainGold = GainGold + ", req.GainGold, ", UseGold = UseGold + ", req.UseGold, ", UseSpin = UseSpin + ", req.UseSpin, ", "+
				"TryRaid = TryRaid + ", req.TryRaid, ", TryAttack = TryAttack + ", req.TryAttack, ", TryEventGame = TryEventGame + ", req.TryEventGame, ", "+
				"CompleteChapter = ", req.CompleteChapter, ", StructureAction = StructureAction + ", req.StructureAction, ", MoveFurniture = MoveFurniture + ", req.MoveFurniture, ", "+
				"GainFurniture = GainFurniture + ", req.GainFurniture)
		err = fmt.Errorf("Update UserQuestData Fail.")
		if rowsAffected == 0 {
			tx.Rollback()
			return rsp, err, 204
		} else {
			return rsp, err, 500
		}
	}

	err = tx.Get(&rsp, "SELECT AccountUID, GainGold, UseGold, UseSpin, "+
		"TryRaid, TryAttack, TryEventGame, CompleteChapter, "+
		"StructureAction, MoveFurniture, GainFurniture, CompleteQuest FROM UserQuestData WHERE AccountUID=?", req.ID)
	if err != nil {
		log.Println("[DBAPI:UpdateUserQuest] Select UserQuestData Info Fail : " + err.Error())
		log.Println("Query : SELECT AccountUID, GainGold, UseGold, UseSpin, TryRaid, TryAttack, TryEventGame, CompleteChapter, StructureAction, MoveFurniture, GainFurniture, CompleteQuest FROM UserQuestData WHERE AccountUID=", req.ID)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Found UserQuest Info.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Select UserQuest Fail.")
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:UpdateUserQuest] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func UpdateUserQuestComplete(req message.QuestCompleteRequest, mydb *sqlx.DB) (message.QuestCompleteResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.QuestCompleteResponse{}

	// 유저 퀘스트 완료 업데이트
	result, err := tx.Exec("Update UserQuestData Set CompleteQuest = CompleteQuest | (1 << (? - 1)), UpdateTime = Now(6) where AccountUID=?", req.CompleteQuest, req.ID)
	if err != nil {
		log.Println("[DBAPI:UpdateUserQuestComplete] Update UserQuestData Fail : " + err.Error())
		log.Println("Query : Update UserQuestData Set CompleteQuest = CompleteQuest | (1 << (", req.CompleteQuest, " - 1)), UpdateTime = Now(6) where AccountUID=", req.ID)
		err = fmt.Errorf("Update UserQuestData Fail.")
		return rsp, err, 500
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateUserQuestComplete] Update UserQuestData Fail : " + GetUpdateError(err))
		log.Println("Query : Update UserQuestData Set CompleteQuest = CompleteQuest | (1 << (", req.CompleteQuest, " - 1)), UpdateTime = Now(6) where AccountUID=", req.ID)
		err = fmt.Errorf("Update UserQuestData Fail.")
		if rowsAffected == 0 {
			tx.Rollback()
			return rsp, err, 204
		} else {
			return rsp, err, 500
		}
	}

	// 보상 정보 가져옴
	err = tx.Get(&rsp, "Select MansionNumber, OpenRoomNumber, NextRoomOpen, MissionRewardGold, MissionRewardSpin, MissionRewardRandomBox from DataDB.MansionMission Where MissionNumber=?", req.CompleteQuest)
	if err != nil {
		log.Println("[DBAPI:UpdateUserQuestComplete] Select MansionMission Info Fail : " + err.Error())
		log.Println("Query : Select MansionNumber, OpenRoomNumber, NextRoomOpen, MissionRewardGold, MissionRewardSpin, MissionRewardRandomBox from DataDB.MansionMission Where MissionNumber=", req.CompleteQuest)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Found MansionMission Info.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Select MansionMission Fail.")
			return rsp, err, 500
		}
	}

	// 멘션 해금 퀘스트라면
	if rsp.IsMansionOpen == 1 {
		type MyInfo struct {
			Mymansion int8 `db:"Mansion"`
			Myroom    int8 `db:"MansionRoom"`
		}
		myinfo := MyInfo{}

		// 현재 유저의 멘션, 룸 정보 확인한 뒤
		err = tx.Get(&myinfo, "Select Mansion, MansionRoom From UserMansion Where AccountUID=?", req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateUserQuestComplete] Select MansionMission Info Fail : " + err.Error())
			log.Println("Query : Select Mansion, MansionRoom From UserMansion Where AccountUID=", req.ID)

			if MyErrCode(err) == 4040 {
				err = fmt.Errorf("Not Found UserMansion Info.")
				return rsp, err, 204
			} else {
				err = fmt.Errorf("Select UserMansion Fail.")
				return rsp, err, 500
			}
		}

		// 해금 멘션 번호와 현재 멘션 번호가 같으면 룸 업데이트 케이스
		if myinfo.Mymansion == rsp.MansionNumber {
			if myinfo.Myroom < rsp.RoomNumber { // 내 룸 번호가 보상 룸 번호보다 작아야 업데이트.. 왜냐.. 돈으로 샀을수도 있으니까
				result, err = tx.Exec("Update UserMansion Set MansionRoom = ? Where AccountUID=?", rsp.RoomNumber, req.ID)
				if err != nil {
					log.Println("[DBAPI:UpdateUserQuestComplete] Update UserMansion Fail : " + err.Error())
					log.Println("Query :Update UserMansion Set MansionRoom = ", rsp.RoomNumber, " Where AccountUID=", rsp.ID)
					err = fmt.Errorf("Update UserMansion Fail.")
					return rsp, err, 500
				}
				rowsAffected, err = result.RowsAffected()
				if err != nil || rowsAffected == 0 {
					log.Println("[DBAPI:UpdateUserQuestComplete] Update UserMansion Fail : " + GetUpdateError(err))
					log.Println("Query :Update UserMansion Set MansionRoom = ", rsp.RoomNumber, " Where AccountUID=", rsp.ID)
					err = fmt.Errorf("Update UserMansion Fail.")
					if rowsAffected == 0 {
						tx.Rollback()
						return rsp, err, 204
					} else {
						return rsp, err, 500
					}
				}
			} else {
				// 내 룸 번호가 보상 룸 번호보다 같거나 크면 그냥 스킵..
			}
		} else if myinfo.Mymansion < rsp.MansionNumber { // 보상 멘션 값이 현재 나보다 크면 ROOM과 Mansion 둘다 업데이트, 이건 무조건임
			result, err = tx.Exec("Update UserMansion Set Mansion = ?, MansionRoom = ? Where AccountUID=?", rsp.MansionNumber, rsp.RoomNumber, req.ID)
			if err != nil {
				log.Println("[DBAPI:UpdateUserQuestComplete] Update UserMansion Fail : " + err.Error())
				log.Println("Query :Update UserMansion Set Mansion = ", rsp.MansionNumber, " MansionRoom = ", rsp.RoomNumber, " Where AccountUID=", rsp.ID)
				err = fmt.Errorf("Update UserMansion Fail.")
				return rsp, err, 500
			}
			rowsAffected, err = result.RowsAffected()
			if err != nil || rowsAffected == 0 {
				log.Println("[DBAPI:UpdateUserQuestComplete] Update UserMansion Fail : " + GetUpdateError(err))
				log.Println("Query :Update UserMansion Set Mansion = ", rsp.MansionNumber, " MansionRoom = ", rsp.RoomNumber, " Where AccountUID=", rsp.ID)
				err = fmt.Errorf("Update UserMansion Fail.")
				if rowsAffected == 0 {
					tx.Rollback()
					return rsp, err, 204
				} else {
					return rsp, err, 500
				}
			}
		} else { // 보상 멘션 값이 현재 나보다 작으면.. 스킵.. 근데 그럴 수 있나?? 로그나 찍어보자
			log.Println("[DBAPI:UpdateUserQuestComplete] Mymansion:", myinfo.Mymansion, " RewardMansion:", rsp.MansionNumber)
		}
	} // 멘션 해금 조건 END

	// 보상정보 업데이트
	result, err = tx.Exec("Update User Set Gold = Gold + ?, Spin = Spin + ?, UpdateDate = NOW(6) Where AccountUID=?", rsp.RewardGold, rsp.RewardSpin, req.ID)
	if err != nil {
		log.Println("[DBAPI:UpdateUserQuestComplete] Update User Fail : " + err.Error())
		log.Println("Query :Update User Set Gold = Gold + ", rsp.RewardGold, ", Spin = Spin + ", rsp.RewardSpin, ", UpdateDate = NOW(6) Where AccountUID=", req.ID)
		err = fmt.Errorf("Update User Fail.")
		return rsp, err, 500
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateUserQuestComplete] Update User Fail : " + GetUpdateError(err))
		log.Println("Query :Update User Set Gold = Gold + ", rsp.RewardGold, ", Spin = Spin + ", rsp.RewardSpin, ", UpdateDate = NOW(6) Where AccountUID=", req.ID)
		err = fmt.Errorf("Update User Fail.")
		if rowsAffected == 0 {
			tx.Rollback()
			return rsp, err, 204
		} else {
			return rsp, err, 500
		}
	}

	// 최종 유저 골드,스핀 정보
	err = tx.Get(&rsp, "Select AccountUID, Gold, Spin, LastSpinRechargeTime From User Where AccountUID=?", req.ID)
	if err != nil {
		log.Println("[DBAPI:UpdateUserQuestComplete] Select User Info Fail : " + err.Error())
		log.Println("Query : Select AccountUID, Gold, Spin, LastSpinRechargeTime From User Where AccountUID=", req.ID)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Found User Info.")
			return rsp, err, 204
		} else {
			err = fmt.Errorf("Select User Fail.")
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:UpdateUserQuestComplete] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func UpdateUserQuestClear(req message.QuestCompleteRequest, mydb *sqlx.DB) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	type MissionValue struct {
		MissionType  int8 `db:"MissionType"`
		MissionValue int  `db:"MissionValue"`
	}

	missionValue := MissionValue{}

	err := tx.Get(&missionValue, "Select MissionType, MissionValue From DataDB.MansionMission where MissionNumber=?", req.CompleteQuest)
	if err != nil {
		log.Println("[DBAPI:UpdateUserQuestClear] Select UserQuestData Info Fail : " + err.Error())
		log.Println("Query :Select MissionType, MissionValue From DataDB.MansionMission where MissionNumber=", req.CompleteQuest)
		return
	}

	query := "UPDATE UserQuestData SET "
	switch missionValue.MissionType {
	case common.MISSION_GET_GOLD:
		//query += "GainGold = GainGold - "
		query += "GainGold = 0"
	case common.MISSION_SPEND_GOLD:
		//query += "UseGold = UseGold - "
		query += "UseGold = 0"
	case common.MISSION_SPEND_SPIN:
		//query += "UseSpin = UseSpin - "
		query += "UseSpin = 0"
	case common.MISSION_TRY_RAID:
		//query += "TryRaid = TryRaid - "
		query += "TryRaid = 0"
	case common.MISSION_TRY_ATTACK:
		//query += "TryAttack = TryAttack - "
		query += "TryAttack = 0"
	case common.MISSION_TRY_SUPER_ATTACK:
		//query += "TryEventGame = TryEventGame - "
		query += "TryEventGame = 0"
	case common.MISSION_COMPLETE_CHAPTER:
		log.Println("[DBAPI:UpdateUserQuestClear] Skip UserQuestClear Becuase COMPLETE_CHAPTER")
		return
	case common.MISSION_BUILDING:
		//query += "StructureAction = StructureAction - "
		query += "StructureAction = 0"
	case common.MISSION_PLACEMENT_MANSION_ITEM:
		//query += "MoveFurniture = MoveFurniture - "
		query += "MoveFurniture = 0"
	case common.MISSION_GET_MANSION_ITEM:
		//query += "GainFurniture = GainFurniture - "
		query += "GainFurniture = 0"
	default:
		log.Println("[DBAPI:UpdateUserQuestClear] Unknown MissinoType ", missionValue.MissionType)
		return
	}
	//query += strconv.Itoa(missionValue.MissionValue) + " WHERE AccountUID=" + strconv.Itoa(req.ID)
	query += " WHERE AccountUID=" + strconv.Itoa(req.ID)

	log.Println("Query : ", query)
	result, err := tx.Exec(query)
	if err != nil {
		log.Println("[DBAPI:UpdateUserQuestClear] Update UserQuestData Fail : " + err.Error())
		log.Println("Query : ", query)
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateUserQuestClear] Update UserQuestData Fail : " + GetUpdateError(err))
		log.Println("Query : ", query)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:UpdateUserQuestClear] Commit Fail : " + err.Error())
		return
	}

	log.Println("[DBAPI:UpdateUserQuestClear] Success UserQuest Data Clear. AccountUID(", req.ID, ")")
}

func UpdateUserGoldRoomOpen(req message.GoldRoomOpenRequest, mydb *sqlx.DB) (message.GoldRoomOpenResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.GoldRoomOpenResponse{}

	// 유저 방 오픈, 방은 항상 순차적으로 오픈되기 때문에 오픈 요청 방 번호 = 현재 방 +1 임..
	result, err := tx.Exec("Update UserMansion Set MansionRoom = ? Where AccountUID=? AND MansionRoom = ?-1", req.OpenRoom, req.ID, req.OpenRoom)
	if err != nil {
		log.Println("[DBAPI:UpdateUserGoldRoomOpen] Update UserQuestData Fail : " + err.Error())
		log.Println("Query : Update UserMansion Set MansionRoom = ", req.OpenRoom, " Where AccountUID=", req.ID, " AND MansionRoom = ", req.OpenRoom, "-1")
		err = fmt.Errorf("Update UserMansion Fail.")
		return rsp, err, 500
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateUserGoldRoomOpen] Update UserQuestData Fail : " + GetUpdateError(err))
		log.Println("Query : Update UserMansion Set MansionRoom = ", req.OpenRoom, " Where AccountUID=", req.ID, " AND MansionRoom = ", req.OpenRoom, "-1")
		err = fmt.Errorf("Update UserMansion Fail.")
		if rowsAffected == 0 {
			tx.Rollback()
			return rsp, err, 400
		} else {
			return rsp, err, 500
		}
	}

	// 골드 차감
	result, err = tx.Exec("Update User Set Gold = Gold - ?, UpdateDate=NOW(6) Where AccountUID=? ", req.Gold, req.ID)
	if err != nil {
		log.Println("[DBAPI:UpdateUserGoldRoomOpen] Update User Fail : " + err.Error())
		log.Println("Query : Update User Set Gold = Gold - ", req.Gold, ", UpdateDate=NOW(6) Where AccountUID=", req.ID)
		err = fmt.Errorf("Update User Fail.")
		return rsp, err, 500
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		log.Println("[DBAPI:UpdateUserGoldRoomOpen] Update User Fail : " + GetUpdateError(err))
		log.Println("Query : Update User Set Gold = Gold - ", req.Gold, ", UpdateDate=NOW(6) Where AccountUID=", req.ID)
		err = fmt.Errorf("Update User Fail.")
		if rowsAffected == 0 {
			tx.Rollback()
			return rsp, err, 400
		} else {
			return rsp, err, 500
		}
	}

	err = tx.Get(&rsp, "Select U.AccountUID as AccountUID, Gold, MansionRoom From User U INNER JOIN UserMansion M ON U.AccountUID=M.AccountUID Where U.AccountUID=?", req.ID)
	if err != nil {
		log.Println("[DBAPI:UpdateUserGoldRoomOpen] Select User & UserMansion Info Fail : " + err.Error())
		log.Println("Query : Select U.AccountUID, Gold, MansionRoom From User U INNER JOIN UserMansion M ON U.AccountUID=M.AccountUID Where U.AccountUID=", req.ID)

		if MyErrCode(err) == 4040 {
			err = fmt.Errorf("Not Found User & UserMansion Info.")
			return rsp, err, 400
		} else {
			err = fmt.Errorf("Select User & UserMansion Fail.")
			return rsp, err, 500
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:UpdateUserGoldRoomOpen] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func UpdateShopGame(req message.ShopGameRequest, mydb *sqlx.DB) (message.ShopGameResponse, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.ShopGameResponse{}

	// 1. Gold 값이 0이면 무료구매
	if req.Gold == 0 {
		err := tx.Get(&rsp, "Select G.AccountUID, ItemID, BuyTime, Gold from UserShopGame G INNER JOIN User U ON G.AccountUID = U.AccountUID Where G.AccountUID=? And ItemID=?", req.ID, req.ItemID)
		if err != nil { // 구매 이력이 없거나 에러일 경우
			if MyErrCode(err) != 4040 { // 미존재 빼고 다른 에러면 문제
				log.Println("[DBAPI:UpdateShopGame] Select UserShopGame Info Fail : " + err.Error())
				log.Println("Query : Select G.AccountUID, ItemID, BuyTime, Gold from UserShopGame G INNER JOIN User U ON G.AccountUID = U.AccountUID Where AccountUID=", req.ID, " And ItemID=", req.ItemID)
				err = fmt.Errorf("Select User & UserMansion Fail.")
				return rsp, err, 500
			} else { // 없으면 한번도 구매 한적 없으니 그냥 사면 됨
				// 구매이력 INSERT
				result, err := tx.Exec("Insert into UserShopGame Values(?, ?, NOW())", req.ID, req.ItemID)
				if err != nil {
					log.Println("[DBAPI:UpdateShopGame] Insert UserShopGame Fail : " + err.Error())
					log.Println("Query : Insert into UserShopGame Values(?, ?, NOW())", req.ID, req.ItemID)
					err = fmt.Errorf("Insert UserShopGame Fail.")
					return rsp, err, 500
				}
				rowsAffected, err := result.RowsAffected()
				if err != nil || rowsAffected == 0 {
					log.Println("[DBAPI:UpdateShopGame] Insert UserShopGame Fail : " + GetUpdateError(err))
					log.Println("Query : Insert into UserShopGame Values(?, ?, NOW())", req.ID, req.ItemID)
					err = fmt.Errorf("Insert UserShopGame Fail.")
					if rowsAffected == 0 {
						tx.Rollback()
						return rsp, err, 400
					} else {
						return rsp, err, 500
					}
				}
				// 유저 보유골드 및 아이템 번호 리턴
				err = tx.Get(&rsp, "Select G.AccountUID, ItemID, BuyTime, Gold from UserShopGame G INNER JOIN User U ON G.AccountUID = U.AccountUID Where G.AccountUID=? And ItemID=?", req.ID, req.ItemID)
				if err != nil { // 어떤 에러든 나오면 안됨
					log.Println("[DBAPI:UpdateShopGame] Select UserShopGame Info Fail : " + err.Error())
					log.Println("Query : Select G.AccountUID, ItemID, BuyTime, Gold from UserShopGame G INNER JOIN User U ON G.AccountUID = U.AccountUID Where AccountUID=", req.ID, " And ItemID=", req.ItemID)
					err = fmt.Errorf("Select User & UserMansion Fail.")
					return rsp, err, 500
				}
				err = fmt.Errorf("ShopGame Update Success")
				goto EndOfUpdateShopGame
			}
		} // 구매 이력이 없거나 에러 END

		// 무료 구매 시간 값 확인
		var freeAdTime int
		err = tx.Get(&freeAdTime, "Select FreeADTime From DataDB.ShopInfo S INNER JOIN DataDB.ShopInfoDetail D ON S.ShopDetailGroupNumber = D.GroupNumber "+
			" Where S.CostType = 1 AND D.ShopRewardValue = ? ", req.ItemID)
		if err != nil {
			log.Println("[DBAPI:UpdateShopGame] Select ShopInfo Fail : " + err.Error())
			log.Println("Query : Select FreeADTime From DataDB.ShopInfo S INNER JOIN DataDB.ShopInfoDetail D ON S.ShopDetailGroupNumber = D.GroupNumber "+
				" Where S.CostType = 1 AND D.ShopRewardValue =", req.ItemID)
			if MyErrCode(err) == 4040 {
				err = fmt.Errorf("Not Found Item Info in ShopInfo.")
				return rsp, err, 400
			} else {
				err = fmt.Errorf("Select ShopInfo Fail.")
				return rsp, err, 500
			}
		}

		// 구매 이력이 있으니 무료가 가능한지 확인
		if common.TimeCompare(rsp.BuyTime, time.Now(), common.TIME_COMPARE_OLDER, freeAdTime, common.TIME_HOUR) { // 구매이력 시간이 freeAdTime 보다 과거면 구매 가능
			// 구매시간 업데이트 후
			result, err := tx.Exec("Update UserShopGame Set BuyTime = Now() Where AccountUID=? AND ItemID=?", req.ID, req.ItemID)
			if err != nil {
				log.Println("[DBAPI:UpdateShopGame] Update UserShopGame Fail : " + err.Error())
				log.Println("Query : Update UserShopGame Set BuyTime = Now() Where AccountUID=", req.ID, " AND ItemID=", req.ItemID)
				err = fmt.Errorf("Update UpdateShopGame Fail.")
				return rsp, err, 500
			}
			rowsAffected, err := result.RowsAffected()
			if err != nil || rowsAffected == 0 {
				log.Println("[DBAPI:UpdateShopGame] Update User Fail : " + GetUpdateError(err))
				log.Println("Query : Update UserShopGame Set BuyTime = Now() Where AccountUID=", req.ID, " AND ItemID=", req.ItemID)
				err = fmt.Errorf("Update UpdateShopGame Fail.")
				if rowsAffected == 0 {
					tx.Rollback()
					return rsp, err, 400
				} else {
					return rsp, err, 500
				}
			}
			// ITEM 리턴
			err = tx.Get(&rsp, "Select G.AccountUID, ItemID, BuyTime, Gold from UserShopGame G INNER JOIN User U ON G.AccountUID = U.AccountUID Where G.AccountUID=? And ItemID=?", req.ID, req.ItemID)
			if err != nil { // 어떤 에러든 나오면 안됨
				log.Println("[DBAPI:UpdateShopGame] Select UserShopGame Info Fail : " + err.Error())
				log.Println("Query : Select G.AccountUID, ItemID, BuyTime, Gold from UserShopGame G INNER JOIN User U ON G.AccountUID = U.AccountUID Where AccountUID=", req.ID, " And ItemID=", req.ItemID)
				err = fmt.Errorf("Select User & UserMansion Fail.")
				return rsp, err, 500
			}
			err = fmt.Errorf("ShopGame Update Success")
			goto EndOfUpdateShopGame
		} else { // // 구매이력 시간이 freeAdTime 보다 과거가 아니면 에러처리
			err = fmt.Errorf("Can't Free Buying. Cause Not Exceed Free Time")
			return rsp, err, 400
		}
	} else { // 2. 골드 값이 0이 아니라면 골드 차감 후 아이템 리턴
		// 골드 차감
		result, err := tx.Exec("Update User Set Gold = Gold - ?, UpdateDate=NOW(6) Where AccountUID=? ", req.Gold, req.ID)
		if err != nil {
			log.Println("[DBAPI:UpdateShopGame] Update User Fail : " + err.Error())
			log.Println("Query : Update User Set Gold = Gold - ", req.Gold, ", UpdateDate=NOW(6) Where AccountUID = ", req.ID)
			err = fmt.Errorf("Update User Fail.")
			return rsp, err, 500
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected == 0 {
			log.Println("[DBAPI:UpdateShopGame] Update User Fail : " + err.Error())
			log.Println("Query : Update User Set Gold = Gold - ", req.Gold, ", UpdateDate=NOW(6) Where AccountUID = ", req.ID)
			err = fmt.Errorf("Update User Fail.")
			if rowsAffected == 0 {
				tx.Rollback()
				return rsp, err, 400
			} else {
				return rsp, err, 500
			}
		}
		// ITEM 리턴
		err = tx.Get(&rsp, "Select G.AccountUID, ItemID, BuyTime, Gold from UserShopGame G INNER JOIN User U ON G.AccountUID = U.AccountUID Where G.AccountUID=? And ItemID=?", req.ID, req.ItemID)
		if err != nil { // 어떤 에러든 나오면 안됨
			log.Println("[DBAPI:UpdateShopGame] Select UserShopGame Info Fail : " + err.Error())
			log.Println("Query : Select G.AccountUID, ItemID, BuyTime, Gold from UserShopGame G INNER JOIN User U ON G.AccountUID = U.AccountUID Where AccountUID=", req.ID, " And ItemID=", req.ItemID)
			err = fmt.Errorf("Select User & UserMansion Fail.")
			return rsp, err, 500
		}
		err = fmt.Errorf("ShopGame Update Success")
	}

EndOfUpdateShopGame:
	err := tx.Commit()
	if err != nil {
		log.Println("[DBAPI:UpdateShopGame] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}

func SelectShopGameList(id int, mydb *sqlx.DB) (message.UserShopGameList, error, int) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	// mysql connect
	tx := mydb.MustBegin()
	defer tx.Rollback() // 중간에 에러가 나면 롤백

	rsp := message.UserShopGameList{}
	list := []message.UserShopGameInfo{}

	err := tx.Select(&list, "Select ItemID, BuyTime from UserShopGame Where AccountUID=? ", id)
	if err != nil {
		log.Println("[DBAPI:SelectShopGameList] Select UserShopGame Info Fail : " + err.Error())
		log.Println("Query : Select ItemID, BuyTime from UserShopGame Where AccountUID=", id)
		if MyErrCode(err) == 4040 {
			err := fmt.Errorf("Not Found UserShopGame Info.")
			return rsp, err, 400
		} else {
			err = fmt.Errorf("Select UserShopGame Fail.")
			return rsp, err, 500
		}
	}
	rsp.ShopGameList = list

	err = tx.Commit()
	if err != nil {
		log.Println("[DBAPI:SelectShopGameList] Commit Fail : " + err.Error())
		err = fmt.Errorf("DB Commit Error.")
		return rsp, err, 500
	}

	return rsp, err, 200
}
