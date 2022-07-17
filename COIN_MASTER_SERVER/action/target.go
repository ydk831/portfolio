package action

import (
	"db"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"

	"httpPkg"
)

type TargetResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (TargetResource) Uri() string {
	return "/target"
}

func (TargetResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()
	var keyidlist []string
	var id int
	var err error

	log.Println("Target : Request Data (", qry, ")")

	if qry["acc_uid"] != nil {
		id, err = strconv.Atoi(strings.Join(qry["acc_uid"], " "))
		if err != nil {
			log.Println("[Target:Get] acc_uid can't convert to integer error.. " + string(id))
			err := fmt.Errorf("Target : Account ID convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
		}
	} else {
		log.Println("[Target:Get] acc_uid is empty")
		return httpPkg.Response{Code: 400, Msg: "Target : Your ID is empty.", Data: nil}
	}

	if qry["keyid[]"] != nil {
		for _, v := range qry["keyid[]"] {
			keyidlist = append(keyidlist, v)
		}

		rsp, err, code := db.FindTarget(keyidlist, mydb)
		if err != nil {
			log.Println("Target : ", err.Error())
			return httpPkg.Response{Code: code, Msg: "Target : " + err.Error(), Data: nil}
		}
		log.Println("Target : Response Data (", rsp, ")")
		return httpPkg.Response{Code: code, Msg: "Find Target Info Success", Data: rsp}
	} else {
		rsp, err, code := db.FindRandTarget(id, mydb)

		if err != nil {
			log.Println("Target : ", err.Error())
			return httpPkg.Response{Code: code, Msg: "Target : " + err.Error(), Data: nil}
		}
		log.Println("Target : Response Data (", rsp, ")")
		return httpPkg.Response{Code: code, Msg: "Find Target Info Success", Data: rsp}
	}
}

func (TargetResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	// /TileActionType 에 따른 처리
	//packetdata := make(map[string]interface{})
	//packetdata := message.TargetResultRequest{}
	//err := ReadOnJSONBody(r, &packetdata)
	//fmt.Println(body)
	//if err != nil {
	//	return httpPkg.Response{Code: 400, Msg: "Target : " + err.Error(), Data: nil}
	//} else { // url과 body의 유효성 체크가 끝났으니 db 작업 수행
	//	rsp, err := db.UpdateTargetResult(packetdata, mydb)
	//	if err != nil {
	//		return httpPkg.Response{Code: 500, Msg: "Target : " + err.Error(), Data: rsp}
	//	}
	// rsp := LoginResponse{result.ID, result.LastLoginTime}
	//	return httpPkg.Response{Code: 200, Msg: "Target Result Update Success", Data: rsp}
	//return httpPkg.Response{Code: 200, Msg: "test", Data: rsp}
	//}
	return httpPkg.Response{Code: 405, Msg: "Not Defined POST Method", Data: nil}
}
