package datastore

import (
	// "compress/gzip"
	// "bytes"
	"context"
	"encoding/json"
	"fmt"
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

const projectID string = "parkingproject-2-283415"

//PubSubMessage gcp pub/sub payload
type PubSubMessage struct {
	Data []byte `json:"data"`
}

//UpdateParkingInfo consumes a Pub/Sub message 並更新停車位資訊
func UpdateParkingInfo(ctx context.Context, m PubSubMessage) error {
	log.Println(string(m.Data))
	//取open data

	//fileURL := "https://tcgbusfs.blob.core.windows.net/blobtcmsv/TCMSV_roadquery.gz"
	//TPEParkingInfo, err := getParkingInfo(fileURL)
	NTPCParkingInfo, err := getParkingInfo("https://data.ntpc.gov.tw/api/datasets/54A507C4-C038-41B5-BF60-BBECB9D052C6/json")
	//fmt.Printf(*TPEParkingInfo)
	if err != nil {
		log.Print(err)
	}

	var NTPC parking.NTPC
	roadKeys := []*datastore.Key{}

	//json轉struct
	if err := json.Unmarshal([]byte(*NTPCParkingInfo), &NTPC.Cells); err != nil {
		log.Fatalf("error: %v", err)
	}

	//以roadID產生entity key
	for _, cell := range NTPC.Cells {

		parentKey := datastore.NameKey("NTPCRoadName", cell.RoadID, nil)
		roadKey := datastore.NameKey("NTPCParkings", strconv.Itoa(cell.ID), parentKey)
		roadKeys = append(roadKeys, roadKey)

	}
	//log.Print(&NTPC)
	putParkingInfo(ctx, roadKeys, &NTPC)

	return nil
}

//put路段資訊
func putParkingInfo(ctx context.Context, roadKeys []*datastore.Key, parkings interface{}) {

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	//fmt.Print(parkings.(*parking.NTPC))

	var tmp = 0
	n := math.Ceil(float64(len(roadKeys)) / 500) //一次最多put500筆

	for i := 1; i <= int(n); i++ {
		var size int
		if size = len(roadKeys); i*500 < len(roadKeys) {
			size = i * 500
		}

		switch parkings.(type) {
		case *parking.TPE:
			if _, err := client.PutMulti(ctx, roadKeys[tmp:size-1], parkings.(*parking.TPE).Roads[tmp:size-1]); err != nil {
				log.Fatalf("PutMulti TPE: %v", err)
			}
		case *parking.NTPC:
			//fmt.Print("----------", i, parkings.(*parking.NTPC).Cells[tmp:size-1])
			if _, err := client.PutMulti(ctx, roadKeys[tmp:size-1], parkings.(*parking.NTPC).Cells[tmp:size-1]); err != nil {
				log.Fatalf("PutMulti NTPC: %v", err)
			}
		}

		tmp = size - 1
	}

	fmt.Printf("Info Saved sucess")

}

//GetParkingInfo 取得停車格資訊(TPE)
func getParkingInfo(url string) (*string, error) {
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
	// data = strings.ReplaceAll(data, "\"CELLSTATUS\""+":"+"\"Y\"", "\"CELLSTATUS\""+":"+"\"true\"")
	// data = strings.ReplaceAll(data, "\"CELLSTATUS\""+":"+"\"N\"", "\"CELLSTATUS\""+":"+"\"false\"")

	// fmt.Printf(data)

	// var temp string
	// temp=url+"?page="+"0"+"&size=1000"
	// // fmt.Printf("################# %s",temp)

	// resp, err := http.Get(temp)
	// if err != nil {
	// 		return nil, err
	// }

	// var data string ?page=1&size=30900
	// if strings.Contains(url, ".gz") {
	// 	reader, err := gzip.NewReader(resp.Body)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	defer reader.Close()

	// 	body, err := ioutil.ReadAll(reader)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	var xml = strings.NewReader(string(body))
	// 	json, err := xml2json.Convert(xml)
	// 	if err != nil {
	// 		log.Print("Failed to convert xml to json")
	// 	}

	// 	data = json.String()
	// } else {
	// 	body, err := ioutil.ReadAll(resp.Body)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	data = string(body)
	// 	data = strings.ReplaceAll(data, "\"CELLSTATUS\""+":"+"\"Y\"", "\"CELLSTATUS\""+":"+"\"true\"")
	// 	data = strings.ReplaceAll(data, "\"CELLSTATUS\""+":"+"\"N\"", "\"CELLSTATUS\""+":"+"\"false\"")
	// 	fmt.Print()
	// }

	// defer resp.Body.Close()
	// var tt string
	// tt = "123"
	return &data, nil
}

//gcloud functions deploy PutParkingInfo --source https://source.developers.google.com/projects/parkingproject-261207/repos/github_wei02427_linebotproject/moveable-aliases/master/paths/datastore --runtime=go113 --trigger-topic=updateInfo

/*單一車格資訊，因缺少座標故先不用*/

// var cells CellList
// if err := json.Unmarshal(road.CellStatusList, &cells); err == nil {
// 	cellKeys := []*datastore.Key{}
// 	if len(cells.Cells) != 0 {
// 		for i := 0; i < len(cells.Cells); i++ {
// 			cellKey := datastore.IncompleteKey("Cells", roadKey)
// 			cellKeys = append(cellKeys, cellKey)
// 		}

// 		if _, err := client.PutMulti(ctx, cellKeys, cells.Cells); err != nil {
// 			log.Fatalf("PutMulti: %v", err)
// 		} else {
// 			fmt.Printf("%s cells Saved sucess", road.RoadSegName)
// 		}
// 	}
// }
