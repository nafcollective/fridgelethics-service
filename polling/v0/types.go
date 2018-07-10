package main // import "github.com/nafcollective/fridgelethics-service/polling/v0"

import (
	"golang.org/x/oauth2"
)

type UserData struct {
	T *oauth2.Token
	D uint64
	C uint64
}
