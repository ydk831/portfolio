package action

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"db"

	"github.com/jmoiron/sqlx"
)

type MansionFriendResource struct {
	httpPkg.PutNotSupported
	//httpPkg.DeleteNotSupported
}

func (MansionFriendResource) Uri() string {
	return "/mansionfriend"
}

func (MansionFriendResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()
	log.Println("MansionFriend : Request Data  (", qry, ")")

	var id int
	var err error
	if qry["acc_uid"] == nil {
		log.Println("[MansionFriend:Get] acc_uid param is null.. (" + strings.Join(qry["acc_uid"], " ") + ")")
		err := fmt.Errorf("MansionFriend : Account ID is null!")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	} else {
		id, err = strconv.Atoi(strings.Join(qry["acc_uid"], " "))
		if err != nil {
			log.Println("[MansionFriend:Get] acc_uid can't convert to integer error.. " + string(id))
			err := fmt.Errorf("MansionFriend : Account ID convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
		}
	}

	// 멘션 친구 정보 획득
	rsp, err, code := db.SelectMansionFriend(id, mydb)
	if err != nil {
		log.Println("MansionFriend : ", err.Error())
		return httpPkg.Response{Code: code, Msg: "MansionFriend : " + err.Error(), Data: nil}
	}

	log.Println("MansionFriend : Response Data (", rsp, ")")
	return httpPkg.Response{Code: 200, Msg: "Select MansionFriend Success", Data: rsp}
}

func (MansionFriendResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	req := make(map[string]interface{})
	err := ReadOnJSONBody(r, &req)

	log.Println("MansionFriend : Request Data (", req, ")")
	if err != nil {
		log.Println("MansionFriend : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "MansionFriend : " + err.Error(), Data: nil}
	}

	id := int(req["acc_uid"].(float64))
	friend := int(req["friend_id"].(float64))

	// 멘션 친구 추가
	rsp, err, code := db.InsertMansionFriend(id, friend, mydb)
	if err != nil {
		log.Println("MansionFriend : ", err.Error())
		return httpPkg.Response{Code: code, Msg: "MansionFriend : " + err.Error(), Data: nil}
	}
	log.Println("MansionFriend : Response Data (", rsp, ")")
	return httpPkg.Response{Code: code, Msg: "MansionFriend Result Update Success", Data: rsp}
}

func (MansionFriendResource) Delete(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	req := make(map[string]interface{})
	err := ReadOnJSONBody(r, &req)

	log.Println("MansionFriend : Request Data (", req, ")")
	if err != nil {
		return httpPkg.Response{Code: 400, Msg: "MansionFriend : " + err.Error(), Data: nil}
	}

	id := int(req["acc_uid"].(float64))
	friend := int(req["friend_id"].(float64))

	// 멘션 친구 삭제
	rsp, err, code := db.DeleteMansionFriend(id, friend, mydb)
	if err != nil {
		log.Println("MansionFriend : ", err.Error())
		return httpPkg.Response{Code: code, Msg: "MansionFriend : " + err.Error(), Data: nil}
	}
	log.Println("MansionFriend : Response Data (", rsp, ")")
	return httpPkg.Response{Code: code, Msg: "MansionFriend Result Update Success", Data: rsp}
}
