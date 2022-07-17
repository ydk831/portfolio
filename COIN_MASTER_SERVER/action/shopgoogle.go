package action

import (
	"context"
	"db"
	"encoding/json"
	"io/ioutil"
	"log"
	"message"

	"common"
	"net/http"

	"github.com/awa/go-iap/playstore"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"

	"httpPkg"
)

type ShopGResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (ShopGResource) Uri() string {
	return "/shopg"
}

func (ShopGResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (ShopGResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	packetdata := make(map[string]interface{})
	err := ReadOnJSONBody(r, &packetdata)

	log.Println("ShopG : Request Data (", packetdata, ")")

	if err != nil {
		log.Println("ShopG :", err.Error())
		return httpPkg.Response{Code: 400, Msg: "ShopG : Request Data Format Error.", Data: nil}
	}

	uid := int(packetdata["acc_uid"].(float64))
	receipt := message.Receipt{}
	receipt, err = common.ParseReceipt(packetdata)
	if err != nil {
		log.Println("ShopG :", err.Error())
		return httpPkg.Response{Code: 400, Msg: "ShopG : Receipt Data Parse Error.", Data: nil}
	} else {
		// You need to prepare a public key for your Android app's in app billing
		// at https://console.developers.google.com.
		//jsonKey, err := ioutil.ReadFile("C:\\Users\\ydk831\\Desktop\\Coin\\project_coin_server\\coin2srv.json")
		jsonKey, err := ioutil.ReadFile("/home/ec2-user/project_coin_server/coin2srv.json")
		if err != nil {
			log.Println("ShopG : Read Config File Error.", err.Error())
			return httpPkg.Response{Code: 500, Msg: "ShopG : authentication file load error.", Data: nil}
		}

		client, _ := playstore.New(jsonKey)
		ctx := context.Background()
		resp, err := client.VerifyProduct(ctx, receipt.Payload.Json.PackageName, receipt.Payload.Json.ProductId, receipt.Payload.Json.PurchaseToken)
		if err != nil {
			log.Println("ShopG : VerifyProduct Error. ", err.Error())
			return httpPkg.Response{Code: 400, Msg: "ShopG : GSE Verifying error.", Data: nil}
		} else {
			// resp를 그대로 쓸 수도 있었을것 같은데.. 아마 개발할때 \n같은 특수문자때문에 다시 인코딩 한거같음
			rsp, _ := resp.MarshalJSON()
			auth := make(map[string]interface{})
			err := json.Unmarshal(rsp, &auth)
			if err != nil {
				log.Println("ShopG :", err.Error())
				return httpPkg.Response{Code: 400, Msg: "ShopG : GSE Verifying data is worng.", Data: nil}
			}
			// auth data : https://developers.google.com/android-publisher/api-ref/purchases/products

			purchaseType := int(auth["purchaseType"].(float64))
			// 0 : Purchased
			// 1 : Canceled
			// 2 : Pending
			switch purchaseType {
			case 0:
				// 1. 영수증 저장.
				err = db.SaveReceiptGoogle(uid, receipt, mydb)
				if err != nil {
					log.Println("ShopG : Receipt Save Error!! Write Log.")
					r, err := json.Marshal(receipt)
					if err != nil {
						log.Println("ShopG : receipt convert error. ", err.Error())
					}
					log.Println("ShopG : ", string(r))
				}
				// 2. 보상 지급
				item, err, code := db.RewardShop(uid, receipt, mydb)
				if err != nil {
					log.Println("ShopG :", err.Error())
					return httpPkg.Response{Code: code, Msg: "ShopG : " + err.Error(), Data: nil}
				}
				switch item.RewardType {
				case 1: // Spin
					response, err, code := db.UpdateSpinItem(uid, item.RewardValue, mydb)
					if err != nil {
						log.Println("ShopG :", err.Error())
						return httpPkg.Response{Code: code, Msg: "ShopG : " + err.Error(), Data: nil}
					} else {
						log.Println("ShopG : Response Data (", response, ")")
						return httpPkg.Response{Code: code, Msg: "ShopG : Update Success", Data: response}
					}
				case 2: // Gold
					response, err, code := db.UpdateGoldItem(uid, item.RewardValue, mydb)
					if err != nil {
						log.Println("ShopG :", err.Error())
						return httpPkg.Response{Code: code, Msg: "ShopG : " + err.Error(), Data: nil}
					} else {
						log.Println("ShopG : Response Data (", response, ")")
						return httpPkg.Response{Code: code, Msg: "ShopG : Update Success", Data: response}
					}
				case 3: // multiple spin count AdvUpdateDate
					log.Println("ShopG : Not Yet..")
					return httpPkg.Response{Code: 200, Msg: "ShopG : Not Yet..", Data: nil}
				case 4: // RandomBox
					response := message.ShopUpdateResponse{Box: item.RewardValue, IsUpdate: true}
					log.Println("ShopG : Response Data (", response, ")")
					return httpPkg.Response{Code: code, Msg: "ShopG : Update Success", Data: response}
				default: // error
					log.Println("ShopG : Not Defined Reward Type.")
					return httpPkg.Response{Code: 500, Msg: "ShopG : Not Defined Reward Type.", Data: nil}
				}
			case 1:
				log.Println("ShopG : Google receipt was canceled.")
				return httpPkg.Response{Code: 400, Msg: "ShopG : Google receipt was canceled.", Data: nil}
			case 2:
				log.Println("ShopG : Google receipt is pending.")
				return httpPkg.Response{Code: 400, Msg: "ShopG : Google receipt is pending.", Data: nil}
			default:
				log.Println("ShopG : Google authentication server error.")
				return httpPkg.Response{Code: 400, Msg: "ShopG : Google authentication server error.", Data: nil}
			}
		}
	}
}
