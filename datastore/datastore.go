/*$env:GOOGLE_APPLICATION_CREDENTIALS=
"C:\Users\wei14\Desktop\ParkingProject\[ParkingProject-9785d4e7adb0].json"*/

package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/datastore"

	xml2json "github.com/basgys/goxml2json"
)

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

func main() {
	fileURL := "https://tcgbusfs.blob.core.windows.net/blobtcmsv/TCMSV_roadquery.gz"
	TPEParkingInfo, err := GetParkingInfo(fileURL)
	if err != nil {
		panic(err)
	}
	//	fmt.Printf(TPEParkingInfo)

	var xml = strings.NewReader(TPEParkingInfo)
	pjson, err := xml2json.Convert(xml)
	if err != nil {
		panic("That's embarrassing...")
	}

	ctx := context.Background()

	// Set your Google Cloud Platform project ID.
	projectID := "parkingproject-261207"

	// Creates a client.
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	//fmt.Println(pjson.String()[9 : len(pjson.String())-2])
	var data Data
	// var cells CellList
	roadKeys := []*datastore.Key{}
	// cellKeys := []*datastore.Key{}

	if err := json.Unmarshal([]byte(pjson.String()[9:len(pjson.String())-2]), &data); err != nil {
		fmt.Println("error:", err)
	}
	for index, road := range data.ROAD {
		fmt.Printf("%s\n", road.RoadSegName)
		roadKey := datastore.NameKey("Parkings", road.RoadSegName, nil)
		roadKeys = append(roadKeys, roadKey)
		var cells CellList
		er := json.Unmarshal(road.CellStatusList, &cells)
		if er != nil {
			cellKeys := []*datastore.Key{}
			if len(cells.Cells) != 0 {
				for _, cell := range cells.Cells {
					cellKey := datastore.IncompleteKey("Cells", roadKey)
					cellKeys = append(cellKeys, cellKey)
					fmt.Printf("第%s筆cells:%+v\n", road.RoadSegName, cell)
				}
				if _, err := client.PutMulti(ctx, cellKeys, cells.Cells); err != nil {
					log.Fatalf("Failed to save cell: %v", err)
				}
			}
		}

		if index == 4 {
			break
		}
	}

	// Saves the new entity.
	if _, err := client.PutMulti(ctx, roadKeys[:5], data.ROAD[:5]); err != nil {
		log.Fatalf("Failed to save road: %v", err)
	}

	fmt.Printf("Saved sucess")
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
