package service

import (
	"encoding/json"
	"io"
)

func NewJsonDecoder(r io.Reader) *json.Decoder {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()

	return d
}
