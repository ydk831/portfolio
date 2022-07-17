package action

import (
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"github.com/jmoiron/sqlx"
)

type ResRefresh struct {
	Gold   int64     `json:"gold" db:"Gold"`
	Shield int8      `json:"shield" db:"Shield"`
	Coin   int       `json:"coin" db:"Coin"`
	Now    time.Time `json:"current_time" db:"Now"`
}

type RefreshResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (RefreshResource) Uri() string {
	return "/refresh"
}

func (RefreshResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	req := r.URL.Query()

	log.Println("Refresh : Request Data (", req, ")")

	// 현재 튜토리얼 확인
	refresh := ResRefresh{}
	err := db.Get(&refresh, "select Gold, Shield, Coin, NOW() as Now from User where AccountUID = ?;", req["acc_uid"][0])
	if err != nil {
		log.Println("invalid acc_uid : ", req["acc_uid"][0])
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
	}

	log.Println("Refresh : Response Data (", refresh, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: refresh}
}

func (RefreshResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined POST Method", Data: nil}
}
