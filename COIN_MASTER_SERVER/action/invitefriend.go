package action

import (
	"common"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"

	"httpPkg"
)

type DBInviteFriend struct {
	Spin                 int16     `json:"spin" db:"Spin"`
	MaxSpin              int16     `json:"max_spin" db:"MaxSpin"`
	LastSpinRechargeTime time.Time `json:"spin_recharge_time" db:"LastSpinRechargeTime"`
	InviteCount          int8      `json:"invite_count" db:"InviteCount"`
	Now                  time.Time `json:"now" db:"Now"`
}

type ResInviteFriend struct {
	ResultCode           int32     `json:"result_code"`
	Spin                 int16     `json:"spin"`
	LastSpinRechargeTime time.Time `json:"spin_recharge_time"`
	InviteCount          int8      `json:"invite_count"`
}

type InviteFriendResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (InviteFriendResource) Uri() string {
	return "/invitefriend"
}

func (InviteFriendResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (InviteFriendResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	packetdata := make(map[string]interface{})
	log.Println("inviteFriend : Request Data  (", packetdata, ")")

	err := ReadOnJSONBody(r, &packetdata)
	if err != nil {
		log.Println("inviteFriend : )", err.Error())
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	// 현재 친구 초대 정보 확인
	userinfo := DBInviteFriend{}
	err = db.Get(&userinfo, "SELECT Spin, MaxSpin, LastSpinRechargeTime, IFNULL((CASE WHEN DATE(InviteUpdateDate) = DATE(NOW()) THEN InviteCount ELSE 0 END), 0 ) InviteCount, Now() Now FROM Coin.User WHERE AccountUID = ?;", packetdata["acc_uid"])
	if err != nil {
		log.Println("invalid acc_uid : ", packetdata["acc_uid"])
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	// 친구 초대 최대 횟수 확인
	if userinfo.InviteCount >= 5 {
		return httpPkg.Response{Code: 200, Msg: "invite friend count over!", Data: ResInviteFriend{1, userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.InviteCount}}
	}

	userinfo.InviteCount++

	beforeSpin := userinfo.Spin
	if userinfo.Spin < userinfo.MaxSpin {
		userinfo.Spin, userinfo.LastSpinRechargeTime = common.CalcRechargeSP(userinfo.Spin, userinfo.MaxSpin, userinfo.Now, userinfo.LastSpinRechargeTime)
	}

	inviteReward := int16(5)
	userinfo.Spin = common.AddGoods(common.GoodsTypeSpin, inviteReward, userinfo.Spin).(int16)

	// 친구 초대 정보 및 보상 업데이트
	result, err := db.Exec("UPDATE User SET Spin = ?, LastSpinRechargeTime = ?, InviteCount = ?, InviteUpdateDate = ? WHERE AccountUID = ? AND Spin = ?;",
		userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.InviteCount, userinfo.Now, packetdata["acc_uid"], beforeSpin)
	if err != nil {
		log.Println("failed to request : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	rowsAffected, err := result.RowsAffected()

	if rowsAffected == 0 {
		log.Println("Tutorial Complete Request RowsAffected 0 rows! acc_uid : ", packetdata["acc_uid"])
		return httpPkg.Response{Code: 400, Msg: "success", Data: nil}
	}

	log.Println("inviteFriend : Response Data  (", ResInviteFriend{0, userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.InviteCount}, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: ResInviteFriend{0, userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.InviteCount}}
}
