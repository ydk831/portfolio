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

type QuestUpdateResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (QuestUpdateResource) Uri() string {
	return "/questupdate"
}

func (QuestUpdateResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (QuestUpdateResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	req := message.QuestUpdateRequest{}
	err := ReadOnJSONBody(r, &req)

	log.Println("QuestUpdate : Request Data (", req, ")")

	if err != nil {
		log.Println("QuestUpdate : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "QuestUpdate : " + err.Error(), Data: nil}
	} else {
		// 퀘스트 데이터 누적 업데이트 및 최종 누적 데이터 응답
		rsp, err, code := db.UpdateUserQuest(req, mydb)
		if err != nil {
			log.Println("QuestUpdate : ", err.Error())
			return httpPkg.Response{Code: code, Msg: "QuestUpdate : " + err.Error(), Data: nil}
		}

		log.Println("QuestUpdate : Response Data (", rsp, ")")
		return httpPkg.Response{Code: code, Msg: "UserQuest Update Success", Data: rsp}
	}
}
