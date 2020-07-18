package table

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/datastore"
)

/*上傳roadID-roadNmae talbe*/
type ID2Name struct {
	RoadID   string
	RoadName string
}
type TTT struct {
	IDs []*ID2Name
}

//a
func Upload(ctx context.Context) error {

	var datas TTT
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
		var a ID2Name
		a.RoadID = dataSlice[1]
		a.RoadName = dataSlice[4]
		// log.Print(a.RoadID, a.RoadName)
		datas.IDs = append(datas.IDs, &a)
	}

	//以roadID產生entity key
	for _, ID := range datas.IDs {
		roadKey := datastore.NameKey("NTPCRoadName", ID.RoadID, nil)
		roadKeys = append(roadKeys, roadKey)

	}
	B(ctx, roadKeys, &datas)
	return nil
}

//put路段資訊
func B(ctx context.Context, roadKeys []*datastore.Key, datas interface{}) {
	// fmt.Print("@@@", datas)
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	if _, err := client.PutMulti(ctx, roadKeys[0:], datas.(*TTT).IDs[0:]); err != nil {
		log.Fatalf("PutMulti ID: %v", err)
	}
	fmt.Printf("table Info Saved sucess")

}
