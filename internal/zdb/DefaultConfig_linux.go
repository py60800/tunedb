package zdb

import (
	_ "embed"
)

//go:embed config.yml
var defaultConfiguration []byte
