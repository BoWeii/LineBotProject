package fulfillment

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"cloud.google.com/go/datastore"
	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/line/line-bot-sdk-go/linebot"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	"project.com/fulfillment/carouselmessage"

)

//road 路段停車格
type road struct {
	RoadSegAvail    string `json:"roadSegAvail"`    //路段剩餘格位數
	RoadSegFee      string `json:"roadSegFee"`      //收費標準
	RoadSegID       string `json:"roadSegID"`       //路段ID
	RoadSegName     string `json:"roadSegName"`     //路段名稱
	RoadSegTmEnd    string `json:"roadSegTmEnd"`    //收費結束時間
	RoadSegTmStart  string `json:"roadSegTmStart"`  //收費開始時間
	RoadSegTotal    string `json:"roadSegTotal"`    //路段總格位數
	RoadSegUpdateTm string `json:"roadSegUpdateTm"` //資料更新時間
	RoadSegUsage    string `json:"roadSegUsage"`    //路段使用率
}

// dialogflowProcessor has all the information for connecting with Dialogflow
type dialogflowProcessor struct {
	projectID        string
	authJSONFilePath string
	lang             string
	timeZone         string
	sessionClient    *dialogflow.SessionsClient
	ctx              context.Context
}

// datastoreProcessor 存取 datastore
type datastoreProcessor struct {
	projectID string
	client    *datastore.Client
	ctx       context.Context
}

// nlpResponse is webhook回應
type nlpResponse struct {
	Intent     string            `json:"intent"`
	Confidence float32           `json:"confidence"`
	Entities   map[string]string `json:"entities"`
}

const projectID string = "parkingproject-261207"

var dialogflowProc dialogflowProcessor
var datastoreProc datastoreProcessor
var bot *linebot.Client

var err error

//response webhook回應
type response struct {
	FulfillmentText string `json:"fulfillmentText"`
}

// init 初始化權限
func init() {
	bot, err = linebot.New("57cc60c3fc1530cc32ba896e1c4b7856", "GiKIwKk+Lwku0WeGEGnlEDBDDGC67tQVCSIMbcQaKpA2IyZPU6OgVSIdI0h1HUUG2Ky/psNLEEkjfnEZGITnJolxlEScGgLoWT/iKpwyinf/IJDgeB5gnIB0zmuag0vYlcs7WgOYdUg0CwbGXlWKIwdB04t89/1O/w1cDnyilFU=")
	dialogflowProc.init(projectID, "parkingproject-261207-2933e4112308.json", "zh-TW", "Asia/Hong_Kong")
	datastoreProc.init(projectID)

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

	var respText string
	//可能不只一位使用者傳送訊息
	for _, event := range events {
		//訊息事件 https://developers.line.biz/en/reference/messaging-api/#common-properties
		if event.Type == linebot.EventTypeMessage {
			//訊息種類
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				response := dialogflowProc.processNLP(message.Text, "testUser")

				if response.Intent == "FindParking" {
					respText = getData(response.Entities["RoadName"], response.Intent)
				} else {
					respText = "我聽不太懂"
				}

			case *linebot.ImageMessage:
				fmt.Print("image")
			case *linebot.LocationMessage:
				fmt.Print("location:", message.Address)
			}
			//追隨事件
		} else if event.Type == linebot.EventTypeFollow {
			respText = "還敢加我好友啊"
		}

		fmt.Print(respText)
		var roads []map[string]string
		roads = append(roads, map[string]string{"roadName": "五權路", "roadAvail": "10"})
		roads = append(roads, map[string]string{"roadName": "忠孝東路", "roadAvail": "50"})
		container := carouselmessage.Carouselmesage(roads)
		if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewFlexMessage("車位資訊。", container)).Do(); err != nil {
			log.Print(err)
		}
	}
}

//初始化 dialogflow (pointer receiver)
func (dp *dialogflowProcessor) init(data ...string) (err error) {
	dp.projectID = data[0]
	dp.authJSONFilePath = data[1]
	dp.lang = data[2]
	dp.timeZone = data[3]
	// Auth process: https://dialogflow.com/docs/reference/v2-auth-setup

	dp.ctx = context.Background()
	dp.sessionClient, err = dialogflow.NewSessionsClient(dp.ctx, option.WithCredentialsFile(dp.authJSONFilePath))

	return
}

func (ds *datastoreProcessor) init(data string) (err error) {
	ds.projectID = data
	ds.ctx = context.Background()
	ds.client, err = datastore.NewClient(ds.ctx, ds.projectID)
	return
}

//dialogflow 分析語意 (pointer receiver)
func (dp *dialogflowProcessor) processNLP(rawMessage string, username string) (r nlpResponse) {
	//DetectIntentRequest struct https://godoc.org/google.golang.org/genproto/googleapis/cloud/dialogflow/v2#StreamingDetectIntentRequest
	sessionID := username
	request := dialogflowpb.DetectIntentRequest{
		Session: fmt.Sprintf("projects/%s/agent/sessions/%s", dp.projectID, sessionID),
		QueryInput: &dialogflowpb.QueryInput{
			Input: &dialogflowpb.QueryInput_Text{
				Text: &dialogflowpb.TextInput{
					Text:         rawMessage,
					LanguageCode: dp.lang,
				},
			},
		},
		QueryParams: &dialogflowpb.QueryParameters{
			TimeZone: dp.timeZone,
		},
	}
	//DetectIntent https://godoc.org/cloud.google.com/go/dialogflow/apiv2#SessionsClient.DetectIntent
	response, err := dp.sessionClient.DetectIntent(dp.ctx, &request)
	if err != nil {
		log.Fatalf("Error in communication with Dialogflow %s", err.Error())
		return
	}
	queryResult := response.GetQueryResult()
	if queryResult.Intent != nil {
		//The name of this Intent
		r.Intent = queryResult.Intent.DisplayName
		//Values range from 0.0 (completely uncertain) to 1.0 (completely certain).
		// This value is for informational purpose only and is only used to
		// help match the best intent within the classification threshold.
		r.Confidence = float32(queryResult.IntentDetectionConfidence)

	}
	r.Entities = make(map[string]string)
	//The collection of extracted parameters.
	params := queryResult.Parameters.GetFields()
	if len(params) > 0 {
		for paramName, entity := range params {
			extractedValue := extractDialogflowEntities(entity)
			log.Printf("paramName= %s, entity= %s\n", paramName, extractedValue)
			r.Entities[paramName] = extractedValue
		}
	}
	return
}

// func (ds *datastoreProcessor) processDB()
// 解碼 Protobuf 格式
func extractDialogflowEntities(p *structpb.Value) (extractedEntity string) {
	kind := p.GetKind()
	switch kind.(type) {
	case *structpb.Value_StringValue:
		return p.GetStringValue()
	case *structpb.Value_NumberValue:
		return strconv.FormatFloat(p.GetNumberValue(), 'f', 6, 64)
	case *structpb.Value_BoolValue:
		return strconv.FormatBool(p.GetBoolValue())
	case *structpb.Value_StructValue:
		s := p.GetStructValue()
		fields := s.GetFields()
		extractedEntity = ""
		for key, value := range fields {
			log.Printf("key: %s, value: %s", key, value)
			// @TODO: Other entity types can be added here
		}
		return extractedEntity
	case *structpb.Value_ListValue:
		list := p.GetListValue()
		if len(list.GetValues()) > 1 {
			// @TODO: Extract more values
		}
		extractedEntity = extractDialogflowEntities(list.GetValues()[0])
		return extractedEntity
	default:
		return ""
	}
}

// getData  找車位資料
func getData(roadName string, intent string) (data string) {
	if roadName == "" {
		data = "哪一條路上的車位呢?"
		return
	}

	// ctx := context.Background()
	// projectID := "parkingproject-261207"
	// client, err := datastore.NewClient(ctx, projectID)
	// if err != nil {
	// 	log.Fatalf("Failed to create client: %v", err)
	// }

	// log.Printf("roadName: %s", roadName)

	//datastore 查詢路段資料
	query := datastore.NewQuery("Parkings").Filter("RoadSegName=", roadName)
	it := datastoreProc.client.Run(datastoreProc.ctx, query)
	for {
		var road road
		_, err := it.Next(&road)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error fetching road: %v", err)
		}
		fmt.Printf("RoadName %s, RoadSegAvail %s\n", road.RoadSegName, road.RoadSegAvail)

		data = road.RoadSegName + "有 " + road.RoadSegAvail + " 個車位"
	}
	return

}
