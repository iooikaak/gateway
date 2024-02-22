package config

import (
	"os"
	"testing"

	"github.com/iooikaak/frame/config/paladin/apollo"
	"github.com/iooikaak/frame/paladin"
)

var (
//confPath string
)

func TestApolloInit(t *testing.T) {

}

func TestGetAppllo(t *testing.T) {
	apolloAppName := "gateway"
	_ = os.Setenv("APOLLO_NAMESPACES", apolloAppName)
	if err := paladin.Init(apollo.PaladinDriverApollo); err != nil {
		t.Log(err.Error())
	}
	conf := Config{}
	if err := paladin.Get("gateway").UnmarshalYAML(&Conf); err != nil {
		t.Log(err.Error())
	}
	t.Log(conf)
}
