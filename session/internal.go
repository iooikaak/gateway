package session

import (
	"net/http"

	"github.com/iooikaak/gateway/config"
)

/*
	内网访问认证方式，验证
	header：
	x-appid:1
	x-apptoken:123
*/
type Internal struct {
}

// NewInternal new internal
func NewInternal(cfg *config.Config) *Internal {
	return new(Internal)
}

// Do do auth
func (internal *Internal) Do(r *http.Request) error {
	return nil
}
