package action

import (
	"common"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"github.com/jmoiron/sqlx"
)

type DBBuildingReward struct {
	Gold                         int64     `db:"Gold"`
	Spin                         int16     `db:"Spin"`
	MaxSpin                      int16     `db:"MaxSpin"`
	LastSpinRechargeTime         time.Time `db:"LastSpinRechargeTime"`
	RewardGold                   int64     `db:"RewardGold"`
	RewardSpin                   int16     `db:"RewardSpin"`
	BuildingRewardGoldUpdateDate time.Time `db:"BuildingRewardGoldUpdateDate"`
	BuildingRewardSpinUpdateDate time.Time `db:"BuildingRewardSpinUpdateDate"`
	Now                          time.Time `db:"Now"`
}

type ResBuildingReward struct {
	IsReward                     int8      `json:"is_reward"`
	RewardGold                   int64     `json:"reward_gold"`
	RewardSpin                   int16     `json:"reward_spin"`
	Gold                         int64     `json:"current_gold"`
	Spin                         int16     `json:"current_spin"`
	LastSpinRechargeTime         time.Time `json:"spin_recharge_time"`
	BuildingRewardGoldUpdateDate time.Time `json:"building_reward_gold_date"`
	BuildingRewardSpinUpdateDate time.Time `json:"building_reward_spin_date"`
}

type BuildingRewardResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (BuildingRewardResource) Uri() string {
	return "/buildingreward"
}

func (BuildingRewardResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (BuildingRewardResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	packetdata := make(map[string]interface{})
	err := ReadOnJSONBody(r, &packetdata)
	//fmt.Println(body)

	log.Println("buildingReward : Request Data  (", packetdata, ")")
	if err != nil {
		log.Println("buildingReward : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "BuildingReward : " + err.Error(), Data: nil}
	}

	// 챕터 보상 정보
	userinfo := DBBuildingReward{}
	err = db.Get(&userinfo, "call spGetBuildingReward(?);", packetdata["acc_uid"])
	if err != nil {
		log.Println("call spGetBuildingReward Fail : ", err.Error())
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 보상이 있는지 확인
	if userinfo.RewardGold == 0 && userinfo.RewardSpin == 0 {
		return httpPkg.Response{Code: 200, Msg: "success", Data: ResBuildingReward{0, 0, 0, userinfo.Gold, userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.BuildingRewardGoldUpdateDate, userinfo.BuildingRewardSpinUpdateDate}}
	}

	beforeGold := userinfo.Gold
	beforeSpin := userinfo.Spin

	// SP 회복 가능
	if userinfo.Spin < userinfo.MaxSpin {
		userinfo.Spin, userinfo.LastSpinRechargeTime = common.CalcRechargeSP(userinfo.Spin, userinfo.MaxSpin, userinfo.Now, userinfo.LastSpinRechargeTime)
	}

	userinfo.Gold = common.AddGoods(common.GoodsTypeGold, userinfo.RewardGold, userinfo.Gold).(int64)
	userinfo.Spin = common.AddGoods(common.GoodsTypeSpin, userinfo.RewardSpin, userinfo.Spin).(int16)

	// 보상 지급
	result, err := db.Exec("call spProcessBuildingReward(?,?,?,?,?,?,?,?);", packetdata["acc_uid"], beforeGold, beforeSpin, userinfo.Gold, userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.BuildingRewardGoldUpdateDate, userinfo.BuildingRewardSpinUpdateDate)
	if err != nil {
		log.Println("call spProcessBuildingReward Fail : ", err.Error())
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	RowsAffected, err := result.RowsAffected()

	if RowsAffected == 0 {
		log.Println("call spProcessBuildingReward RowsAffected 0 rows!")
		return httpPkg.Response{Code: 204, Msg: "success", Data: ResBuildingReward{0, userinfo.RewardGold, userinfo.RewardSpin, userinfo.Gold, userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.BuildingRewardGoldUpdateDate, userinfo.BuildingRewardSpinUpdateDate}}
	}

	log.Println("buildingReward : Response Data  (", ResBuildingReward{1, userinfo.RewardGold, userinfo.RewardSpin, userinfo.Gold, userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.BuildingRewardGoldUpdateDate, userinfo.BuildingRewardSpinUpdateDate}, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: ResBuildingReward{1, userinfo.RewardGold, userinfo.RewardSpin, userinfo.Gold, userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.BuildingRewardGoldUpdateDate, userinfo.BuildingRewardSpinUpdateDate}}
}
