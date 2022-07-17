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

type NewsResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (NewsResource) Uri() string {
	return "/news"
}

func (NewsResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()

	log.Println("News : Request Data (", qry, ")")

	if qry["acc_uid"] == nil {
		log.Println("[News:Get] acc_uid param is null.. (" + strings.Join(qry["acc_uid"], " ") + ")")
		err := fmt.Errorf("News : Account ID is null.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	}

	id, err := strconv.Atoi(strings.Join(qry["acc_uid"], " "))
	if err != nil {
		log.Println("[News:Get] acc_uid can't convert to integer error.. " + string(id))
		err := fmt.Errorf("News : Account ID convert error. invalid format.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	}

	rsp, err, code := db.GetNews(id, mydb)
	if err != nil {
		log.Println("News : ", err.Error())
		return httpPkg.Response{Code: code, Msg: err.Error(), Data: nil}
	}
	log.Println("News : Response Data (", rsp, ")")
	return httpPkg.Response{Code: code, Msg: "GET News Information Success", Data: rsp}
}

func (NewsResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined POST Method", Data: nil}
}
