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
	"github.com/thedevsaddam/gojsonq"
	"google.golang.org/api/iterator"

)

const rangeLon float64 = 0.009
const rangeLat float64 = 0.008
const googleMapAPIKey string = "AIzaSyCzGP7dIwrOEuWxN8w40tBvwA_rvnbqudE"

type address struct {
	Original    string
	Destination string
}

// RouteWithParkings 導航地址
type RouteWithParkings struct {
	Address address
	Spaces  []ParkingSpace
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

//ParkingLot 新北市停車場
type ParkingLot struct {
	ID          int     //停車場序號
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

//ParkingLotAvailNTPC 新北市停車場剩餘數量
type ParkingLotAvailNTPC struct {
	ID           int //停車場序號
	AvailableCar int //剩餘數量
}
type roadName struct {
	RoadID   string
	RoadName string
}

type road struct {
	RoadID string
}
type userFavor struct {
	RoadID []string
}

func floatToString(num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(num, 'f', 6, 64)
}

func getAboutDist(userLat float64, userLon float64, lat float64, lon float64) (dist float64) {
	dist = math.Abs(userLat-lat) + math.Abs(userLon-lon)
	return
}

func getMapDist(userLat float64, userLon float64, lat float64, lon float64) (dist float64) {
	origins := floatToString(userLat) + "," + floatToString(userLon)
	destinations := floatToString(lat) + "," + floatToString(lon)
	// log.Printf("origins===",origins)
	// log.Printf("destinations===",destinations)
	url := "https://maps.googleapis.com/maps/api/distancematrix/json?origins=" + origins + "&destinations=" + destinations + "&key=" + googleMapAPIKey
	//fmt.Print(url)
	resp, _ := http.Get(url)
	body, _ := ioutil.ReadAll(resp.Body)
	jq := gojsonq.New().FromString(string(body)) //gojsonq解析json
	res := jq.Find("rows.[0].elements.[0].distance")
	dis := res.(map[string]interface{})
	dist = dis["value"].(float64)
	// distValue = dis["value"].(float64)
	// log.Print("valueeeee=", distValue)
	return
}

//GetGPS 路名轉gps
func GetGPS(roadName string) (lat float64, lon float64) {

	geocoding := "https://maps.googleapis.com/maps/api/geocode/json?address=" + roadName + "&key=" + googleMapAPIKey
	log.Print(geocoding)
	resp, _ := http.Get(geocoding)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Print(string(body))
	jq := gojsonq.New().FromString(string(body))    //gojsonq解析json
	res := jq.Find("results.[0].geometry.location") //可以直接點網址了解json結構
	gps := res.(map[string]interface{})             //interface型態轉回map
	lat = gps["lat"].(float64)
	lon = gps["lng"].(float64)
	return
}

func getRoadName(id string) (name string) {

	// query := datastore.NewQuery("NTPCRoadName").
	// 	Filter("RoadID =", id)
	// it := datastoreProc.client.Run(datastoreProc.ctx, query)

	key := datastore.NameKey("NTPCRoadName", id, nil)
	road := new(roadName)
	if err := DatastoreProc.client.Get(DatastoreProc.ctx, key, road); err != nil {

		log.Fatalf("Error fetching road name:%s %v", id, err)
	}

	name = road.RoadName

	return
}

func getUserFavor(userID string) (favor []string) {
	var favorRoads userFavor
	key := datastore.NameKey("userFavor", userID, nil)
	if err := DatastoreProc.client.Get(DatastoreProc.ctx, key, &favorRoads); err != nil && err != datastore.ErrNoSuchEntity {
		log.Fatalf("Error fetching favor road: %v", err)
	}
	return favorRoads.RoadID
}

//GetFeeInfo get fee info
func GetFeeInfo(carID string) (fees []FeeInfo) {
	var temp FeeInfo
	query := datastore.NewQuery("NTPCFeeInfo").
		Filter("CarID=", carID)
	it := DatastoreProc.client.Run(DatastoreProc.ctx, query)
	for {
		_, err := it.Next(&temp)
		if err == iterator.Done {
			break
		} else if err != nil {
			log.Fatalf("Error fetching fee info: %v", err)
		}
		fees = append(fees, temp)
	}
	return
}

//GetParkingsByFavor 以 favor 查車格
func GetParkingsByFavor(userID string) (result []ParkingSpace) {

	//query := datastore.NewQuery("userFavor").
	//Filter("__key__ =", key)
	favorRoads := getUserFavor(userID)

	//datastore 查詢剩餘車位
	var parkingSpaceList []ParkingSpace //儲存各路段離使用者最近且為空位的車格(一個路段一個) ex:[RoadID][distance,lat,lon,剩餘數量]
	for index, roadID := range favorRoads {
		for _, status := range []int{2, 3} { //2為空位,3為非收費時段,datastore查詢沒有or的方法，所以須查詢兩次
			query := datastore.NewQuery("NTPCParkingSpaces").
				Filter("RoadID=", roadID)

			var parking []ParkingSpace
			if len(parkingSpaceList) == index {
				if _, err := DatastoreProc.client.GetAll(DatastoreProc.ctx, query.Limit(1), &parking); err != nil {
					log.Fatalf("Error fetching favor road parking: %v", err)
				} else if len(parking) > 0 {
					parking[0].RoadName = getRoadName(roadID)
					parking[0].Distance = -1
					parkingSpaceList = append(parkingSpaceList, parking[0])
				}
			}

			if num, err := DatastoreProc.client.Count(DatastoreProc.ctx, query.Filter("CellStatus =", false).Filter("ParkingStatus =", status)); err != nil {
				log.Fatalf("Error counting favor road parking: %v", err)
			} else {
				fmt.Print(status, num)
				parkingSpaceList[index].Avail += num
			}
		}
	}

	return parkingSpaceList
}

// GetParkingSpacesByGPS  以GPS找車位資料
func GetParkingSpacesByGPS(lat float64, lon float64, IsOnlyEmpty bool) (result []ParkingSpace) {

	//datastore 查詢剩餘車位
	parkingSpaceList := make(map[string]ParkingSpace) //儲存各路段離使用者最近且為空位的車格(一個路段一個) ex:[RoadID][distance,lat,lon,剩餘數量]

	for _, i := range []int{2, 3} { //2為空位,3為非收費時段,datastore查詢沒有or的方法，所以須查詢兩次
		query := datastore.NewQuery("NTPCParkingSpaces").
			Filter("ParkingStatus =", i).
			Filter("Lat >", lat-rangeLat).
			Filter("Lat <", lat+rangeLat)
		if IsOnlyEmpty {
			query.Filter("CellStatus =", false) //false代表沒有車，但必須確認ParkingStatus必須為2或3才可停
		}
		it := DatastoreProc.client.Run(DatastoreProc.ctx, query)

		for {
			var parking ParkingSpace
			_, err := it.Next(&parking) //查詢後的結果一一迭代儲存到車格的struct

			if err == iterator.Done || len(parkingSpaceList) == 5 {
				break
			} else if err != nil {
				log.Fatalf("Error fetching road: %v", err)
			}
			if parking.Lon < lon-rangeLon || parking.Lon > lon+rangeLon { //datastore 只能對同一屬性作不等式filter 故需再次判斷lon
				continue
			}

			dist := getAboutDist(lat, lon, parking.Lat, parking.Lon) //經位度計算直線距離
			parking.Distance = dist
			parking.RoadName = getRoadName(parking.RoadID)

			if roadParkingSpace, ok := parkingSpaceList[parking.RoadID]; ok { //確認車格是否已在list內，有則比較直線距離，無則直接儲存
				parking.Avail = parkingSpaceList[parking.RoadID].Avail + 1
				if dist < roadParkingSpace.Distance { //比較同路段車格距離，若距離較小，則復寫到list
					parkingSpaceList[parking.RoadID] = parking
				}
			} else {
				parking.Avail = 1
				parkingSpaceList[parking.RoadID] = parking
			}
		}
	}

	var queryRes []ParkingSpace
	for _, tmp := range parkingSpaceList {
		tmp.Distance = getMapDist(lat, lon, tmp.Lat, tmp.Lon)
		if tmp.Distance > 1000 {
			continue
		}
		len := len(queryRes)
		if len == 0 {
			queryRes = append(queryRes, tmp)
		} else {
			for i := len; i >= 0; i-- {
				if i == 0 {
					queryRes = append([]ParkingSpace{tmp}, queryRes...)
				} else if tmp.Distance > queryRes[i-1].Distance {
					queryRes = append(queryRes, ParkingSpace{})
					copy(queryRes[i+1:], queryRes[i:])
					queryRes[i] = tmp
					break
				}
			}
		}
	}
	return queryRes
}

// GetParkingLotsByGPS  以GPS找停車場資料
func GetParkingLotsByGPS(lat float64, lon float64) (result []ParkingLot) {

	//datastore 查詢剩餘車位
	parkingLotList := make(map[int]ParkingLot) //儲存各路段離使用者最近且為空位的車格(一個路段一個) ex:[RoadID][distance,lat,lon,剩餘數量]

	query := datastore.NewQuery("NTPCParkingLots").
		Filter("Lat >", lat-rangeLat).
		Filter("Lat <", lat+rangeLat)

	it := DatastoreProc.client.Run(DatastoreProc.ctx, query)

	for {
		var lot ParkingLot
		_, err := it.Next(&lot) //查詢後的結果一一迭代儲存到車格的struct

		if err == iterator.Done || len(parkingLotList) == 5 {
			break
		} else if err != nil {
			log.Fatalf("Error fetching road: %v", err)
		}
		if lot.Lon < lon-rangeLon || lot.Lon > lon+rangeLon { //datastore 只能對同一屬性作不等式filter 故需再次判斷lon
			continue
		}

		dist := getAboutDist(lat, lon, lot.Lat, lot.Lon) //經位度計算直線距離
		lot.Distance = dist

		lotKey := datastore.NameKey("NTPCParkingLotsAvail", strconv.Itoa(lot.ID), nil)
		query := datastore.NewQuery("NTPCParkingLotsAvail").Ancestor(lotKey)

		var lotsAvail []ParkingLotAvailNTPC

		if _, err := DatastoreProc.client.GetAll(DatastoreProc.ctx, query.Limit(1), &lotsAvail); err != nil {
			log.Fatalf("Error fetching favor road parking: %v", err)
		} else if len(lotsAvail) == 1 {
			lot.Avail = lotsAvail[0].AvailableCar
		}
		parkingLotList[lot.ID] = lot

	}

	var queryRes []ParkingLot
	for _, tmp := range parkingLotList {
		tmp.Distance = getMapDist(lat, lon, tmp.Lat, tmp.Lon)
		if tmp.Distance > 1000 {
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
	return queryRes
}
func unmarshalFavorPostback(data string) map[string]string {
	postBack := make(map[string]string)
	fmt.Printf(data)
	tmp := strings.Split(data, "&")
	action := strings.Split(tmp[0], "=")[1]
	roadID := strings.Split(tmp[1], "=")[1]
	postBack["action"] = action
	postBack["roadID"] = roadID

	return postBack
}

//UserFavorModify 修改使用者 FAVOR
func UserFavorModify(userID string, datap string) (resp string) {
	key := datastore.NameKey("userFavor", userID, nil)
	data := unmarshalFavorPostback(datap)

	favorRoads := getUserFavor(userID)

	fmt.Print(favorRoads)

	switch data["action"] {
	case "加入最愛":
		if _, res := findFavorIndex(favorRoads, data["roadID"]); res == false {
			favorRoads = append(favorRoads, data["roadID"])
			resp = "新增成功"
		} else {
			resp = "已經在最愛裡面囉！"
		}
	case "移除":
		if index, res := findFavorIndex(favorRoads, data["roadID"]); res == true {
			favorRoads = append(favorRoads[:index], favorRoads[index+1:]...)
			resp = "移除成功"
		}

	}

	if _, err := DatastoreProc.client.Put(DatastoreProc.ctx, key, &userFavor{favorRoads}); err != nil {
		log.Fatalf("Put favor err: %v", err)
	}

	return
}

func findFavorIndex(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

/*查詢各路段 ID*/
//  for _, i := range id {
//  	query = datastore.NewQuery("NTPCParkings").
//  		Filter("RoadID =", i).
//  		Order("RoadID").
//  		Limit(1)
//  	it = datastoreProc.client.Run(datastoreProc.ctx, query)
//  	for {
//  		var road road
//  		_, err := it.Next(&road)
//  		if err == iterator.Done {
//  			break
//  		}
//  		if err != nil {
//  			log.Fatalf("Error fetching road: %v", err)
//  		}
// /*geocoding gps 轉路名*/
//  		fmt.Printf("RoadID %s ,%f ,%f ", road.RoadID, road.Lat, road.Lon)
//  		geo := "https://maps.googleapis.com/maps/api/geocode/json?latlng=" + fmt.Sprintf("%f", road.Lat) + "," + fmt.Sprintf("%f", road.Lon) + "&result_type=route&language=zh-tw&key=AIzaSyAhsij-kCTyOzK9Vq83zemmxJXTdNJVkV8"
//  		resp, _ := http.Get(geo)
//  		body, _ := ioutil.ReadAll(resp.Body)
//  		jq := gojsonq.New().FromString(string(body))
//  		res := jq.From("results.[0].address_components").Where("types.[0]", "=", "route").Get()
//  		fmt.Println(res.([]interface{})[0].(map[string]interface{})["long_name"].(string))
//  	}
//  }
