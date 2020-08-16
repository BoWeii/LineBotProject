package query

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

//DialogflowProc  Dialogflow實體
var DialogflowProc dialogflowProcessor

// nlpResponse is webhook回應
type nlpResponse struct {
	Intent                   string
	Confidence               float32
	Entities                 map[string]string
	Prompts                  string
	Response                 string
	AllRequiredParamsPresent bool
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
func (dp *dialogflowProcessor) Init(data ...string) (err error) {
	dp.projectID = data[0]
	dp.authJSONFilePath = data[1]
	dp.lang = data[2]
	dp.timeZone = data[3]
	// Auth process: https://dialogflow.com/docs/reference/v2-auth-setup

	dp.ctx = context.Background()
	dp.sessionClient, err = dialogflow.NewSessionsClient(dp.ctx, option.WithCredentialsJSON([]byte(`{
  "type": "service_account",
  "project_id": "parkingproject-2-283415",
  "private_key_id": "d6156d24b1d3f038ee51cc6e7ce6db2d2e5b5c20",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDN4Jq2Z6CkTHUo\n5Tcd/OtCYC1pit83/q7TP6ozcV59WpAjZr1wkJA/9aAeAz8N2vOxPLwvwW0PFwE1\nDrl8HaQ5TUc+ZEFyDJUVWQk0oku/TyivE7/SwtZ75lhHEL0j9QeGoLoRb4+cwUfH\ncP3y64oek5lihJRfrHHVVpbyC1siLITesdTuziexPX5XMDqLk/aCSFVpq2tkj08M\nGScjIh3NXAoDv1tFr1xQqynkQLPSRAWhNiKTkp7R1y8JIJDm5QGKdUgrWP9tU/iU\nM9FxX5Vv8Nmzujaoa21VGyI+3Mgy8mnE85nWcJZsbFeV+MZejLbKOtPkhcGBmNhO\n/6CInnjxAgMBAAECggEADJCIiqkvhHVy2UeHbgvocPspDFncioifY269UMFKj66Q\nMUvo8/qrpg6j2m8O6SXfpWWIM+crdG4+TW5YD0FJG3LqHh584MSwMM7HtTvSMQNY\nD5rORELuN36w7KnirDQNJPgQUQX4qxu/6WXtxrZ9rgTqqZi/LClcY2Vy5gPfyImY\nVUDAAR0KUGaaL3FgHy5kgVnrckmfS/jI7skrDQLMxl+iktxFSP+Spn2SRy6T3pbJ\nykrUACPV9xbJQuym9hUMs4oaNI6z4DVWk6EsQbDjqhQXTDmXCLtOzgkr/PejRwP6\nvTm7C1U6ryK+qUez5TsdGpI+IuJWHBiAkEJd+pR7oQKBgQDx0HVkuCOxnLGxmhKN\ncVgFhrr2L5Mk+Af+g0zJ+rMIzZpF8noVflktYUsiT/GITn/+d6IWAawARG2xmAnN\nEYQTXmzJMsadhdFhw/NVk790qhONb5PnvzWFGBB7Gdv3VVx10o2MxWOsmUceEeNL\nQyU2PjF6IuMJ6z2Sm5lF1oM84QKBgQDZ9HLh0ropwhCkACtVRqcW+2OW9DqdyvFW\nRCukTM9wc3rNwKRrLzKmqxl+skjrDxtFQQ11uF3ILCX8foP4BMH0rszKqaS+gFMJ\nAB6h6ttxIRZUjfyRgvf87hjRBNcybGJssfewbF0z2u8/294SxWm6mIDZxTCMcb4U\n8LN4FNcuEQKBgQDxmLqRZMCUxd9reGos0x+EdfX53dJ/zyf9i6V+73FMzE7kr7x2\nGQR0KI7uuzywWO3cih5xKj51Deki1KqGLofs6hx6cLarz3VA3owR5koU/5AFcYMu\nuYV5cm+U7mMtHCYCudke2mAZpBK/4lNbcLyPE1hPlOeNk3CzN67NbeM1QQKBgH77\nIitbAEbv344M4zItlY+YKq953uSrpetikCKK9ZhIT1WsVJ51wwbDTHk6Ga2JAZRZ\nkCPzo//JaOAwPWa0LuQFKx8vsuGiFb56qV4gXHUOl9nvVyTXru9XMHImZdHkv3sg\nPHQ2zh42AYms6Tb6eNzTmM5HSj+ozNuaWJUvXyZhAoGBAONjZnTqg5+3ELbJbt57\nu9IVt9bhg+pba+HS0k0yllwWJYvsvHHWO2noRcXJ2/AA3pSZnbkgVwqFCujAaFEc\naka/Fh7rZ5dS6iHNh/01fxqqDjHor5ac95AvOvHEte04XSw/ghUDM3znpYV7WRYy\n5cEdW+zWYqQi5DYsVL1vOdl3\n-----END PRIVATE KEY-----\n",
  "client_email": "dialogflow@parkingproject-2-283415.iam.gserviceaccount.com",
  "client_id": "111257058616178367598",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/dialogflow%40parkingproject-2-283415.iam.gserviceaccount.com"
}`)))

	return
}

//dialogflow 分析語意 (pointer receiver)
func (dp *dialogflowProcessor) ProcessNLP(rawMessage string, username string) (r nlpResponse) {
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
		r.Response = queryResult.GetFulfillmentText()
		r.AllRequiredParamsPresent = queryResult.AllRequiredParamsPresent
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
		var extractedEntity string
		for _, value := range fields {
			//log.Printf("key: %s, value: %s", key, value)
			if value.String() != "" {
				extractedEntity += value.GetStringValue()
			}
		}
		//	extractedEntity := fields["street-address"].GetStringValue() //取得地址這欄
		return extractedEntity

	case *structpb.Value_ListValue:
		list := p.GetListValue()
		if len(list.GetValues()) == 1 {
			extractedEntity = extractDialogflowEntities(list.GetValues()[0])
		}

		return extractedEntity
	default:
		return ""
	}
}
