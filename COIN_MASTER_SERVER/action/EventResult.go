package action

import (
	"db"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"message"

	"github.com/jmoiron/sqlx"
)

type EventResultResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (EventResultResource) Uri() string {
	return "/eventresult"
}

func (EventResultResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (EventResultResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	req := message.EventResultRequest{}
	err := ReadOnJSONBody(r, &req)
	log.Println("EventResult : Request Data (", req, ")")
	if err != nil {
		log.Println("EventResult : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	// 유저 Spin Recharge 갱신.. Spin값을 내려주는 url에 다 넣자..
	err, code := db.RechargeSP(req.ID, mydb)
	if err != nil {
		log.Println("EventResult : ", err.Error())
		return httpPkg.Response{Code: code, Msg: "EventResult : " + err.Error(), Data: nil}
	}

	// 유저 이벤트 정산처리
	rsp, err, code := db.UpdateEventResult(req, mydb)
	if err != nil {
		log.Println("EventResult : ", err.Error())
		return httpPkg.Response{Code: code, Msg: "EventResult : " + err.Error(), Data: nil}
	}

	log.Println("EventResult : Response Data (", rsp, ")")
	return httpPkg.Response{Code: code, Msg: "EventResult Update Success", Data: rsp}

}
