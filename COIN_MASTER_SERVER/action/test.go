package action

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"

	"httpPkg"
)

type TestResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (TestResource) Uri() string {
	return "/Test"
}

func (TestResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 200, Msg: "Test : https Test", Data: nil}
}

func (TestResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	// /TileActionType 에 따른 처리
	//packetdata := make(map[string]interface{})
	//packetdata := message.TestResultRequest{}
	//err := ReadOnJSONBody(r, &packetdata)
	//fmt.Println(body)
	//if err != nil {
	//	return httpPkg.Response{Code: 400, Msg: "Test : " + err.Error(), Data: nil}
	//} else { // url과 body의 유효성 체크가 끝났으니 db 작업 수행
	//	rsp, err := db.UpdateTestResult(packetdata, mydb)
	//	if err != nil {
	//		return httpPkg.Response{Code: 500, Msg: "Test : " + err.Error(), Data: rsp}
	//	}
	// rsp := LoginResponse{result.ID, result.LastLoginTime}
	//	return httpPkg.Response{Code: 200, Msg: "Test Result Update Success", Data: rsp}
	//return httpPkg.Response{Code: 200, Msg: "test", Data: rsp}
	//}
	return httpPkg.Response{Code: 405, Msg: "Not Defined POST Method", Data: nil}
}
