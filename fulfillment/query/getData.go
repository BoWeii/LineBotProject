package query

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/goinggo/mapstructure"
	"github.com/thedevsaddam/gojsonq"
	"google.golang.org/api/iterator"
)

const (
	rangeLon             float64 = 0.0045
	rangeLat             float64 = 0.004
	googleMapAPIKey      string  = "AIzaSyCzGP7dIwrOEuWxN8w40tBvwA_rvnbqudE"
	parkingSpacesData    string  = "https://data.ntpc.gov.tw/api/datasets/54A507C4-C038-41B5-BF60-BBECB9D052C6/json"
	parkingLotsAvailData string  = "https://data.ntpc.gov.tw/api/datasets/E09B35A5-A738-48CC-B0F5-570B67AD9C78/json"
	feeData              string  = "https://data.ntpc.gov.tw/api/datasets/A676AF8E-D143-4D7A-95FE-99BB8DB5BCA0/json"
)
const (
	parkingSpaces = iota
	parkingLotsAvail
)

type address struct {
	Original    string
	Destination string
}

// RouteWithParkings 導航地址
type RouteWithParkings struct {
	Address address
	Spaces  []ParkingSpace
	Lots    []ParkingLot
}

//FeeInfo 繳費資訊
type FeeInfo struct {
	TicketNo     string //收費編號
	CarID        string //車牌號碼
	Parkdt       string //開單日
	Paylim       string //繳費截止日
	AmountTicket string //費用
	CarType      string //車種
}

//ParkingSpace 停車格
type ParkingSpace struct {
	ID            string  //車格序號
	CELLID        string  //車格編號
	Name          string  //車格類型
	Day           string  //收費天
	Hour          string  //收費時段
	Pay           string  //收費形式
	PayCash       string  //費率
	Memo          string  //車格備註
	RoadID        string  //路段代碼
	CellStatus    string  //車格狀態判斷 Y有車 N空位
	IsNowCash     string  //收費時段判斷
	ParkingStatus string  //車格狀態 　1：有車、2：空位、3：非收費時段、4：時段性禁停、5：施工（民眾申請施工租用車格時使用）
	Lat           float64 //緯度
	Lon           float64 //經度
	Distance      float64 //距離
	Avail         int
	RoadName      string
}

//ParkingLot 新北市停車場
type ParkingLot struct {
	ID          string  //停車場序號
	Name        string  //停車場名稱
	Type        int     //1：剩餘車位數 2：靜態停車場資料
	Tel         string  //停車場電話
	Pay         string  //停車場收費資訊
	ServiceTime string  //服務時間
	TotalCar    int     //總汽車數
	TotalMotor  int     //總機車數
	Lat         float64 //緯度
	Lon         float64 //經度
	Distance    float64 //距離
	Avail       int
}

type roadName struct {
	RoadID   string
	RoadName string
}

type userFavor struct {
	RoadID []string
	LotID  []string
}

const (
	roadID      = "roadID"
	lotID       = "lotID"
	parkingType = "type"
)

func floatToString(num float64) string {
	return fmt.Sprintf("%f", num)
}

func getAboutDist(userLat float64, userLon float64, lat float64, lon float64) (dist float64) {
	dist = math.Abs(userLat-lat) + math.Abs(userLon-lon)
	return
}

func getMapDist(userLat float64, userLon float64, lat float64, lon float64) (dist float64) {
	origins := floatToString(userLat) + "," + floatToString(userLon)
	destinations := floatToString(lat) + "," + floatToString(lon)

	url := "https://maps.googleapis.com/maps/api/distancematrix/json?origins=" + origins + "&destinations=" + destinations + "&key=" + googleMapAPIKey

	resp, _ := http.Get(url)
	body, _ := ioutil.ReadAll(resp.Body)
	jq := gojsonq.New().FromString(string(body)) //gojsonq解析json
	res := jq.Find("rows.[0].elements.[0].distance")
	dis := res.(map[string]interface{})
	dist = dis["value"].(float64)

	return
}

//GetGPS 路名轉gps
func GetGPS(roadName string) (lat float64, lon float64) {

	geocoding := "https://maps.googleapis.com/maps/api/geocode/json?address=" + roadName + "&key=" + googleMapAPIKey

	resp, _ := http.Get(geocoding)
	body, _ := ioutil.ReadAll(resp.Body)

	jq := gojsonq.New().FromString(string(body))    //gojsonq解析json
	res := jq.Find("results.[0].geometry.location") //可以直接點網址了解json結構
	gps := res.(map[string]interface{})             //interface型態轉回map
	lat = gps["lat"].(float64)
	lon = gps["lng"].(float64)
	return
}

func getRoadName(id string) (name string) {

	key := datastore.NameKey("NTPCRoadName", id, nil)
	road := new(roadName)
	if err := DatastoreProc.client.Get(DatastoreProc.ctx, key, road); err != nil {

		log.Fatalf("Error fetching road name:%s %v", id, err)
	}

	name = road.RoadName

	return
}

func getUserFavor(userID string) (favor userFavor) {

	key := datastore.NameKey("userFavor", userID, nil)
	if err := DatastoreProc.client.Get(DatastoreProc.ctx, key, &favor); err != nil && err != datastore.ErrNoSuchEntity {
		log.Fatalf("Error fetching favor road: %v", err)
	}
	return favor
}

//GetFeeInfo get fee info
func GetFeeInfo(carID string) (fees []FeeInfo) {

	info := getFeeInfo(feeData)
	jq := gojsonq.New().FromString(*info)
	res := jq.Where("CarID", "=", carID).OrWhere("id", "=", nil).Get()

	var result FeeInfo
	for _, item := range res.([]interface{}) {
		err := mapstructure.Decode(item.(map[string]interface{}), &result)
		fees = append(fees, result)
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(fees) > 10 {
		fees = fees[:10]
	}
	return
}

//GetParkingsByFavor 以 favor 查車格
func GetParkingsByFavor(userID string) ([]ParkingSpace, []ParkingLot) {

	favorRoads := getUserFavor(userID)

	var spacesJq *gojsonq.JSONQ
	if len(favorRoads.LotID)+len(favorRoads.RoadID) != 0 {
		if NTPCParkingSpaceInfo, err := getNTPCParkingsInfo(parkingSpacesData, parkingSpaces, 31); err != nil {
			log.Fatalf("error: %v", err)
		} else {
			spacesJq = gojsonq.New().FromString(*NTPCParkingSpaceInfo)
		}
	}

	//datastore 查詢剩餘車位
	var parkingSpaceList []ParkingSpace //儲存各路段離使用者最近且為空位的車格(一個路段一個) ex:[RoadID][distance,lat,lon,剩餘數量]
	for _, roadID := range favorRoads.RoadID {
		spacesJq.Reset()
		res := spacesJq.Where("ROADID", "=", roadID).First()

		var parkingSpace ParkingSpace

		err := mapstructure.Decode(res.(map[string]interface{}), &parkingSpace)
		parkingSpace.RoadName = getRoadName(roadID)
		parkingSpace.Distance = -1
		parkingSpace.Avail = spacesJq.Where("CELLSTATUS", "=", "N").
			WhereIn("ParkingStatus", []string{"2", "3"}).Count()
		parkingSpaceList = append(parkingSpaceList, parkingSpace)

		if err != nil {
			log.Fatal("Decode:", err)
		}

	}

	var lotsJq *gojsonq.JSONQ
	if NTPCParkingLotsAvail, err := getNTPCParkingsInfo(parkingLotsAvailData, parkingLotsAvail, 1); err != nil {
		log.Fatalf("error: %v", err)
	} else {
		lotsJq = gojsonq.New().FromString(*NTPCParkingLotsAvail)
	}

	var parkingLotList []ParkingLot
	for _, LotID := range favorRoads.LotID {

		key := datastore.NameKey("NTPCParkingLots", LotID, nil)
		var lot ParkingLot

		if err := DatastoreProc.client.Get(DatastoreProc.ctx, key, &lot); err != nil {
			log.Fatalf("Error fetching favor parking lot: %v", err)
		} else {
			lot.Distance = -1
			if lot.Type == 1 {
				lotsJq.Reset()
				res := lotsJq.Where("id", "=", LotID).First()
				lot.Avail, _ = strconv.Atoi(res.(map[string]interface{})["availableCar"].(string))
			}

			parkingLotList = append(parkingLotList, lot)
		}

	}
	return parkingSpaceList, parkingLotList
}

// GetParkingSpacesByGPS  以GPS找車位資料
func GetParkingSpacesByGPS(lat float64, lon float64, IsOnlyEmpty bool, maxLen int) (result []ParkingSpace) {

	if maxLen > 10 {
		log.Fatalln("GetParkingSpacesByGPS MaxLen <=10 ")
	}

	if NTPCParkingSpaceInfo, err := getNTPCParkingsInfo(parkingSpacesData, parkingSpaces, 31); err != nil {
		log.Fatalf("error: %v", err)
	} else {
		parkingSpaceList := make(map[string]ParkingSpace)

		status := []string{"N"}
		if !IsOnlyEmpty {
			status = append(status, "Y")
		}

		jq := gojsonq.New().FromString(*NTPCParkingSpaceInfo)
		res := jq.WhereIn("CELLSTATUS", status).
			WhereIn("ParkingStatus", []string{"2", "3"}).
			Where("lat", ">", lat-rangeLat).Where("lat", "<", lat+rangeLat).
			Where("lon", ">", lon-rangeLon).Where("lon", "<", lon+rangeLon).
			Get()

		var parkingSpace ParkingSpace

		for _, item := range res.([]interface{}) {

			err := mapstructure.Decode(item.(map[string]interface{}), &parkingSpace)

			if roadParkingSpace, ok := parkingSpaceList[parkingSpace.RoadID]; ok { //確認車格是否已在list內，有則比較直線距離，無則直接儲存
				parkingSpace.Avail = parkingSpaceList[parkingSpace.RoadID].Avail + 1
				dist := getAboutDist(lat, lon, parkingSpace.Lat, parkingSpace.Lon) //經位度計算直線距離

				if dist < roadParkingSpace.Distance { //比較同路段車格距離，若距離較小，則復寫到list
					parkingSpaceList[parkingSpace.RoadID] = parkingSpace
				} else {
					tmp := parkingSpaceList[parkingSpace.RoadID]
					tmp.Avail++
					parkingSpaceList[parkingSpace.RoadID] = tmp
				}

			} else {
				parkingSpace.Avail = 1
				parkingSpaceList[parkingSpace.RoadID] = parkingSpace
			}
			if err != nil {
				log.Fatal("Decode:", err)
			}
		}

		for _, tmp := range parkingSpaceList {
			tmp.Distance = getMapDist(lat, lon, tmp.Lat, tmp.Lon)
			if tmp.Distance > 500 {
				continue
			}
			tmp.RoadName = getRoadName(tmp.RoadID)
			len := len(result)
			if len == 0 {
				result = append(result, tmp)
			} else {
				for i := len; i >= 0; i-- {
					if i == 0 {
						result = append([]ParkingSpace{tmp}, result...)
					} else if tmp.Distance > result[i-1].Distance {
						result = append(result, ParkingSpace{})
						copy(result[i+1:], result[i:])
						result[i] = tmp
						break
					}
				}
			}
		}
	}

	if len(result) > maxLen {
		result = result[:maxLen]
	}

	return
}

// GetParkingLotsByGPS  以GPS找停車場資料
func GetParkingLotsByGPS(lat float64, lon float64, maxLen int) (result []ParkingLot) {

	if maxLen > 10 {
		log.Fatalln("GetParkingSpacesByGPS MaxLen <=10 ")
	}
	//datastore 查詢剩餘車位
	parkingLotList := make(map[string]ParkingLot) //儲存各路段離使用者最近且為空位的車格(一個路段一個) ex:[RoadID][distance,lat,lon,剩餘數量]

	query := datastore.NewQuery("NTPCParkingLots").
		Filter("Lat >", lat-rangeLat).
		Filter("Lat <", lat+rangeLat)

	it := DatastoreProc.client.Run(DatastoreProc.ctx, query)

	var lotsJq *gojsonq.JSONQ
	if NTPCParkingLotsAvail, err := getNTPCParkingsInfo(parkingLotsAvailData, parkingLotsAvail, 1); err != nil {
		log.Fatalf("error: %v", err)
	} else {
		lotsJq = gojsonq.New().FromString(*NTPCParkingLotsAvail)
	}

	for {
		var lot ParkingLot
		_, err := it.Next(&lot) //查詢後的結果一一迭代儲存到車格的struct

		if err == iterator.Done {
			break
		} else if err != nil {
			log.Fatalf("Error fetching road: %v", err)
		}
		if lot.Lon < lon-rangeLon || lot.Lon > lon+rangeLon { //datastore 只能對同一屬性作不等式filter 故需再次判斷lon
			continue
		}

		dist := getAboutDist(lat, lon, lot.Lat, lot.Lon) //經位度計算直線距離
		lot.Distance = dist
		if lot.Type == 1 {
			lotsJq.Reset()
			res := lotsJq.Where("id", "=", lot.ID).First()
			lot.Avail, _ = strconv.Atoi(res.(map[string]interface{})["availableCar"].(string))
		}
		parkingLotList[lot.ID] = lot

	}

	var queryRes []ParkingLot
	for _, tmp := range parkingLotList {
		//fmt.Printf("%v\n", tmp)
		tmp.Distance = getMapDist(lat, lon, tmp.Lat, tmp.Lon)
		if tmp.Distance > 500 {
			continue
		}
		len := len(queryRes)
		if len == 0 {
			queryRes = append(queryRes, tmp)
		} else {
			for i := len; i >= 0; i-- {
				if i == 0 {
					queryRes = append([]ParkingLot{tmp}, queryRes...)
				} else if tmp.Distance > queryRes[i-1].Distance {
					queryRes = append(queryRes, ParkingLot{})
					copy(queryRes[i+1:], queryRes[i:])
					queryRes[i] = tmp
					break
				}
			}
		}
	}

	if len(queryRes) > maxLen {
		queryRes = queryRes[:maxLen]
	}

	return queryRes

}
func unmarshalFavorPostback(data string) map[string]string {
	postBack := make(map[string]string)
	//fmt.Printf(data)
	tmp := strings.Split(data, "&")
	action := strings.Split(tmp[0], "=")[1]

	parkingType := strings.Split(tmp[1], "=")[0]
	id := strings.Split(tmp[1], "=")[1]
	if parkingType == "roadID" {
		postBack["roadID"] = id
		postBack["type"] = roadID
	} else {
		postBack["lotID"] = id
		postBack["type"] = lotID
	}
	postBack["action"] = action

	return postBack
}

//UserFavorModify 修改使用者 FAVOR
func UserFavorModify(userID string, postBack string) (resp string) {
	key := datastore.NameKey("userFavor", userID, nil)
	postBackdata := unmarshalFavorPostback(postBack)

	favor := getUserFavor(userID)
	modifyType := postBackdata[parkingType]

	var modifyFavor *[]string

	if modifyType == roadID {
		modifyFavor = &favor.RoadID
	} else {
		modifyFavor = &favor.LotID
	}

	switch postBackdata["action"] {
	case "加入最愛":
		if _, res := findFavorIndex(favor, postBackdata); res == false {
			*modifyFavor = append(*modifyFavor, postBackdata[modifyType])
			resp = "新增成功"
		} else {
			resp = "已經在最愛裡面囉！"
		}
	case "移除":
		if index, res := findFavorIndex(favor, postBackdata); res == true {
			(*modifyFavor) = append((*modifyFavor)[:index], (*modifyFavor)[index+1:]...)
			resp = "移除成功"
		}

	}

	if _, err := DatastoreProc.client.Put(DatastoreProc.ctx, key, &favor); err != nil {
		log.Fatalf("Put favor err: %v", err)
	}

	return
}

func findFavorIndex(favor userFavor, postback map[string]string) (int, bool) {
	var slice []string
	var val string
	if postback[parkingType] == roadID {
		slice = favor.RoadID
		val = postback[roadID]
	} else {
		slice = favor.LotID
		val = postback[lotID]
	}

	for i, item := range slice {
		if item == val {
			return i, true
		}
	}

	return -1, false
}

// 取得停車費用
func getFeeInfo(url string) *string {
	var data string
	data = "["
	ch := make(chan string)
	for i := 0; i <= 40; i++ {
		go fetchFee(url+"?page="+strconv.Itoa(i)+"&size=1000", ch)
	}
	for i := 0; i <= 40; i++ {
		if i == 0 {
			data = data + <-ch
		} else {
			data = data + "," + <-ch
		}
	}
	data += "]"
	return &data
}

//getNTPCParkingsInfo
func getNTPCParkingsInfo(url string, dataType int, page int) (*string, error) {
	var data string
	data = "["

	ch := make(chan string)
	for i := 0; i <= page; i++ {
		go fetchParking(url+"?page="+strconv.Itoa(i)+"&size=1000", dataType, ch)
	}

	for i := 0; i <= page; i++ {
		if i == 0 {
			data = data + <-ch
		} else {
			data = data + "," + <-ch
		}
	}
	data += "]"
	return &data, nil
}

func fetchParking(url string, dataType int, ch chan<- string) {

	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	bodyString := string(body)
	bodyString = strings.Replace(bodyString, "[", "", 1)
	bodyString = strings.Replace(bodyString, "]", "", 1)
	if dataType == parkingSpaces {
		bodyString = strings.ReplaceAll(bodyString, "\"lat\":\"", "\"lat\":")
		bodyString = strings.ReplaceAll(bodyString, "\",\"lon\"", ",\"lon\"")
		bodyString = strings.ReplaceAll(bodyString, "\"lon\":\"", "\"lon\":")
		bodyString = strings.ReplaceAll(bodyString, "\"}", "}")
	}
	ch <- bodyString
}
func fetchFee(url string, ch chan<- string) {

	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	bodyString := string(body)
	bodyString = strings.Replace(bodyString, "[", "", 1)
	bodyString = strings.Replace(bodyString, "]", "", 1)
	bodyString = strings.ReplaceAll(bodyString, "   ", "")
	bodyString = strings.ReplaceAll(bodyString, "Amount_Ticket", "AmountTicket")

	ch <- bodyString
}
