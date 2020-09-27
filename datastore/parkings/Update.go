package parkings

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	parking "project.com/datastore/parkingstruct"

)

const googleMapAPIKey string = "AIzaSyCzGP7dIwrOEuWxN8w40tBvwA_rvnbqudE"
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

//UpdateNTPCParkingLotsInfo consumes a Pub/Sub message 並更新停車位資訊
func UpdateNTPCParkingLotsInfo(ctx context.Context, m PubSubMessage) error {
	var parkingLotsJSONData []*parking.ParkingLotNTPC
	var parkingLots []*parking.ParkingLotNTPC
	parkingLotsKeys := []*datastore.Key{}

	//新北市停車場
	if NTPCParkingLotInfo, err := GetNTPCParkingsInfo("https://data.ntpc.gov.tw/api/datasets/B1464EF0-9C7C-4A6F-ABF7-6BDF32847E68/json"); err != nil {
		log.Fatalf("error: %v", err)
	} else {
		if err := json.Unmarshal([]byte(*NTPCParkingLotInfo), &parkingLotsJSONData); err != nil {
			log.Fatalf("error: %v", err)
		} else {

			isExist := make(map[string]bool)
			for _, lot := range parkingLotsJSONData {
				if lot.TotalCar != 0 {
					if _, ok := isExist[lot.ID]; ok {
						continue //排除重複公開資料
					} else {
						isExist[lot.ID] = true
						lotKey := datastore.NameKey("NTPCParkingLots", lot.ID, nil)
						lot.Lat, lot.Lon = twd97ToWgs84(lot.Lon, lot.Lat)
						parkingLots = append(parkingLots, lot)
						parkingLotsKeys = append(parkingLotsKeys, lotKey)
						time.Sleep(500 * time.Millisecond)
					}

				}

			}

			log.Println("Update NTPC parking lots")
			//putParkingInfo(context.Background(), parkingLotsKeys, parkingLots)
		}
	}

	return nil
}

//put路段資訊
func putParkingInfo(ctx context.Context, keys []*datastore.Key, parkings interface{}) {

	client, err := datastore.NewClient(ctx, projectID)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	switch parkings.(type) {
	case []*parking.ParkingLotNTPC:
		for i := 0; i <= 1; i++ {

			var tmp = 0
			n := math.Ceil(float64(len(keys)) / 500) //一次最多put500筆

			for j := 1; j <= int(n); j++ {
				var size int
				if size = len(keys); j*500 < len(keys) {
					size = j * 500
				}

				if _, err := client.PutMulti(ctx, keys[tmp:size-1], parkings.([]*parking.ParkingLotNTPC)[tmp:size-1]); err != nil {
					log.Fatalf("PutMulti NTPCParkingLot: %v", err)
				}

				tmp = size - 1
			}
		}
		log.Println("Parking Lots Info Saved sucess")

	}

}

//GetNTPCParkingsInfo 取得公開資料
func GetNTPCParkingsInfo(url string) (*string, error) {
	var data string
	data = "["
	i := 0
	for {

		resp, err := http.Get(url + "?page=" + strconv.Itoa(i) + "&size=1000")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(body)
		bodyString = strings.Replace(bodyString, "[", "", 1)
		bodyString = strings.Replace(bodyString, "]", "", 1)

		if bodyString == "" {
			break
		} else {

			if i != 0 {
				bodyString = "," + bodyString
			}
			data = data + bodyString
			i++
		}

	}

	data += "]"
	return &data, nil
}

//getGPS twd97轉gps
func twd97ToWgs84(tx float64, ty float64) (lat float64, lon float64) {

	deocoding := "http://taibif.tw/BDTools/proj4/convert.php?source=5&destination=1&x=" + fmt.Sprintf("%f", tx) + "&y=" + fmt.Sprintf("%f", ty)

	resp, _ := http.Get(deocoding)
	body, _ := ioutil.ReadAll(resp.Body)
	//log.Println(tx, ty)
	html := string(body)
	//log.Println(deocoding)
	conversion := strings.Split(html, "<br>")[1]
	res := strings.Split(conversion, " ")

	lon, _ = strconv.ParseFloat(res[2], 64)
	lat, _ = strconv.ParseFloat(res[3], 64)

	return
}
