package action

import (
	"common"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"message"

	"db"

	"github.com/jmoiron/sqlx"
)

type SlotResultResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (SlotResultResource) Uri() string {
	return "/slotresult"
}

func (SlotResultResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (SlotResultResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	packetdata := make(map[string]interface{})
	err := ReadOnJSONBody(r, &packetdata)

	log.Println("slotResult : Request Data (", packetdata, ")")

	if err != nil {
		log.Println("slotResult : )", err.Error())
		return httpPkg.Response{Code: 400, Msg: "SlotResult : " + err.Error(), Data: nil}
	}

	/*
			 req == packetdata
			 type slotResultRequest struct {
			 ID int `json:"acc_uid"`
			 ResultType int `json:"result_type"`
			 UseSpin int `json:"use_spin"`
			 UseShield int `json:"use_shield"`
			 AddSpin int `json:"add_spin"`
			 AddGold int `json:"add_gold"`
			 AddSheield int `json:"add_shield"`
			 MansionItem int `json:"Mansion_item"`
			 UseCoin int `json:"use_coin"`
		 	}
	*/
	rsp := message.SlotResultResponse{}

	// 유저 정보
	userinfo := message.DBParamUser{}
	err = mydb.Get(&userinfo, "select AccountUID,	Name, Gold, Spin, MaxSpin, LastSpinRechargeTime, Shield, CurrentChapterIDX, Now() Now, Coin FROM User WHERE AccountUID = ?;", packetdata["acc_uid"])
	if err != nil {
		log.Println("Select User Fail : ", err.Error())
		log.Println("Query : select AccountUID,	Name, Gold, Spin, MaxSpin, LastSpinRechargeTime, Shield, CurrentChapterIDX, Now() Now, Coin FROM User WHERE AccountUID = ", packetdata["acc_uid"])
		err = fmt.Errorf("SlotResult : Find user info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// SP 회복 가능
	if userinfo.Spin < userinfo.MaxSpin {
		userinfo.Spin, userinfo.LastSpinRechargeTime = common.CalcRechargeSP(userinfo.Spin, userinfo.MaxSpin, userinfo.Now, userinfo.LastSpinRechargeTime)
	}

	// 파라미터 값 체크
	if packetdata["use_spin"].(float64) < 0 || packetdata["use_shield"].(float64) < 0 ||
		packetdata["add_gold"].(float64) < 0 || packetdata["add_spin"].(float64) < 0 ||
		packetdata["add_shield"].(float64) < 0 || packetdata["use_coin"].(float64) < 0 {
		log.Println("SlotResult Parameter Invalid : ", packetdata)
		err = fmt.Errorf("SlotResult : Invalid Parameter.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	if float64(userinfo.Spin)+packetdata["add_spin"].(float64)-packetdata["use_spin"].(float64) < 0 {
		log.Println("invalid spin value!", packetdata)
		err = fmt.Errorf("SlotResult : Invalid spin value.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	if float64(userinfo.Shield)+packetdata["add_shield"].(float64)-packetdata["use_shield"].(float64) < 0 {
		log.Println("invalid shield value!", packetdata)
		err = fmt.Errorf("SlotResult : Invalid shield value.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	spin := int16(packetdata["add_spin"].(float64)) - int16(packetdata["use_spin"].(float64))
	shield := int16(packetdata["add_shield"].(float64)) - int16(packetdata["use_shield"].(float64))

	coin := int32(packetdata["use_coin"].(float64))

	userinfo.Gold = common.AddGoods(common.GoodsTypeGold, int64(packetdata["add_gold"].(float64)), userinfo.Gold).(int64)
	userinfo.Spin = common.AddGoods(common.GoodsTypeSpin, int16(spin), userinfo.Spin).(int16)
	userinfo.Shield = common.AddGoods(common.GoodsTypeShield, int8(shield), userinfo.Shield).(int8)
	userinfo.Coin = userinfo.Coin - coin

	// Spin 및 접속 정보 갱신
	result, err := mydb.Exec("update User Set Gold = ?, Spin = ?, Shield = ?, LastSpinRechargeTime = ?, UpdateDate = NOW(6), Coin = ?"+
		" where AccountUID = ? and Gold + ? >= 0 and Spin + ? >= 0 and Shield + ? >= 0",
		userinfo.Gold, userinfo.Spin, userinfo.Shield, userinfo.LastSpinRechargeTime, userinfo.Coin, userinfo.AccountUID, packetdata["add_gold"], spin, shield)
	if err != nil {
		log.Println("Update User Fail : ", err.Error())
		log.Println("Query : update User Set Gold = ", userinfo.Gold, ", Spin = ", userinfo.Spin, ", Shield = ", userinfo.Shield, ", LastSpinRechargeTime = ", userinfo.LastSpinRechargeTime, ", UpdateDate = NOW(6)", ", Coin = ", userinfo.Coin, " where AccountUID = ", userinfo.AccountUID, " and Gold + ", packetdata["add_gold"], " >= 0 and Spin + ", spin, " >= 0 and Shield + ", shield, " >= 0;")
		err = fmt.Errorf("SlotResult : Update user info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		if err != nil {
			log.Println("Update User Fail(RowsAffected) : ", err.Error())
		} else {
			log.Println("Update User Fail(RowsAffected) : No Rows Update")
		}
		log.Println("Query : update User Set Gold = ", userinfo.Gold, ", Spin = ", userinfo.Spin, ", Shield = ", userinfo.Shield, ", LastSpinRechargeTime = ", userinfo.LastSpinRechargeTime, ", UpdateDate = NOW(6)", ", Coin = ", userinfo.Coin, " where AccountUID = ", userinfo.AccountUID, " and Gold + ", packetdata["add_gold"], " >= 0 and Spin + ", spin, " >= 0 and Shield + ", shield, " >= 0;")
		err = fmt.Errorf("SlotResult : Update user info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 멘션아이템(가구) 얻었을 경우 DB 저장
	var gold int64
	var code int
	if int(packetdata["mansion_item"].(float64)) != 0 { // 0이면 가구 못얻음
		var id int = int(packetdata["acc_uid"].(float64))
		var item int = int(packetdata["mansion_item"].(float64))

		gold, err, code = db.InsertMansionItem(id, item, mydb)
		if err != nil {
			log.Println("SlotResult : ", err.Error())
			return httpPkg.Response{Code: code, Msg: err.Error(), Data: nil}
		}
	}

	rsp.ResultType = int(packetdata["result_type"].(float64))
	rsp.Gold = userinfo.Gold + gold // 중복시 보상 골드
	rsp.Spin = userinfo.Spin
	rsp.Shield = userinfo.Shield
	rsp.Coin = userinfo.Coin
	rsp.SpinRechargeTime = userinfo.LastSpinRechargeTime
	rsp.MansionItem = int(packetdata["mansion_item"].(float64))

	log.Println("slotResult : Response Data (", rsp, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: rsp}

}
