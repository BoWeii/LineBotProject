package carouselmessage

import (
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"

)

func bubbleContainer(roadName string, avail string, num int) (container *linebot.BubbleContainer) {
	container = &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,
		Header: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   linebot.FlexComponentTypeText,
					Text:   "No " + strconv.Itoa(num),
					Size:   linebot.FlexTextSizeTypeXxl,
					Weight: linebot.FlexTextWeightTypeBold,
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: roadName,
					Size: linebot.FlexTextSizeTypeXl,
				},
			},
		},
		Styles: &linebot.BubbleStyle{
			Header: &linebot.BlockStyle{
				BackgroundColor: "#FF6B6E",
			},
		},
		Body: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: "剩餘 " + avail + " 個",
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:  linebot.FlexComponentTypeButton,
					Style: linebot.FlexButtonStyleTypeLink,
					Action: &linebot.URIAction{
						Label: "導航",
						URI:   "https://www.google.com/maps/search/?api=1&query=47.5951518,-122.3316393",
					},
				},
			},
		},
		Size: linebot.FlexBubbleSizeTypeNano,
	}

	return
}

//Carouselmesage 產生訊息
func Carouselmesage(roads []map[string]string) (container *linebot.CarouselContainer) {
	var bubbleConts []*linebot.BubbleContainer

	for i, info := range roads {
		bubbleConts = append(bubbleConts, bubbleContainer(info["roadName"], info["roadAvail"], i+1))
	}
	container = &linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: bubbleConts,
	}
	return
}
