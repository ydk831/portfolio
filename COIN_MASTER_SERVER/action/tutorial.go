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

type DBTutorial struct {
	Gold                 int64     `json:"gold" db:"Gold"`
	Spin                 int16     `json:"spin" db:"Spin"`
	MaxSpin              int16     `json:"max_spin" db:"MaxSpin"`
	LastSpinRechargeTime time.Time `json:"spin_recharge_time" db:"LastSpinRechargeTime"`
	Tutorial             int8      `json:"tutorial" db:"Tutorial"`
	RewardType1          int       `json:"reward_type1" db:"DBRewardType1"`
	RewardValue1         int       `json:"reward_value1" db:"DBRewardValue1"`
	RewardType2          int       `json:"reward_type2" db:"DBRewardType2"`
	RewardValue2         int       `json:"reward_value2" db:"DBRewardValue2"`
	Now                  time.Time `json:"now" db:"Now"`
}

type ResTutorial struct {
	Gold                 int64     `json:"gold"`
	Spin                 int16     `json:"spin"`
	LastSpinRechargeTime time.Time `json:"spin_recharge_time"`
	Now                  time.Time `json:"now"`
	TutorialUniqueID     int8      `json:"tutorial_uid"`
	RewardType1          int       `json:"reward_type1"`
	RewardValue1         int       `json:"reward_value1"`
	RewardType2          int       `json:"reward_type2"`
	RewardValue2         int       `json:"reward_value2"`
}

type TutorialResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (TutorialResource) Uri() string {
	return "/tutorial"
}

func (TutorialResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (TutorialResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	packetdata := make(map[string]interface{})
	err := ReadOnJSONBody(r, &packetdata)

	log.Println("Tutorial : Request Data (", packetdata, ")")

	if err != nil {
		log.Println("Tutorial : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	// 현재 튜토리얼 확인
	userinfo := DBTutorial{}
	err = db.Get(&userinfo, "call spGetTutorialReward(?,?);", packetdata["acc_uid"], packetdata["tutorial_uid"])
	if err != nil {
		log.Println("invalid acc_uid : ", packetdata["acc_uid"])
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	if userinfo.Tutorial >= int8(packetdata["tutorial_uid"].(float64)) {
		return httpPkg.Response{Code: 204, Msg: "Already Completed Tutorial.", Data: nil}
	}

	beforeGold := userinfo.Gold
	beforeSpin := userinfo.Spin

	// 첫번째보상
	if userinfo.RewardType1 == common.GoodsTypeGold {
		userinfo.Gold = common.AddGoods(common.GoodsTypeGold, int64(userinfo.RewardValue1), userinfo.Gold).(int64)
	} else if userinfo.RewardType1 == common.GoodsTypeSpin {
		if userinfo.Spin < userinfo.MaxSpin {
			userinfo.Spin, userinfo.LastSpinRechargeTime = common.CalcRechargeSP(userinfo.Spin, userinfo.MaxSpin, userinfo.Now, userinfo.LastSpinRechargeTime)
		}

		userinfo.Spin = common.AddGoods(common.GoodsTypeSpin, int16(userinfo.RewardValue1), userinfo.Spin).(int16)
	} else if userinfo.RewardType1 == -1 {
		// 보상 정보 없음
		//log.Println("invalid reward data! ", packetdata)
		//return httpPkg.Response{Code: 204, Msg: "success", Data: nil}
	} else {
		log.Println("invalid reward data! ", packetdata)
		return httpPkg.Response{Code: 204, Msg: "success", Data: nil}
	}

	// 두번째 보상
	if userinfo.RewardType2 == common.GoodsTypeGold {
		userinfo.Gold = common.AddGoods(common.GoodsTypeGold, int64(userinfo.RewardValue2), userinfo.Gold).(int64)
	} else if userinfo.RewardType2 == common.GoodsTypeSpin {
		if userinfo.Spin < userinfo.MaxSpin {
			userinfo.Spin, userinfo.LastSpinRechargeTime = common.CalcRechargeSP(userinfo.Spin, userinfo.MaxSpin, userinfo.Now, userinfo.LastSpinRechargeTime)
		}

		userinfo.Spin = common.AddGoods(common.GoodsTypeSpin, int16(userinfo.RewardValue2), userinfo.Spin).(int16)
	} else if userinfo.RewardType2 == -1 {
		// 보상 정보 없음
		//log.Println("invalid reward data! ", packetdata)
		//return httpPkg.Response{Code: 204, Msg: "success", Data: nil}
	} else {
		log.Println("invalid reward data! ", packetdata)
		return httpPkg.Response{Code: 204, Msg: "success", Data: nil}
	}

	// 유저 정보 업데이트
	result, err := db.Exec("update User set Gold = ?, Spin = ?, LastSpinRechargeTime = ?, Tutorial = ? where AccountUID = ? and Gold = ? and Spin = ? and Tutorial = ?;",
		userinfo.Gold, userinfo.Spin, userinfo.LastSpinRechargeTime, packetdata["tutorial_uid"], packetdata["acc_uid"], beforeGold, beforeSpin, userinfo.Tutorial)
	if err != nil {
		log.Println("failed to tutorial complete : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	rowsAffected, err := result.RowsAffected()

	if rowsAffected == 0 {
		log.Println("Tutorial Complete Request RowsAffected 0 rows! acc_uid : ", packetdata["acc_uid"])
		return httpPkg.Response{Code: 204, Msg: "success", Data: nil}
	}

	log.Println("Tutorial : Response Data (", ResTutorial{userinfo.Gold, userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.Now, int8(packetdata["tutorial_uid"].(float64)), userinfo.RewardType1, userinfo.RewardValue1, userinfo.RewardType2, userinfo.RewardValue2}, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: ResTutorial{userinfo.Gold, userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.Now, int8(packetdata["tutorial_uid"].(float64)), userinfo.RewardType1, userinfo.RewardValue1, userinfo.RewardType2, userinfo.RewardValue2}}
}
