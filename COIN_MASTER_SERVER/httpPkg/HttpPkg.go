package httpPkg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	// xsql mysql driver
	"github.com/jmoiron/sqlx"
)

// ToJSON : URL에 담긴 파라미터를 json 형식으로 변환, 라이언트에게 내려줄 데이터가 있을 때 쓰면 됨
func ToJSON(m interface{}) string {
	js, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}
	return strings.ReplaceAll(string(js), ",", ", ")
}

// Response : 클라이언트 응답 데이터 구조체
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

// Resource : 클라이언트 요청 처리 API
type Resource interface {
	Uri() string
	Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) Response
	Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) Response
	Put(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) Response
	Delete(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) Response
}

// 사용하지 않을 API의 Flag
type (
	// GetNotSupported ...
	GetNotSupported struct{}
	// PostNotSupported ...
	PostNotSupported struct{}
	// PutNotSupported ...
	PutNotSupported struct{}
	// DeleteNotSupported ...
	DeleteNotSupported struct{}
)

// Get Method 정의
func (GetNotSupported) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) Response {
	return Response{405, "", nil}
}

// Post Method 정의
func (PostNotSupported) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) Response {
	return Response{405, "", nil}
}

// Put Method 정의
func (PutNotSupported) Put(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) Response {
	return Response{405, "", nil}
}

// Delete Method 정의
func (DeleteNotSupported) Delete(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, db *sqlx.DB) Response {
	return Response{405, "", nil}
}

func abort(rw http.ResponseWriter, statusCode int) {
	rw.WriteHeader(statusCode)
}

// HttpResponse : 클라이언트 응답 수행 API
func HttpResponse(rw http.ResponseWriter, req *http.Request, res Response) {
	content, err := json.Marshal(res)

	if err != nil {
		abort(rw, 500)
	}

	rw.WriteHeader(res.Code)
	rw.Write(content)
}

// AddResource : httprouter에 Resource API 등록
func AddResource(router *httprouter.Router, resource Resource, db *sqlx.DB) {
	fmt.Println("\"" + resource.Uri() + "\" api is registerd")

	router.GET(resource.Uri(), func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		res := resource.Get(rw, r, ps, db)
		HttpResponse(rw, r, res)
	})
	router.POST(resource.Uri(), func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		res := resource.Post(rw, r, ps, db)
		HttpResponse(rw, r, res)
	})
	router.PUT(resource.Uri(), func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		res := resource.Put(rw, r, ps, db)
		HttpResponse(rw, r, res)
	})
	router.DELETE(resource.Uri(), func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		res := resource.Delete(rw, r, ps, db)
		HttpResponse(rw, r, res)
	})
}
