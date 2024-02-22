package session

import (
	"net/http"
)

// Authorization 认证
type Authorization interface {
	Do(r *http.Request) error
}
