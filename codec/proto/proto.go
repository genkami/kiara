// Package msgpack provides a Codec for Protocol Buffers
package proto

import (
	"fmt"
	"google.golang.org/protobuf/proto"

	"github.com/genkami/kiara/types"
)

type codec struct{}

// Codec is a `types.Codec` that converts data into Protocol Buffers.
var Codec types.Codec = &codec{}

func (c *codec) Marshal(v interface{}) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("expected proto.Message but got %T", v)
	}
	return proto.Marshal(msg)
}

func (c *codec) Unmarshal(src []byte, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("expected proto.Message but got %T", v)
	}
	return proto.Unmarshal(src, msg)
}
