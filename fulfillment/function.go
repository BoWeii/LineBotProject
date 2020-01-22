package fulfillment

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tidwall/gjson"
)

//Fulfillment 查詢車位
func Fulfillment(w http.ResponseWriter, r *http.Request) {
	buf, _ := ioutil.ReadAll(r.Body)
	roadName := gjson.Get(string(buf), "queryResult.parameters.RoadName")
	defer r.Body.Close()
	log.Printf("roadName: %s", roadName)
	// buf, _ := ioutil.ReadAll(r.Body)
	// body := ioutil.NopCloser(bytes.NwBuffer(buf))
	// log.Printf("BODY: %q", body)
}
