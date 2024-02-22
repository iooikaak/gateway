package enum

type Apollo string

const (
	ApolloNamespaces    Apollo = "APOLLO_NAMESPACES"
	ApolloAppID         Apollo = "APOLLO_APP_ID"
	ApolloCluster       Apollo = "APOLLO_CLUSTER"
	ApolloClusterValue  Apollo = "Dev"
	ApolloMetaAddr      Apollo = "APOLLO_META_ADDR"
	ApolloMetaAddrValue Apollo = "CYRIL_DEVELOP"
	ApolloCacheDir      Apollo = "APOLLO_CACHE_DIR"
	ApolloCacheDirValue Apollo = "/tmp"
	ApolloAppName       Apollo = "TEST1.gateway.json"
)

func (a Apollo) String() string {
	return string(a)
}
