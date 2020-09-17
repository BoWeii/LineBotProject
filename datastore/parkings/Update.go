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
const (
	parkingSpaces = iota
	parkingLots
	parkingLotsAvail
)

//PubSubMessage gcp pub/sub payload
type PubSubMessage struct {
	Data []byte `json:"data"`
}

//UpdateNTPCParkingLotInfo consumes a Pub/Sub message 並更新停車位資訊
func UpdateNTPCParkingLotInfo(ctx context.Context, m PubSubMessage) error {
	var NTPCJsonData parking.NTPC
	var NTPC parking.NTPC
	NTPCKeys := [3][]*datastore.Key{}
	// //新北市路邊停車格
	if NTPCParkingSpaceInfo, err := getNTPCParkingsInfo("https://data.ntpc.gov.tw/api/datasets/54A507C4-C038-41B5-BF60-BBECB9D052C6/json", 30, parkingSpaces); err != nil {
		log.Fatalf("error: %v", err)
	} else {

		//roadKeysNTPC := []*datastore.Key{}
		if err := json.Unmarshal([]byte(*NTPCParkingSpaceInfo), &NTPCJsonData.Spaces); err != nil {
			log.Fatalf("error: %v", err)
		} else {
			//以roadID產生entity key
			for _, cell := range NTPCJsonData.Spaces {
				parentKey := datastore.NameKey("NTPCRoadName", cell.RoadID, nil)
				roadKey := datastore.NameKey("NTPCParkingSpaces", strconv.Itoa(cell.ID), parentKey)
				NTPCKeys[parkingSpaces] = append(NTPCKeys[parkingSpaces], roadKey)
			}
			log.Println("Update NTPC parking spaces")
		}
		NTPC.Spaces = NTPCJsonData.Spaces
	}

	//新北市停車場
	if NTPCParkingLotInfo, err := getNTPCParkingsInfo("https://data.ntpc.gov.tw/api/datasets/B1464EF0-9C7C-4A6F-ABF7-6BDF32847E68/json", 1, parkingLots); err != nil {
		log.Fatalf("error: %v", err)
	} else {

		if err := json.Unmarshal([]byte(*NTPCParkingLotInfo), &NTPCJsonData.Lot); err != nil {
			log.Fatalf("error: %v", err)
		} else {
			//以lotID產生entity key
			for _, lot := range NTPCJsonData.Lot {
				//只存有汽車車位停車場
				if lot.TotalCar != 0 {
					lotKey := datastore.NameKey("NTPCParkingLots", strconv.Itoa(lot.ID), nil)
					NTPC.Lot = append(NTPC.Lot, lot)
					NTPCKeys[parkingLots] = append(NTPCKeys[parkingLots], lotKey)
				}

			}

			log.Println("Update NTPC parking lots")

		}
	}

	//putParkingInfo(ctx, NTPCKeys, &NTPC)
	//新北市停車場剩餘數量
	if NTPCParkingLotsAvailInfo, err := getNTPCParkingsInfo("https://data.ntpc.gov.tw/api/datasets/E09B35A5-A738-48CC-B0F5-570B67AD9C78/json", 1, parkingLots); err != nil {
		log.Fatalf("error: %v", err)
	} else {

		var NTPCParkingLotsAvailJSON []parking.ParkingLotAvailNTPC
		var NTPCParkingLotsAvail []*parking.ParkingLotAvailNTPC
		if err := json.Unmarshal([]byte(*NTPCParkingLotsAvailInfo), &NTPCParkingLotsAvailJSON); err != nil {
			log.Fatalf("error: %v", err)
		} else {

			for _, lot := range NTPCParkingLotsAvailJSON {
				if lot.AvailableCar != -9 {
					lotKey := datastore.NameKey("NTPCParkingLots", strconv.Itoa(lot.ID), nil)
					lotAvailKey := datastore.NameKey("NTPCParkingLotsAvail", strconv.Itoa(lot.ID), lotKey)
					NTPCParkingLotsAvail = append(NTPCParkingLotsAvail, &lot)
					NTPCKeys[parkingLotsAvail] = append(NTPCKeys[parkingLotsAvail], lotAvailKey)
				}
			}

			log.Println("Update NTPC parking lots Avail")
			//fmt.Println(NTPCParkingLotsAvail)
			putParkingInfo(ctx, NTPCKeys, &NTPCParkingLotsAvail)

		}
	}

	return nil
}

//put路段資訊
func putParkingInfo(ctx context.Context, keys [3][]*datastore.Key, parkings interface{}) {

	client, err := datastore.NewClient(ctx, projectID)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	switch parkings.(type) {
	case *parking.NTPC:
		for i := 0; i <= 1; i++ {

			var tmp = 0
			n := math.Ceil(float64(len(keys[i])) / 500) //一次最多put500筆

			for j := 1; j <= int(n); j++ {
				var size int
				if size = len(keys[i]); j*500 < len(keys[i]) {
					size = j * 500
				}

				if i == parkingSpaces {
					if _, err := client.PutMulti(ctx, keys[i][tmp:size-1], parkings.(*parking.NTPC).Spaces[tmp:size-1]); err != nil {
						log.Fatalf("PutMulti NTPCParkingSpaces: %v", err)
					}
				} else {

					if _, err := client.PutMulti(ctx, keys[i][tmp:size-1], parkings.(*parking.NTPC).Lot[tmp:size-1]); err != nil {
						log.Fatalf("PutMulti NTPCParkingLot: %v", err)
					}
				}
				tmp = size - 1
			}
		}
		log.Println("Parking Info Saved sucess")
	case *[]parking.ParkingLotAvailNTPC:
		var tmp = 0
		n := math.Ceil(float64(len(keys[parkingLotsAvail])) / 500) //一次最多put500筆

		for j := 1; j <= int(n); j++ {
			var size int
			if size = len(keys[parkingLotsAvail]); j*500 < len(keys[parkingLotsAvail]) {
				size = j * 500
			}
			avails := parkings.(*[]parking.ParkingLotAvailNTPC)
			if _, err := client.PutMulti(ctx, keys[parkingLotsAvail][tmp:size-1], (*avails)[tmp:size-1]); err != nil {
				log.Fatalf("PutMulti NTPCParkingLot Avail: %v", err)
			}
			tmp = size - 1
		}
	}

	log.Println("Parking Lots Avail Info Saved sucess")

}

//getNTPCParkingsInfo
func getNTPCParkingsInfo(url string, page int, parkingsType int) (*string, error) {
	var data string
	data = "["
	for i := 0; i <= page; i++ {

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
		if parkingsType == parkingSpaces {
			temp = strings.ReplaceAll(temp, "\"CELLSTATUS\""+":"+"\"Y\"", "\"CELLSTATUS\""+":"+"\"true\"")
			temp = strings.ReplaceAll(temp, "\"CELLSTATUS\""+":"+"\"N\"", "\"CELLSTATUS\""+":"+"\"false\"")
		}
		data = data + temp + ","
	}
	data = data[:len(data)-1]
	data += "]"
	return &data, nil
}
