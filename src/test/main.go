package main

import (
	"context"
	linebotProject "project/datastore"

)

//main 測試function執行
func main() {
	linebotProject.PutParkingInfo(context.Background(), linebotProject.PubSubMessage{Data: []byte("update")})
}
