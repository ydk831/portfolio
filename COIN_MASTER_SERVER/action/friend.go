package action

import (
	"db"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"

	"httpPkg"
)

type FriendResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (FriendResource) Uri() string {
	return "/friend"
}

func (FriendResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	qry := r.URL.Query()
	var keyidlist []string

	log.Println("Friend : Request Data (", qry, ")")

	if qry["keyid[]"] != nil {
		for _, v := range qry["keyid[]"] {
			keyidlist = append(keyidlist, v)
		}

		rsp, err, code := db.FindFriend(keyidlist, mydb)
		if err != nil {
			log.Println("Friend : )", err.Error())
			return httpPkg.Response{Code: code, Msg: "Friend : " + err.Error(), Data: nil}
		}
		log.Println("Friend : Response Data (", rsp, ")")
		return httpPkg.Response{Code: code, Msg: "Find Friend Info Success", Data: rsp}
	} else {
		log.Println("Friend : Can't recived firend list)")
		return httpPkg.Response{Code: 400, Msg: "Friend : Can't recived firend list", Data: nil}
	}
}

func (FriendResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined POST Method", Data: nil}
}
