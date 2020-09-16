package datastore

import (
	"context"

	"project.com/datastore/parkings"

)

//UpdateParkingInfo 更新新北市車位資訊
func UpdateParkingInfo(ctx context.Context, m parkings.PubSubMessage) {
	parkings.UpdateNTPCParkingLotInfo(ctx, m)
}
