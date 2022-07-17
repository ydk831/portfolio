package common

import (
	"encoding/json"
	"log"
	"message"
)

func ParseReceipt(src map[string]interface{}) (message.Receipt, error) {
	defer func() {
		// 런타임 에러가 발생하면 recover 함수가 실행됨
		if r := recover(); r != nil { // log.Panic 함수에서 출력한 에러 문자열 리턴
			log.Println("Panic Caused : ", r) // 에러 문자열 출력
		}
	}()

	receipt := message.Receipt{}
	rPayload := message.ReceiptPayload{}
	rPLJson := message.PayloadJson{}
	rPLSku := message.PayloadSkuDetails{}
	rPLJsonDevPL := message.PayloadJsonDeveloperPayload{}

	//recipt convert
	{
		receipt.Store = src["Store"].(string)
		receipt.TransactionID = src["TransactionID"].(string)

		Payload := src["Payload"].(string)
		Payload2Byte := []byte(Payload)
		rPayloadByte := make(map[string]interface{})
		err := json.Unmarshal(Payload2Byte, &rPayloadByte)
		if err != nil {
			log.Println("parseReceipt : Payload convert Error.")
			log.Println(err.Error())
			return receipt, err
			//return httpPkg.Response{Code: 400, Msg: "ShopG : receipt paylaod is wrong.", Data: nil}
		}

		// receipt.Payload convert
		{
			rPayload.IsPurchaseHistorySupported = rPayloadByte["isPurchaseHistorySupported"].(bool)
			rPayload.Signature = rPayloadByte["signature"].(string)

			// recipt.Json convert
			{
				rPLBJson := rPayloadByte["json"].(string)
				rPLBJson2Byte := []byte(rPLBJson)
				rPLJsonByte := make(map[string]interface{})
				err = json.Unmarshal(rPLBJson2Byte, &rPLJsonByte)
				if err != nil {
					log.Println("parseReceipt : Payload Json convert Error.")
					log.Println(err.Error())
					return receipt, err
					//return httpPkg.Response{Code: 400, Msg: "ShopG : receipt paylaod is wrong.", Data: nil}
				}

				rPLJson.OrderId = rPLJsonByte["orderId"].(string)
				rPLJson.PackageName = rPLJsonByte["packageName"].(string)
				rPLJson.ProductId = rPLJsonByte["productId"].(string)
				rPLJson.PurchaseTime = uint64(rPLJsonByte["purchaseTime"].(float64))
				rPLJson.PurchaseState = uint64(rPLJsonByte["purchaseState"].(float64))
				rPLJson.PurchaseToken = rPLJsonByte["purchaseToken"].(string)

				// receipt.Json.developerPayload convert
				{
					rPLBJsonDevPL := rPLJsonByte["developerPayload"].(string)
					rPLBJsonDevPL2Byte := []byte(rPLBJsonDevPL)
					rPLJsonDevPLByte := make(map[string]interface{})
					err = json.Unmarshal(rPLBJsonDevPL2Byte, &rPLJsonDevPLByte)
					if err != nil {
						log.Println("parseReceipt : Payload Json DeveloperPayload convert Error.")
						log.Println(err.Error())
						return receipt, err
						//return httpPkg.Response{Code: 400, Msg: "ShopG : receipt paylaod is wrong.", Data: nil}
					}

					rPLJsonDevPL.DeveloperPayload = rPLJsonDevPLByte["developerPayload"].(string)
					rPLJsonDevPL.Isfreetrial = rPLJsonDevPLByte["is_free_trial"].(bool)
					rPLJsonDevPL.HasIntroductoryPriceTrial = rPLJsonDevPLByte["has_introductory_price_trial"].(bool)
					rPLJsonDevPL.IsUpdate = rPLJsonDevPLByte["is_updated"].(bool)
					rPLJsonDevPL.AccountID = rPLJsonDevPLByte["accountId"].(string)
				} // end of receipt.Json.developerPayload
				rPLJson.DeveloperPayload = rPLJsonDevPL
			} // end of receipt.Json
			rPayload.Json = rPLJson

			// recipt.SkuDetails convert
			{
				rPLBSku := rPayloadByte["skuDetails"].(string)
				rPLBSku2Byte := []byte(rPLBSku)
				rPLSkuByte := make(map[string]interface{})
				err = json.Unmarshal(rPLBSku2Byte, &rPLSkuByte)
				if err != nil {
					log.Println("parseReceipt : Payload SkuDetails convert Error.")
					log.Println(err.Error())
					return receipt, err
					//return httpPkg.Response{Code: 400, Msg: "ShopG : receipt paylaod is wrong.", Data: nil}
				}

				rPLSku.SkuDetailsToken = rPLSkuByte["skuDetailsToken"].(string)
				rPLSku.ProductId = rPLSkuByte["productId"].(string)
				rPLSku.Type = rPLSkuByte["type"].(string)
				rPLSku.Price = rPLSkuByte["price"].(string)
				rPLSku.PriceAmountMicros = rPLSkuByte["price_amount_micros"].(float64)
				rPLSku.PriceCurrencyCode = rPLSkuByte["price_currency_code"].(string)
				rPLSku.Title = rPLSkuByte["title"].(string)
				rPLSku.Description = rPLSkuByte["description"].(string)
			} // end of recipt.SkuDetails
			rPayload.SkuDetails = rPLSku

		} // end of recipt.Payload2Byte
		receipt.Payload = rPayload

	} // end of recipt
	return receipt, nil
}
