package roadname

import (
	"context"
	"log"
	"math"

	"cloud.google.com/go/datastore"

)

/*上傳roadID-roadNmae talbe*/
type roadInfo struct {
	RoadID   string
	RoadName string
}
type roads struct {
	info []*roadInfo
}

//Update 更新路名
func Update(ctx context.Context) {

	log.Printf("Update RoadName Info")

	roadKeys := []*datastore.Key{}

	roadData := gpsToRoadName()

	//以roadID產生entity key
	for _, info := range roadData.info {
		roadKey := datastore.NameKey("NTPCRoadName", info.RoadID, nil)
		roadKeys = append(roadKeys, roadKey)

	}
	putRoadName(ctx, roadKeys, &roadData)

}

//put路段資訊
func putRoadName(ctx context.Context, roadKeys []*datastore.Key, datas *roads) {

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	n := math.Ceil(float64(len(roadKeys)) / 500) //一次最多put500筆
	tmp := 0
	for i := 1; i <= int(n); i++ {
		var size int
		if size = len(roadKeys); i*500 < len(roadKeys) {
			size = i * 500
		}

		if _, err := client.PutMulti(ctx, roadKeys[tmp:size-1], datas.info[tmp:size-1]); err != nil {
			log.Fatalf("PutMulti ID: %v", err)
		}
		tmp = size - 1
	}
	log.Printf("RoadName Info Saved sucess")

}
