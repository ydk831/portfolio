package action

import (
	"db"
	"log"
	"message"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"

	"httpPkg"
)

type GoldRoomOpenResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (GoldRoomOpenResource) Uri() string {
	return "/goldroomopen"
}

func (GoldRoomOpenResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (GoldRoomOpenResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	req := message.GoldRoomOpenRequest{}
	err := ReadOnJSONBody(r, &req)

	log.Println("GoldRoomOpen : Request Data (", req, ")")

	if err != nil {
		log.Println("GoldRoomOpen : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "GoldRoomOpen : " + err.Error(), Data: nil}
	} else {

		// Room Open 응답
		rsp, err, code := db.UpdateUserGoldRoomOpen(req, mydb)
		if err != nil {
			log.Println("GoldRoomOpen : ", err.Error())
			return httpPkg.Response{Code: code, Msg: "GoldRoomOpen : " + err.Error(), Data: nil}
		}

		log.Println("GoldRoomOpen : Response Data (", rsp, ")")
		return httpPkg.Response{Code: code, Msg: "GoldRoomOpen Update Success", Data: rsp}
	}
}
