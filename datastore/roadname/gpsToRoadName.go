package roadname

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/thedevsaddam/gojsonq"
	"project.com/datastore/parkings"

)

// var datastoreProc datastoreProcessor

const projectID string = "exalted-yeti-289303"
const googleMapAPIKey string = "AIzaSyCzGP7dIwrOEuWxN8w40tBvwA_rvnbqudE"

/*GpsToRoadName 查詢各路段 ID*/
func gpsToRoadName() (roadData roads) {

	const parkingSpacesData string = "https://data.ntpc.gov.tw/api/datasets/54A507C4-C038-41B5-BF60-BBECB9D052C6/json"
	var spacesJq *gojsonq.JSONQ
	if NTPCParkingSpaceInfo, err := parkings.GetNTPCParkingsInfo(parkingSpacesData); err != nil {
		log.Fatalf("error: %v", err)
	} else {
		spacesJq = gojsonq.New().FromString(*NTPCParkingSpaceInfo)
	}

	res := spacesJq.GroupBy("ROADID").Get()

	for id, spaces := range res.(map[string][]interface{}) {

		space := spaces[0].(map[string]interface{})

		geo := "https://maps.googleapis.com/maps/api/geocode/json?latlng=" + space["lat"].(string) + "," + space["lon"].(string) + "&result_type=route&language=zh-tw&key=" + googleMapAPIKey
		resp, _ := http.Get(geo)
		body, _ := ioutil.ReadAll(resp.Body)
		jq := gojsonq.New().FromString(string(body))

		res := jq.From("results.[0].address_components").Where("types.[0]", "=", "route").Get()
		roadName := res.([]interface{})[0].(map[string]interface{})["long_name"].(string)

		roadData.info = append(roadData.info,
			&roadInfo{RoadID: id,
				RoadName: roadName})
	}

	return roadData

}
