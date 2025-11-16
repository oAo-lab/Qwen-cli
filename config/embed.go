package config

import (
	_ "embed"
)

//go:embed default.json
var defaultConfigJSON []byte