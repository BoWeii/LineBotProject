package parkings

import (
	// "compress/gzip"
	// "bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	// "strings"
	// "reflect"

	"cloud.google.com/go/datastore"
	// xml2json "github.com/basgys/goxml2json"
	parking "project.com/datastore/parkingstruct"

)

const projectID string = "exalted-yeti-289303"

//PubSubMessage gcp pub/sub payload
type PubSubMessage struct {
	Data []byte `json:"data"`
}

//UpdateNTPCParkingLotInfo consumes a Pub/Sub message 並更新停車位資訊
func UpdateNTPCParkingLotInfo(ctx context.Context, m PubSubMessage) error {
	var NTPC parking.NTPC
	NTPCKeys := [2][]*datastore.Key{}
	// //新北市路邊停車格
	// if NTPCParkingSpaceInfo, err := getNTPCParkingSpaceInfo("https://data.ntpc.gov.tw/api/datasets/54A507C4-C038-41B5-BF60-BBECB9D052C6/json"); err != nil {
	// 	log.Fatalf("error: %v", err)
	// } else {

	// 	//roadKeysNTPC := []*datastore.Key{}
	// 	if err := json.Unmarshal([]byte(*NTPCParkingSpaceInfo), &NTPC.Spaces); err != nil {
	// 		log.Fatalf("error: %v", err)
	// 	} else {
	// 		//以roadID產生entity key
	// 		for _, cell := range NTPC.Spaces {
	// 			parentKey := datastore.NameKey("NTPCRoadName", cell.RoadID, nil)
	// 			roadKey := datastore.NameKey("NTPCParkings", strconv.Itoa(cell.ID), parentKey)
	// 			NTPCKeys[0] = append(NTPCKeys[0], roadKey)
	// 		}
	// 		log.Println("Update NTPC parking spaces")
	// 		putParkingInfo(ctx, roadKeysNTPC, &NTPC)
	// 	}
	// }
	//新北市停車場
	if NTPCParkingLotInfo, err := getNTPCParkingLotInfo("https://data.ntpc.gov.tw/api/datasets/B1464EF0-9C7C-4A6F-ABF7-6BDF32847E68/json"); err != nil {
		log.Fatalf("error: %v", err)
	} else {

		//lotKeysNTPC := []*datastore.Key{}

		if err := json.Unmarshal([]byte(*NTPCParkingLotInfo), &NTPC.Lot); err != nil {
			log.Fatalf("error: %v", err)
		} else {
			//以lotID產生entity key
			for _, lot := range NTPC.Lot {
				lotKey := datastore.NameKey("NTPCParkingLots", strconv.Itoa(lot.ID), nil)
				NTPCKeys[1] = append(NTPCKeys[1], lotKey)
			}
			log.Println("Update NTPC parking lots")
			putParkingInfo(ctx, NTPCKeys, &NTPC)
		}
	}

	return nil
}

//put路段資訊
func putParkingInfo(ctx context.Context, keys [2][]*datastore.Key, parkings interface{}) {

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	//fmt.Print(parkings.(*parking.NTPC))

	var tmp = 0
	int maxLen

	n := math.Ceil(float64(len(keys[1])) / 500) //一次最多put500筆

	for i := 1; i <= int(n); i++ {
		var size int
		if size = len(keys[1]); i*500 < len(keys[1]) {
			size = i * 500
		}

		switch parkings.(type) {
		case *parking.NTPC:
			if _, err := client.PutMulti(ctx, keys[1][tmp:size-1], parkings.(*parking.NTPC).Lot[tmp:size-1]); err != nil {
				log.Fatalf("PutMulti NTPCParkingLot: %v", err)
			}
		}

		tmp = size - 1
	}

	log.Println("Parkings Info Saved sucess")

}

//getNTPCParkingSpaceInfo
func getNTPCParkingSpaceInfo(url string) (*string, error) {
	var data string
	data = "["
	for i := 0; i <= 30; i++ {

		resp, err := http.Get(url + "?page=" + strconv.Itoa(i) + "&size=1000")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		temp := string(body)
		temp = strings.Replace(temp, "[", "", -1)
		temp = strings.Replace(temp, "]", "", -1)
		temp = strings.ReplaceAll(temp, "\"CELLSTATUS\""+":"+"\"Y\"", "\"CELLSTATUS\""+":"+"\"true\"")
		temp = strings.ReplaceAll(temp, "\"CELLSTATUS\""+":"+"\"N\"", "\"CELLSTATUS\""+":"+"\"false\"")
		data = data + temp + ","
	}
	data = data[:len(data)-1]
	data += "]"
	return &data, nil
}

//getNTPCParkingSpaceInfo
func getNTPCParkingLotInfo(url string) (*string, error) {
	var data string
	data = "["
	for i := 0; i <= 1; i++ {

		resp, err := http.Get(url + "?page=" + strconv.Itoa(i) + "&size=1000")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		temp := string(body)
		temp = strings.Replace(temp, "[", "", -1)
		temp = strings.Replace(temp, "]", "", -1)

		data = data + temp + ","
	}
	data = data[:len(data)-1]
	data += "]"
	return &data, nil
}
