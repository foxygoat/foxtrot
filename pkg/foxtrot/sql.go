package foxtrot

import _ "embed" //nolint:golint // allow blank import

//go:embed sql/schema.sql
var schema string

//go:embed sql/sample_data.sql
var sampleData string
