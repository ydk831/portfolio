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

type MansionInfoResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (MansionInfoResource) Uri() string {
	return "/mansioninfo"
}

func (MansionInfoResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()
	log.Println("MansionInfo : Request Data  (", qry, ")")

	var id int
	var rand int
	var keyid string
	var err error
	if qry["acc_uid"] == nil {
		log.Println("[MansionInfo:Get] acc_uid param is null.. (" + strings.Join(qry["acc_uid"], " ") + ")")
		err := fmt.Errorf("MansionInfo : Account ID is null!")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	} else {
		id, err = strconv.Atoi(strings.Join(qry["acc_uid"], " "))
		if err != nil {
			log.Println("[MansionInfo:Get] acc_uid can't convert to integer error.. " + string(id))
			err := fmt.Errorf("MansionInfo : Account ID convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
		}
	}

	if qry["rand_num"] != nil && qry["keyid"] != nil {
		log.Println("[MansionInfo:Get] Don't Request Prameter rand_num and keyid together")
		err := fmt.Errorf("MansionInfo : Don't Request Prameter rand_num and keyid together")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	if qry["rand_num"] != nil {
		rand, err = strconv.Atoi(strings.Join(qry["rand_num"], " "))
		if err != nil {
			log.Println("[MansionInfo:Get] rand_num can't convert to integer error.. " + strings.Join(qry["rand_num"], " "))
			err := fmt.Errorf("MansionInfo : Random Number convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
		}

		rsp, err, code := db.SelectUserMansionInfo(id, rand, mydb)
		if err != nil {
			log.Println("MansionInfo : )", err.Error())
			return httpPkg.Response{Code: code, Msg: "MansionInfo : " + err.Error(), Data: nil}
		}

		log.Println("MansionInfo : Response Data (", rsp, ")")
		return httpPkg.Response{Code: 200, Msg: "Select MansionInfo Success", Data: rsp}

	} else if qry["keyid"] != nil {
		keyid = strings.Join(qry["keyid"], " ")

		rsp, err, code := db.SelectKeyIDMansionInfo(id, keyid, mydb)
		if err != nil {
			log.Println("MansionInfo : )", err.Error())
			return httpPkg.Response{Code: code, Msg: "MansionInfo : " + err.Error(), Data: nil}
		}
		log.Println("MansionInfo : Response Data (", rsp, ")")
		return httpPkg.Response{Code: 200, Msg: "Select MansionInfo Success", Data: rsp}

	} else {
		rsp, err, code := db.SelectUserMansionInfo(id, 0, mydb)
		if err != nil {
			log.Println("MansionInfo : )", err.Error())
			return httpPkg.Response{Code: code, Msg: "MansionInfo : " + err.Error(), Data: nil}
		}
		log.Println("MansionInfo : Response Data (", rsp, ")")
		return httpPkg.Response{Code: 200, Msg: "Select MansionInfo Success", Data: rsp}
	}
}

func (MansionInfoResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined POST Method", Data: nil}
}
