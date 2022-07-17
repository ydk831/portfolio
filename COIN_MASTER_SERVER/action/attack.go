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

type AttackResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (AttackResource) Uri() string {
	return "/attack"
}

func (AttackResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()

	log.Println("Attack : Request Data  (", qry, ")")

	if qry["acc_uid"] == nil {
		log.Println("[Attack:Get] acc_uid param is null.. (" + strings.Join(qry["acc_uid"], " ") + ")")
		err := fmt.Errorf("Attack : Account ID is null!")
		return httpPkg.Response{Code: 400, Msg: err.Error(), Data: qry}
	} else {

		req := message.AttackRequest{}

		id, err := strconv.Atoi(strings.Join(qry["acc_uid"], " "))
		if err != nil {
			log.Println("[Attack:Get] acc_uid can't convert to integer error.. " + string(id))
			err := fmt.Errorf("Attack : Account ID convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
		}

		mv, err := strconv.Atoi(strings.Join(qry["multiple_value"], " "))
		if err != nil {
			log.Println("[Attack:Get] multiple_value can't convert to integer error.. " + string(mv))
			err := fmt.Errorf("Attack : Multiple Value convert error. invalid format.")
			return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
		}

		req.ID = uint(id)
		req.MultipleValue = mv

		if qry["target_id"] != nil {
			tid, err := strconv.Atoi(strings.Join(qry["target_id"], " "))
			if err != nil {
				log.Println("[Attack:Get] target_id can't convert to integer error.. " + string(tid))
				err := fmt.Errorf("Attack : Target ID convert error. invalid format")
				return httpPkg.Response{Code: 400, Msg: err.Error(), Data: nil}
			}

			req.TargetID = uint(tid)

			rsp, err, code := db.FindUserForAttack(req, mydb)
			if err != nil {
				log.Println("[Attack:Get] ", err.Error())
				return httpPkg.Response{Code: code, Msg: "Attack : " + err.Error(), Data: nil}
			}
			log.Println("[Attack:Get] : Response Data (", rsp, ")")
			return httpPkg.Response{Code: code, Msg: "Select Target Info Success", Data: rsp}
		} else {
			req.TargetID = 0
			rsp, err, code := db.FindRandUserForAttack(req, mydb)
			if err != nil {
				log.Println("[Attack:Get] ", err.Error())
				return httpPkg.Response{Code: code, Msg: "Attack : " + err.Error(), Data: nil}
			}
			log.Println("[Attack:Get] Response Data (", rsp, ")")
			return httpPkg.Response{Code: code, Msg: "Select Random Target Info Success", Data: rsp}
		}
	}
}

func (AttackResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	// /TileActionType 에 따른 처리
	//packetdata := make(map[string]interface{})
	packetdata := message.AttackResultRequest{}
	err := ReadOnJSONBody(r, &packetdata)

	log.Println("[Attack:Post] Request Data (", packetdata, ")")

	if err != nil {
		log.Println("[Attack:Post] ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "Attack : " + err.Error(), Data: nil}
	} else { // url과 body의 유효성 체크가 끝났으니 db 작업 수행
		rsp, err, code := db.UpdateAttackResult(packetdata, mydb)
		if err != nil {
			log.Println("[Attack:Post] ", err.Error())
			return httpPkg.Response{Code: code, Msg: "Attack : " + err.Error(), Data: rsp}
		}
		log.Println("[Attack:Post] Response Data (", rsp, ")")
		return httpPkg.Response{Code: code, Msg: "Attack Result Update Success", Data: rsp}
	}
}
