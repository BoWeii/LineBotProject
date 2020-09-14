package table

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/thedevsaddam/gojsonq"
	"google.golang.org/api/iterator"

)

type Parking struct {
	ID            int     //車格序號
	CELLID        float64 //車格編號
	Name          string  //車格類型
	Day           string  //收費天
	Hour          string  //收費時段
	Pay           string  //收費形式
	PayCash       string  //費率
	Memo          string  //車格備註
	RoadID        string  //路段代碼
	CellStatus    bool    //車格狀態判斷 Y有車 N空位
	IsNowCash     bool    //收費時段判斷
	ParkingStatus int     //車格狀態 　1：有車、2：空位、3：非收費時段、4：時段性禁停、5：施工（民眾申請施工租用車格時使用）
	Lat           float64 //緯度
	Lon           float64 //經度
	Distance      float64 //距離
	Avail         int
	RoadName      string
}

var DatastoreProc datastoreProcessor

const projectID string = "exalted-yeti-289303"
const googleMapAPIKey string = "AIzaSyCzGP7dIwrOEuWxN8w40tBvwA_rvnbqudE"


type datastoreProcessor struct {
	projectID string
	client    *datastore.Client
	ctx       context.Context
}

func (ds *datastoreProcessor) Init(data string) (err error) {
	ds.projectID = data
	ds.ctx = context.Background()
	ds.client, err = datastore.NewClient(ds.ctx, ds.projectID)
	return
}

/*GpsToRoadName 查詢各路段 ID*/
func GpsToRoadName() {
	var roads []string

	// for _, i := range id {
	DatastoreProc.Init(projectID)
	query := datastore.NewQuery("NTPCParkings") //.Filter("CellStatus =", false)
	// existStr=append(existStr,"init")
	var tmpRoadID string
	it := DatastoreProc.client.Run(DatastoreProc.ctx, query)
	for {

		var road Parking
		_, err := it.Next(&road)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error fetching road: %v", err)
		}

		if road.RoadID != tmpRoadID {

			/*geocoding gps 轉路名*/

			geo := "https://maps.googleapis.com/maps/api/geocode/json?latlng=" + fmt.Sprintf("%f", road.Lat) + "," + fmt.Sprintf("%f", road.Lon) + "&result_type=route&language=zh-tw&key="+googleMapAPIKey
			resp, _ := http.Get(geo)
			body, _ := ioutil.ReadAll(resp.Body)
			jq := gojsonq.New().FromString(string(body))
			// log.Print(string(body))
			res := jq.From("results.[0].address_components").Where("types.[0]", "=", "route").Get()

			fmt.Printf("RoadID %s ,%f ,%f ", road.RoadID, road.Lat, road.Lon)
			fmt.Println(res.([]interface{})[0].(map[string]interface{})["long_name"].(string))
			roads = append(roads, road.RoadID)
		}
		tmpRoadID = road.RoadID
	}
	fmt.Printf("%v", roads)
}
