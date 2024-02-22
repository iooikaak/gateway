package enum

type Gateway string

const (
	ServiceName Gateway = "gateway"
	GatewayYaml Gateway = "config/gateway.yaml"
)

func (g Gateway) String() string {
	return string(g)
}
