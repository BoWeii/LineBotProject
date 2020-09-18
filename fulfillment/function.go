package fulfillment

import (
	"fmt"
	"log"
	"net/http"

	//"cloud.google.com/go/datastore"
	//dialogflow "cloud.google.com/go/dialogflow/apiv2"
	//structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/line/line-bot-sdk-go/linebot"
	// "github.com/thedevsaddam/gojsonq"
	"project.com/fulfillment/query"
	//"google.golang.org/api/option"
	//dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	//"project.com/fulfillment/carouselmessage"

)

// dialogflowProcessor has all the information for connecting with Dialogflow

// datastoreProcessor 存取 datastore

const projectID string = "exalted-yeti-289303"

var bot *linebot.Client

var err error

const feeURL string = "https://data.ntpc.gov.tw/api/datasets/A676AF8E-D143-4D7A-95FE-99BB8DB5BCA0/json"

//response webhook回應
// type response struct {
// 	FulfillmentText string `json:"fulfillmentText"`
// }

// Pair A data structure to hold a key/value pair.

// init 初始化權限
func init() {
	bot, err = linebot.New("57cc60c3fc1530cc32ba896e1c4b7856", "GiKIwKk+Lwku0WeGEGnlEDBDDGC67tQVCSIMbcQaKpA2IyZPU6OgVSIdI0h1HUUG2Ky/psNLEEkjfnEZGITnJolxlEScGgLoWT/iKpwyinf/IJDgeB5gnIB0zmuag0vYlcs7WgOYdUg0CwbGXlWKIwdB04t89/1O/w1cDnyilFU=")
	query.DialogflowProc.Init(projectID, "exalted-yeti-289303-753a4c88e472.json", "zh-TW", "Asia/Hong_Kong")

	query.DatastoreProc.Init(projectID)

}

func replyUser(resp interface{}, event *linebot.Event) {
	var respMessg []linebot.SendingMessage
	switch resp.(type) { //確認是何種類型訊息
	case string:
		respMessg = append(respMessg, linebot.NewTextMessage(resp.(string)))
	case *linebot.BubbleContainer:
		respMessg = append(respMessg, linebot.NewFlexMessage("使用介紹", resp.(*linebot.BubbleContainer)))
	case []query.ParkingSpace:
		container := query.CreateCarouselmesage(resp)
		respMessg = append(respMessg, linebot.NewFlexMessage("車位資訊。", container))
	case []query.ParkingLot:
		container := query.CreateCarouselmesage(resp)
		respMessg = append(respMessg, linebot.NewFlexMessage("停車場資訊。", container))
	case query.RouteWithParkings:
		container := query.CreateCarouselmesage(resp)
		respMessg = append(respMessg, linebot.NewFlexMessage("車位資訊。", container))
	case []query.FeeInfo:
		container := query.CreateCarouselmesage(resp)
		respMessg = append(respMessg, linebot.NewFlexMessage("待繳車費。", container))
	default:
		spacesContainer := query.CreateCarouselmesage(resp.([2]interface{})[0])
		lotsContainer := query.CreateCarouselmesage(resp.([2]interface{})[1])
		respMessg = append(respMessg, linebot.NewFlexMessage("車位資訊。", spacesContainer), linebot.NewFlexMessage("停車場資訊。", lotsContainer))
	}

	if len(respMessg) == 1 {
		_, err = bot.ReplyMessage(event.ReplyToken, respMessg[0]).Do()
	} else {
		_, err = bot.ReplyMessage(event.ReplyToken, respMessg[0], respMessg[1]).Do()
	}

	if err != nil {
		log.Println(err)
	}
}

func processByDialogflow(message string, UserID string) (resp interface{}) {

	response := query.DialogflowProc.ProcessNLP(message, UserID) //解析使用者所傳文字
	if response.Intent == "GetRoute" {
		if response.AllRequiredParamsPresent {
			lat, lon := query.GetGPS(response.Entities["destination"]) //路名轉GPS
			spaces := query.GetParkingSpacesByGPS(lat, lon, false, 3)
			lots := query.GetParkingLotsByGPS(lat, lon, 3)
			route := query.RouteWithParkings{
				Spaces: spaces,
				Lots:   lots,
			}

			route.Address.Original = response.Entities["original"]
			route.Address.Destination = response.Entities["destination"]
			if len(route.Spaces) == 0 {
				resp = query.EmptyParkingBubbleMsg(route.Address)
			} else {
				//log.Print(reflect.TypeOf(route))
				resp = route //查詢車格資訊
			}

		} else {
			resp = response.Prompts
		}
	} else if response.Intent == "GetFee" {
		if response.AllRequiredParamsPresent {
			result := query.GetFeeInfo(message)
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
				spaces := query.GetParkingSpacesByGPS(message.Latitude, message.Longitude, true, 10)
				lots := query.GetParkingLotsByGPS(message.Latitude, message.Longitude, 10)
				//fmt.Printf("%v", lots)
				if len(spaces) == 0 && len(lots) == 0 {
					resp = "你附近沒有空車位哦 😢"
				} else {
					if len(spaces) > 0 && len(lots) > 0 {
						resp = [2]interface{}{spaces, lots}
					} else if len(spaces) > 0 {
						resp = spaces
					} else {
						resp = lots
					}
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
				spaces, lots := query.GetParkingsByFavor(UserID)
				if len(spaces) > 0 && len(lots) > 0 {
					resp = [2]interface{}{spaces, lots}
				} else if len(spaces) == 0 && len(lots) == 0 {
					resp = "你還沒有最愛哦 😜"
				} else if len(spaces) > 0 {
					resp = spaces
				} else if len(lots) > 0 {
					resp = lots
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
