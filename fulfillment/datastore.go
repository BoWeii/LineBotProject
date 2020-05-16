package fulfillment

import (
	"context"

	"cloud.google.com/go/datastore"
)

var datastoreProc datastoreProcessor

type datastoreProcessor struct {
	projectID string
	client    *datastore.Client
	ctx       context.Context
}

func (ds *datastoreProcessor) init(data string) (err error) {
	ds.projectID = data
	ds.ctx = context.Background()
	ds.client, err = datastore.NewClient(ds.ctx, ds.projectID)
	return
}
