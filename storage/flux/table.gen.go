// Generated by tmpl
// https://github.com/benbjohnson/tmpl
//
// DO NOT EDIT!
// Source: table.gen.go.tmpl

package storageflux

import (
	"sync"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/models"
	storage "github.com/influxdata/influxdb/v2/storage/reads"
	"github.com/influxdata/influxdb/v2/tsdb/cursors"
)

//
// *********** Float ***********
//

type floatTable struct {
	table
	mu    sync.Mutex
	cur   cursors.FloatArrayCursor
	alloc *memory.Allocator
}

func newFloatTable(
	done chan struct{},
	cur cursors.FloatArrayCursor,
	bounds execute.Bounds,
	key flux.GroupKey,
	cols []flux.ColMeta,
	tags models.Tags,
	defs [][]byte,
	cache *tagsCache,
	alloc *memory.Allocator,
) *floatTable {
	t := &floatTable{
		table: newTable(done, bounds, key, cols, defs, cache, alloc),
		cur:   cur,
	}
	t.readTags(tags)
	t.advance()

	return t
}

func (t *floatTable) Close() {
	t.mu.Lock()
	if t.cur != nil {
		t.cur.Close()
		t.cur = nil
	}
	t.mu.Unlock()
}

func (t *floatTable) Statistics() cursors.CursorStats {
	t.mu.Lock()
	defer t.mu.Unlock()
	cur := t.cur
	if cur == nil {
		return cursors.CursorStats{}
	}
	cs := cur.Stats()
	return cursors.CursorStats{
		ScannedValues: cs.ScannedValues,
		ScannedBytes:  cs.ScannedBytes,
	}
}

func (t *floatTable) Do(f func(flux.ColReader) error) error {
	return t.do(f, t.advance)
}

func (t *floatTable) advance() bool {
	a := t.cur.Next()
	l := a.Len()
	if l == 0 {
		return false
	}

	// Retrieve the buffer for the data to avoid allocating
	// additional slices. If the buffer is still being used
	// because the references were retained, then we will
	// allocate a new buffer.
	cr := t.allocateBuffer(l)
	cr.cols[timeColIdx] = arrow.NewInt(a.Timestamps, t.alloc)
	cr.cols[valueColIdx] = t.toArrowBuffer(a.Values)
	t.appendTags(cr)
	t.appendBounds(cr)
	return true
}

// group table

type floatGroupTable struct {
	table
	mu  sync.Mutex
	gc  storage.GroupCursor
	cur cursors.FloatArrayCursor
}

func newFloatGroupTable(
	done chan struct{},
	gc storage.GroupCursor,
	cur cursors.FloatArrayCursor,
	bounds execute.Bounds,
	key flux.GroupKey,
	cols []flux.ColMeta,
	tags models.Tags,
	defs [][]byte,
	cache *tagsCache,
	alloc *memory.Allocator,
) *floatGroupTable {
	t := &floatGroupTable{
		table: newTable(done, bounds, key, cols, defs, cache, alloc),
		gc:    gc,
		cur:   cur,
	}
	t.readTags(tags)
	t.advance()

	return t
}

func (t *floatGroupTable) Close() {
	t.mu.Lock()
	if t.cur != nil {
		t.cur.Close()
		t.cur = nil
	}
	if t.gc != nil {
		t.gc.Close()
		t.gc = nil
	}
	t.mu.Unlock()
}

func (t *floatGroupTable) Do(f func(flux.ColReader) error) error {
	return t.do(f, t.advance)
}

func (t *floatGroupTable) advance() bool {
RETRY:
	a := t.cur.Next()
	l := a.Len()
	if l == 0 {
		if t.advanceCursor() {
			goto RETRY
		}

		return false
	}

	// Retrieve the buffer for the data to avoid allocating
	// additional slices. If the buffer is still being used
	// because the references were retained, then we will
	// allocate a new buffer.
	cr := t.allocateBuffer(l)
	cr.cols[timeColIdx] = arrow.NewInt(a.Timestamps, t.alloc)
	cr.cols[valueColIdx] = t.toArrowBuffer(a.Values)
	t.appendTags(cr)
	t.appendBounds(cr)
	return true
}

func (t *floatGroupTable) advanceCursor() bool {
	t.cur.Close()
	t.cur = nil
	for t.gc.Next() {
		cur := t.gc.Cursor()
		if cur == nil {
			continue
		}

		if typedCur, ok := cur.(cursors.FloatArrayCursor); !ok {
			// TODO(sgc): error or skip?
			cur.Close()
			t.err = &influxdb.Error{
				Code: influxdb.EInvalid,
				Err: &GroupCursorError{
					typ:    "float",
					cursor: cur,
				},
			}
			return false
		} else {
			t.readTags(t.gc.Tags())
			t.cur = typedCur
			return true
		}
	}
	return false
}

func (t *floatGroupTable) Statistics() cursors.CursorStats {
	if t.cur == nil {
		return cursors.CursorStats{}
	}
	cs := t.cur.Stats()
	return cursors.CursorStats{
		ScannedValues: cs.ScannedValues,
		ScannedBytes:  cs.ScannedBytes,
	}
}

//
// *********** Integer ***********
//

type integerTable struct {
	table
	mu    sync.Mutex
	cur   cursors.IntegerArrayCursor
	alloc *memory.Allocator
}

func newIntegerTable(
	done chan struct{},
	cur cursors.IntegerArrayCursor,
	bounds execute.Bounds,
	key flux.GroupKey,
	cols []flux.ColMeta,
	tags models.Tags,
	defs [][]byte,
	cache *tagsCache,
	alloc *memory.Allocator,
) *integerTable {
	t := &integerTable{
		table: newTable(done, bounds, key, cols, defs, cache, alloc),
		cur:   cur,
	}
	t.readTags(tags)
	t.advance()

	return t
}

func (t *integerTable) Close() {
	t.mu.Lock()
	if t.cur != nil {
		t.cur.Close()
		t.cur = nil
	}
	t.mu.Unlock()
}

func (t *integerTable) Statistics() cursors.CursorStats {
	t.mu.Lock()
	defer t.mu.Unlock()
	cur := t.cur
	if cur == nil {
		return cursors.CursorStats{}
	}
	cs := cur.Stats()
	return cursors.CursorStats{
		ScannedValues: cs.ScannedValues,
		ScannedBytes:  cs.ScannedBytes,
	}
}

func (t *integerTable) Do(f func(flux.ColReader) error) error {
	return t.do(f, t.advance)
}

func (t *integerTable) advance() bool {
	a := t.cur.Next()
	l := a.Len()
	if l == 0 {
		return false
	}

	// Retrieve the buffer for the data to avoid allocating
	// additional slices. If the buffer is still being used
	// because the references were retained, then we will
	// allocate a new buffer.
	cr := t.allocateBuffer(l)
	cr.cols[timeColIdx] = arrow.NewInt(a.Timestamps, t.alloc)
	cr.cols[valueColIdx] = t.toArrowBuffer(a.Values)
	t.appendTags(cr)
	t.appendBounds(cr)
	return true
}

// group table

type integerGroupTable struct {
	table
	mu  sync.Mutex
	gc  storage.GroupCursor
	cur cursors.IntegerArrayCursor
}

func newIntegerGroupTable(
	done chan struct{},
	gc storage.GroupCursor,
	cur cursors.IntegerArrayCursor,
	bounds execute.Bounds,
	key flux.GroupKey,
	cols []flux.ColMeta,
	tags models.Tags,
	defs [][]byte,
	cache *tagsCache,
	alloc *memory.Allocator,
) *integerGroupTable {
	t := &integerGroupTable{
		table: newTable(done, bounds, key, cols, defs, cache, alloc),
		gc:    gc,
		cur:   cur,
	}
	t.readTags(tags)
	t.advance()

	return t
}

func (t *integerGroupTable) Close() {
	t.mu.Lock()
	if t.cur != nil {
		t.cur.Close()
		t.cur = nil
	}
	if t.gc != nil {
		t.gc.Close()
		t.gc = nil
	}
	t.mu.Unlock()
}

func (t *integerGroupTable) Do(f func(flux.ColReader) error) error {
	return t.do(f, t.advance)
}

func (t *integerGroupTable) advance() bool {
RETRY:
	a := t.cur.Next()
	l := a.Len()
	if l == 0 {
		if t.advanceCursor() {
			goto RETRY
		}

		return false
	}

	// Retrieve the buffer for the data to avoid allocating
	// additional slices. If the buffer is still being used
	// because the references were retained, then we will
	// allocate a new buffer.
	cr := t.allocateBuffer(l)
	cr.cols[timeColIdx] = arrow.NewInt(a.Timestamps, t.alloc)
	cr.cols[valueColIdx] = t.toArrowBuffer(a.Values)
	t.appendTags(cr)
	t.appendBounds(cr)
	return true
}

func (t *integerGroupTable) advanceCursor() bool {
	t.cur.Close()
	t.cur = nil
	for t.gc.Next() {
		cur := t.gc.Cursor()
		if cur == nil {
			continue
		}

		if typedCur, ok := cur.(cursors.IntegerArrayCursor); !ok {
			// TODO(sgc): error or skip?
			cur.Close()
			t.err = &influxdb.Error{
				Code: influxdb.EInvalid,
				Err: &GroupCursorError{
					typ:    "integer",
					cursor: cur,
				},
			}
			return false
		} else {
			t.readTags(t.gc.Tags())
			t.cur = typedCur
			return true
		}
	}
	return false
}

func (t *integerGroupTable) Statistics() cursors.CursorStats {
	if t.cur == nil {
		return cursors.CursorStats{}
	}
	cs := t.cur.Stats()
	return cursors.CursorStats{
		ScannedValues: cs.ScannedValues,
		ScannedBytes:  cs.ScannedBytes,
	}
}

//
// *********** Unsigned ***********
//

type unsignedTable struct {
	table
	mu    sync.Mutex
	cur   cursors.UnsignedArrayCursor
	alloc *memory.Allocator
}

func newUnsignedTable(
	done chan struct{},
	cur cursors.UnsignedArrayCursor,
	bounds execute.Bounds,
	key flux.GroupKey,
	cols []flux.ColMeta,
	tags models.Tags,
	defs [][]byte,
	cache *tagsCache,
	alloc *memory.Allocator,
) *unsignedTable {
	t := &unsignedTable{
		table: newTable(done, bounds, key, cols, defs, cache, alloc),
		cur:   cur,
	}
	t.readTags(tags)
	t.advance()

	return t
}

func (t *unsignedTable) Close() {
	t.mu.Lock()
	if t.cur != nil {
		t.cur.Close()
		t.cur = nil
	}
	t.mu.Unlock()
}

func (t *unsignedTable) Statistics() cursors.CursorStats {
	t.mu.Lock()
	defer t.mu.Unlock()
	cur := t.cur
	if cur == nil {
		return cursors.CursorStats{}
	}
	cs := cur.Stats()
	return cursors.CursorStats{
		ScannedValues: cs.ScannedValues,
		ScannedBytes:  cs.ScannedBytes,
	}
}

func (t *unsignedTable) Do(f func(flux.ColReader) error) error {
	return t.do(f, t.advance)
}

func (t *unsignedTable) advance() bool {
	a := t.cur.Next()
	l := a.Len()
	if l == 0 {
		return false
	}

	// Retrieve the buffer for the data to avoid allocating
	// additional slices. If the buffer is still being used
	// because the references were retained, then we will
	// allocate a new buffer.
	cr := t.allocateBuffer(l)
	cr.cols[timeColIdx] = arrow.NewInt(a.Timestamps, t.alloc)
	cr.cols[valueColIdx] = t.toArrowBuffer(a.Values)
	t.appendTags(cr)
	t.appendBounds(cr)
	return true
}

// group table

type unsignedGroupTable struct {
	table
	mu  sync.Mutex
	gc  storage.GroupCursor
	cur cursors.UnsignedArrayCursor
}

func newUnsignedGroupTable(
	done chan struct{},
	gc storage.GroupCursor,
	cur cursors.UnsignedArrayCursor,
	bounds execute.Bounds,
	key flux.GroupKey,
	cols []flux.ColMeta,
	tags models.Tags,
	defs [][]byte,
	cache *tagsCache,
	alloc *memory.Allocator,
) *unsignedGroupTable {
	t := &unsignedGroupTable{
		table: newTable(done, bounds, key, cols, defs, cache, alloc),
		gc:    gc,
		cur:   cur,
	}
	t.readTags(tags)
	t.advance()

	return t
}

func (t *unsignedGroupTable) Close() {
	t.mu.Lock()
	if t.cur != nil {
		t.cur.Close()
		t.cur = nil
	}
	if t.gc != nil {
		t.gc.Close()
		t.gc = nil
	}
	t.mu.Unlock()
}

func (t *unsignedGroupTable) Do(f func(flux.ColReader) error) error {
	return t.do(f, t.advance)
}

func (t *unsignedGroupTable) advance() bool {
RETRY:
	a := t.cur.Next()
	l := a.Len()
	if l == 0 {
		if t.advanceCursor() {
			goto RETRY
		}

		return false
	}

	// Retrieve the buffer for the data to avoid allocating
	// additional slices. If the buffer is still being used
	// because the references were retained, then we will
	// allocate a new buffer.
	cr := t.allocateBuffer(l)
	cr.cols[timeColIdx] = arrow.NewInt(a.Timestamps, t.alloc)
	cr.cols[valueColIdx] = t.toArrowBuffer(a.Values)
	t.appendTags(cr)
	t.appendBounds(cr)
	return true
}

func (t *unsignedGroupTable) advanceCursor() bool {
	t.cur.Close()
	t.cur = nil
	for t.gc.Next() {
		cur := t.gc.Cursor()
		if cur == nil {
			continue
		}

		if typedCur, ok := cur.(cursors.UnsignedArrayCursor); !ok {
			// TODO(sgc): error or skip?
			cur.Close()
			t.err = &influxdb.Error{
				Code: influxdb.EInvalid,
				Err: &GroupCursorError{
					typ:    "unsigned",
					cursor: cur,
				},
			}
			return false
		} else {
			t.readTags(t.gc.Tags())
			t.cur = typedCur
			return true
		}
	}
	return false
}

func (t *unsignedGroupTable) Statistics() cursors.CursorStats {
	if t.cur == nil {
		return cursors.CursorStats{}
	}
	cs := t.cur.Stats()
	return cursors.CursorStats{
		ScannedValues: cs.ScannedValues,
		ScannedBytes:  cs.ScannedBytes,
	}
}

//
// *********** String ***********
//

type stringTable struct {
	table
	mu    sync.Mutex
	cur   cursors.StringArrayCursor
	alloc *memory.Allocator
}

func newStringTable(
	done chan struct{},
	cur cursors.StringArrayCursor,
	bounds execute.Bounds,
	key flux.GroupKey,
	cols []flux.ColMeta,
	tags models.Tags,
	defs [][]byte,
	cache *tagsCache,
	alloc *memory.Allocator,
) *stringTable {
	t := &stringTable{
		table: newTable(done, bounds, key, cols, defs, cache, alloc),
		cur:   cur,
	}
	t.readTags(tags)
	t.advance()

	return t
}

func (t *stringTable) Close() {
	t.mu.Lock()
	if t.cur != nil {
		t.cur.Close()
		t.cur = nil
	}
	t.mu.Unlock()
}

func (t *stringTable) Statistics() cursors.CursorStats {
	t.mu.Lock()
	defer t.mu.Unlock()
	cur := t.cur
	if cur == nil {
		return cursors.CursorStats{}
	}
	cs := cur.Stats()
	return cursors.CursorStats{
		ScannedValues: cs.ScannedValues,
		ScannedBytes:  cs.ScannedBytes,
	}
}

func (t *stringTable) Do(f func(flux.ColReader) error) error {
	return t.do(f, t.advance)
}

func (t *stringTable) advance() bool {
	a := t.cur.Next()
	l := a.Len()
	if l == 0 {
		return false
	}

	// Retrieve the buffer for the data to avoid allocating
	// additional slices. If the buffer is still being used
	// because the references were retained, then we will
	// allocate a new buffer.
	cr := t.allocateBuffer(l)
	cr.cols[timeColIdx] = arrow.NewInt(a.Timestamps, t.alloc)
	cr.cols[valueColIdx] = t.toArrowBuffer(a.Values)
	t.appendTags(cr)
	t.appendBounds(cr)
	return true
}

// group table

type stringGroupTable struct {
	table
	mu  sync.Mutex
	gc  storage.GroupCursor
	cur cursors.StringArrayCursor
}

func newStringGroupTable(
	done chan struct{},
	gc storage.GroupCursor,
	cur cursors.StringArrayCursor,
	bounds execute.Bounds,
	key flux.GroupKey,
	cols []flux.ColMeta,
	tags models.Tags,
	defs [][]byte,
	cache *tagsCache,
	alloc *memory.Allocator,
) *stringGroupTable {
	t := &stringGroupTable{
		table: newTable(done, bounds, key, cols, defs, cache, alloc),
		gc:    gc,
		cur:   cur,
	}
	t.readTags(tags)
	t.advance()

	return t
}

func (t *stringGroupTable) Close() {
	t.mu.Lock()
	if t.cur != nil {
		t.cur.Close()
		t.cur = nil
	}
	if t.gc != nil {
		t.gc.Close()
		t.gc = nil
	}
	t.mu.Unlock()
}

func (t *stringGroupTable) Do(f func(flux.ColReader) error) error {
	return t.do(f, t.advance)
}

func (t *stringGroupTable) advance() bool {
RETRY:
	a := t.cur.Next()
	l := a.Len()
	if l == 0 {
		if t.advanceCursor() {
			goto RETRY
		}

		return false
	}

	// Retrieve the buffer for the data to avoid allocating
	// additional slices. If the buffer is still being used
	// because the references were retained, then we will
	// allocate a new buffer.
	cr := t.allocateBuffer(l)
	cr.cols[timeColIdx] = arrow.NewInt(a.Timestamps, t.alloc)
	cr.cols[valueColIdx] = t.toArrowBuffer(a.Values)
	t.appendTags(cr)
	t.appendBounds(cr)
	return true
}

func (t *stringGroupTable) advanceCursor() bool {
	t.cur.Close()
	t.cur = nil
	for t.gc.Next() {
		cur := t.gc.Cursor()
		if cur == nil {
			continue
		}

		if typedCur, ok := cur.(cursors.StringArrayCursor); !ok {
			// TODO(sgc): error or skip?
			cur.Close()
			t.err = &influxdb.Error{
				Code: influxdb.EInvalid,
				Err: &GroupCursorError{
					typ:    "string",
					cursor: cur,
				},
			}
			return false
		} else {
			t.readTags(t.gc.Tags())
			t.cur = typedCur
			return true
		}
	}
	return false
}

func (t *stringGroupTable) Statistics() cursors.CursorStats {
	if t.cur == nil {
		return cursors.CursorStats{}
	}
	cs := t.cur.Stats()
	return cursors.CursorStats{
		ScannedValues: cs.ScannedValues,
		ScannedBytes:  cs.ScannedBytes,
	}
}

//
// *********** Boolean ***********
//

type booleanTable struct {
	table
	mu    sync.Mutex
	cur   cursors.BooleanArrayCursor
	alloc *memory.Allocator
}

func newBooleanTable(
	done chan struct{},
	cur cursors.BooleanArrayCursor,
	bounds execute.Bounds,
	key flux.GroupKey,
	cols []flux.ColMeta,
	tags models.Tags,
	defs [][]byte,
	cache *tagsCache,
	alloc *memory.Allocator,
) *booleanTable {
	t := &booleanTable{
		table: newTable(done, bounds, key, cols, defs, cache, alloc),
		cur:   cur,
	}
	t.readTags(tags)
	t.advance()

	return t
}

func (t *booleanTable) Close() {
	t.mu.Lock()
	if t.cur != nil {
		t.cur.Close()
		t.cur = nil
	}
	t.mu.Unlock()
}

func (t *booleanTable) Statistics() cursors.CursorStats {
	t.mu.Lock()
	defer t.mu.Unlock()
	cur := t.cur
	if cur == nil {
		return cursors.CursorStats{}
	}
	cs := cur.Stats()
	return cursors.CursorStats{
		ScannedValues: cs.ScannedValues,
		ScannedBytes:  cs.ScannedBytes,
	}
}

func (t *booleanTable) Do(f func(flux.ColReader) error) error {
	return t.do(f, t.advance)
}

func (t *booleanTable) advance() bool {
	a := t.cur.Next()
	l := a.Len()
	if l == 0 {
		return false
	}

	// Retrieve the buffer for the data to avoid allocating
	// additional slices. If the buffer is still being used
	// because the references were retained, then we will
	// allocate a new buffer.
	cr := t.allocateBuffer(l)
	cr.cols[timeColIdx] = arrow.NewInt(a.Timestamps, t.alloc)
	cr.cols[valueColIdx] = t.toArrowBuffer(a.Values)
	t.appendTags(cr)
	t.appendBounds(cr)
	return true
}

// group table

type booleanGroupTable struct {
	table
	mu  sync.Mutex
	gc  storage.GroupCursor
	cur cursors.BooleanArrayCursor
}

func newBooleanGroupTable(
	done chan struct{},
	gc storage.GroupCursor,
	cur cursors.BooleanArrayCursor,
	bounds execute.Bounds,
	key flux.GroupKey,
	cols []flux.ColMeta,
	tags models.Tags,
	defs [][]byte,
	cache *tagsCache,
	alloc *memory.Allocator,
) *booleanGroupTable {
	t := &booleanGroupTable{
		table: newTable(done, bounds, key, cols, defs, cache, alloc),
		gc:    gc,
		cur:   cur,
	}
	t.readTags(tags)
	t.advance()

	return t
}

func (t *booleanGroupTable) Close() {
	t.mu.Lock()
	if t.cur != nil {
		t.cur.Close()
		t.cur = nil
	}
	if t.gc != nil {
		t.gc.Close()
		t.gc = nil
	}
	t.mu.Unlock()
}

func (t *booleanGroupTable) Do(f func(flux.ColReader) error) error {
	return t.do(f, t.advance)
}

func (t *booleanGroupTable) advance() bool {
RETRY:
	a := t.cur.Next()
	l := a.Len()
	if l == 0 {
		if t.advanceCursor() {
			goto RETRY
		}

		return false
	}

	// Retrieve the buffer for the data to avoid allocating
	// additional slices. If the buffer is still being used
	// because the references were retained, then we will
	// allocate a new buffer.
	cr := t.allocateBuffer(l)
	cr.cols[timeColIdx] = arrow.NewInt(a.Timestamps, t.alloc)
	cr.cols[valueColIdx] = t.toArrowBuffer(a.Values)
	t.appendTags(cr)
	t.appendBounds(cr)
	return true
}

func (t *booleanGroupTable) advanceCursor() bool {
	t.cur.Close()
	t.cur = nil
	for t.gc.Next() {
		cur := t.gc.Cursor()
		if cur == nil {
			continue
		}

		if typedCur, ok := cur.(cursors.BooleanArrayCursor); !ok {
			// TODO(sgc): error or skip?
			cur.Close()
			t.err = &influxdb.Error{
				Code: influxdb.EInvalid,
				Err: &GroupCursorError{
					typ:    "boolean",
					cursor: cur,
				},
			}
			return false
		} else {
			t.readTags(t.gc.Tags())
			t.cur = typedCur
			return true
		}
	}
	return false
}

func (t *booleanGroupTable) Statistics() cursors.CursorStats {
	if t.cur == nil {
		return cursors.CursorStats{}
	}
	cs := t.cur.Stats()
	return cursors.CursorStats{
		ScannedValues: cs.ScannedValues,
		ScannedBytes:  cs.ScannedBytes,
	}
}
