package fulfillment

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/tidwall/gjson"
	"google.golang.org/api/iterator"

)

//road 路段停車格
type road struct {
	RoadSegAvail    string `json:"roadSegAvail"`    //路段剩餘格位數
	RoadSegFee      string `json:"roadSegFee"`      //收費標準
	RoadSegID       string `json:"roadSegID"`       //路段ID
	RoadSegName     string `json:"roadSegName"`     //路段名稱
	RoadSegTmEnd    string `json:"roadSegTmEnd"`    //收費結束時間
	RoadSegTmStart  string `json:"roadSegTmStart"`  //收費開始時間
	RoadSegTotal    string `json:"roadSegTotal"`    //路段總格位數
	RoadSegUpdateTm string `json:"roadSegUpdateTm"` //資料更新時間
	RoadSegUsage    string `json:"roadSegUsage"`    //路段使用率
}

//response webhook回應
type response struct {
	FulfillmentText string `json:"fulfillmentText"`
}

//Fulfillment 查詢車位
func Fulfillment(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	projectID := "parkingproject-261207"
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	//取得 roadName entity
	buf, _ := ioutil.ReadAll(r.Body)
	roadName := gjson.Get(string(buf), "queryResult.parameters.RoadName")
	log.Printf("roadName: %s", roadName)

	//datastore 查詢路段資料
	query := datastore.NewQuery("Parkings").Filter("RoadSegName=", roadName.String())
	it := client.Run(ctx, query)
	for {
		var road road
		_, err := it.Next(&road)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error fetching road: %v", err)
		}
		fmt.Printf("RoadName %s, RoadSegAvail %s\n", road.RoadSegName, road.RoadSegAvail)

		//回應訊息
		w.Header().Set("Content-Type", "application/json")
		response := response{
			FulfillmentText: road.RoadSegName + "有 " + road.RoadSegAvail + " 個車位",
		}
		json.NewEncoder(w).Encode(response)

	}

	defer r.Body.Close()

}
