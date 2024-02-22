package enum

type Gateway string

const (
	ServiceName Gateway = "gateway"
	GatewayJson Gateway = "config/gateway.json"
)

func (g Gateway) String() string {
	return string(g)
}
