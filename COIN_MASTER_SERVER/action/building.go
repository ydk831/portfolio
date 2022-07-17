package action

import (
	"common"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"github.com/jmoiron/sqlx"
)

type ResBuilding struct {
	ResultCode     int32 `json:"result_code"`
	IsChapterClear int8  `json:"is_chapter_clear"`
	TileIdx        int8  `json:"tile_index"`
	Gold           int64 `json:"gold"`
	Spin           int16 `json:"spin"`
	MaxSpin        int16 `json:"max_spin"`
	Shield         int8  `json:"shield"`
	RewardType1    int   `json:"reward_type1" db:"RewardType1"`
	RewardCount1   int   `json:"reward_count1" db:"RewardCount1"`
	RewardType2    int   `json:"reward_type2" db:"RewardType2"`
	RewardCount2   int   `json:"reward_count2" db:"RewardCount2"`
}

type DBParamTile struct {
	TileIdx        int8      `db:"TileIdx"`
	TileStatus     int8      `db:"TileStatus"`
	TileChargeTime time.Time `db:"TileChargeTime"`
}

type DBParamBuilding struct {
	AccountUID           uint32    `db:"AccountUID"`
	Name                 string    `db:"Name"`
	Gold                 int64     `db:"Gold"`
	Spin                 int16     `db:"Spin"`
	MaxSpin              int16     `db:"MaxSpin"`
	LastSpinRechargeTime time.Time `db:"LastSpinRechargeTime"`
	Shield               int8      `db:"Shield"`
	CurrentChapterIDX    int8      `db:"CurrentChapterIDX"`
	Now                  time.Time `db:"Now"`
	RefairBuildingCount  int8      `db:"RefairBuildingCount"`
}

type DataTilePassive struct {
	TileIdx               int8 `db:"TileOrder"`
	TileRequireGold       int  `db:"TileRequireGold"`
	BuildingRequireGold   int  `db:"BuildingRequireGold"`
	BuildingFixGold       int  `db:"BuildingFixGold"`
	BuildingRewardMaxSpin int  `db:"BuildingRewardMaxSpin"`
}

type BuildingResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (BuildingResource) Uri() string {
	return "/building"
}

func (BuildingResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (BuildingResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	packetdata := make(map[string]interface{})
	err := ReadOnJSONBody(r, &packetdata)
	//fmt.Println(body)

	log.Println("building : Request Data  (", packetdata, ")")
	if err != nil {
		log.Println("building : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "Building : " + err.Error(), Data: nil}
	}

	resBuilding := ResBuilding{}
	resBuilding.ResultCode = 0

	// 유저 정보
	userinfo := DBParamBuilding{}
	err = mydb.Get(&userinfo, "select AccountUID, Name, Gold, Spin, MaxSpin, LastSpinRechargeTime, Shield, CurrentChapterIDX, Now() Now, ( select count(*) from UserTile where AccountUID = ? and TileStatus = 3) RefairBuildingCount FROM User WHERE AccountUID = ?;", packetdata["acc_uid"], packetdata["acc_uid"])
	if err != nil {
		log.Println("Select User Table Fail : ", err.Error())
		log.Println("Query : select AccountUID,	Name, Gold, Spin, MaxSpin, LastSpinRechargeTime, Shield, CurrentChapterIDX, Now() Now, ( select count(*) from UserTile where AccountUID = ", packetdata["acc_uid"], " and TileStatus = 3) RefairBuildingCount FROM User WHERE AccountUID = ", packetdata["acc_uid"])
		err = fmt.Errorf("Building : Find user info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 유저 타일 정보
	isTile := 0
	err = mydb.Get(&isTile, "select count(*) from UserTile where AccountUID = ? and TileIdx = ? and TileStatus = 1;", packetdata["acc_uid"], packetdata["tile_index"])
	if err != nil {
		log.Println("Select UserTile Table Fail : ", err.Error())
		log.Println("Query : select count(*) from UserTile where AccountUID = ", packetdata["acc_uid"], " and TileIdx = ", packetdata["tile_index"], " and TileStatus = 1")
		err = fmt.Errorf("Building : Find tile info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	if isTile == 0 {
		log.Println("Select Success. But No Count Tile Info..")
		log.Println("Query : select count(*) from UserTile where AccountUID = ", packetdata["acc_uid"], " and TileIdx = ", packetdata["tile_index"], " and TileStatus = 1")
		return httpPkg.Response{Code: 400, Msg: "Building : You must buy tile before construction.", Data: nil}
	}

	// 타일 구매에 필요한 정보
	datatile := DataTilePassive{}
	err = mydb.Get(&datatile, "select TileOrder, TileRequireGold, BuildingRequireGold, BuildingFixGold, BuildingRewardMaxSpin from DataDB.ChapterMapTileInfo where MapTileGroupIDX = ? and TileOrder = ?;", userinfo.CurrentChapterIDX, packetdata["tile_index"])
	if err != nil {
		log.Println("Select DataDB.ChapterMapTileInfo Table Fail : ", err.Error())
		log.Println("Query : select TileOrder, TileRequireGold, BuildingRequireGold, BuildingFixGold, BuildingRewardMaxSpin from DataDB.ChapterMapTileInfo where MapTileGroupIDX = ", userinfo.CurrentChapterIDX, " and TileOrder = ", packetdata["tile_index"])
		err = fmt.Errorf("Building : Find chapter info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 골드 소모
	// 파라미터 값 체크
	if datatile.BuildingRequireGold < 0 || userinfo.Gold-int64(datatile.BuildingRequireGold) < 0 {
		resBuilding.ResultCode = 1
		log.Println("Not Enough Gold. User : ", userinfo.AccountUID, ", Gold : ", userinfo.Gold, ", Building Cost : ", datatile.BuildingRequireGold)
		return httpPkg.Response{Code: 400, Msg: "Building : Not Enough Gold.", Data: resBuilding}
	}

	userinfo.Gold -= int64(datatile.BuildingRequireGold)

	// 임시로 사용할 챕터 별 건설해야 할 타일 개수
	needBuilding := 0
	err = mydb.Get(&needBuilding, "select count(*) from DataDB.ChapterMapTileInfo where MapTileGroupIDX = ?;", userinfo.CurrentChapterIDX)
	if err != nil {
		log.Println("Select DataDB.ChapterMapTileInfo Fail : ", err.Error())
		log.Println("Query : select count(*) from DataDB.ChapterMapTileInfo where MapTileGroupIDX = ", userinfo.CurrentChapterIDX)
		err = fmt.Errorf("Building : Find chapter info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 다음 챕터로 넘길지 확인
	if needBuilding == int(packetdata["tile_index"].(float64)) && userinfo.RefairBuildingCount == int8(0) {
		resBuilding.IsChapterClear = 1

		// 챕터 클리어 보상
		err = mydb.Get(&resBuilding, "call spGetChapterClearReward(?);", userinfo.CurrentChapterIDX)
		if err != nil {
			log.Println("Failed to call spGetChapterClearReward : ", err.Error())
			return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
		}

		common.AddGoodsFromAllType(resBuilding.RewardType1, resBuilding.RewardCount1, &userinfo.Gold, &userinfo.Spin, &userinfo.Shield)
		common.AddGoodsFromAllType(resBuilding.RewardType2, resBuilding.RewardCount2, &userinfo.Gold, &userinfo.Spin, &userinfo.Shield)
	} else {
		resBuilding.IsChapterClear = 0
	}

	userinfo.MaxSpin = int16(50) + int16(datatile.BuildingRewardMaxSpin)

	// update
	result, err := mydb.Exec("call spProcessBuildingOrRefair(?,?,?,?,?,?,?,?);", packetdata["acc_uid"], resBuilding.IsChapterClear, packetdata["tile_index"], userinfo.Gold, userinfo.Spin, userinfo.Shield, userinfo.MaxSpin, userinfo.LastSpinRechargeTime)
	if err != nil {
		log.Println("failed to spProcessBuildingOrRefair : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	rowsAffected, err := result.RowsAffected()

	if rowsAffected == 0 {
		log.Println("call spProcessBuilding RowsAffected 0 rows! ", packetdata)
		return httpPkg.Response{Code: 204, Msg: "success", Data: nil}
	}

	resBuilding.TileIdx = int8(packetdata["tile_index"].(float64))
	resBuilding.Gold = userinfo.Gold
	resBuilding.Spin = userinfo.Spin
	resBuilding.MaxSpin = userinfo.MaxSpin
	resBuilding.Shield = userinfo.Shield

	log.Println("building : Response Data  (", resBuilding, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: resBuilding}
}
