package roadname

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/thedevsaddam/gojsonq"
	"google.golang.org/api/iterator"
	"project.com/datastore/parkingstruct"

)

var datastoreProc datastoreProcessor

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
func gpsToRoadName() (roadData roads) {

	datastoreProc.Init(projectID)
	query := datastore.NewQuery("NTPCParkings").Project("RoadID", "Lat", "Lon").DistinctOn("RoadID") //.Filter("CellStatus =", false)

	it := datastoreProc.client.Run(datastoreProc.ctx, query)

	for {

		var parkingSpace parkingstruct.ParkingSpaceNTPC
		_, err := it.Next(&parkingSpace)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error fetching road: %v", err)
		}

		geo := "https://maps.googleapis.com/maps/api/geocode/json?latlng=" + fmt.Sprintf("%f", parkingSpace.Lat) + "," + fmt.Sprintf("%f", parkingSpace.Lon) + "&result_type=route&language=zh-tw&key=" + googleMapAPIKey
		resp, _ := http.Get(geo)
		body, _ := ioutil.ReadAll(resp.Body)
		jq := gojsonq.New().FromString(string(body))

		res := jq.From("results.[0].address_components").Where("types.[0]", "=", "route").Get()
		roadName := res.([]interface{})[0].(map[string]interface{})["long_name"].(string)
		//	log.Printf("RoadID %s,RoadName %s ,%f ,%f ", parking.RoadID, roadName, parking.Lat, parking.Lon)

		roadData.info = append(roadData.info,
			&roadInfo{RoadID: parkingSpace.RoadID,
				RoadName: roadName})
	}

	return roadData

}
