package fulfillment

import (
	"context"
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"

	"cloud.google.com/go/datastore"
	// "github.com/tidwall/gjson"
	"google.golang.org/api/iterator"

	"github.com/line/line-bot-sdk-go/linebot"

	"github.com/golang/protobuf/ptypes/struct"
	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	"google.golang.org/api/option"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
)

var bot *linebot.Client
var err error

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

// DialogflowProcessor has all the information for connecting with Dialogflow
type DialogflowProcessor struct {
	projectID        string
	authJSONFilePath string
	lang             string
	timeZone         string
	sessionClient    *dialogflow.SessionsClient
	ctx              context.Context
}

// NLPResponse is webhook回應
type NLPResponse struct {
	Intent     string            `json:"intent"`
	Confidence float32           `json:"confidence"`
	Entities   map[string]string `json:"entities"`
}

var dp DialogflowProcessor

//response webhook回應
type response struct {
	FulfillmentText string `json:"fulfillmentText"`
}

//Fulfillment 查詢車位
func Fulfillment(w http.ResponseWriter, r *http.Request) {
	bot, err = linebot.New("6156c2512e6a30274dd536947bc6fe9b", "GiKIwKk+Lwku0WeGEGnlEDBDDGC67tQVCSIMbcQaKpA2IyZPU6OgVSIdI0h1HUUG2Ky/psNLEEkjfnEZGITnJolxlEScGgLoWT/iKpwyinf/IJDgeB5gnIB0zmuag0vYlcs7WgOYdUg0CwbGXlWKIwdB04t89/1O/w1cDnyilFU=")
	dp.init("parkingproject-261207", "parkingproject-261207-2933e4112308.json", "zh-TW", "Asia/Hong_Kong")
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				log.Println(reflect.TypeOf(message.Text))
				response := dp.processNLP(message.Text, "testUser")
				fmt.Printf("before reply====>%#v", response)
				text:=response.Intent
				fmt.Print("#######",text)
				fmt.Print("&&&",response.Entities["RoadName"])
				if text == "FindParking"{
					text = GetData(response.Entities["RoadName"],response.Intent)
				} else{
					text = "我聽不太懂"
				}	
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(text)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}

}

func (dp *DialogflowProcessor) init(data ...string) (err error) {
	dp.projectID = data[0]
	dp.authJSONFilePath = data[1]
	dp.lang = data[2]
	dp.timeZone = data[3]
	log.Printf("secret::::: %s",dp.authJSONFilePath)
	// Auth process: https://dialogflow.com/docs/reference/v2-auth-setup

	dp.ctx = context.Background()
	sessionClient, err := dialogflow.NewSessionsClient(dp.ctx, option.WithCredentialsFile(dp.authJSONFilePath))
	if err != nil {
		log.Fatal("Error in auth with Dialogflow===",err)
	}
	dp.sessionClient = sessionClient

	return
}

func (dp *DialogflowProcessor) processNLP(rawMessage string, username string) (r NLPResponse) {
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
	response, err := dp.sessionClient.DetectIntent(dp.ctx, &request)
	if err != nil {
		log.Fatalf("Error in communication with Dialogflow %s", err.Error())
		return
	}
	queryResult := response.GetQueryResult()
	if queryResult.Intent != nil {
		r.Intent = queryResult.Intent.DisplayName
		r.Confidence = float32(queryResult.IntentDetectionConfidence)
	}
	r.Entities = make(map[string]string)
	params := queryResult.Parameters.GetFields()
	fmt.Println("parmas=",params)
	if len(params) > 0 {
		for paramName, p := range params {
			fmt.Println("type of p ===",reflect.TypeOf(p))
			fmt.Println("Param %s", paramName)
			fmt.Println("two ==== %s", p.GetStringValue())
			fmt.Println("three ==== %s", p.String())
			extractedValue := extractDialogflowEntities(p)
			r.Entities[paramName] = extractedValue
		}
	}
	return
}

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
			if key == "amount" {
				extractedEntity = fmt.Sprintf("%s%s", extractedEntity, strconv.FormatFloat(value.GetNumberValue(), 'f', 6, 64))
			}
			if key == "unit" {
				extractedEntity = fmt.Sprintf("%s%s", extractedEntity, value.GetStringValue())
			}
			if key == "date_time" {
				extractedEntity = fmt.Sprintf("%s%s", extractedEntity, value.GetStringValue())
			}
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
// GetData  找車位資料
func GetData(roadName string,intent string) (resp string){
	if roadName == ""{
		resp="哪一條路上的車位呢?"
		return
	} 
	
	ctx := context.Background()
 	projectID := "parkingproject-261207"
 	client, err := datastore.NewClient(ctx, projectID)
 	if err != nil {
 		log.Fatalf("Failed to create client: %v", err)
 	}
	
 	//取得 roadName entity
 	// buf, _ := ioutil.ReadAll(r.Body)
 	// roadName := gjson.Get(string(buf), "events.0.message.text")
 	// roadName := gjson.Get(string(buf), "queryResult.parameters.RoadName")
 	log.Printf("roadName: %s", roadName)
	//datastore 查詢路段資料
 	query := datastore.NewQuery("Parkings").Filter("RoadSegName=", roadName)
 	it := client.Run(ctx, query)
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
		
		resp=road.RoadSegName + "有 " + road.RoadSegAvail + " 個車位"
 		// w.Header().Set("Content-Type", "application/json")
 		// response := response{
 		// 	FulfillmentText: road.RoadSegName + "有 " + road.RoadSegAvail + " 個車位",
 		// }
 		// json.NewEncoder(w).Encode(response)

	 }
	return
 	// defer r.Body.Close()
}
