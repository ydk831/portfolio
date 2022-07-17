package action

import (
	"fmt"
	"log"
	"message"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"db"

	"github.com/jmoiron/sqlx"
)

type MansionRankResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (MansionRankResource) Uri() string {
	return "/mansionrank"
}

func (MansionRankResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()

	log.Println("MansionRank : Request Data  (", qry, ")")

	req := message.MansionRankRequest{}
	var err error

	// 파라미터 파싱
	if qry["acc_uid"] == nil {
		log.Println("[MansionRank:Get] acc_uid param is null.. (" + strings.Join(qry["acc_uid"], " ") + ")")
		err := fmt.Errorf("MansionRank : Account ID is null!")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	} else {
		req.ID, err = strconv.Atoi(strings.Join(qry["acc_uid"], " "))
		if err != nil {
			log.Println("[MansionRank:Get] acc_uid can't convert to integer error.. " + strings.Join(qry["acc_uid"], " "))
			err := fmt.Errorf("MansionRank : Account ID convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
		}
	}
	if qry["rank_min"] == nil {
		log.Println("[MansionRank:Get] rank_min param is null.. (" + strings.Join(qry["rank_min"], " ") + ")")
		err := fmt.Errorf("MansionRank : RankMin is null!")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	} else {
		req.RankMin, err = strconv.Atoi(strings.Join(qry["rank_min"], " "))
		if err != nil {
			log.Println("[MansionRank:Get] rank_min can't convert to integer error.. " + strings.Join(qry["rank_min"], " "))
			err := fmt.Errorf("MansionRank : RankMin convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
		}
	}
	if qry["rank_max"] == nil {
		log.Println("[MansionRank:Get] rank_max param is null.. (" + strings.Join(qry["rank_max"], " ") + ")")
		err := fmt.Errorf("MansionRank : RankMax is null!")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	} else {
		req.RankMax, err = strconv.Atoi(strings.Join(qry["rank_max"], " "))
		if err != nil {
			log.Println("[MansionRank:Get] rank_max can't convert to integer error.. " + strings.Join(qry["rank_max"], " "))
			err := fmt.Errorf("MansionRank : RankMax convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
		}
	}
	// 파라미터 파싱 END

	rsp, err, code := db.SelectMainsionRanking(req, mydb)
	if err != nil {
		log.Println("MansionRank : ", err.Error())
		return httpPkg.Response{Code: code, Msg: "MansionRank : " + err.Error(), Data: nil}
	}

	log.Println("MansionRank : Response Data (", rsp, ")")
	return httpPkg.Response{Code: 200, Msg: "Select MansionRank Success", Data: rsp}
}

func (MansionRankResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined POST Method", Data: nil}
}
