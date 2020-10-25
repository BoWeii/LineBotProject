package main

import (
	"context"

	"project.com/datastore"
	"project.com/datastore/parkings"

)

//main 測試function執行
func main() {

	datastore.UpdateParkingInfo(context.Background(), parkings.PubSubMessage{Data: []byte("update")})

}

//部署指令
//gcloud functions deploy PutParkingInfo --source https://source.developers.google.com/projects/parkingproject-261207/repos/github_wei02427_linebotproject/moveable-aliases/master/paths/datastore --runtime=go113 --trigger-topic=updateInfo
