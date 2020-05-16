package fulfillment

import (
	"fmt"
	"log"
	"net/http"

	//"cloud.google.com/go/datastore"
	//dialogflow "cloud.google.com/go/dialogflow/apiv2"
	//structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/line/line-bot-sdk-go/linebot"
	"project.com/fulfillment/query"
	//"google.golang.org/api/option"
	//dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	//"project.com/fulfillment/carouselmessage"
)

// dialogflowProcessor has all the information for connecting with Dialogflow

// datastoreProcessor 存取 datastore

const projectID string = "parkingproject-261207"

var bot *linebot.Client

var err error

//response webhook回應
// type response struct {
// 	FulfillmentText string `json:"fulfillmentText"`
// }

// Pair A data structure to hold a key/value pair.

// init 初始化權限
func init() {
	bot, err = linebot.New("57cc60c3fc1530cc32ba896e1c4b7856", "GiKIwKk+Lwku0WeGEGnlEDBDDGC67tQVCSIMbcQaKpA2IyZPU6OgVSIdI0h1HUUG2Ky/psNLEEkjfnEZGITnJolxlEScGgLoWT/iKpwyinf/IJDgeB5gnIB0zmuag0vYlcs7WgOYdUg0CwbGXlWKIwdB04t89/1O/w1cDnyilFU=")
	query.DialogflowProc.Init(projectID, "parkingproject-261207-2933e4112308.json", "zh-TW", "Asia/Hong_Kong")
	query.DatastoreProc.Init(projectID)

}

func replyUser(resp interface{}, event *linebot.Event) {

	var respMessg linebot.SendingMessage
	switch resp.(type) { //確認是何種類型訊息
	case string:
		respMessg = linebot.NewTextMessage(resp.(string))
	case []query.Parking:
		parkings := resp.([]query.Parking)

		var container *linebot.CarouselContainer

		if parkings[0].Distance > 0 {
			container = query.Carouselmesage(parkings, "加入最愛")
		} else {
			container = query.Carouselmesage(parkings, "移除")
		}
		respMessg = linebot.NewFlexMessage("車位資訊。", container)
	case *linebot.BubbleContainer:
		respMessg = linebot.NewFlexMessage("使用介紹", resp.(*linebot.BubbleContainer))
	}

	if _, err = bot.ReplyMessage(event.ReplyToken, respMessg).Do(); err != nil {
		log.Print(err)
	}

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

				response := query.DialogflowProc.ProcessNLP(message.Text, event.Source.UserID) //解析使用者所傳文字

				if response.Intent == "FindParking" {
					if _, ok := response.Entities["location"]; ok {
						lat, lon := query.GetGPS(response.Entities["location"]) //路名轉GPS
						resp = query.GetParkingsByGPS(lat, lon)                 //查詢車格資訊
					} else {
						resp = response.Prompts //如果偵測到intent卻沒有entity，回傳提示輸入訊息
					}
				} else {
					resp = "我聽不太懂"
				}

			case *linebot.LocationMessage: //位置訊息
				fmt.Printf("gps %f,%f\n", message.Latitude, message.Longitude)

				parkings := query.GetParkingsByGPS(message.Latitude, message.Longitude)

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

			if postbackData == "favor" {
				parkings := query.GetParkingsByFavor(UserID)

				if len(parkings) == 0 {
					resp = "你還沒有最愛哦 😜"
				} else {
					resp = parkings
				}

			} else if postbackData == "intro" {
				resp = query.IntroBubbleMsg()
			} else {
				resp = query.UserFavorModify(UserID, postbackData)
			}

		}

		replyUser(resp, event) //回復使用者訊息
	}
}
