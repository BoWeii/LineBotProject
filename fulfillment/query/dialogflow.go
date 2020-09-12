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
		"project_id": "exalted-yeti-289303",
		"private_key_id": "753a4c88e472a80fbc4ca710f971cadf844c6394",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCLsRTDxp2mRbqL\ngfzqyyevM2G48pOAiQ8h+8Ag4DceO0M1XOwiNn3S9y6U2NQcn71L0kxCpD0QhdyX\nZFyMgLWoPaDYqfjUx6Wi+lnIvya29+ugVz/7ILnyAYpek46KxIL0D3ZyY3JHmB/b\nS6W+P5paG3wxd23NyqslG1Dw2jLKDI5Qx6H5jpCHgazw9A2QpEjxvAE+nXzTyf7C\n/BvHxMeWjOsOE1oN65+vi2kbW9/7H0qhnut5dZ5GJNgPVMVa0w63BT/nbRRUnv2V\nrkSQOFzoNQFRsg6vWdHQdhr8ojLGmrB8Goe7HlnkP8yUKSoSBGImuLuSgHiPQ5dR\nSo56NPNlAgMBAAECggEAHN+LNTQWXcnH1xIVHsrC9xmdt0acalLqB9IrUiBGBD+n\nkp4USLYOp85jmGyA2zHBRKzBCn08CfBBOiGKZ5gq9A1Y86+eOQzTFa69ZBrue58e\n2tNt7BBFntnmqcnCCri3gI65oscfVeEBpKgsJ/iphLptnyPaVFDxswfEYSQYn15O\nyq7bMAx4CMQIqSWSgNsxuiTPOo3+ihyV2r8tOwJZmSKB4sWOmmiJLceylfWiaCkz\nwEK3Li81WbKMVOPElfP/G6l1MLnH28IvBQdbGrmfNTQxIkCN+QvD6X6AWygBsM2o\nyGYG5qgCfmIMDqOfS63pqfCunYyLLJvKuR28XtBlQQKBgQDD+6i/MYAKtY/vN2mC\nUM6mz4YxgwvzbBfLgDN2qUROFgvJycKUv0V+1DcV+qvXHoqnvsYG/gBRfiQXHnj2\nkxHOzKjjdYevTYSIgiE/xvPgflak4M1lHFtWXz+PHOsFzFwzQ6k/z7HSjacv09S9\nJ7IbqAxuVHAuQ8oWwm3RhY1SGQKBgQC2eGRcBIA3BwxJGSndkDwXINbgfSRu+5zi\nH3/izPGNWu6b5Pa/8NCewMd8VUJg0ccsNpjMk9R7XjXvBrQXxZblljay9YZv7Aow\nmlQOTz6g8qFGSJHOjB60ffsOeC6smmcg/o9OGv5KXTNOw0i1pnPjsiVuWS4ZNl2V\nLZn/gYxNLQKBgEUhJuPSBRVB9/RselYSROKIPxCF5hhGH5qTrROFH2ff1OanuXAY\ni93x40BofGbUChqja1yaCei08uwIvqhTNivY+xXIpkTKth4ksK+7cNjWF5/u/+RT\nfVBZJqVHkQspi7g1fKTakSHw/3EaamcScnvY6hGczTk0hjtC99O5JSE5AoGAV1SL\ng9bLnoqLALlAZkBP4infbZW3SD64OSkmReAcg7DPnmEZD4gr8K8HSqRrnncIQkrn\nGpJuEZVnbrzgmLgCnmMkTsZfz1VDEzvpmuema9V0BnVZA2fgkjXxYF14yTckwI/U\n+mbE6cZtdfbU58uAj6uFaqjX/U0dwPYQTE8uXQkCgYBl030qm90GGRr9qUE94xPq\n0MtTgabfyujvzbrPxB1QO1/N4sBbK78iX4+5nXk1k6Uu4vf6YVi1ON/DhF7Ukkfn\n2GEN9pf4uWFKcaEaw7/yXXesLjza1dokaTfFJAqLN7JBpoJUh20QpEWeQpLKT+26\nCGqXTskGpYk33ReCO9junA==\n-----END PRIVATE KEY-----\n",
		"client_email": "dialogflow@exalted-yeti-289303.iam.gserviceaccount.com",
		"client_id": "109523558446951167395",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/dialogflow%40exalted-yeti-289303.iam.gserviceaccount.com"
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
