package table

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"cloud.google.com/go/datastore"

)

/*上傳roadID-roadNmae talbe*/
type roadInfo struct {
	RoadID   string
	RoadName string
}
type roads struct {
	IDs []*roadInfo
}

//a
func UploadRoads(ctx context.Context) error {

	var datas roads
	roadKeys := []*datastore.Key{}

	file, err := os.Open("data.txt")

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	file.Close()

	for _, eachline := range txtlines {
		dataSlice := strings.Split(eachline, " ")
		var a roadInfo
		a.RoadID = dataSlice[1]
		a.RoadName = dataSlice[4]
		log.Print(a.RoadID, a.RoadName)
		datas.IDs = append(datas.IDs, &a)
	}

	//以roadID產生entity key
	for _, ID := range datas.IDs {
		roadKey := datastore.NameKey("NTPCRoadName", ID.RoadID, nil)
		roadKeys = append(roadKeys, roadKey)

	}
	putRoadName(ctx, roadKeys, &datas)
	return nil
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

		if _, err := client.PutMulti(ctx, roadKeys[tmp:size-1], datas.IDs[tmp:size-1]); err != nil {
			log.Fatalf("PutMulti ID: %v", err)
		}
		tmp = size - 1
	}
	fmt.Printf("table Info Saved sucess")

}
