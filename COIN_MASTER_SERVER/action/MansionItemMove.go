package action

import (
	"db"
	"log"
	"message"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver

	"github.com/jmoiron/sqlx"
)

type MansionItemMoveResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (MansionItemMoveResource) Uri() string {
	return "/mansionitemmove"
}

func (MansionItemMoveResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (MansionItemMoveResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	req := message.MansionItemMoveRequest{}
	err := ReadOnJSONBody(r, &req)

	log.Println("MansionItemMove : Request Data (", req, ")")
	if err != nil {
		return httpPkg.Response{Code: 400, Msg: "MansionItemMove : " + err.Error(), Data: nil}
	}

	rsp, err, code := db.UpdateMoveMansionItem(req, mydb)
	if err != nil {
		log.Println("MansionItemMove : ", err.Error())
		return httpPkg.Response{Code: code, Msg: "MansionItemMove : " + err.Error(), Data: nil}
	}
	log.Println("MansionItemMove : Response Data (", rsp, ")")
	return httpPkg.Response{Code: code, Msg: "MansionItemMove Result Update Success", Data: rsp}
}
