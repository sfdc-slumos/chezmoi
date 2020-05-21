package chezmoi

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

type yamlFormat struct{}

// YAMLFormat is the YAML serialization format.
var YAMLFormat yamlFormat

func (yamlFormat) Decode(data []byte, value interface{}) error {
	return yaml.NewDecoder(bytes.NewBuffer(data)).Decode(value)
}

func (yamlFormat) Name() string {
	return "yaml"
}

func (yamlFormat) Marshal(value interface{}) ([]byte, error) {
	return yaml.Marshal(value)
}

func (yamlFormat) Unmarshal(data []byte) (interface{}, error) {
	var result interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func init() {
	Formats[YAMLFormat.Name()] = YAMLFormat
}
