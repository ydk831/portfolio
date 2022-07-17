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

type RaidResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (RaidResource) Uri() string {
	return "/raid"
}

func (RaidResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()

	log.Println("Raid : Request Data (", qry, ")")

	if qry["acc_uid"] == nil {
		log.Println("[Raid:Get] acc_uid param is null.. (" + strings.Join(qry["acc_uid"], " ") + ")")
		err := fmt.Errorf("Raid : Account ID is null!")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	} else {

		req := message.AttackRequest{}

		id, err := strconv.Atoi(strings.Join(qry["acc_uid"], " "))
		if err != nil {
			log.Println("[Raid:Get] acc_uid can't convert to integer error.. " + string(id))
			err := fmt.Errorf("Raid : Account ID convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
		}

		mv, err := strconv.Atoi(strings.Join(qry["multiple_value"], " "))
		if err != nil {
			log.Println("[Raid:Get] multiple_value can't convert to integer error.. " + string(mv))
			err := fmt.Errorf("Raid : Multiple Value convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
		}

		req.ID = uint(id)
		req.MultipleValue = mv

		if qry["target_id"] != nil {
			tid, err := strconv.Atoi(strings.Join(qry["target_id"], " "))
			if err != nil {
				log.Println("[Raid:Get] target_id can't convert to integer error.. " + string(tid))
				err := fmt.Errorf("Raid : Target ID convert error. invalid format.")
				return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
			}

			req.TargetID = uint(tid)

			rsp, err, code := db.FindUserForRaid(req, mydb)
			if err != nil {
				log.Println("Raid : ", err.Error())
				return httpPkg.Response{Code: code, Msg: "Raid : " + err.Error(), Data: nil}
			}
			log.Println("Raid : Response Data (", rsp, ")")
			return httpPkg.Response{Code: code, Msg: "Select Raid Target Info Success", Data: rsp}
		} else {
			req.TargetID = 0
			rsp, err, code := db.FindRandUserForRaid(req, mydb)
			if err != nil {
				log.Println("Raid : ", err.Error())
				return httpPkg.Response{Code: code, Msg: "Raid : " + err.Error(), Data: nil}
			}
			log.Println("Raid : Response Data (", rsp, ")")
			return httpPkg.Response{Code: code, Msg: "Select Random Raid Target Info Success", Data: rsp}
		}
	}
}

func (RaidResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	packetdata := message.RaidResultRequest{}
	err := ReadOnJSONBody(r, &packetdata)

	log.Println("Raid : Request Data (", packetdata, ")")

	if err != nil {
		log.Println("Raid : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "Raid : " + err.Error(), Data: nil}
	} else { // url과 body의 유효성 체크가 끝났으니 db 작업 수행
		rsp, err, code := db.UpdateRaidResult(packetdata, mydb)
		if err != nil {
			log.Println("Raid : ", err.Error())
			return httpPkg.Response{Code: code, Msg: "Raid : " + err.Error(), Data: rsp}
		}
		log.Println("Raid : Response Data (", rsp, ")")
		return httpPkg.Response{Code: code, Msg: "Raid Result Update Success", Data: rsp}
	}
}
