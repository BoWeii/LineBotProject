package query

import (
	"fmt"
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"

)

func getSpaceBubbleContents(space ParkingSpace) (component []linebot.FlexComponent) {
	component = []linebot.FlexComponent{
		&linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   space.RoadName,
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
					Text: strconv.Itoa(space.Avail) + " 個",
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
					Text: space.Day + "\n" + space.Hour,
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
					Text: space.Pay + "\n" + space.PayCash,
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
				Text: strconv.Itoa(int(space.Distance)) + " 公尺",
			},
		},
	}
	if space.Distance >= 0 {
		component = append(component, &linebot.BoxComponent{})
		copy(component[2:], component[1:])
		component[1] = distComp
	}
	return
}

func getLotBubbleContents(lot ParkingLot) (component []linebot.FlexComponent) {
	var avail = func() string {
		if lot.Type == 2 {
			return "暫無資料"
		} else {
			return strconv.Itoa(lot.Avail) + " 個"
		}
	}
	var fee = func() string {
		if lot.Pay == "" {
			return "暫無資料"
		} else {
			return lot.Pay
		}
	}
	component = []linebot.FlexComponent{
		&linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   lot.Name,
			Size:   linebot.FlexTextSizeTypeXl,
			Weight: linebot.FlexTextWeightTypeBold,
		},
		&linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeBaseline,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  "總車位數",
					Color: "#aaaaaa",
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: strconv.Itoa(lot.TotalCar) + " 個",
				},
			},
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
					Text: avail(),
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
					Text: lot.ServiceTime,
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
					Text: fee(),
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
				Text: strconv.Itoa(int(lot.Distance)) + " 公尺",
			},
		},
	}
	if lot.Distance >= 0 {
		component = append(component, &linebot.BoxComponent{})
		copy(component[2:], component[1:])
		component[1] = distComp
	}
	return
}
func getFeeBubbleContents(info FeeInfo) (component []linebot.FlexComponent) {
	component = []linebot.FlexComponent{
		&linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   string(info.AmountTicket) + "元",
			Size:   linebot.FlexTextSizeTypeXl,
			Weight: linebot.FlexTextWeightTypeBold,
		},
		&linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeBaseline,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  "停車日期",
					Color: "#aaaaaa",
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: info.Parkdt,
				},
			},
		},
		&linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeBaseline,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  "截止日期",
					Color: "#aaaaaa",
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: info.Paylim,
				},
			},
		},
		&linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeBaseline,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:  linebot.FlexComponentTypeText,
					Text:  "車牌",
					Color: "#aaaaaa",
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: info.CarID,
					Wrap: true,
				},
			},
		},
	}
	return
}
func createFeeInfoContainer(info FeeInfo) (container *linebot.BubbleContainer) {
	container = &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,

		Hero: &linebot.ImageComponent{
			Type: linebot.FlexComponentTypeImage,
			URL:  "https://upload.cc/i1/2020/08/16/eBihZU.png",
			Size: linebot.FlexImageSizeType3xl,
		},
		Styles: &linebot.BubbleStyle{
			Hero: &linebot.BlockStyle{
				BackgroundColor: "#00CCB2",
			},
		},
		Body: &linebot.BoxComponent{
			Type:     linebot.FlexComponentTypeBox,
			Layout:   linebot.FlexBoxLayoutTypeVertical,
			Spacing:  linebot.FlexComponentSpacingTypeSm,
			Contents: getFeeBubbleContents(info),
		},
		Size: linebot.FlexBubbleSizeTypeKilo,
	}

	return
}

func createSpaceBubbleContainer(space ParkingSpace, action string, route ...address) (container *linebot.BubbleContainer) {
	var uri string
	if len(route) != 0 {
		// &origin=" + route[0].Original +
		//https://www.google.com/maps/dir/?api=1&origin= &destination= &waypoints=
		uri = "https://www.google.com/maps/dir/?api=1&origin=" + route[0].Original + "&destination=" + route[0].Destination + "&waypoints=" + fmt.Sprintf("%f", space.Lat) + "," + fmt.Sprintf("%f", space.Lon)
		//println(uri)
	} else {
		uri = "https://www.google.com/maps/search/?api=1&query=" + fmt.Sprintf("%f", space.Lat) + "," + fmt.Sprintf("%f", space.Lon)
	}
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
			Contents: getSpaceBubbleContents(space),
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
						URI:   uri,
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
						Label: action,
						Data:  "action=" + action + "&roadID=" + space.RoadID,
					},
				},
			},
		},
		Size: linebot.FlexBubbleSizeTypeKilo,
	}

	return
}

func createLotBubbleContainer(lot ParkingLot, action string, route ...address) (container *linebot.BubbleContainer) {
	var uri string
	if len(route) != 0 {
		// &origin=" + route[0].Original +
		//https://www.google.com/maps/dir/?api=1&origin= &destination= &waypoints=
		uri = "https://www.google.com/maps/dir/?api=1&origin=" + route[0].Original + "&destination=" + route[0].Destination + "&waypoints=" + fmt.Sprintf("%f", lot.Lat) + "," + fmt.Sprintf("%f", lot.Lon)
		println(uri)
	} else {
		uri = "https://www.google.com/maps/search/?api=1&query=" + fmt.Sprintf("%f", lot.Lat) + "," + fmt.Sprintf("%f", lot.Lon)
	}
	container = &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,

		Hero: &linebot.ImageComponent{
			Type: linebot.FlexComponentTypeImage,
			URL:  "https://upload.cc/i1/2020/09/18/8A7iY5.png",
			Size: linebot.FlexImageSizeType3xl,
		},
		Styles: &linebot.BubbleStyle{
			Hero: &linebot.BlockStyle{
				BackgroundColor: "#7092BE",
			},
		},
		Body: &linebot.BoxComponent{
			Type:     linebot.FlexComponentTypeBox,
			Layout:   linebot.FlexBoxLayoutTypeVertical,
			Spacing:  linebot.FlexComponentSpacingTypeSm,
			Contents: getLotBubbleContents(lot),
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
						URI:   uri,
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
						Label: action,
						Data:  "action=" + action + "&lotID=" + lot.ID,
					},
				},
			},
		},
		Size: linebot.FlexBubbleSizeTypeKilo,
	}

	return
}

//CreateCarouselmesage 產生訊息
func CreateCarouselmesage(info interface{}) (container *linebot.CarouselContainer) {
	var bubbleConts []*linebot.BubbleContainer

	var action string

	switch info.(type) {
	case []ParkingSpace:
		spaces := info.([]ParkingSpace)
		if int(spaces[0].Distance) < 0 {
			action = "移除"
		} else {
			action = "加入最愛"
		}

		for _, space := range spaces {
			bubbleConts = append(bubbleConts, createSpaceBubbleContainer(space, action))
		}
	case []ParkingLot:
		lots := info.([]ParkingLot)
		if int(lots[0].Distance) < 0 {
			action = "移除"
		} else {
			action = "加入最愛"
		}

		for _, lot := range lots {
			bubbleConts = append(bubbleConts, createLotBubbleContainer(lot, action))
		}
	case []FeeInfo:
		feeInfos := info.([]FeeInfo)
		for _, feeInfo := range feeInfos {
			bubbleConts = append(bubbleConts, createFeeInfoContainer(feeInfo))
		}

	case RouteWithParkings:
		routeWithParkings := info.(RouteWithParkings)
		spaces := routeWithParkings.Spaces
		lots := routeWithParkings.Lots
		route := routeWithParkings.Address
		action = "加入最愛"

		for _, space := range spaces {
			bubbleConts = append(bubbleConts, createSpaceBubbleContainer(space, action, route))
		}
		for _, lot := range lots {
			bubbleConts = append(bubbleConts, createLotBubbleContainer(lot, action, route))
		}
	}

	container = &linebot.CarouselContainer{
		Type:     linebot.FlexContainerTypeCarousel,
		Contents: bubbleConts,
	}

	return
}

//IntroBubbleMsg 介紹訊息
func IntroBubbleMsg() (container *linebot.BubbleContainer) {
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
					Text: "只要傳送位置訊息給我，小幫手就會搜尋附近 500 公尺內有空位的路邊停車格，若有常停的路段，也可以加入至最愛哦！😘",
					Wrap: true,
				},
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: "按下開始使用，即刻體驗更多功能 👇🏻",
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
					Action: &linebot.PostbackAction{
						Label: "開始使用",
						Data:  "query",
					},
				},
			},
		},
		Size: linebot.FlexBubbleSizeTypeMega,
	}

	return
}

//SearchBubbleMsg 搜尋訊息
func SearchBubbleMsg() (container *linebot.BubbleContainer) {
	container = &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,

		Hero: &linebot.ImageComponent{
			Type: linebot.FlexComponentTypeImage,
			URL:  "https://upload.cc/i1/2020/08/16/8dsYrK.png",
			Size: linebot.FlexImageSizeType3xl,
		},
		Styles: &linebot.BubbleStyle{
			Hero: &linebot.BlockStyle{
				BackgroundColor: "#ECEE9F",
			},
		},
		Body: &linebot.BoxComponent{
			Type:    linebot.FlexComponentTypeBox,
			Layout:  linebot.FlexBoxLayoutTypeVertical,
			Spacing: linebot.FlexComponentSpacingTypeLg,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: "需要什麼協助呢？",
					Wrap: true,
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{

				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypeLink,
					Margin: linebot.FlexComponentMarginTypeSm,
					Height: linebot.FlexButtonHeightTypeSm,
					Flex:   linebot.IntPtr(2),
					Action: &linebot.URIAction{
						Label: "搜尋車位",
						URI:   "line://nv/location",
					},
				},
				&linebot.ButtonComponent{
					Type:  linebot.FlexComponentTypeButton,
					Style: linebot.FlexButtonStyleTypeLink,

					Height: linebot.FlexButtonHeightTypeSm,
					Margin: linebot.FlexComponentMarginTypeSm,
					Flex:   linebot.IntPtr(2),
					Action: &linebot.PostbackAction{
						Label:       "規劃路線",
						DisplayText: "規劃路線",
						Data:        "route",
					},
				},
				&linebot.ButtonComponent{
					Type:  linebot.FlexComponentTypeButton,
					Style: linebot.FlexButtonStyleTypeLink,

					Height: linebot.FlexButtonHeightTypeSm,
					Margin: linebot.FlexComponentMarginTypeSm,
					Flex:   linebot.IntPtr(2),
					Action: &linebot.PostbackAction{
						Label:       "查詢欠費",
						DisplayText: "查詢欠費",
						Data:        "fee",
					},
				},
			},
		},
		Size: linebot.FlexBubbleSizeTypeKilo,
	}

	return
}

//EmptyParkingBubbleMsg 查無車位訊息
func EmptyParkingBubbleMsg(route address) (container *linebot.BubbleContainer) {
	container = &linebot.BubbleContainer{
		Type: linebot.FlexContainerTypeBubble,

		Hero: &linebot.ImageComponent{
			Type: linebot.FlexComponentTypeImage,
			URL:  "https://upload.cc/i1/2020/08/11/AKUjRz.png",
			Size: linebot.FlexImageSizeType3xl,
		},
		Styles: &linebot.BubbleStyle{
			Hero: &linebot.BlockStyle{
				BackgroundColor: "#d7b082",
			},
		},
		Body: &linebot.BoxComponent{
			Type:    linebot.FlexComponentTypeBox,
			Layout:  linebot.FlexBoxLayoutTypeVertical,
			Spacing: linebot.FlexComponentSpacingTypeLg,
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type: linebot.FlexComponentTypeText,
					Text: "目的地附近沒有空車位哦 ，直接幫你導航好嗎😢",
					Wrap: true,
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   linebot.FlexComponentTypeBox,
			Layout: linebot.FlexBoxLayoutTypeVertical,
			Contents: []linebot.FlexComponent{

				&linebot.ButtonComponent{
					Type:   linebot.FlexComponentTypeButton,
					Style:  linebot.FlexButtonStyleTypeLink,
					Margin: linebot.FlexComponentMarginTypeSm,
					Height: linebot.FlexButtonHeightTypeSm,
					Flex:   linebot.IntPtr(2),
					Action: &linebot.URIAction{
						Label: "好",
						URI:   "https://www.google.com/maps/dir/?api=1&origin=" + route.Original + "&destination=" + route.Destination,
					},
				},
			},
		},
		Size: linebot.FlexBubbleSizeTypeKilo,
	}

	return
}

//"line://nv/location",
