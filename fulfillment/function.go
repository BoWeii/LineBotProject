package fulfillment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	//"cloud.google.com/go/datastore"
	//dialogflow "cloud.google.com/go/dialogflow/apiv2"
	//structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/thedevsaddam/gojsonq"
	"project.com/fulfillment/query"
	//"google.golang.org/api/option"
	//dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	//"project.com/fulfillment/carouselmessage"

)

// dialogflowProcessor has all the information for connecting with Dialogflow

// datastoreProcessor 存取 datastore

const projectID string = "parkingproject-2-283415"

var bot *linebot.Client

var err error

const FeeURL string = "https://data.ntpc.gov.tw/api/datasets/A676AF8E-D143-4D7A-95FE-99BB8DB5BCA0/json"

//response webhook回應
// type response struct {
// 	FulfillmentText string `json:"fulfillmentText"`
// }

// Pair A data structure to hold a key/value pair.

// init 初始化權限
func init() {
	bot, err = linebot.New("57cc60c3fc1530cc32ba896e1c4b7856", "GiKIwKk+Lwku0WeGEGnlEDBDDGC67tQVCSIMbcQaKpA2IyZPU6OgVSIdI0h1HUUG2Ky/psNLEEkjfnEZGITnJolxlEScGgLoWT/iKpwyinf/IJDgeB5gnIB0zmuag0vYlcs7WgOYdUg0CwbGXlWKIwdB04t89/1O/w1cDnyilFU=")
	err := query.DialogflowProc.Init(projectID, "fulfillment/parkingproject-2-283415-746d5d4c4c37.json", "zh-TW", "Asia/Hong_Kong")
	log.Println("init-------------------", err)
	query.DatastoreProc.Init(projectID)

}

func replyUser(resp interface{}, event *linebot.Event) {
	var respMessg linebot.SendingMessage
	switch resp.(type) { //確認是何種類型訊息
	case string:
		respMessg = linebot.NewTextMessage(resp.(string))
	case *linebot.BubbleContainer:
		respMessg = linebot.NewFlexMessage("使用介紹", resp.(*linebot.BubbleContainer))
	default:
		var container *linebot.CarouselContainer
		container = query.CreateCarouselmesage(resp)
		respMessg = linebot.NewFlexMessage("車位資訊。", container)
	}

	if _, err = bot.ReplyMessage(event.ReplyToken, respMessg).Do(); err != nil {
		log.Println(respMessg)
		log.Print("ReplyMessage Error ", err)
	}

}

func processByDialogflow(message string, UserID string) (resp interface{}) {

	response := query.DialogflowProc.ProcessNLP(message, UserID) //解析使用者所傳文字
	if response.Intent == "GetRoute" {
		if response.AllRequiredParamsPresent {
			lat, lon := query.GetGPS(response.Entities["destination"]) //路名轉GPS
			result := query.GetParkingsByGPS(lat, lon, false)
			route := query.RouteWithParkings{
				Parkings: result,
			}

			route.Address.Original = response.Entities["original"]
			route.Address.Destination = response.Entities["destination"]
			if len(route.Parkings) == 0 {
				resp = query.EmptyParkingBubbleMsg(route.Address)
			} else {
				log.Print(reflect.TypeOf(route))
				resp = route //查詢車格資訊
			}

		} else {
			resp = response.Prompts
		}
	} else if response.Intent == "GetFee" {
		if response.AllRequiredParamsPresent {
			var result []query.FeeInfo
			for i := 0; i <= 100; i++ {
				resp2, err := http.Get(FeeURL + "?page=" + strconv.Itoa(i) + "&size=1000")
				if err != nil {
					log.Fatal(err)
				}
				defer resp2.Body.Close()
				body, _ := ioutil.ReadAll(resp2.Body)
				bodyStr := "{\"Fee\":" + string(body) + "}"
				bodyStr = strings.ReplaceAll(bodyStr, "Amount_Ticket", "AmountTicket")
				jq := gojsonq.New().JSONString(string(bodyStr)) //gojsonq解析json
				feeInfo := jq.From("Fee").WhereContains("CarID", message).Get()
				feeString, err := json.Marshal(feeInfo)
				var temp []query.FeeInfo
				json.Unmarshal([]byte(feeString), &temp)
				result = append(result, temp...)
			}
			if len(result) == 0 {
				resp = "尚無此紀錄 😢"
			} else {
				resp = result
			}

		} else {
			resp = response.Prompts
		}
	} else {
		resp = response.Response
	}

	return
}

//Fulfillment 查詢車位
func Fulfillment(w http.ResponseWriter, r *http.Request) {

	var events []*linebot.Event
	events, err = bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
			log.Print(err)
		} else {
			w.WriteHeader(500)
			log.Print(err)
		}

	} else {
		w.WriteHeader(200)
	}

	var resp interface{} //回傳的訊息，可能為text、Carouselmesage，故用interface

	//可能不只一位使用者傳送訊息
	for _, event := range events {
		//訊息事件 https://developers.line.biz/en/reference/messaging-api/#common-properties

		if event.Type == linebot.EventTypeMessage {
			//訊息種類
			switch message := event.Message.(type) {
			case *linebot.TextMessage: //文字訊息
				resp = processByDialogflow(message.Text, event.Source.UserID)
			case *linebot.LocationMessage: //位置訊息
				fmt.Printf("gps %f,%f\n", message.Latitude, message.Longitude)
				parkings := query.GetParkingsByGPS(message.Latitude, message.Longitude, true)

				if len(parkings) == 0 {
					resp = "你附近沒有空車位哦 😢"
				} else {
					resp = parkings
				}
			}

			//加好友事件
		} else if event.Type == linebot.EventTypeFollow {
			resp = "歡迎加我好友 😍，本專案目前還在進行中，按下 ？ 讓我告訴你怎麼做"
		} else if event.Type == linebot.EventTypePostback {
			UserID := event.Source.UserID
			postbackData := event.Postback.Data
			log.Println("UserID", UserID, "  ", postbackData)

			switch postbackData {
			case "favor":
				parkings := query.GetParkingsByFavor(UserID)

				if len(parkings) == 0 {
					resp = "你還沒有最愛哦 😜"
				} else {
					resp = parkings
				}
			case "intro":
				resp = query.IntroBubbleMsg()
			case "query":
				resp = query.SearchBubbleMsg()
			case "route":
				resp = processByDialogflow("導航", event.Source.UserID)
			case "fee":
				resp = processByDialogflow("車費", event.Source.UserID)
			default:
				resp = query.UserFavorModify(UserID, postbackData)
			}

		}

		replyUser(resp, event) //回復使用者訊息
	}
}
