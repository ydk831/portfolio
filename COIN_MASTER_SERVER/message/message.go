package message

import (
	"time"
)

type DBParamAccount struct {
	AccountUID  uint32 `db:"AccountUID"`
	Name        string `db:"Name"`
	Type        int8   `db:"Type"`
	KeyID       string `db:"KeyID"`
	CountryCode string `db:"CountryCode"`
	Gender      int8   `db:"Gender"`
	Status      int8   `db:"Status"`
}

type DBParamUser struct {
	AccountUID           uint32    `db:"AccountUID"`
	Name                 string    `db:"Name"`
	Gold                 int64     `db:"Gold"`
	Spin                 int16     `db:"Spin"`
	MaxSpin              int16     `db:"MaxSpin"`
	LastSpinRechargeTime time.Time `db:"LastSpinRechargeTime"`
	Shield               int8      `db:"Shield"`
	CurrentChapterIDX    int8      `db:"CurrentChapterIDX"`
	Now                  time.Time `db:"Now"`
	Gender               int8      `db"Gender"`
	AdvGoldRewardCnt     int32     `db:"AdvGoldRewardCnt"`
	AdvSpinRewardCnt     int32     `db:"AdvSpinRewardCnt"`
	Coin                 int32     `db:"Coin"` // 여기다 넣긴 싫었는데.. 나중에 고치던지 그냥 쓰던지..
}

/*
type UserInfo struct {
	AccountUID             uint32    `db:"AccountUID"`
	Name                   string    `db:"Name"`
	Gold                   int64     `db:"Gold"`
	Spin                   int16     `db:"Spin"`
	MaxSpin                int16     `db:"MaxSpin"`
	LastSpinRechargeTime   time.Time `db:"LastSpinRechargeTime"`
	Shield                 int8      `db:"Shield"`
	CurrentChapterIDX      int8      `db:"CurrentChapterIDX"`
	BuildingRewardGoldDate time.Time `db:"BuildingRewardGoldDate"`
	BuildingRewardSpinDate time.Time `db:"BuildingRewardSpinDate"`
	AdvGoldRewardCnt       int32     `db:"AdvGoldRewardCnt"`
	AdvSpinRewardCnt       int32     `db:"AdvSpinRewardCnt"`
	Now                    time.Time `db:"Now"`
}
*/

type ResLogin struct {
	AccountUID             uint32    `db:"AccountUID"`
	Name                   string    `db:"Name"`
	Gold                   int64     `db:"Gold"`
	Spin                   int16     `db:"Spin"`
	MaxSpin                int16     `db:"MaxSpin"`
	LastSpinRechargeTime   time.Time `db:"LastSpinRechargeTime"`
	Shield                 int8      `db:"Shield"`
	CurrentChapterIDX      int8      `db:"CurrentChapterIDX"`
	BuildingRewardGoldDate time.Time `db:"BuildingRewardGoldDate"`
	BuildingRewardSpinDate time.Time `db:"BuildingRewardSpinDate"`
	AdvGoldRewardCnt       int32     `db:"AdvGoldRewardCnt"`
	AdvSpinRewardCnt       int32     `db:"AdvSpinRewardCnt"`
	Tutorial               int8      `db:"Tutorial"`
	InviteCount            int8      `db:"InviteCount"`
	Now                    time.Time `db:"Now"`
	Gender                 int8      `db:"Gender"`
	Coin                   int       `db:"Coin"`
}

type CrtMansion struct {
	AccountUID int
	KeyID      string
	Name       string
	Gender     int8
	Type       int8
	HaveLike   int
	RemainLike int
}

type AttackRequest struct {
	ID            uint
	MultipleValue int
	TargetID      uint
}

type AttackRspUserInfo struct {
	TargetID         uint   `json:"target_id" db:"AccountUID"`
	TargetNick       string `json:"target_name" db:"Name"`
	TargetGold       uint64 `json:"target_gold" db:"Gold"`
	TargetSpin       int    `json:"target_spin" db:"Spin"`
	TargetShield     int    `json:"target_shield" db:"Shield"`
	TargetChapterIdx int    `json:"target_chapter_index" db:"CurrentChapterIDX"`
	TargetGender     int8   `json:"target_gender" db:"Gender"`
	TargetAppType    int8   `json:"target_apptype" db:"Type"`
	TargetKeyid      string `json:"target_keyid" db:"Keyid"`
}

type AttackRspTileInfo struct {
	TileIdx        int       `json:"TileIdx" db:"TileIdx" `
	TileStatus     int       `json:"TileStatus" db:"TileStatus"`
	TileChargeTime time.Time `json:"TileChargeTime" db:"TileChargeTime"`
}

type AttackResponse struct {
	MultipleValue int                 `json:"multiple_value"`
	UserInfo      AttackRspUserInfo   `json:"UserInfo"`
	TileInfo      []AttackRspTileInfo `json:"TileInfo"`
}

type AttackResultRequest struct {
	ID            uint   `json:"acc_uid"`
	TargetID      uint   `json:"target_id"`
	StealGold     uint32 `json:"steal_gold"`
	IsSuccess     bool   `json:"is_success"`
	AttackTileIdx int    `json:"attack_tile_index"`
}

type AttackResultResponse struct {
	IsUpdate bool `json:"is_update"`
}

type RaidResultRequest struct {
	ID            uint   `json:"acc_uid"`
	TargetID      uint   `json:"target_id"`
	StealGold     uint32 `json:"steal_gold"`
	MultipleValue uint   `json:"multiple_value"`
}

type News struct {
	MyNews []NewsInfo `json:"news_info"`
}

type NewsInfo struct {
	EnemyID       uint      `json:"enemy_id" db:"AttackUser"`
	EnemyName     string    `json:"enemy_name" db:"AttackName"`
	ResultType    int8      `json:"result_type" db:"Result"`
	StolenGold    int       `json:"stolen_gold" db:"StolenGold"`
	AttackTime    time.Time `json:"attack_time" db:"AttackTime"`
	AttackGender  int8      `json:"enemy_gender" db:"AttackGender"`
	AttackAppType int8      `json:"enemy_apptype" db:"AttackAppType"`
	EnemyKeyid    string    `json:"enemy_keyid" db:"AttackKeyid"`
}

type AttackerInfo struct {
	Name    string `db:"Name"`
	Gender  int8   `db:"Gender"`
	AppType int8   `db:"Type"`
	Keyid   string `db:"Keyid"`
}

type AdvRewardRequest struct {
	ID         int    `json:"acc_uid"`
	RewardType string `json:"type"`
	Reward     int    `json:"reward"`
}

type AdvRewardResponse struct {
	IsUpdate      bool `json:"is_update"`
	GoldRewardCnt int  `json:"gold_reward_cnt" db:"AdvGoldRewardCnt"`
	SpinRewardCnt int  `json:"spin_reward_cnt" db:"AdvSpinRewardCnt"`
}

type TargetResponse struct {
	TargetID      uint   `json:"target_id" db:"AccountUID"`
	TargetNick    string `json:"target_name" db:"Name"`
	TargetGender  int8   `json:"target_gender" db:"Gender"`
	TargetAppType int8   `json:"target_apptype" db:"Type"`
	TargetGold    uint32 `json:"target_gold" db:"Gold"`
	TargetKeyid   string `json:"target_keyid" db:"Keyid"`
}

type FriendResponse struct {
	MyFriends []FriendInfo
}

type FriendInfo struct {
	FriendID      uint   `json:"friend_id" db:"AccountUID"`
	FriendNick    string `json:"friend_name" db:"Name"`
	FriendGender  int8   `json:"friend_gender" db:"Gender"`
	FriendAppType int8   `json:"friend_apptype" db:"Type"`
	FriendKeyid   string `json:"friend_keyid" db:"Keyid"`
}

type ShopUpdateResponse struct {
	Gold     uint64 `json:"update_gold" db:"Gold"`
	Spin     int    `json:"update_spin" db:"Spin"`
	Box      int    `json:"random_box" db:"Box"`
	IsUpdate bool   `json:"is_update"`
}

type NotiInfoResponse struct {
	Notice []NotiInfo `json:"noti_info"`
}

type NotiReward struct {
	RewardGold int `json:"reward_gold" db:"RewardGold"`
	RewardSpin int `json:"reward_spin" db:"RewardSpin"`
}

type NotiInfo struct {
	NotiID         int        `json:"noti_id" db:"NotiID"`
	NotiType       int        `json:"noti_type" db:"NotiType"`
	NotiRewardInfo NotiReward `json:"noti_reward" `
	NotiStartDate  time.Time  `json:"start_date" db:"StartDate"`
	NotiEndDate    time.Time  `json:"end_date" db:"EndDate"`
	NotiMsg        string     `json:"noti_msg" db:"NotiMsg"`
}

type EventResultRequest struct {
	ID            int `json:"acc_uid" db:""`
	EventID       int `json:"event_id" db:""`
	Score         int `json:"score" db:""`
	MultipleValue int `json:"multiple_value" db:""`
}

type EventResultResponse struct {
	ID                   int       `json:"acc_uid" db:""`
	EventID              int       `json:"event_id" db:""`
	PlusGold             int       `json:"plus_gold" db:"RewardGold"`
	PlusSpin             int       `json:"plus_spin" db:"RewardSpin"`
	PlusCoin             int       `json:"plus_coin" db:"RewardEventCoin"`
	UserGold             int       `json:"user_gold" db:""`
	UserSpin             int       `json:"user_spin" db:""`
	UserCoin             int       `json:"user_coin" db:""`
	LastSpinRechargeTime time.Time `json:"spin_recharge_time" db:""`
}

type SlotResultResponse struct {
	ResultType       int
	Gold             int64
	Spin             int16
	Shield           int8
	Coin             int32
	SpinRechargeTime time.Time
	MansionItem      int
}

type MansionInfoResponse struct {
	ID            int           `json:"acc_uid" db:"AccountUID"`
	MansionIndex  int           `json:"mansion_index" db:"Mansion"`
	MansionRoom   int           `json:"mansion_room" db:"MansionRoom"`
	KeyID         string        `json:"keyid" db:"KeyID"`
	Name          string        `json:"name" db:"Name"`
	Gender        int8          `json:"gender" db:"Gender"`
	AppType       int8          `json:"apptype" db:"AppType"`
	HaveLike      int           `json:"have_like" db:"HaveLike"`
	RemainLike    int           `json:"reamin_like" db:"RemainLike"`
	MyMansionItem []MansionItem `json:"mansion_item"`
}

type MansionItem struct {
	ItemIndex    int `json:"item_index" db:"ItemIndex"`
	ItemPosition int `json:"item_position" db:"ItemPosition"`
}

type MansionItemMoveRequest struct {
	ID              int           `json:"acc_uid"`
	MoveMansionItem []MansionItem `json:"mansion_item"`
}

type MansionItemMoveResponse struct {
	ID              int           `json:"acc_uid"`
	MoveMansionItem []MansionItem `json:"mansion_item"`
}

type MansionFirendResponse struct {
	MyMansionFirend []MansionFirend `json:"mansion_friend"`
}

type MansionFirend struct {
	ID       int    `json:"friend_uid" db:"AccountUID"`
	KeyID    string `json:"friend_keyid" db:"KeyID"`
	Name     string `json:"friend_name" db:"Name"`
	Gender   int8   `json:"friend_gender" db:"Gender"`
	AppType  int8   `json:"friend_apptype" db:"AppType"`
	HaveLike int    `json:"friend_like" db:"HaveLike"`
}

type MansionLikeResponse struct {
	RemainLike int `json:"remain_like"`
}

type MansionRankRequest struct {
	ID      int `json:"acc_uid"`
	RankMin int `json:"rank_min"`
	RankMax int `json:"rank_max"`
}

type MansionRank struct {
	ID       int    `json:"acc_uid" db:"AccountUID"`
	KeyID    string `json:"keyid" db:"KeyID"`
	Name     string `json:"name" db:"Name"`
	Gender   int8   `json:"gender" db:"Gender"`
	AppType  int8   `json:"apptype" db:"AppType"`
	Rank     int    `json:"rank" db:"Rank"`
	HaveLike int    `json:"have_like" db:"HaveLike"`
}

type MansionRankResponse struct {
	TotalUser       int           `json:"total_user"`
	MyRank          int           `json:"my_rank"`
	MansionRankInfo []MansionRank `json:"rank_list"`
}

type MansionRankRewardResponse struct {
	ID       int    `json:"acc_uid" db:"AccountUID"`
	KeyID    string `json:"keyid" db:"KeyID"`
	Name     string `json:"name" db:"Name"`
	Gender   int8   `json:"gender" db:"Gender"`
	AppType  int8   `json:"apptype" db:"AppType"`
	HaveLike int    `json:"have_like" db:"HaveLike"`
	Rank     int    `json:"rank" db:"Rank"`
	RankID   int    `json:"rank_id" db:"RankID"`
	//RewardGold      int    `json:"reward_gold" db:"RewardGold"`
	//RewardSpin      int    `json:"reward_spin" db:"RewardSpin"`
	//RewardRandomBox int    `json:"reward_box" db:"RewardRandomBox"`
}

type QuestInfoResponse struct {
	ID              int  `json:"acc_uid" db:"AccountUID"`
	GainGold        int  `json:"gain_gold" db:"GainGold"`
	UseGold         int  `json:"use_gold" db:"UseGold"`
	UseSpin         int  `json:"use_spin" db:"UseSpin"`
	TryRaid         int  `json:"try_raid" db:"TryRaid"`
	TryAttack       int  `json:"try_attack" db:"TryAttack"`
	TryEventGame    int  `json:"try_event" db:"TryEventGame"`
	CompleteChapter int8 `json:"complete_chapter" db:"CompleteChapter"`
	StructureAction int  `json:"structure_action" db:"StructureAction"`
	MoveFurniture   int  `json:"move_furniture" db:"MoveFurniture"`
	GainFurniture   int  `json:"gain_furniture" db:"GainFurniture"`
	CompleteQuest   int  `json:"complete_quest" db:"CompleteQuest"`
}

type QuestUpdateRequest struct {
	ID              int  `json:"acc_uid" db:"AccountUID"`
	GainGold        int  `json:"gain_gold" db:"GainGold"`
	UseGold         int  `json:"use_gold" db:"UseGold"`
	UseSpin         int  `json:"use_spin" db:"UseSpin"`
	TryRaid         int  `json:"try_raid" db:"TryRaid"`
	TryAttack       int  `json:"try_attack" db:"TryAttack"`
	TryEventGame    int  `json:"try_event" db:"TryEventGame"`
	CompleteChapter int8 `json:"complete_chapter" db:"CompleteChapter"`
	StructureAction int  `json:"structure_action" db:"StructureAction"`
	MoveFurniture   int  `json:"move_furniture" db:"MoveFurniture"`
	GainFurniture   int  `json:"gain_furniture" db:"GainFurniture"`
	CompleteQuest   int  `json:"complete_quest" db:"CompleteQuest"`
}

type QuestCompleteRequest struct {
	ID            int  `json:"acc_uid" db:"AccountUID"`
	CompleteQuest uint `json:"complete_quest" db:"CompleteQuest"`
}

type QuestCompleteResponse struct {
	ID                   int       `json:"acc_uid" db:"AccountUID"`
	RewardGold           int       `json:"reward_gold" db:"MissionRewardGold"`
	RewardSpin           int       `json:"reward_spin" db:"MissionRewardSpin"`
	RewardBox            int       `json:"reward_box" db:"MissionRewardRandomBox"`
	UserGold             int       `json:"user_gold" db:"Gold"`
	UserSpin             int       `json:"user_spin" db:"Spin"`
	MansionNumber        int8      `json:"mansion_num" db:"MansionNumber"`
	RoomNumber           int8      `json:"room_num" db:"OpenRoomNumber"`
	IsMansionOpen        int8      `json:"is_mansion_open" db:"NextRoomOpen"`
	LastSpinRechargeTime time.Time `json:"spin_recharge_time" db:"LastSpinRechargeTime"`
}

type GoldRoomOpenRequest struct {
	ID       int `json:"acc_uid" `
	OpenRoom int `json:"open_room" `
	Gold     int `json:"gold" `
}

type GoldRoomOpenResponse struct {
	ID         int `json:"acc_uid" db:"AccountUID"`
	OpenedRoom int `json:"opened_room" db:"MansionRoom"`
	UserGold   int `json:"user_gold" db:"Gold"`
}

type ShopGameRequest struct {
	ID     int `json:"acc_uid" `
	ItemID int `json:"item_id" `
	Gold   int `json:"gold" `
}

type ShopGameResponse struct {
	ID      int       `json:"acc_uid" db:"AccountUID"`
	ItemID  int       `json:"item_id" db:"ItemID"`
	Gold    int       `json:"user_gold" db:"Gold"`
	BuyTime time.Time `json:"buy_time" db:"BuyTime"`
}

type UserShopGameInfo struct {
	ItemID  int       `json:"item_id" db:"ItemID"`
	BuyTime time.Time `json:"buy_time" db:"BuyTime"`
}

type UserShopGameList struct {
	ShopGameList []UserShopGameInfo
}
