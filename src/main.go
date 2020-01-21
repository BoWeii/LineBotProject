package main

import (
	"context"

	"example.com/test/datastore"

)

func main() {
	datastore.PutParkingInfo(context.Background(), datastore.PubSubMessage{Data: []byte("update")})
}
