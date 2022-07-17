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

type MansionRankRewardResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (MansionRankRewardResource) Uri() string {
	return "/mansionrankreward"
}

func (MansionRankRewardResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()

	log.Println("MansionRankReward : Request Data  (", qry, ")")

	var err error
	var id int

	// 파라미터 파싱
	if qry["acc_uid"] == nil {
		log.Println("[MansionRankReward:Get] acc_uid param is null.. (" + strings.Join(qry["acc_uid"], " ") + ")")
		err := fmt.Errorf("MansionRankReward : Account ID is null!")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	} else {
		id, err = strconv.Atoi(strings.Join(qry["acc_uid"], " "))
		if err != nil {
			log.Println("[MansionRankReward:Get] acc_uid can't convert to integer error.. " + strings.Join(qry["acc_uid"], " "))
			err := fmt.Errorf("MansionRankReward : Account ID convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
		}
	}
	// 파라미터 파싱 END

	rsp, err, code := db.SelectMainsionRankReward(id, mydb)
	if err != nil {
		log.Println("MansionRankReward : ", err.Error())
		return httpPkg.Response{Code: code, Msg: "MansionRankReward : " + err.Error(), Data: nil}
	}

	log.Println("MansionRankReward : Response Data (", rsp, ")")
	return httpPkg.Response{Code: 200, Msg: "Select MansionRankReward Success", Data: rsp}
}

func (MansionRankRewardResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined POST Method", Data: nil}
}
