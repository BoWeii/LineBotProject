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
	RoadSegAvail    string   `json:"roadSegAvail"`    //路段剩餘格位數
	RoadSegFee      string   `json:"roadSegFee"`      //收費標準
	RoadSegID       string   `json:"roadSegID"`       //路段ID
	RoadSegName     string   `json:"roadSegName"`     //路段名稱
	RoadSegTmEnd    string   `json:"roadSegTmEnd"`    //收費結束時間
	RoadSegTmStart  string   `json:"roadSegTmStart"`  //收費開始時間
	RoadSegTotal    string   `json:"roadSegTotal"`    //路段總格位數
	RoadSegUpdateTm string   `json:"roadSegUpdateTm"` //資料更新時間
	RoadSegUsage    string   `json:"roadSegUsage"`    //路段使用率
	CellStatusList  CellList `json:"cellStatusList"`  //單一停車格資訊
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

	//fmt.Println(pjson.String()[9 : len(pjson.String())-2])
	var data Data
	keys := []*datastore.Key{}

	er := json.Unmarshal([]byte(pjson.String()[9:len(pjson.String())-2]), &data)
	if er != nil {
		fmt.Println("error:", er)
	}
	for index, element := range data.ROAD {
		fmt.Printf("第%d筆:%+v\n", index, element)
		keys = append(keys, datastore.IncompleteKey("Parkings", nil))
	}

	ctx := context.Background()

	// Set your Google Cloud Platform project ID.
	projectID := "parkingproject-261207"

	// Creates a client.
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Saves the new entity.
	if _, err := client.PutMulti(ctx, keys[:100], data.ROAD[:100]); err != nil {
		log.Fatalf("Failed to save task: %v", err)
	}

	fmt.Printf("Saved sucess")
}

//GetParkingInfo 取得停車格資訊
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
