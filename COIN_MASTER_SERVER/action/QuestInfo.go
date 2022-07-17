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

type QuestInfoResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (QuestInfoResource) Uri() string {
	return "/questinfo"
}

func (QuestInfoResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()

	log.Println("QuestInfo : Request Data (", qry, ")")

	if qry["acc_uid"] == nil {
		log.Println("QuestInfo : acc_uid param is null.. (" + strings.Join(qry["acc_uid"], " ") + ")")
		err := fmt.Errorf("QuestInfo : Account ID is null.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	}

	id, err := strconv.Atoi(strings.Join(qry["acc_uid"], " "))
	if err != nil {
		log.Println("QuestInfo : acc_uid can't convert to integer error.. " + string(id))
		err := fmt.Errorf("QuestInfo : Account ID convert error. invalid format.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	}

	rsp, err, code := db.SelectQuestInfo(id, mydb)
	if err != nil {
		log.Println("QuestInfo : ", err.Error())
		return httpPkg.Response{Code: code, Msg: err.Error(), Data: nil}
	}
	log.Println("QuestInfo : Response Data (", rsp, ")")
	return httpPkg.Response{Code: code, Msg: "GET QuestInfo Success", Data: rsp}
}

func (QuestInfoResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined POST Method", Data: nil}
}
