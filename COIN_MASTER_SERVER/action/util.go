package action

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func ReadJSONBody(r *http.Request) ([]byte, error) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			fmt.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Panic(err)
		return nil, fmt.Errorf("HttpRequest Body ReadAll Fail!")
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Panic(err)
		return nil, fmt.Errorf("HttpRequest Body JSON Unmarshal Fail!")
	}

	for key, value := range data {
		if value == nil || value == "" {
			log.Println(err)
			return nil, fmt.Errorf("JSON Body Key[%s] is Empty", key)
		}
	}
	return body, nil
}

func ReadOnJSONBody(r *http.Request, data interface{}) error {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			fmt.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Panic("[ReadOnJSONBody] " + err.Error())
		return fmt.Errorf("HttpRequest Body ReadAll Fail!")
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Panic("[ReadOnJSONBody] " + err.Error())
		return fmt.Errorf("HttpRequest Body JSON Unmarshal Fail!")
	}

	return nil
}
