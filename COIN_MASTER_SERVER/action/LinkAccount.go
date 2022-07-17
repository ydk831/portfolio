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

type res_linkAccount struct {
	Keyid string
}

type LinkAccountResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (LinkAccountResource) Uri() string {
	return "/linkaccount"
}

func (LinkAccountResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (LinkAccountResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	packetdata := make(map[string]interface{})
	log.Println("LinkAccount : Request Data  (", packetdata, ")")

	err := ReadOnJSONBody(r, &packetdata)
	if err != nil {
		log.Println("LinkAccount : )", err.Error())
		return httpPkg.Response{Code: 400, Msg: "LinkAccount : " + err.Error(), Data: nil}
	}

	// 키값 확인
	keyid, exists := packetdata["new_keyid"]
	if !exists || keyid == "" {
		log.Println("new_keyid is invalid. (", packetdata["new_keyid"], ")")
		err = fmt.Errorf("LinkAccount : new_keyid is invalid.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	// 닉네임 확인
	name, exists := packetdata["name"]
	if !exists || name == "" {
		log.Println("new_keyid is invalid. (", packetdata["name"], ")")
		err = fmt.Errorf("LinkAccount : name is invalid.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	// 연동하려는 키값이 있는지 확인 ( 마을을 성장시킨 정보가 있을 경우에만 기존 정보로 연동)
	var rowcount int
	err = db.Get(&rowcount, "select count(*) from Account A inner join UserTile T on A.AccountUID = T.AccountUID and A.KeyID = ?;", packetdata["new_keyid"])
	if err != nil {
		log.Println("Select Account join UserTile Fail : ", err.Error())
		log.Println("Query : select count(*) from Account A inner join UserTile T on A.AccountUID = T.AccountUID and A.KeyID = ", packetdata["new_keyid"])
		err = fmt.Errorf("LinkAccount : Find account info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 	계정 연동
	if rowcount == 0 {
		// 이미 연동된 계정이 있다면 KeyID 초기화
		// status ( 0 = 일반, 100 = 계정 연동으로 삭제 된 계정 )
		result, err := db.Exec("update Account Set KeyID = concat('_',?), Status = 100 where KeyID = ?;", packetdata["keyid"], packetdata["new_keyid"])
		if err != nil {
			log.Println("Update Account Fail : ", err.Error())
			log.Println("Query : update Account Set KeyID = '', Status = 100 where KeyID = ", packetdata["new_keyid"])
			err = fmt.Errorf("LinkAccount : Update account info fail.")
			return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
		}

		result, err = db.Exec("update Account Set Name = ?, Type = ?, KeyID = ? where KeyID = ?;",
			packetdata["name"], packetdata["apptype"], packetdata["new_keyid"], packetdata["keyid"])
		if err != nil {
			log.Println("Update Account Fail : ", err.Error())
			log.Println("Query : update Account Set Name = ", packetdata["name"], ", Type = ", packetdata["apptype"], ", KeyID = ", packetdata["new_keyid"], " where KeyID = ", packetdata["keyid"])
			err = fmt.Errorf("LinkAccount : Update account info fail.")
			return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("Update Account Fail(RowsAffected) : ", err.Error())
			log.Println("Query : update Account Set Name = ", packetdata["name"], ", Type = ", packetdata["apptype"], ", KeyID = ", packetdata["new_keyid"], " where KeyID = ", packetdata["keyid"])
			err = fmt.Errorf("LinkAccount : Update account info fail.")
			return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
		}

		if rowsAffected == 0 {
			log.Println("Update Account Fail(RowsAffected) : No Rows Update..")
			log.Println("Query : update Account Set Name = ", packetdata["name"], ", Type = ", packetdata["apptype"], ", KeyID = ", packetdata["new_keyid"], " where KeyID = ", packetdata["keyid"])
			err = fmt.Errorf("LinkAccount : Update account info fail.")
			return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
		}
	}

	// 새로운 계정 정보
	accinfo := message.DBParamAccount{}
	err = db.Get(&accinfo, "select AccountUID, Name, Type, KeyID, CountryCode, Status FROM Account WHERE KeyID = ?;", packetdata["new_keyid"])
	if err != nil {
		log.Println("Select Account Fail : ", err.Error())
		log.Println("Query : select AccountUID, Name, Type, KeyID, CountryCode, Status FROM Account WHERE KeyID = ", packetdata["new_keyid"])
		err = fmt.Errorf("LinkAccount : Find account info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 계정 상태값 체크
	if accinfo.Status > 0 {
		log.Println("invalid account status.(", accinfo.Status, ")")
		err = fmt.Errorf("LinkAccount : Invalid account Status.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	// User 테이블 닉네임 변경
	result, err := db.Exec("update User Set Name = ? where AccountUID = ?;", packetdata["name"], accinfo.AccountUID)
	if err != nil {
		log.Println("Update User Fail : ", err.Error())
		log.Println("Query : update User Set Name = ", packetdata["name"], " where Name = ", accinfo.AccountUID)
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Query : update User Set Name = ", packetdata["name"], " where AccountUID = ", accinfo.AccountUID)
		return httpPkg.Response{Code: 500, Msg: "failed request", Data: nil}
	}

	if rowsAffected == 0 {
		// log.Println("facebook link request. not change name!")
		// err = fmt.Errorf("LinkAccount : Update user nickname fail.")
		// return httpPkg.Response{Code: 500, Msg: "failed request", Data: nil}
	}

	log.Println("LinkAccount : Response Data  (", res_linkAccount{Keyid: accinfo.KeyID}, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: res_linkAccount{Keyid: accinfo.KeyID}}
}
