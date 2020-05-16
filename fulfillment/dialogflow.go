package fulfillment

import (
	"context"
	"fmt"
	"log"
	"strconv"

	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/api/option"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"

)

// nlpResponse is webhook回應
type nlpResponse struct {
	Intent     string
	Confidence float32
	Entities   map[string]string
	Prompts    string
}

type dialogflowProcessor struct {
	projectID        string
	authJSONFilePath string
	lang             string
	timeZone         string
	sessionClient    *dialogflow.SessionsClient
	ctx              context.Context
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
			extractedValue := extractDialogflowEntities(entity) //解析entities type
			log.Printf("paramName= %s, entity= %s\n", paramName, extractedValue)
			if extractedValue != "" {
				r.Entities[paramName] = extractedValue
			} else {
				r.Prompts = queryResult.GetFulfillmentText() //因entity為必要參數，若為空則取得prompts提示文字
			}
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
	case *structpb.Value_StructValue: //sys.location 為內建entity，回傳格式為struct，可以在dialogflow上輸入測試地址看完整結構
		s := p.GetStructValue()
		fields := s.GetFields()

		// for key, value := range fields {
		// 	log.Printf("key: %s, value: %s", key, value)
		// 	// @TODO: Other entity types can be added here
		// }
		extractedEntity := fields["street-address"].GetStringValue() //取得地址這欄
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
