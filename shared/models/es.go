package models

import (
	"bytes"
	"encoding/gob"
)

// ESPayload is the payload to send to es upload topic
type ESPayload struct {
	Event        GCSEvent `json:"event"`
	DetectedText []string `json:"detectedText"`
}

// Encode encodes the ESPayload in byte array
func (p *ESPayload) Encode() (bytes.Buffer, error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)

	err := e.Encode(p)
	return b, err
}

// Decode encodes the ESPayload from byte array
func (p *ESPayload) Decode(b []byte) error {
	var bb bytes.Buffer
	bb.Write(b)

	d := gob.NewDecoder(&bb)
	return d.Decode(p)
}
