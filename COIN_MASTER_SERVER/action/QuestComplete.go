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

type QuestCompleteResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (QuestCompleteResource) Uri() string {
	return "/questcomplete"
}

func (QuestCompleteResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (QuestCompleteResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	req := message.QuestCompleteRequest{}
	err := ReadOnJSONBody(r, &req)

	log.Println("QuestComplete : Request Data (", req, ")")

	if err != nil {
		log.Println("QuestComplete : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "QuestComplete : " + err.Error(), Data: nil}
	} else {
		// spin recharge
		err, code := db.RechargeSP(req.ID, mydb)
		if err != nil {
			log.Println("QuestComplete : ", err.Error())
			return httpPkg.Response{Code: code, Msg: "QuestComplete : " + err.Error(), Data: nil}
		}

		// 퀘스트 완료 업데이트 및 완료 보상 응답
		rsp, err, code := db.UpdateUserQuestComplete(req, mydb)
		if err != nil {
			log.Println("QuestComplete : ", err.Error())
			return httpPkg.Response{Code: code, Msg: "QuestComplete : " + err.Error(), Data: nil}
		}

		// 퀘스트 완료 후 누적 데이터 클리어.. 메시지 처리와는 별개이므로 고루틴 처리 함
		// 퀘스트 완료 업데이트가 성공해야지 하므로 위치는 여기가 맞다.
		go db.UpdateUserQuestClear(req, mydb)

		log.Println("QuestComplete : Response Data (", rsp, ")")
		return httpPkg.Response{Code: code, Msg: "QuestComplete Update Success", Data: rsp}
	}
}
