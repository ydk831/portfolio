package action

import (
	"context"
	"db"
	"encoding/json"
	"log"
	"message"

	"net/http"

	"github.com/awa/go-iap/appstore"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"

	"httpPkg"
)

type ShopAResource struct {
	httpPkg.PutNotSupported
	httpPkg.DeleteNotSupported
}

func (ShopAResource) Uri() string {
	return "/shopa"
}

func (ShopAResource) Get(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	return httpPkg.Response{Code: 405, Msg: "Not Defined GET Method", Data: nil}
}

func (ShopAResource) Post(rw http.ResponseWriter, r *http.Request, ps httprouter.Params, mydb *sqlx.DB) httpPkg.Response {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	packetdata := make(map[string]interface{})
	err := ReadOnJSONBody(r, &packetdata)

	log.Println("ShopA : Request Data (", packetdata, ")")

	receipt := message.ReceiptIOS{
		AccountUID:        int(packetdata["acc_uid"].(float64)),
		Store:             packetdata["Store"].(string),
		TransactionID:     packetdata["TransactionID"].(string),
		Price:             int(packetdata["Price"].(float64)),
		PriceCurrencyCode: packetdata["PriceCurrencyCode"].(string),
		Description:       packetdata["Description"].(string)}

	/*
		receipt := message.ReceiptIOS{
			AccountUID:        int(packetdata["acc_uid"].(float64)),
			Store:             packetdata["Store"].(string),
			TransactionID:     packetdata["TransactionID"].(string),
			Price:             1000,
			PriceCurrencyCode: "KRW",
			Description:       "TEST"}
	*/

	if err != nil {
		log.Println("ShopA : ", err.Error())
		return httpPkg.Response{Code: 400, Msg: "ShopA : " + err.Error(), Data: nil}
	} else {
		client := appstore.New()
		req := appstore.IAPRequest{
			ReceiptData: packetdata["Payload"].(string),
		}
		resp := &appstore.IAPResponse{}
		ctx := context.Background()
		err = client.Verify(ctx, req, resp)

		if err != nil {
			log.Println("ShopA : VerifyProduct Error. ", err.Error())
			return httpPkg.Response{Code: 400, Msg: "ShopA : AppStore Verifying error.", Data: nil}
		} else {
			/*
				https://developer.apple.com/library/archive/releasenotes/General/ValidateAppStoreReceipt/Chapters/ValidateRemotely.html
				latest_receipt및 latest_receipt_info키 값은 자동 갱신 가능한 구독이 현재 활성화되어 있는지 확인할 때 유용합니다.
				latest_expired_receipt_info키 값은 자동 갱신 가능한 구독이 만료되었는지 여부를 확인할 때 유용합니다. 만기 이유를 얻으려면 서브 스크립 션 만기 의도 값과 함께이를 사용하십시오 .
				pending_renewal_info키 값은 자동 갱신 가능 구독의 보류중인 갱신 트랜잭션에 대한 중요한 정보를 얻는 데 유용합니다.
				구독에 대한 앱 영수증 또는 거래 영수증을 제공하고 이러한 값을 확인하면 현재 활성화 된 구독 기간에 대한 정보를 얻을 수 있습니다. 확인중인 영수증이 최신 갱신 용인 경우의 값 latest_receipt은 receipt-data(요청시)와 latest_receipt_info같고 의 값은 입니다 receipt.
			*/
			switch resp.Status {
			case 0: // 성공
				// 1. 영수증 저장
				receipt.ProductID = resp.Receipt.InApp[0].ProductID
				receipt.Payload = resp.Receipt

				err = db.SaveReceiptApple(&receipt, mydb)
				if err != nil {
					log.Println("ShopA : Receipt Save Error!! Write Log.")
					r, err := json.Marshal(receipt)
					if err != nil {
						log.Println("ShopA : receipt convert error. ", err.Error())
					}
					log.Println("ShopA : ", string(r))
				}

				// 2. 보상지급
				item, err, code := db.RewardShop(receipt.AccountUID, receipt, mydb)
				if err != nil {
					log.Println("ShopA :", err.Error())
					return httpPkg.Response{Code: code, Msg: "ShopA : " + err.Error(), Data: nil}
				}
				switch item.RewardType {
				case 1: // Spin
					response, err, code := db.UpdateSpinItem(receipt.AccountUID, item.RewardValue, mydb)
					if err != nil {
						log.Println("ShopA :", err.Error())
						return httpPkg.Response{Code: code, Msg: "ShopA : " + err.Error(), Data: nil}
					} else {
						log.Println("ShopA : Response Data (", response, ")")
						return httpPkg.Response{Code: code, Msg: "ShopA : Update Success", Data: response}
					}
				case 2: // Gold
					response, err, code := db.UpdateGoldItem(receipt.AccountUID, item.RewardValue, mydb)
					if err != nil {
						log.Println("ShopA :", err.Error())
						return httpPkg.Response{Code: code, Msg: "ShopA : " + err.Error(), Data: nil}
					} else {
						log.Println("ShopG : Response Data (", response, ")")
						return httpPkg.Response{Code: code, Msg: "ShopA : Update Success", Data: response}
					}
				case 3: // multiple spin count AdvUpdateDate
					log.Println("ShopA : Not Yet..")
					return httpPkg.Response{Code: 200, Msg: "ShopA : Not Yet..", Data: nil}
				case 4: // RandomBox
					response := message.ShopUpdateResponse{Box: item.RewardValue, IsUpdate: true}
					log.Println("ShopA : Response Data (", response, ")")
					return httpPkg.Response{Code: code, Msg: "ShopG : Update Success", Data: response}
				default: // error
					log.Println("ShopA : Not Defined Reward Type.")
					return httpPkg.Response{Code: 500, Msg: "ShopA : Not Defined Reward Type.", Data: nil}
				}

			case 21000: // App Store가 제공 한 JSON 객체를 읽을 수 없습니다.
				log.Println("ShopA : The App Store could not read the JSON object you provided.")
				return httpPkg.Response{Code: 400, Msg: "ShopA : The App Store could not read the JSON object you provided.", Data: nil}
			case 21002: // receipt-data속성 의 데이터 가 잘못되었거나 누락되었습니다.
				log.Println("ShopA : The data in the receipt-data property was malformed or missing.")
				return httpPkg.Response{Code: 400, Msg: "ShopA : The data in the receipt-data property was malformed or missing.", Data: nil}
			case 21003: // 영수증을 인증 할 수 없습니다.
				log.Println("ShopA : The receipt could not be authenticated.")
				return httpPkg.Response{Code: 400, Msg: "ShopA : The receipt could not be authenticated.", Data: nil}
			case 21004: // 제공 한 공유 비밀번호가 계정의 파일에있는 공유 비밀번호와 일치하지 않습니다.
				log.Println("ShopA : The shared secret you provided does not match the shared secret on file for your account.")
				return httpPkg.Response{Code: 400, Msg: "ShopA : The shared secret you provided does not match the shared secret on file for your account.", Data: nil}
			case 21005: // 영수증 서버를 현재 사용할 수 없습니다.
				log.Println("ShopA : The receipt server is not currently available.")
				return httpPkg.Response{Code: 400, Msg: "ShopA : The receipt server is not currently available.", Data: nil}
			case 21006: // 이 영수증은 유효하지만 구독이 만료되었습니다. 이 상태 코드가 서버로 반환되면 영수증 데이터도 디코딩되어 응답의 일부로 반환됩니다. 자동 갱신 구독의 iOS 6 스타일 거래 영수증에 대해서만 반환됩니다.
				log.Println("ShopA : This receipt is valid but the subscription has expired. When this status code is returned to your server, the receipt data is also decoded and returned as part of the response. Only returned for iOS 6 style transaction receipts for auto-renewable subscriptions.")
				return httpPkg.Response{Code: 400, Msg: "ShopA : This receipt is valid but the subscription has expired. When this status code is returned to your server, the receipt data is also decoded and returned as part of the response. Only returned for iOS 6 style transaction receipts for auto-renewable subscriptions.", Data: nil}
			case 21007: // 이 영수증은 테스트 환경에서 제공되었지만 검증을 위해 프로덕션 환경으로 전송되었습니다. 대신 테스트 환경으로 보내십시오.
				log.Println("ShopA : This receipt is from the test environment, but it was sent to the production environment for verification. Send it to the test environment instead.")
				return httpPkg.Response{Code: 400, Msg: "ShopA : This receipt is from the test environment, but it was sent to the production environment for verification. Send it to the test environment instead.", Data: nil}
			case 21008: // 이 영수증은 프로덕션 환경에서 제공되었지만 검증을 위해 테스트 환경으로 전송되었습니다. 대신 프로덕션 환경으로 보내십시오.
				log.Println("ShopA : This receipt is from the production environment, but it was sent to the test environment for verification. Send it to the production environment instead.")
				return httpPkg.Response{Code: 400, Msg: "ShopA : This receipt is from the production environment, but it was sent to the test environment for verification. Send it to the production environment instead.", Data: nil}
			case 21009: // 이 영수증을 승인 할 수 없습니다. 구매 한 적이없는 것처럼 취급하십시오.
				log.Println("ShopA : This receipt could not be authorized. Treat this the same as if a purchase was never made.")
				return httpPkg.Response{Code: 400, Msg: "ShopA : This receipt could not be authorized. Treat this the same as if a purchase was never made.", Data: nil}
			default: //21100-21199 : 내부 데이터 액세스 오류
				log.Println("ShopA : Internal data access error(", resp.Status, ")")
				return httpPkg.Response{Code: 400, Msg: "ShopA : Internal data access error.", Data: nil}
			}
		}
	}
}
