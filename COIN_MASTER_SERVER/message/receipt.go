package message

import (
	"github.com/awa/go-iap/appstore"
)

/* Receipt For Goolge Store */
type Receipt struct {
	Payload       ReceiptPayload `json:"Payload"`
	Store         string         `json:"Store"`
	TransactionID string         `json:"TransactionID"`
}

type ReceiptPayload struct {
	IsPurchaseHistorySupported bool              `json:"isPurchaseHistorySupported"`
	Signature                  string            `json:"signature"`
	Json                       PayloadJson       `json:"json"`
	SkuDetails                 PayloadSkuDetails `json:"skuDetails"`
}

type PayloadJson struct {
	OrderId          string
	PackageName      string
	ProductId        string
	PurchaseTime     uint64
	PurchaseState    uint64
	DeveloperPayload PayloadJsonDeveloperPayload
	PurchaseToken    string
}

type PayloadSkuDetails struct {
	SkuDetailsToken   string
	ProductId         string
	Type              string
	Price             string
	PriceAmountMicros float64
	PriceCurrencyCode string
	Title             string
	Description       string
}

type PayloadJsonDeveloperPayload struct {
	DeveloperPayload          string
	Isfreetrial               bool
	HasIntroductoryPriceTrial bool
	IsUpdate                  bool
	AccountID                 string
}

/* Receipt For Goolge Store END*/

/* Receipt For Apple Store */
type ReceiptIOS struct {
	AccountUID        int
	Store             string
	TransactionID     string
	ProductID         string
	Price             int
	PriceCurrencyCode string
	Description       string
	Payload           appstore.Receipt
}

/* Receipt For Apple Store END*/
