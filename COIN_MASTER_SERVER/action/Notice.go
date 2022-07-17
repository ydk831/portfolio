package action

import (
	"db"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"github.com/jmoiron/sqlx"
)

type NoticeInfoResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (NoticeInfoResource) Uri() string {
	return "/notice"
}

func (NoticeInfoResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()
	log.Println("Notice : Request Data (", qry, ")")

	if qry["acc_uid"] == nil {
		log.Println("[Notice:Get] acc_uid param is null.. (" + strings.Join(qry["acc_uid"], " ") + ")")
		err := fmt.Errorf("Notice : Account ID is null.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	id, err := strconv.Atoi(strings.Join(qry["acc_uid"], " "))
	if err != nil {
		log.Println("[Notice:Get] acc_uid can't convert to integer error.. " + string(id))
		err := fmt.Errorf("Notice : Account ID convert error. invalid format.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	rsp, err, code := db.GetNotice(id, mydb)
	if err != nil {
		log.Println("Notice : ", err.Error())
		return httpPkg.Response{Code: code, Msg: err.Error(), Data: nil}
	}
	log.Println("Notice : Response Data (", rsp, ")")
	return httpPkg.Response{Code: code, Msg: "GET Notice Information Success", Data: rsp}

}

func (NoticeInfoResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined POST Method", Data: nil}
}
