package main

import (
	"action"
	"common"
	"httpPkg"

	"log"
	"net/http"
	_ "time"

	"github.com/julienschmidt/httprouter"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"

	// xsql mysql driver
	"github.com/jmoiron/sqlx"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	//AccountDB := "root:coin123!@#@tcp(jb.ctkb2mflzwwh.ap-northeast-2.rds.amazonaws.com:3306)/Coin?parseTime=true"
	GameDB := "root:coin123!@#@tcp(jb.ctkb2mflzwwh.ap-northeast-2.rds.amazonaws.com:3306)/Coin?parseTime=true"
	//var DataDB string = "root:coin123!@#@tcp(jb.ctkb2mflzwwh.ap-northeast-2.rds.amazonaws.com:3306)/DataDB"

	// mysql connect
	// accountdb := sqlx.MustConnect("mysql", AccountDB)
	// defer accountdb.Close()
	gamedb := sqlx.MustConnect("mysql", GameDB)
	defer gamedb.Close()

	router := httprouter.New()

	httpPkg.AddResource(router, new(action.CrtAccResource), gamedb)
	httpPkg.AddResource(router, new(action.SlotResultResource), gamedb)
	httpPkg.AddResource(router, new(action.LoginResource), gamedb)
	httpPkg.AddResource(router, new(action.LinkAccountResource), gamedb)
	httpPkg.AddResource(router, new(action.ChapterResource), gamedb)
	httpPkg.AddResource(router, new(action.BuyTileResource), gamedb)
	httpPkg.AddResource(router, new(action.BuildingResource), gamedb)
	httpPkg.AddResource(router, new(action.RefairBuildingResource), gamedb)
	httpPkg.AddResource(router, new(action.BuildingRewardResource), gamedb)
	httpPkg.AddResource(router, new(action.TutorialResource), gamedb)
	httpPkg.AddResource(router, new(action.RefreshResource), gamedb)
	httpPkg.AddResource(router, new(action.InviteFriendResource), gamedb)

	// ydk test
	httpPkg.AddResource(router, new(action.AttackResource), gamedb)
	httpPkg.AddResource(router, new(action.RaidResource), gamedb)
	httpPkg.AddResource(router, new(action.NewsResource), gamedb)
	httpPkg.AddResource(router, new(action.AdvRewardResource), gamedb)
	httpPkg.AddResource(router, new(action.TargetResource), gamedb)
	httpPkg.AddResource(router, new(action.FriendResource), gamedb)
	httpPkg.AddResource(router, new(action.ShopGResource), gamedb)
	httpPkg.AddResource(router, new(action.ShopAResource), gamedb)
	httpPkg.AddResource(router, new(action.ShopGameResource), gamedb)
	httpPkg.AddResource(router, new(action.NoticeInfoResource), gamedb)
	httpPkg.AddResource(router, new(action.EventResultResource), gamedb)
	httpPkg.AddResource(router, new(action.MansionInfoResource), gamedb)
	httpPkg.AddResource(router, new(action.MansionItemMoveResource), gamedb)
	httpPkg.AddResource(router, new(action.MansionFriendResource), gamedb)
	httpPkg.AddResource(router, new(action.MansionRankResource), gamedb)
	httpPkg.AddResource(router, new(action.MansionRankRewardResource), gamedb)
	httpPkg.AddResource(router, new(action.MansionLikeResource), gamedb)
	httpPkg.AddResource(router, new(action.QuestInfoResource), gamedb)
	httpPkg.AddResource(router, new(action.QuestUpdateResource), gamedb)
	httpPkg.AddResource(router, new(action.QuestCompleteResource), gamedb)
	httpPkg.AddResource(router, new(action.GoldRoomOpenResource), gamedb)

	//httpPkg.AddResource(router, new(action.TestResource), gamedb)

	//log.Fatal(http.ListenAndServeTLS(":8081", "/home/ec2-user/cert.pem", "/home/ec2-user/key.pem", router))
	//log.Fatal(http.ListenAndServe(":8080", router))

	common.Init()

	errs := make(chan error)

	// Starting HTTP server
	go func() {
		log.Println("Staring HTTP service ...")
		if err := http.ListenAndServe(":8080", router); err != nil {
			errs <- err
		}
	}()

	// Starting HTTPS server
	/*
		go func() {
			log.Printf("Staring HTTPS service ...")
			// C:\\Users\\ydk831\\Desktop\\Coin\\project_coin_server\\
			// /home/ec2-user/cert/

			//if err := http.ListenAndServeTLS(":8081", "/home/ec2-user/cert/private.crt", "/home/ec2-user/cert/private.key", nil); err != nil {
			if err := http.ListenAndServeTLS(":8081", "C:\\Users\\ydk831\\Desktop\\cert\\private.crt", "C:\\Users\\ydk831\\Desktop\\cert\\private.key", nil); err != nil {
				errs <- err
			}
		}()
	*/

	select {
	case err := <-errs:
		log.Printf("Could not start serving service due to (error: %s)", err)
	}
	/*
	   for {
	   				time.Sleep(1000)
	   }
	*/

}
