package chezmoi

import (
	"bytes"
	"encoding/json"
	"strings"
)

type jsonFormat struct{}

// JSONFormat is the JSON serialization format.
var JSONFormat jsonFormat

func (jsonFormat) Decode(data []byte, value interface{}) error {
	return json.NewDecoder(bytes.NewBuffer(data)).Decode(value)
}

func (jsonFormat) Name() string {
	return "json"
}

func (jsonFormat) Marshal(value interface{}) ([]byte, error) {
	sb := &strings.Builder{}
	e := json.NewEncoder(sb)
	e.SetIndent("", "  ")
	if err := e.Encode(value); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}

func (jsonFormat) Unmarshal(data []byte) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func init() {
	Formats[JSONFormat.Name()] = JSONFormat
}
