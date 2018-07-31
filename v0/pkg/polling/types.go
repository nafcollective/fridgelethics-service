package polling

import (
	"bytes"
	"encoding/gob"

	"golang.org/x/oauth2"
)

type UserData struct {
	T oauth2.Token
	D uint64
	C uint64
}

func (ud UserData) ToBytes() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(ud)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ud *UserData) FromBytes(b []byte) error {
	buf := bytes.NewBuffer(b)
	err := gob.NewDecoder(buf).Decode(ud)
	if err != nil {
		return err
	}
	return nil
}
