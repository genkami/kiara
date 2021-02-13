// Package msgpack provides a Codec for JSON
package msgpack

import (
	"encoding/json"

	"github.com/genkami/kiara/types"
)

type codec struct{}

// Codec is a `types.Codec` that converts data into JSON.
var Codec types.Codec = &codec{}

func (c *codec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *codec) Unmarshal(src []byte, v interface{}) error {
	return json.Unmarshal(src, v)
}
