package action

import (
	"db"
	"fmt"
	"log"
	"message"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"

	"httpPkg"
)

type ShopGameResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (ShopGameResource) Uri() string {
	return "/shopgame"
}

func (ShopGameResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()

	log.Println("ShopGame[Get] : Request Data (", qry, ")")

	if qry["acc_uid"] == nil {
		log.Println("ShopGame[Get] acc_uid param is null.. (" + strings.Join(qry["acc_uid"], " ") + ")")
		err := fmt.Errorf("ShopGame[Get] : Account ID is null.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	}

	id, err := strconv.Atoi(strings.Join(qry["acc_uid"], " "))
	if err != nil {
		log.Println("ShopGame[Get] acc_uid can't convert to integer error.. " + string(id))
		err := fmt.Errorf("ShopGame : Account ID convert error. invalid format.")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	}

	rsp, err, code := db.SelectShopGameList(id, mydb)
	if err != nil {
		log.Println("ShopGame[Get] : ", err.Error())
		return httpPkg.Response{Code: code, Msg: err.Error(), Data: nil}
	}

	log.Println("ShopGame[Get] : Response Data (", rsp, ")")
	return httpPkg.Response{Code: code, Msg: "GET UserShopGameList Success", Data: rsp}
}

func (ShopGameResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {

	req := message.ShopGameRequest{}
	err := ReadOnJSONBody(r, &req)

	log.Println("ShopGame[Post] : Request Data (", req, ")")

	if err != nil {
		log.Println("ShopGame[Post] : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "ShopGame : " + err.Error(), Data: nil}
	} else {
		rsp, err, code := db.UpdateShopGame(req, mydb)
		if err != nil {
			log.Println("ShopGame[Post] : ", err.Error())
			return httpPkg.Response{Code: code, Msg: "ShopGame : " + err.Error(), Data: rsp}
		}

		log.Println("ShopGame[Post] : Response Data (", rsp, ")")
		return httpPkg.Response{Code: code, Msg: "ShopGame Update Success", Data: rsp}
	}
}
