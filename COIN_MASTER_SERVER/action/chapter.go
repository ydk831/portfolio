package action

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	"httpPkg"
	// xsql mysql driver
	"github.com/jmoiron/sqlx"
)

type ResChapter struct {
	CurrentChapterIDX      uint8     `db:"CurrentChapterIDX"`
	BuildingRewardGoldDate time.Time `db:"BuildingRewardGoldDate"`
	BuildingRewardSpinDate time.Time `db:"BuildingRewardSpinDate"`
	Now                    time.Time `db:"Now"`
	TileList               []DBParamTileList
}

type DBParamTileList struct {
	TileIdx    int8 `db:"TileIdx"`
	TileStatus int8 `db:"TileStatus"`
}

type ChapterResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (ChapterResource) Uri() string {
	return "/chapter"
}

func (ChapterResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	req := r.URL.Query()
	log.Println("chapter : Request Data  (", req, ")")

	// 현재 챕터 정보
	ResChapter := ResChapter{}
	err := db.Get(&ResChapter, "select CurrentChapterIDX, BuildingRewardGoldDate, BuildingRewardSpinDate, Now() Now from User where AccountUID = ?;", req["acc_uid"][0])
	if err != nil {
		log.Println("Select User Fail : ", err.Error())
		log.Println("Query : select CurrentChapterIDX, BuildingRewardDate, Now() Now FROM User WHERE AccountUID = ", req["acc_uid"][0])
		err = fmt.Errorf("Chapter : Find user info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 유저 타일 리스트 정보
	err = db.Select(&ResChapter.TileList, "select TileIdx, TileStatus FROM UserTile WHERE AccountUID = ? and TileStatus > 0;", req["acc_uid"][0])

	log.Println("chapter : Response Data  (", ResChapter, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: ResChapter}
}

func (ChapterResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) httpPkg.Response {
	packetdata := make(map[string]interface{})
	err := ReadOnJSONBody(r, &packetdata)
	//fmt.Println(body)

	log.Println("chapter : Request Data  (", packetdata, ")")
	if err != nil {
		log.Println("chapter : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "Chapter : " + err.Error(), Data: nil}
	}

	// 현재 챕터 정보
	ResChapter := ResChapter{}
	err = db.Get(&ResChapter, "select CurrentChapterIDX, BuildingRewardGoldDate, BuildingRewardSpinDate, Now() Now FROM User WHERE AccountUID = ?;", packetdata["acc_uid"])
	if err != nil {
		log.Println("Select User Fail : ", err.Error())
		log.Println("Query : select CurrentChapterIDX, BuildingRewardDate, Now() Now FROM User WHERE AccountUID = ", packetdata["acc_uid"])
		err = fmt.Errorf("Chapter : Find user info fail.")
		return httpPkg.Response{Code: 500, Msg: err.Error(), Data: nil}
	}

	// 유저 타일 리스트 정보
	err = db.Select(&ResChapter.TileList, "select TileIdx, TileStatus FROM UserTile WHERE AccountUID = ? and TileStatus > 0 ;", packetdata["acc_uid"])

	log.Println("chapter : Response Data  (", ResChapter, ")")
	return httpPkg.Response{Code: 200, Msg: "success", Data: ResChapter}
	//return httpPkg.Response{Code: 405, Msg: "Not Defined POST Method", Data: nil}
}
