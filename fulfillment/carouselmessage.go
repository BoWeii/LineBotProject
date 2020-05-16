package fulfillment

import (
	"fmt"
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"

)

func getBubbleInfo(parking parking) (component []linebot.FlexComponent) {
	component = []linebot.FlexComponent{

		&linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   parking.RoadName,
			Size:   linebot.FlexTextSizeTypeXl,
			Weight: linebot.FlexTextWeightTypeBold,
		},
		&linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeBaseline,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  "剩餘 ",
					Color: "#aaaaaa",
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: strconv.Itoa(parking.Avail) + " 個",
				},
			},
		},
		&linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeBaseline,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  "收費時段 ",
					Color: "#aaaaaa",
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: parking.Day + "\n" + parking.Hour,
					Wrap: true,
				},
			},
		},
		&linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeBaseline,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  "費率 ",
					Color: "#aaaaaa",
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: parking.Pay + "\n" + parking.PayCash,
					Wrap: true,
				},
			},
		},
	}

	distComp := &linebot.BoxComponent{
		Type:   linebot.FlexComponentTypeBox,
		Layout: linebot.FlexBoxLayoutTypeBaseline,
		Contents: []linebot.FlexComponent{
			&linebot.TextComponent{
				Type:  linebot.FlexComponentTypeText,
				Text:  "距離 ",
				Color: "#aaaaaa",
			},
			&linebot.TextComponent{
				Type: linebot.FlexComponentTypeText,
				Text: strconv.Itoa(int(parking.Distance)) + " 公尺",
			},
		},
	}
	if parking.Distance >= 0 {
		component = append(component, &linebot.BoxComponent{})
		copy(component[2:], component[1:])
		component[1] = distComp
	}
	return
}
func createBubbleContainer(parking parking, postBack string) (container *linebot.BubbleContainer) {
	container = &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,

		Hero: &linebot.ImageComponent{
			Type: linebot.FlexComponentTypeImage,
			URL:  "https://upload.cc/i1/2020/05/16/RMFJkO.png",
			Size: linebot.FlexImageSizeType3xl,
		},
		Styles: &linebot.BubbleStyle{
			Hero: &linebot.BlockStyle{
				BackgroundColor: "#a5dee5",
			},
		},
		Body: &linebot.BoxComponent{
			Type:     linebot.FlexComponentTypeBox,
			Layout:   linebot.FlexBoxLayoutTypeVertical,
			Spacing:  linebot.FlexComponentSpacingTypeSm,
			Contents: getBubbleInfo(parking),
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeHorizontal,
			Contents: []linebot.FlexComponent{

				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypePrimary,
					Color:  "#292b3b",
					Height: linebot.FlexButtonHeightTypeSm,
					Flex:   linebot.IntPtr(2),
					Action: &linebot.URIAction{
						Label: "導航",
						URI:   "https://www.google.com/maps/search/?api=1&query=" + fmt.Sprintf("%f", parking.Lat) + "," + fmt.Sprintf("%f", parking.Lon),
					},
				},
				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypeSecondary,
					Color:  "#ffc90e",
					Height: linebot.FlexButtonHeightTypeSm,
					Flex:   linebot.IntPtr(3),
					Margin: linebot.FlexComponentMarginTypeXl,
					Action: &linebot.PostbackAction{
						Label: postBack,
						Data:  "action=" + postBack + "&roadID=" + parking.RoadID,
					},
				},
			},
		},
		Size: linebot.FlexBubbleSizeTypeKilo,
	}

	return
}

//Carouselmesage 產生訊息
func queryCarouselmesage(roads []parking, action string) (container *linebot.CarouselContainer) {
	var bubbleConts []*linebot.BubbleContainer

	for _, info := range roads {
		bubbleConts = append(bubbleConts, createBubbleContainer(info, action))
	}
	container = &linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: bubbleConts,
	}
	return
}

func introBubbleMsg() (container *linebot.BubbleContainer) {
	container = &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,

		Hero: &linebot.ImageComponent{
			Type: linebot.FlexComponentTypeImage,
			URL:  "https://upload.cc/i1/2020/05/16/xmi5qs.png",
			Size: linebot.FlexImageSizeType3xl,
		},
		Styles: &linebot.BubbleStyle{
			Hero: &linebot.BlockStyle{
				BackgroundColor: "#324a5e",
			},
		},
		Body: &linebot.BoxComponent{
			Type:    linebot.FlexComponentTypeBox,
			Layout:  linebot.FlexBoxLayoutTypeVertical,
			Spacing: linebot.FlexComponentSpacingTypeLg,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: "只要傳送位置訊息給我，小幫手就會搜尋附近 1 公里內有空位的路邊停車格，若有常停的路段，也可以加入至最愛哦！😘",
					Wrap: true,
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: "按下開始使用，即刻體驗👇🏻",
					Wrap: true,
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeHorizontal,
			Contents: []linebot.FlexComponent{

				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypePrimary,
					Color:  "#292b3b",
					Height: linebot.FlexButtonHeightTypeSm,
					Flex:   linebot.IntPtr(2),
					Action: &linebot.URIAction{
						Label: "開始使用",
						URI:   "line://nv/location",
					},
				},
			},
		},
		Size: linebot.FlexBubbleSizeTypeMega,
	}

	return
}

//"line://nv/location",
