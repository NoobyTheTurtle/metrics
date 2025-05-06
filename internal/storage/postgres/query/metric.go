package query

import "database/sql"

type Metric struct {
	Key        string          `db:"key"`
	ValueFloat sql.NullFloat64 `db:"value_float"`
	ValueInt   sql.NullInt64   `db:"value_int"`
}
