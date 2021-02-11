// Package msgpack provides a Codec for MessagePack
package msgpack

import (
	"github.com/vmihailenco/msgpack/v5"

	"github.com/genkami/kiara/types"
)

type codec struct{}

// Codec is a `types.Codec` that converts data into MessagePack.
var Codec types.Codec = &codec{}

func (c *codec) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (c *codec) Unmarshal(src []byte, v interface{}) error {
	return msgpack.Unmarshal(src, v)
}
