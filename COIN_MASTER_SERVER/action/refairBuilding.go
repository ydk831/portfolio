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

type ResRefairBuilding struct {
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

type DBParamRefairBuilding struct {
	AccountUID           uint32    `db:"AccountUID"`
	Name                 string    `db:"Name"`
	Gold                 int64     `db:"Gold"`
	Spin                 int16     `db:"Spin"`
	MaxSpin              int16     `db:"MaxSpin"`
	LastSpinRechargeTime time.Time `db:"LastSpinRechargeTime"`
	Shield               int8      `db:"Shield"`
	CurrentChapterIDX    int8      `db:"CurrentChapterIDX"`
	Now                  time.Time `db:"Now"`
	BuildingCount        int8      `db:"BuildingCount"`
}

type RefairBuildingResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (RefairBuildingResource) Uri() string {
	return "/refairbuilding"
}

func (RefairBuildingResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (RefairBuildingResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	packetdata := make(map[string]interface{})
	err := ReadOnJSONBody(r, &packetdata)

	log.Println("RefairBuilding : Request Data (", packetdata, ")")

	if err != nil {
		log.Println("RefairBuilding : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "RefairBuilding : " + err.Error(), Data: nil}
	}

	// 유저 정보
	userinfo := DBParamRefairBuilding{}
	err = db.Get(&userinfo, "select AccountUID,	Name, Gold, Spin, MaxSpin, LastSpinRechargeTime, Shield, CurrentChapterIDX, Now() Now, (select count(*) from UserTile where AccountUID = ? and TileStatus = 2) BuildingCount FROM User WHERE AccountUID = ?;", packetdata["acc_uid"], packetdata["acc_uid"])
	if err != nil {
		log.Println("Select User Fail : ", err.Error())
		log.Println("Query : select AccountUID,	Name, Gold, Spin, MaxSpin, LastSpinRechargeTime, Shield, CurrentChapterIDX, Now() Now, (select count(*) from UserTile where AccountUID = ", packetdata["acc_uid"], " and TileStatus = 2) BuildingCount FROM User WHERE AccountUID = ", packetdata["acc_uid"])
		err = fmt.Errorf("RefairBuilding : Find user info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 유저 타일 정보
	isTile := 0
	err = db.Get(&isTile, "select count(*) from UserTile where AccountUID = ? and TileIdx = ? and TileStatus = 3;", packetdata["acc_uid"], packetdata["tile_index"])
	if err != nil {
		log.Println("Select UserTile Fail : ", err.Error())
		log.Println("Query : select count(*) from UserTile where AccountUID = ", packetdata["acc_uid"], " and TileIdx = ", packetdata["tile_index"], " and TileStatus = 3;")
		err = fmt.Errorf("RefairBuilding : Find tile info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	if isTile == 0 {
		log.Println("Query : select count(*) from UserTile where AccountUID = ", packetdata["acc_uid"], " and TileIdx = ", packetdata["tile_index"], " and TileStatus = 3;")
		err = fmt.Errorf("RefairBuilding : Not Found Refair Tile Info.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	//타일 구매에 필요한 정보
	datatile := DataTile{}
	err = db.Get(&datatile, "select TileOrder, TileRequireGold, BuildingRequireGold, BuildingFixGold from DataDB.ChapterMapTileInfo where MapTileGroupIDX = ? and TileOrder = ?;", userinfo.CurrentChapterIDX, packetdata["tile_index"])
	if err != nil {
		log.Println("Select DataDB.ChapterMapTileInfo Fail : ", err.Error())
		log.Println("Query : select TileOrder, TileRequireGold, BuildingRequireGold, BuildingFixGold from DataDB.ChapterMapTileInfo where MapTileGroupIDX = ", userinfo.CurrentChapterIDX, " and TileOrder = ", packetdata["tile_index"])
		err = fmt.Errorf("RefairBuilding : Find chapter info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 골드 소모
	// 파라미터 값 체크
	if datatile.BuildingFixGold < 0 || userinfo.Gold-int64(datatile.BuildingFixGold) < 0 {
		log.Println("Not Enough Gold. User : ", userinfo.AccountUID, ", Gold : ", userinfo.Gold, ", BuildingFixGold : ", datatile.BuildingFixGold)
		err = fmt.Errorf("RefairBuilding : Not enough gold")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	userinfo.Gold -= int64(datatile.BuildingFixGold)

	// 임시로 사용할 챕터 별 건설해야 할 타일 개수
	needBuilding := 0
	err = db.Get(&needBuilding, "select count(*) from DataDB.ChapterMapTileInfo where MapTileGroupIDX = ?;", userinfo.CurrentChapterIDX)
	if err != nil {
		log.Println("Select DataDB.ChapterMapTileInfo Fail : ", err.Error())
		log.Println("Query : select count(*) from DataDB.ChapterMapTileInfo where MapTileGroupIDX = ", userinfo.CurrentChapterIDX)
		err = fmt.Errorf("RefairBuilding : Find chapter info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	resRefairBuilding := ResRefairBuilding{}
	if int8(needBuilding) == (userinfo.BuildingCount + 1) {
		resRefairBuilding.IsChapterClear = 1

		// 챕터 클리어 보상
		err = db.Get(&resRefairBuilding, "call spGetChapterClearReward(?);", userinfo.CurrentChapterIDX)
		if err != nil {
			log.Println("Failed to call spGetChapterClearReward : ", err.Error())
			return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
		}

		common.AddGoodsFromAllType(resRefairBuilding.RewardType1, resRefairBuilding.RewardCount1, &userinfo.Gold, &userinfo.Spin, &userinfo.Shield)
		common.AddGoodsFromAllType(resRefairBuilding.RewardType2, resRefairBuilding.RewardCount2, &userinfo.Gold, &userinfo.Spin, &userinfo.Shield)
	} else {
		resRefairBuilding.IsChapterClear = 2
	}

	// update
	result, err := db.Exec("call spProcessBuildingOrRefair(?,?,?,?,?,?,?,?);", packetdata["acc_uid"], resRefairBuilding.IsChapterClear, packetdata["tile_index"], userinfo.Gold, userinfo.Spin, userinfo.Shield, userinfo.MaxSpin, userinfo.LastSpinRechargeTime)
	if err != nil {
		log.Println("failed to spProcessBuildingOrRefair : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	rowsAffected, err := result.RowsAffected()

	if rowsAffected == 0 {
		log.Println("call spProcessBuilding RowsAffected 0 rows! ", packetdata)
		return httpPkg.Response{Code: 204, Msg: "success", Data: nil}
	}

	resRefairBuilding.TileIdx = int8(packetdata["tile_index"].(float64))
	resRefairBuilding.Gold = userinfo.Gold
	resRefairBuilding.Spin = userinfo.Spin
	resRefairBuilding.Shield = userinfo.Shield

	log.Println("RefairBuilding : Response Data (", resRefairBuilding, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: resRefairBuilding}
}
