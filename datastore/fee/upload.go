package fee

import (
	"context"
	"log"
	"math"
	"strings"
	"cloud.google.com/go/datastore"
	"encoding/json"
	"strconv"
	"net/http"
	"io/ioutil"
)

//FeeInfo 繳費資訊
type feeInfo struct {
	TicketNo     string //收費編號
	CarID        string //車牌號碼
	Parkdt       string //開單日
	Paylim       string //繳費截止日
	AmountTicket string //費用
	CarType      string //車種
}
type Fees struct {
	infos []*feeInfo
}
const FeeURL string = "https://data.ntpc.gov.tw/api/datasets/A676AF8E-D143-4D7A-95FE-99BB8DB5BCA0/json"
const projectID string = "exalted-yeti-289303"


//UpdateParkingInfo consumes a Pub/Sub message 並更新停車位資訊
func UpdateFeeInfo(ctx context.Context) error {
	//取open data
	FeeInfo, err := getParkingInfo(FeeURL)
	//fmt.Printf(*TPEParkingInfo)
	if err != nil {
		log.Print(err)
	}

	var fees Fees
	feeKeys := []*datastore.Key{}

	//json轉struct
	if err := json.Unmarshal([]byte(*FeeInfo), &fees.infos); err != nil {
		log.Fatalf("error: %v", err)
	}

	//以roadID產生entity key
	for _, fee := range fees.infos {
		// log.Print(fee.CarID)
		feeKey := datastore.NameKey("NTPCFeeInfo",fee.TicketNo,nil)
		feeKeys = append(feeKeys, feeKey)
	}
	//log.Print(&NTPC)
	
	putFeeInfo(ctx, feeKeys, &fees)
	return nil
}
//put fee informations
func putFeeInfo(ctx context.Context, feeKeys []*datastore.Key, fees *Fees) {

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	//fmt.Print(parkings.(*parking.NTPC))

	var tmp = 0
	n := math.Ceil(float64(len(feeKeys)) / 500) //一次最多put500筆

	for i := 1; i <= int(n); i++ {
		var size int
		if size = len(feeKeys); i*500 < len(feeKeys) {
			size = i * 500
		}
		//fmt.Print("----------", i, parkings.(*parking.NTPC).Cells[tmp:size-1])
		if _, err := client.PutMulti(ctx, feeKeys[tmp:size-1], fees.infos[tmp:size-1]); err != nil {
			log.Fatalf("PutMulti Fees Info: %v", err)
		}
		tmp = size - 1
	}

	log.Println("Fee Info Saved sucess")

}

// 取得停車費用
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
		data = data + temp + ","
	}
	data = data[:len(data)-1]
	data += "]"
	return &data, nil
}