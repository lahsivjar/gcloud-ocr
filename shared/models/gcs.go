package models

import (
	"bytes"
	"encoding/gob"
	"time"
)

// GCSEvent is the payload of a GCS event
type GCSEvent struct {
	Bucket         string    `json:"bucket"`
	Name           string    `json:"name"`
	Metageneration string    `json:"metageneration"`
	ResourceState  string    `json:"resourceState"`
	TimeCreated    time.Time `json:"timeCreated"`
	Updated        time.Time `json:"updated"`
}

// Encode encodes the GCSEvent in byte array
func (g *GCSEvent) Encode() (bytes.Buffer, error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)

	err := e.Encode(g)
	return b, err
}

// Decode encodes the GCSEvent from byte array
func (g *GCSEvent) Decode(b []byte) error {
	var bb bytes.Buffer
	bb.Write(b)

	d := gob.NewDecoder(&bb)
	return d.Decode(g)
}
