package action

import (
	"common"
	"fmt"
	"log"
	"message"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	actiondb "db"

	"github.com/jmoiron/sqlx"
)

type LoginResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (LoginResource) Uri() string {
	return "/login"
}

func (LoginResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (LoginResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	packetdata := make(map[string]interface{})
	err := ReadOnJSONBody(r, &packetdata)
	log.Println("Login : Request Data  (", packetdata, ")")

	if err != nil {
		log.Println("Login : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "Login : " + err.Error(), Data: nil}
	}

	// 점검 공지시 code:204, Msg: db select msg, data: nil
	bMonitor, MonitorMsg, code := actiondb.CheckMonitorNotice(db)
	if bMonitor != 0 {
		log.Println("Login : Fail. cause ", MonitorMsg)
		return httpPkg.Response{Code: code, Msg: MonitorMsg, Data: nil}
	}

	// 계정 로그인
	accinfo := message.DBParamAccount{}
	err = db.Get(&accinfo, "select AccountUID, Type, CountryCode, Gender, Status, KeyID FROM Account WHERE KeyID = ?;", packetdata["keyid"])
	if err != nil {
		log.Println("Select Account Fail : ", err.Error())
		log.Println("Query : select AccountUID, Type, CountryCode, Gender, Status FROM Account WHERE KeyID = ", packetdata["keyid"])
		err = fmt.Errorf("Login : Find account info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 유저 정보
	userinfo := message.ResLogin{}
	err = db.Get(&userinfo, "call spGetUserInfo(?);", accinfo.AccountUID)
	if err != nil {
		log.Println("failed to procedure spGetUserInfo : ", accinfo.AccountUID)
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	userinfo.Gender = accinfo.Gender

	// SP 회복 가능
	if userinfo.Spin < userinfo.MaxSpin {
		beforeSpin := userinfo.Spin
		userinfo.Spin, userinfo.LastSpinRechargeTime = common.CalcRechargeSP(userinfo.Spin, userinfo.MaxSpin, userinfo.Now, userinfo.LastSpinRechargeTime)

		// 트랜잭션 시작
		//tx := db.MustBegin()

		if beforeSpin != userinfo.Spin {
			// Spin 및 접속 정보 갱신
			result, err := db.Exec("update User Set Spin = ?, LastSpinRechargeTime = ?, LoginDate = ?, UpdateDate = NOW(6) where AccountUID = ?;",
				userinfo.Spin, userinfo.LastSpinRechargeTime, userinfo.Now, userinfo.AccountUID)
			if err != nil {
				//tx.Rollback() // 중간에 에러가 나면 롤백
				log.Println("Update User Fail : ", err.Error())
				log.Println("Query : update User Set Spin = ", userinfo.Spin, ", LastSpinRechargeTime = ", userinfo.LastSpinRechargeTime, ", LoginDate = ", userinfo.Now, ", UpdateDate = NOW(6)", " where AccountUID = ?;", userinfo.AccountUID)
				err = fmt.Errorf("Login : Update user info fail.")
				return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil || rowsAffected == 0 {
				log.Println("Update User Fail : ", err.Error())
				log.Println("Query : update User Set Spin = ", userinfo.Spin, ", LastSpinRechargeTime = ", userinfo.LastSpinRechargeTime, ", LoginDate = ", userinfo.Now, ", UpdateDate = NOW(6)", " where AccountUID = ?;", userinfo.AccountUID)
				err = fmt.Errorf("Login : Update user info fail.")
				return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
			}
		}
	}

	// 닉네임 변경 시 유저 정보 업데이트
	if name, ok := packetdata["name"].(string); ok {
		if name != userinfo.Name && name != "" {
			result, err := db.Exec("update Account Set Name = ? where AccountUID = ?;", packetdata["name"], userinfo.AccountUID)
			if err != nil {
				log.Println("Update Account Fail : ", err.Error())
				log.Println("Query : update Account Set Name = ", packetdata["name"], " where AccountUID = ", userinfo.AccountUID)
				err = fmt.Errorf("Login : Update account nickname fail.")
				return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil || rowsAffected == 0 {
				log.Println("Update Account Fail : ", err.Error())
				log.Println("Query : update Account Set Name = ", packetdata["name"], " where AccountUID = ", userinfo.AccountUID)
				err = fmt.Errorf("Login : Update account nickname fail.")
				return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
			}

			result, err = db.Exec("update User Set Name = ? where AccountUID = ?;", packetdata["name"], userinfo.AccountUID)
			if err != nil {
				log.Println("Update User Fail : ", err.Error())
				log.Println("Query : update User Set Name = ", packetdata["name"], " where AccountUID = ", userinfo.AccountUID)
				err = fmt.Errorf("Login : Update user nickname fail.")
				return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
			}

			rowsAffected, err = result.RowsAffected()
			if err != nil || rowsAffected == 0 {
				log.Println("Update User Fail : ", err.Error())
				log.Println("Query : update User Set Name = ", packetdata["name"], " where AccountUID = ", userinfo.AccountUID)
				err = fmt.Errorf("Login : Update user nickname fail.")
				return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
			}

			userinfo.Name = name
		}
	}

	// 멘션정보 생성
	// 초기정보를 여기서 생성한다... db 함수에 들어가서 이미 존재하면 별짓 안하니까..
	// 근데 이 구조 별로다.. 수정 고려해봐야함
	crtmansion := message.CrtMansion{}
	crtmansion.AccountUID = int(accinfo.AccountUID)
	crtmansion.KeyID = accinfo.KeyID
	crtmansion.Name = userinfo.Name
	crtmansion.Gender = accinfo.Gender
	crtmansion.Type = accinfo.Type
	crtmansion.HaveLike = 0
	crtmansion.RemainLike = common.MANSION_LIKE // 1일 좋아요 제한 5회

	err, code = actiondb.CreateUserMansionInfo(&crtmansion, db)
	if err != nil {
		log.Println("Login : Fail. cause ", err.Error())
		return httpPkg.Response{Code: code, Msg: err.Error(), Data: nil}
	}

	log.Println("Login : Response Data (", userinfo, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: userinfo}
}
