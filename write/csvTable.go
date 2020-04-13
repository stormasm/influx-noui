package write

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"unsafe"
)

const (
	labelFieldName   = "_field"
	labelFieldValue  = "_value"
	labelTime        = "_time"
	labelStart       = "_start"
	labelStop        = "_stop"
	labelMeasurement = "_measurement"
)

const (
	linePartIgnored = iota + 1
	linePartMeasurement
	linePartTag
	linePartField
	linePartTime
)

type annotationComment struct {
	label string
	flag  uint8
	setup func(column *CsvTableColumn, value string)
}

func ignoreLeadingComment(value string) string {
	if len(value) > 0 && value[0] == '#' {
		pos := strings.Index(value, " ")
		if pos > 0 {
			return strings.TrimLeft(value[pos+1:], " ")
		}
		return ""
	}
	return value
}

var supportedAnnotations = []annotationComment{
	{"#group", 1, func(column *CsvTableColumn, value string) {
		if strings.HasSuffix(value, "true") {
			column.LinePart = linePartTag
		}
	}},
	{"#datatype", 2, func(column *CsvTableColumn, value string) {
		val := ignoreLeadingComment(value)
		column.DataType = val
		// use extra data type values to identify line parts
		switch {
		case val == "tag":
			column.LinePart = linePartTag
		case strings.HasPrefix(val, "ignore"):
			column.LinePart = linePartIgnored
		case strings.HasPrefix(val, "dateTime"):
			// dateTime field shall be used only for time line part
			column.LinePart = linePartTime
		case val == "measurement":
			column.LinePart = linePartMeasurement
		case val == "field":
			column.LinePart = linePartField
			column.DataType = ""
		case val == "time": // time is an alias for dateTime
			column.LinePart = linePartTime
			column.DataType = dateTimeDatatype
		}
	}},
	{"#default", 4, func(column *CsvTableColumn, value string) {
		column.DefaultValue = ignoreLeadingComment(value)
	}},
}

// CsvTableColumn represents metadata of a csv column
type CsvTableColumn struct {
	// label such as "_start", "_stop", "_time"
	Label string
	// "string", "long", "dateTime:RFC3339" ...
	DataType string
	// column's line part (0 means not determined), see linePart constants
	LinePart int
	// default value to be used for rows where value is an empty string.
	DefaultValue string
	// index of this column in the table row
	Index int

	escapedLabel string
}

// LineLabel returns escaped column name so that it can be used as tag name or field name in line protocol
func (c *CsvTableColumn) LineLabel() string {
	if len(c.escapedLabel) > 0 {
		return c.escapedLabel
	}
	return c.Label
}

// CsvColumnError indicates conversion in a specific column
type CsvColumnError struct {
	Column string
	Err    error
}

func (e CsvColumnError) Error() string {
	return fmt.Sprintf("column '%s': %v", e.Column, e.Err)
}

// CsvTable gathers metadata about columns
type CsvTable struct {
	// all Table columns
	columns []CsvTableColumn
	// bitmap indicating presence of group, datatype and default comments
	partBits uint8
	// indicates that it is ready to read table data
	readTableData bool
	// indicated whether a table layout has changed
	indexed bool

	/* cached columns are initialized before reading the data rows */

	cachedMeasurement *CsvTableColumn
	cachedTime        *CsvTableColumn
	cachedFieldName   *CsvTableColumn
	cachedFieldValue  *CsvTableColumn
	cachedFields      []CsvTableColumn
	cachedTags        []CsvTableColumn
}

// AddRow adds header, comment or data row
func (t *CsvTable) AddRow(row []string) bool {
	// detect data row or table header row
	if len(row[0]) == 0 || row[0][0] != '#' {
		if !t.readTableData {
			if t.partBits == 0 {
				// create table if it does not exist yet
				t.columns = make([]CsvTableColumn, len(row))
				for i := 0; i < len(row); i++ {
					t.columns[i].Index = i
				}
			}
			for i := 0; i < len(t.columns); i++ {
				col := &t.columns[i]
				if len(col.Label) == 0 && col.Index < len(row) {
					col.Label = row[col.Index]
				}
			}
			t.readTableData = true
			return false
		}
		return true
	}
	// process supported anotation comments
	for i := 0; i < len(supportedAnnotations); i++ {
		supportedAnnotation := &supportedAnnotations[i]
		if strings.HasPrefix(strings.ToLower(row[0]), supportedAnnotation.label) {
			if len(row[0]) > len(supportedAnnotation.label) && row[0][len(supportedAnnotation.label)] != ' ' {
				continue // not a comment from the supported annotation
			}
			t.indexed = false
			t.readTableData = false
			// create new columns when data change
			if t.partBits == 0 || t.partBits&supportedAnnotation.flag == 1 {
				t.partBits = supportedAnnotation.flag
				t.columns = make([]CsvTableColumn, len(row))
				for i := 0; i < len(row); i++ {
					t.columns[i].Index = i
				}
			} else {
				t.partBits = t.partBits | supportedAnnotation.flag
			}
			for j := 0; j < len(t.columns); j++ {
				col := &t.columns[j]
				if col.Index >= len(row) {
					continue // missing value
				} else {
					supportedAnnotation.setup(col, row[col.Index])
				}
			}
			return false
		}
	}
	if !strings.HasPrefix(row[0], "# ") {
		log.Println("WARNING: unsupported annotation: ", row[0])
	}
	// comment row
	return false
}

func (t *CsvTable) computeIndexes() bool {
	if !t.indexed {
		t.recomputeIndexes()
		return true
	}
	return false
}

func (t *CsvTable) recomputeIndexes() {
	t.cachedMeasurement = nil
	t.cachedTime = nil
	t.cachedFieldName = nil
	t.cachedFieldValue = nil
	t.cachedTags = nil
	t.cachedFields = nil
	defaultIsField := t.Column(labelFieldName) == nil
	for i := 0; i < len(t.columns); i++ {
		col := t.columns[i]
		switch {
		case len(strings.TrimSpace(col.Label)) == 0 || col.LinePart == linePartIgnored:
		case col.Label == labelMeasurement || col.LinePart == linePartMeasurement:
			t.cachedMeasurement = &col
		case col.Label == labelTime || col.LinePart == linePartTime:
			if t.cachedTime != nil && t.cachedTime.Label != labelStart && t.cachedTime.Label != labelStop {
				log.Printf("WARNING: at most one dateTime column is expected, '%s' column is ignored\n", t.cachedTime.Label)
			}
			t.cachedTime = &col
		case col.Label == labelFieldName:
			t.cachedFieldName = &col
		case col.Label == labelFieldValue:
			t.cachedFieldValue = &col
		case col.LinePart == linePartTag:
			col.escapedLabel = escapeTag(col.Label)
			t.cachedTags = append(t.cachedTags, col)
		case col.LinePart == linePartField:
			col.escapedLabel = escapeTag(col.Label)
			t.cachedFields = append(t.cachedFields, col)
		default:
			if defaultIsField {
				col.escapedLabel = escapeTag(col.Label)
				t.cachedFields = append(t.cachedFields, col)
			}
		}
	}
	if t.cachedTags != nil && len(t.cachedTags) > 0 {
		sort.Slice(t.cachedTags, func(i, j int) bool {
			return t.cachedTags[i].Label < t.cachedTags[j].Label
		})
	}

	t.indexed = true
}

// CreateLine produces a protocol line out of the supplied row or returns error
func (t *CsvTable) CreateLine(row []string) (line string, err error) {
	buffer := make([]byte, 100)[:0]
	buffer, err = t.AppendLine(buffer, row)
	if err != nil {
		return "", err
	}
	return *(*string)(unsafe.Pointer(&buffer)), nil
}

// AppendLine appends a protocol line to the supplied buffer and returns appended buffer or an error if any
func (t *CsvTable) AppendLine(buffer []byte, row []string) ([]byte, error) {
	if t.computeIndexes() {
		// validate column data types
		if t.cachedFieldValue != nil && !IsTypeSupported(t.cachedFieldValue.DataType) {
			return buffer, CsvColumnError{
				t.cachedFieldValue.Label,
				fmt.Errorf("data type '%s' is not supported", t.cachedFieldValue.DataType),
			}
		}
		for _, c := range t.cachedFields {
			if !IsTypeSupported(c.DataType) {
				return buffer, CsvColumnError{
					c.Label,
					fmt.Errorf("data type '%s' is not supported", c.DataType),
				}
			}
		}
	}

	if t.cachedMeasurement == nil {
		return buffer, errors.New("no measurement column found")
	}
	measurement := orDefault(row[t.cachedMeasurement.Index], t.cachedMeasurement.DefaultValue)
	if measurement == "" {
		return buffer, CsvColumnError{
			t.cachedMeasurement.Label,
			errors.New("no measurement supplied"),
		}
	}
	buffer = append(buffer, escapeMeasurement(measurement)...)
	for _, tag := range t.cachedTags {
		value := orDefault(row[tag.Index], tag.DefaultValue)
		if tag.Index < len(row) && len(value) > 0 {
			buffer = append(buffer, ',')
			buffer = append(buffer, tag.LineLabel()...)
			buffer = append(buffer, '=')
			buffer = append(buffer, escapeTag(value)...)
		}
	}
	buffer = append(buffer, ' ')
	fieldAdded := false
	if t.cachedFieldName != nil && t.cachedFieldValue != nil {
		field := orDefault(row[t.cachedFieldName.Index], t.cachedFieldName.DefaultValue)
		value := orDefault(row[t.cachedFieldValue.Index], t.cachedFieldValue.DefaultValue)
		if len(value) > 0 && len(field) > 0 {
			buffer = append(buffer, escapeTag(field)...)
			buffer = append(buffer, '=')
			var err error
			buffer, err = appendConverted(buffer, value, t.cachedFieldValue.DataType)
			if err != nil {
				return buffer, CsvColumnError{
					t.cachedFieldName.Label,
					err,
				}
			}
			fieldAdded = true
		}
	}
	for _, field := range t.cachedFields {
		if field.Index < len(row) {
			value := orDefault(row[field.Index], field.DefaultValue)
			if len(value) > 0 {
				if !fieldAdded {
					fieldAdded = true
				} else {
					buffer = append(buffer, ',')
				}
				buffer = append(buffer, field.LineLabel()...)
				buffer = append(buffer, '=')
				var err error
				buffer, err = appendConverted(buffer, value, field.DataType)
				if err != nil {
					return buffer, CsvColumnError{
						field.Label,
						err,
					}
				}
			}
		}
	}
	if !fieldAdded {
		return buffer, errors.New("no field data found")
	}

	if t.cachedTime != nil && t.cachedTime.Index < len(row) {
		timeVal := orDefault(row[t.cachedTime.Index], t.cachedTime.DefaultValue)
		if len(timeVal) > 0 {
			var dataType = t.cachedTime.DataType
			if len(dataType) == 0 {
				// assume dateTime data type (number or RFC3339)
				dataType = dateTimeDatatype
			}
			buffer = append(buffer, ' ')
			var err error
			buffer, err = appendConverted(buffer, timeVal, dataType)
			if err != nil {
				return buffer, CsvColumnError{
					t.cachedTime.Label,
					err,
				}
			}
		}
	}
	return buffer, nil
}

func orDefault(val string, defaultValue string) string {
	if len(val) > 0 {
		return val
	}
	return defaultValue
}

// Column returns the first column of the supplied label or nil
func (t *CsvTable) Column(label string) *CsvTableColumn {
	for i := 0; i < len(t.columns); i++ {
		if t.columns[i].Label == label {
			return &t.columns[i]
		}
	}
	return nil
}

// Columns returns available columns
func (t *CsvTable) Columns() []CsvTableColumn {
	return t.columns
}

// Measurement returns measurement column or nil
func (t *CsvTable) Measurement() *CsvTableColumn {
	t.computeIndexes()
	return t.cachedMeasurement
}

// Time returns time column or nil
func (t *CsvTable) Time() *CsvTableColumn {
	t.computeIndexes()
	return t.cachedTime
}

// FieldName returns field name column or nil
func (t *CsvTable) FieldName() *CsvTableColumn {
	t.computeIndexes()
	return t.cachedFieldName
}

// FieldValue returns field value column or nil
func (t *CsvTable) FieldValue() *CsvTableColumn {
	t.computeIndexes()
	return t.cachedFieldValue
}

// Tags returns tags
func (t *CsvTable) Tags() []CsvTableColumn {
	t.computeIndexes()
	return t.cachedTags
}

// Fields returns fields
func (t *CsvTable) Fields() []CsvTableColumn {
	t.computeIndexes()
	return t.cachedFields
}
