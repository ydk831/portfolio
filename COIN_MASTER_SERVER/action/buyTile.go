package action

import (
	"fmt"
	"log"
	"message"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"github.com/jmoiron/sqlx"
)

type ResBuyTile struct {
	TileIdx int8
	Gold    int64
}

type DataTile struct {
	TileIdx             int8 `db:"TileOrder"`
	TileRequireGold     int  `db:"TileRequireGold"`
	BuildingRequireGold int  `db:"BuildingRequireGold"`
	BuildingFixGold     int  `db:"BuildingFixGold"`
}

type BuyTileResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (BuyTileResource) Uri() string {
	return "/buytile"
}

func (BuyTileResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (BuyTileResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	packetdata := make(map[string]interface{})
	err := ReadOnJSONBody(r, &packetdata)
	//fmt.Println(body)

	log.Println("buyTile : Request Data  (", packetdata, ")")
	if err != nil {
		log.Println("buyTile : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "BuyTile : " + err.Error(), Data: nil}
	}

	// 유저 정보
	userinfo := message.DBParamUser{}
	err = db.Get(&userinfo, "select AccountUID,	Name, Gold, Spin, MaxSpin, LastSpinRechargeTime, Shield, CurrentChapterIDX, Now() Now FROM User WHERE AccountUID = ?;", packetdata["acc_uid"])
	if err != nil {
		log.Println("Select User Fail : ", err.Error())
		log.Println("Query : select AccountUID,	Name, Gold, Spin, MaxSpin, LastSpinRechargeTime, Shield, CurrentChapterIDX, Now() Now FROM User WHERE AccountUID = ", packetdata["acc_uid"])
		err = fmt.Errorf("BuyTile : Find user info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 타일 구매가 가능한지 확인
	maxTileIndex := 0
	err = db.Get(&maxTileIndex, "select ifnull(max(TileIdx), 0) from UserTile where AccountUID = ? and TileStatus between 1 and 2;", packetdata["acc_uid"])
	if err != nil {
		log.Println("Select UserTile Fail : ", err.Error())
		log.Println("Query : select count(*) from UserTile where AccountUID = ", packetdata["acc_uid"], " and TileStatus between 1 and 2;")
		err = fmt.Errorf("BuyTile : Find chapter info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	if packetdata["tile_index"] != float64(maxTileIndex+1) {
		log.Println("tile_index mismatch. tile_index : ", packetdata["tile_index"], ", user's tile count : ", maxTileIndex+1)
		log.Println("Query : select count(*) from UserTile where AccountUID = ", packetdata["acc_uid"], " and TileStatus between 1 and 2;")
		err = fmt.Errorf("BuyTile : tile index dosen't match. Your index(%v), Server index(%d)", packetdata["tile_index"], maxTileIndex)
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 수리가 필요한 건물이 있는지 확인
	// needRefair := 0
	// err = db.Get(&needRefair, "select count(*) from UserTile where AccountUID = ? and TileStatus = 3;", packetdata["acc_uid"])
	// if err != nil {
	// 	log.Println("Failed to Request! invalid UserTile Table : ", err.Error())
	// 	return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	// }

	// if needRefair != 0 {
	// 	log.Println("Failed to Request! Need refair building!")
	// 	return httpPkg.Response{Code: 400, Msg: "Failed to Request! need refair building!", Data: nil}
	// }

	buyTileIndex := int8(maxTileIndex + 1)

	// 타일 구매에 필요한 정보
	datatile := DataTile{}
	err = db.Get(&datatile, "select TileOrder, TileRequireGold, BuildingRequireGold, BuildingFixGold from DataDB.ChapterMapTileInfo where MapTileGroupIDX = ? and TileOrder = ?;", userinfo.CurrentChapterIDX, buyTileIndex)
	if err != nil {
		log.Println("Select DataDB.ChapterMapTileInfo Fail : ", err.Error())
		log.Println("Query : select TileOrder, TileRequireGold, BuildingRequireGold, BuildingFixGold from DataDB.ChapterMapTileInfo where MapTileGroupIDX = ", userinfo.CurrentChapterIDX, " and TileOrder = ", buyTileIndex)
		err = fmt.Errorf("BuyTile : Not found chapter info.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 골드 소모
	// 파라미터 값 체크
	if datatile.TileRequireGold < 0 || userinfo.Gold-int64(datatile.TileRequireGold) < 0 {
		log.Println("Not Enough Gold. User : ", userinfo.AccountUID, ", Gold : ", userinfo.Gold, ", Building Cost : ", datatile.TileRequireGold)
		return httpPkg.Response{Code: 400, Msg: "BuyTile : Not Enough Gold.", Data: nil}
	}

	// Spin 및 접속 정보 갱신
	result, err := db.Exec("update User Set Gold = Gold - ?, UpdateDate = NOW(6) where AccountUID = ? and Gold + ? >= 0;", datatile.TileRequireGold, userinfo.AccountUID, datatile.TileRequireGold)
	if err != nil {
		log.Println("Update User Fail : ", err.Error())
		log.Println("Query : update User Set Gold = Gold - ", datatile.TileRequireGold, ", UpdateDate = NOW(6)", " where AccountUID = ", userinfo.AccountUID, " and Gold + ", datatile.TileRequireGold, " >= 0;")
		err = fmt.Errorf("BuyTile : Update user info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		if err != nil {
			log.Println("Update User Fail(RowsAffected) : ", err.Error())
		} else {
			log.Println("Update User Fail(RowsAffected) : No Rows Update")
		}
		log.Println("Query : update User Set Gold = Gold - ", datatile.TileRequireGold, ", UpdateDate = NOW(6)", " where AccountUID = ", userinfo.AccountUID, " and Gold + ", datatile.TileRequireGold, " >= 0;")
		err = fmt.Errorf("BuyTile : Update user info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// Tile 테이블 정보 수정 또는 추가
	result, err = db.Exec("insert into UserTile(AccountUID,TileIdx,TileStatus,TileChargeTime) values(?,?,1,0) ON DUPLICATE KEY UPDATE TileStatus=1, TileChargeTime=0;", userinfo.AccountUID, buyTileIndex)
	if err != nil {
		log.Println("Insert UserTile Fail : ", err.Error())
		log.Println("Query : insert into UserTile(AccountUID,TileIdx,TileStatus,TileChargeTime) values(", userinfo.AccountUID, ",", buyTileIndex, ",1,0) ON DUPLICATE KEY UPDATE TileStatus=1, TileChargeTime=0;")
		err = fmt.Errorf("BuyTile : Update tile info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		if err != nil {
			log.Println("Update UserTile Fail(RowsAffected) : ", err.Error())
		} else {
			log.Println("Update UserTile Fail(RowsAffected) : No Rows Update")
		}
		log.Println("Query : insert into UserTile(AccountUID,TileIdx,TileStatus,TileChargeTime) values(", userinfo.AccountUID, ",", buyTileIndex, ",1,0) ON DUPLICATE KEY UPDATE TileStatus=1, TileChargeTime=0;")
		err = fmt.Errorf("BuyTile : Update tile info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	userinfo.Gold -= int64(datatile.TileRequireGold)

	log.Println("buyTile : Response Data  (", ResBuyTile{buyTileIndex, userinfo.Gold}, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: ResBuyTile{buyTileIndex, userinfo.Gold}}
}
