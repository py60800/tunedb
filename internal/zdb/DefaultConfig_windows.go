package zdb

import (
	_ "embed"
)

//go:embed config_windows.yml
var defaultConfiguration []byte
