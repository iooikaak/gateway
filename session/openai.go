package session

import (
	"net/http"

	"github.com/iooikaak/gateway/config"
)

// Openapi Openapi认证
type Openapi struct {
}

// NewOpenapi new openapi
func NewOpenapi(cfg *config.Config) Authorization {
	return new(Openapi)
}

// Do do auth
func (openapi *Openapi) Do(r *http.Request) error {
	return nil
}
