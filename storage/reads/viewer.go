package reads

import (
	"context"

	"github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/storage"
	"github.com/influxdata/influxdb/v2/tsdb/cursors"
	"github.com/influxdata/influxql"
)

// Viewer is used by the store to query data from time-series files.
type Viewer interface {
	CreateCursorIterator(ctx context.Context) (cursors.CursorIterator, error)
	CreateSeriesCursor(ctx context.Context, orgID, bucketID influxdb.ID, cond influxql.Expr) (storage.SeriesCursor, error)
	TagKeys(ctx context.Context, orgID, bucketID influxdb.ID, start, end int64, predicate influxql.Expr) (cursors.StringIterator, error)
	TagValues(ctx context.Context, orgID, bucketID influxdb.ID, tagKey string, start, end int64, predicate influxql.Expr) (cursors.StringIterator, error)
}
