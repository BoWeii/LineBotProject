package main

import (
	"context"
	"project.com/datastore"

)

//main 測試function執行
func main() {
	datastore.PutParkingInfo(context.Background(), datastore.PubSubMessage{Data: []byte("update")})
}
