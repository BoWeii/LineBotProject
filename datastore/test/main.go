package main

import (
	"context"

	// "project.com/datastore"
	// "project.com/datastore/table"
	"project.com/datastore/fee"
)

//main 測試function執行
func main() {
	// datastore.UpdateParkingInfo(context.Background(), datastore.PubSubMessage{Data: []byte("update")})
	// table.GpsToRoadName()
	// table.UploadRoads(context.Background())
	fee.UpdateFeeInfo(context.Background())

}

//部署指令
//gcloud functions deploy PutParkingInfo --source https://source.developers.google.com/projects/parkingproject-261207/repos/github_wei02427_linebotproject/moveable-aliases/master/paths/datastore --runtime=go113 --trigger-topic=updateInfo
