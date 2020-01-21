/*$env:GOOGLE_APPLICATION_CREDENTIALS=
"C:\Users\wei14\Desktop\ParkingProject\[ParkingProject-9785d4e7adb0].json"*/

package datastore

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"

	"cloud.google.com/go/datastore"
	xml2json "github.com/basgys/goxml2json"

)

//PubSubMessage gcp pub/sub payload
type PubSubMessage struct {
	Data []byte `json:"data"`
}

//Cell 單一停車狀態
type Cell struct {
	CellStatus string `json:"cellStatus"` //停車格位狀態(1：車格有車輛停放；2：車格無車輛停放；3：無訊息)
	CoordX     string `json:"coord_X"`    //停車格位X座標
	CoordY     string `json:"coord_Y"`    //停車格位Y座標
	DataDt     string `json:"data_Dt"`    //?
	PsID       string `json:"psId"`       //停車格位格號
}

//CellList 停車格清單
type CellList struct {
	Cells []*Cell `json:"cell"`
}

//Road 路段停車格
type Road struct {
	RoadSegAvail    string          `json:"roadSegAvail"`                 //路段剩餘格位數
	RoadSegFee      string          `json:"roadSegFee"`                   //收費標準
	RoadSegID       string          `json:"roadSegID"`                    //路段ID
	RoadSegName     string          `json:"roadSegName"`                  //路段名稱
	RoadSegTmEnd    string          `json:"roadSegTmEnd"`                 //收費結束時間
	RoadSegTmStart  string          `json:"roadSegTmStart"`               //收費開始時間
	RoadSegTotal    string          `json:"roadSegTotal"`                 //路段總格位數
	RoadSegUpdateTm string          `json:"roadSegUpdateTm"`              //資料更新時間
	RoadSegUsage    string          `json:"roadSegUsage"`                 //路段使用率
	CellStatusList  json.RawMessage `json:"cellStatusList" datastore:"-"` //單一停車格資訊
}

//Data xml最外層
type Data struct {
	ROAD []*Road `json:"ROAD"`
}

//PutParkingInfo consumes a Pub/Sub message 並更新停車位資訊
func PutParkingInfo(ctx context.Context, m PubSubMessage) error {
	log.Println(string(m.Data))
	//取open data
	fileURL := "https://tcgbusfs.blob.core.windows.net/blobtcmsv/TCMSV_roadquery.gz"
	TPEParkingInfo, err := GetParkingInfo(fileURL)
	if err != nil {
		panic(err)
	}

	//xml轉json
	var xml = strings.NewReader(TPEParkingInfo)
	pjson, err := xml2json.Convert(xml)
	if err != nil {
		panic("Failed to convert xml to json")
	}
	
	//連結datastore
	//ctx := context.Background()

	projectID := "parkingproject-261207"
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	var data Data
	roadKeys := []*datastore.Key{}

	//json轉struct
	if err := json.Unmarshal([]byte(pjson.String()[9:len(pjson.String())-2]), &data); err != nil {
		log.Fatalf("error: %v", err)
	}

	//以roadID產生entity key
	for _, road := range data.ROAD {

		roadKey := datastore.NameKey("Parkings", road.RoadSegID, nil)
		roadKeys = append(roadKeys, roadKey)

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

	}

	//put路段資訊
	var tmp = 0
	n := math.Ceil(float64(len(roadKeys)) / 500) //一次最多put500筆
	for i := 1; i <= int(n); i++ {
		var size int
		if size = len(roadKeys); i*500 < len(roadKeys) {
			size = i * 500
		}
		if _, err := client.PutMulti(ctx, roadKeys[tmp:size-1], data.ROAD[tmp:size-1]); err != nil {
			log.Fatalf("PutMulti: %v", err)
		}
		tmp = size - 1
	}
	fmt.Printf("Roads Saved sucess")
	return nil
}

//GetParkingInfo 取得停車格資訊(TPE)
func GetParkingInfo(url string) (string, error) {

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	
	return string(body), nil
}

//gcloud functions deploy putParkingInfo --source https://source.developers.google.com/projects/parkingproject-261207/repos/github_wei02427_linebotproject/moveable-aliases/master/paths/datastore --runtime=go113 --trigger-topic=updateInfo