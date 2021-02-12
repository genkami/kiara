// Package gob provides a Codec for gob.
package gob

import (
	"bytes"
	"encoding/gob"

	"github.com/genkami/kiara/types"
)

type codec struct{}

// Codec is a `types.Codec` that converts data into gob.
var Codec types.Codec = &codec{}

func (c *codec) Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *codec) Unmarshal(src []byte, v interface{}) error {
	dec := gob.NewDecoder(bytes.NewReader(src))
	return dec.Decode(v)
}
