package server

import (
	"net/http"
)

type Session struct {
	Client       *http.Client
	TargetDomain string
}
