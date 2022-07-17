package action

import (
	"log"
	"message"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"db"

	"github.com/jmoiron/sqlx"
)

type MansionLikeResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (MansionLikeResource) Uri() string {
	return "/mansionlike"
}

func (MansionLikeResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (MansionLikeResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	req := make(map[string]interface{})
	err := ReadOnJSONBody(r, &req)

	log.Println("MansionLike : Request Data (", req, ")")
	if err != nil {
		log.Println("MansionLike : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "MansionFriend : " + err.Error(), Data: nil}
	}

	id := int(req["acc_uid"].(float64))
	target := int(req["like_id"].(float64))

	rsp := message.MansionLikeResponse{}

	// 좋아요 처리
	var code int
	rsp.RemainLike, err, code = db.UpdateMansionLike(id, target, mydb)
	if err != nil {
		log.Println("MansionLike : ", err.Error())
		return httpPkg.Response{Code: code, Msg: "MansionLike : " + err.Error(), Data: nil}
	}
	log.Println("MansionLike : Response Data (", rsp, ")")
	return httpPkg.Response{Code: code, Msg: "MansionLike Result Update Success", Data: rsp}
}
