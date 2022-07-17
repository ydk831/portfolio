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

type AdvRewardResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (AdvRewardResource) Uri() string {
	return "/advreward"
}

func (AdvRewardResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (AdvRewardResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	// /TileActionType 에 따른 처리
	//packetdata := make(map[string]interface{})
	packetdata := message.AdvRewardRequest{}
	err := ReadOnJSONBody(r, &packetdata)
	log.Println("AdvReward : Request Data (", packetdata, ")")
	if err != nil {
		return httpPkg.Response{Code: 400, Msg: "AdvReward : " + err.Error(), Data: nil}
	} else { // url과 body의 유효성 체크가 끝났으니 db 작업 수행
		rsp, err, code := db.UpdateAdvReward(packetdata, mydb)
		if err != nil {
			log.Println("AdvReward : ", err.Error())
			return httpPkg.Response{Code: code, Msg: "AdvReward : " + err.Error(), Data: rsp}
		}
		log.Println("AdvReward : Response Data (", rsp, ")")
		return httpPkg.Response{Code: code, Msg: "Adv Reward Update Success", Data: rsp}
	}
}
