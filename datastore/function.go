package datastore

import (
	"context"

	"project.com/datastore/parkings"
	"project.com/datastore/roadname"

)

//UpdateParkingInfo 更新新北市車位資訊
func UpdateParkingInfo(ctx context.Context, m parkings.PubSubMessage) {
	parkings.UpdateNTPCParkingLotsInfo(ctx, m)
	roadname.Update(ctx)
}
