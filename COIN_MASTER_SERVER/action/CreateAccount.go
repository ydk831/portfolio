package action

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"github.com/jmoiron/sqlx"
)

// /Login 라는 resouce에 사용하지 않을 API 정의
type CrtAccResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

// /createaccount 라는 resource Uri 등록
func (CrtAccResource) Uri() string {
	return "/createaccount"
}

// Get Method 의 행동 정의
func (CrtAccResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

// /createaccount resource에 Post Method 정의
func (CrtAccResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	packetdata := make(map[string]interface{})
	log.Println("CreateAccount : Request Data  (", packetdata, ")")

	err := ReadOnJSONBody(r, &packetdata)
	//fmt.Println(body)
	if err != nil {
		log.Println("CreateAccount : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "CreateAccount : " + err.Error(), Data: nil}
	}

	// 키값 확인
	keyid, exists := packetdata["keyid"]
	if !exists {
		log.Println("keyid is invalid format.")
		err = fmt.Errorf("CreateAccount : keyid is invalid format")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	// 닉네임 확인
	name, exists := packetdata["name"]
	if !exists || name == "" {
		log.Println("name is invalid format.")
		err = fmt.Errorf("CreateAccount : name is invalid format")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	// 기존 계정으로 연동 시도
	if keyid != "" {
		var rowcount int
		err = db.Get(&rowcount, "select count(*) from Account where KeyID = ?;", packetdata["keyid"])
		if err != nil {
			log.Println("Select Account Fail : ", err.Error())
			log.Println("Query : select count(*) from Account where KeyID = ", packetdata["keyid"])
			err = fmt.Errorf("CreateAccount : Exist Account Check Error.")
			return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
		}

		// 이미 생성 된 계정이 있다.
		if rowcount == 1 {
			var AccountUID int64
			err = db.Get(&AccountUID, "select AccountUID from Account where KeyID = ?;", packetdata["keyid"])
			if err != nil {
				log.Println("Select Account Fail : ", err.Error())
				log.Println("Query : select AccountUID from Account where KeyID = ", packetdata["keyid"])
				err = fmt.Errorf("CreateAccount : Exist Account Check Error.")
				return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
			}

			resdata := packetdata
			resdata["acc_uid"] = AccountUID

			return httpPkg.Response{Code: 200, Msg: "success", Data: resdata}
		}

		// 기존 계정 연동 실패
		//return httpPkg.Response{Code: 400, Msg: "invalid parameter : keyid", Data: nil}
	}

	// Account 정보 생성
	result, err := db.Exec("insert into Account(Name, Type, KeyID, CountryCode, Gender) values (?, ?, ?, ?, ?);",
		packetdata["name"], packetdata["apptype"], packetdata["keyid"], packetdata["countrycode"], packetdata["gender"])
	if err != nil {
		log.Println("Insert Account Fail : ", err.Error())
		log.Println("Query : insert into Account(Name, Type, KeyID, CountryCode, Gender) values (", packetdata["name"], ",", packetdata["apptype"], ",", packetdata["keyid"], ",", packetdata["countrycode"], ",", packetdata["gender"], ");")
		err = fmt.Errorf("CreateAccount : Create Guest account fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	AccountUID, err := result.LastInsertId()
	if err != nil {
		log.Println("Insert Account Fail(LastInsertID) : ", err.Error())
		log.Println("Query : insert into Account(Name, Type, KeyID, CountryCode, Gender) values (", packetdata["name"], ",", packetdata["apptype"], ",", packetdata["keyid"], ",", packetdata["countrycode"], ",", packetdata["gender"], ");")
		err = fmt.Errorf("CreateAccount : Create Guest account fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	if packetdata["keyid"] == "" {
		packetdata["keyid"] = "Guest_" + strconv.Itoa(int(AccountUID))

		result, err := db.Exec("update Account set KeyID = ? where AccountUID = ?;", packetdata["keyid"], AccountUID)
		if err != nil {
			log.Println("Update Account Fail : ", err.Error())
			log.Println("Query : update Account set KeyID = ", packetdata["keyid"], " where AccountUID = ", AccountUID)
			err = fmt.Errorf("CreateAccount : Create Guest account fail.")
			return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("Update Account Fail(RowsAffected) : ", err.Error())
			log.Println("Query : update Account set KeyID = ", packetdata["keyid"], " where AccountUID = ", AccountUID)
			err = fmt.Errorf("CreateAccount : Create Guest account fail.")
			return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
		}

		if rowsAffected == 0 {
			log.Println("Update Account Fail(RowsAffected) : No Rows Update..")
			log.Println("Query : update Account set KeyID = ", packetdata["keyid"], " where AccountUID = ", AccountUID)
			err = fmt.Errorf("CreateAccount : Create Guest account fail.")
			return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
		}

	}

	// User 정보 생성
	result, err = db.Exec("insert into User(AccountUID, Name) values (?,?);", AccountUID, packetdata["name"])
	if err != nil {
		log.Println("Insert User Fail : ", err.Error())
		log.Println("Query : insert into User(AccountUID, Name) values (", AccountUID, ",", packetdata["name"], ");")
		err = fmt.Errorf("CreateAccount : Create Guest user fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// rowsAffected, err = result.RowsAffected()
	// if err != nil || rowsAffected == 0 {
	// 	log.Println(err)
	// 	tx.Rollback() // 중간에 에러가 나면 롤백
	// 	return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	// }
	if err != nil {
		log.Println("Insert User Fail : ", err.Error())
		log.Println("Query : insert into User(AccountUID, Name) values (", AccountUID, ",", packetdata["name"], ");")
		err = fmt.Errorf("CreateAccount : Create Guest user fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	resdata := packetdata
	resdata["acc_uid"] = AccountUID

	log.Println("CreateAccount : Response Data  (", resdata, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: resdata}
}
