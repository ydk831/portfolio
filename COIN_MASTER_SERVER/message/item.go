package message

type Item struct {
	RewardType    int `db:"ShopRewardType"`
	RewardValue   int `db:"ShopRewardValue"`
	PurchaseLimit int `db:"PurchaseLimit"`
}
