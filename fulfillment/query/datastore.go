package query

import (
	"context"

	"cloud.google.com/go/datastore"

)

//DatastoreProc datastore 實體
var DatastoreProc datastoreProcessor

type datastoreProcessor struct {
	projectID string
	client    *datastore.Client
	ctx       context.Context
}

func (ds *datastoreProcessor) Init(data string) (err error) {
	ds.projectID = data
	ds.ctx = context.Background()
	ds.client, err = datastore.NewClient(ds.ctx, ds.projectID)
	return
}
