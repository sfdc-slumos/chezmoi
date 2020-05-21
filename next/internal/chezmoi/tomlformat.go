package chezmoi

import (
	"bytes"

	"github.com/pelletier/go-toml"
)

type tomlFormat struct{}

// TOMLFormat is the TOML serialization format.
var TOMLFormat tomlFormat

func (tomlFormat) Decode(data []byte, value interface{}) error {
	return toml.NewDecoder(bytes.NewBuffer(data)).Decode(value)
}

func (tomlFormat) Name() string {
	return "toml"
}

func (tomlFormat) Marshal(value interface{}) ([]byte, error) {
	return toml.Marshal(value)
}

func (tomlFormat) Unmarshal(data []byte) (interface{}, error) {
	var result interface{}
	if err := toml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func init() {
	Formats[TOMLFormat.Name()] = TOMLFormat
}
