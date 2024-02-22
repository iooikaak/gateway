package config

import (
	"flag"
	"io"
	"path/filepath"

	"github.com/iooikaak/gateway/model/enum"
	"gopkg.in/yaml.v3"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"

	//"github.com/apolloconfig/agollo/v4/component/log"
	//
	//"github.com/philchia/agollo/v4"
	//aCfg "github.com/apolloconfig/agollo/v4/env/config"
	"os"

	"github.com/iooikaak/frame/config"
	"github.com/iooikaak/frame/config/paladin"
	"github.com/iooikaak/frame/config/paladin/apollo"
)

var (
	Conf     = &Config{}
	confPath string
	M        paladin.YAML
)

// Config 对应配置文件中的格式定义
type Config struct {
	Authorization []string         `yaml:"authorization" json:"authorization"` // 认证方式，支持多个["app","internal","openapi"]
	IP            string           `yaml:"ip" json:"ip"`
	Port          string           `yaml:"port" json:"port"`
	Base          *config.BaseCfg  `yaml:"baseConfig" json:"baseConfig"`
	Redis         *config.RedisCfg `yaml:"redisConfig" json:"redisConfig"`
	//TIDB          *config.TIDBCfg     `yaml:"tidbConfig" json:"tidbConfig"`
	Mysql   *config.MysqlCfg    `yaml:"mysqlConfig" json:"mysqlConfig"`
	Env     *config.Environment `yaml:"env" json:"env"`
	AuthRaw string              `yaml:"authRaw" json:"authRaw"`
	Jaeger  string              `yaml:"jaeger" json:"jaeger"`
}

// Config struct
//
//	type Config2 struct {
//		Log         *xlog.Config                   `yaml:"log" json:"log"`
//		Tracer      *tracer.TracingConfig          `yaml:"tracer" json:"tracer"`
//		GinServer   *gins.Config                   `yaml:"ginServer" json:"ginServer"`
//		MicroServer *micro.Options                 `yaml:"microServer" json:"microServer"`
//		DB          *gorm.Config                   `yaml:"db" json:"db"`
//		Redis       *redis.Config                  `yaml:"redis" json:"redis"`
//		HttpClient  *bm.ClientConfig               `yaml:"httpClient" json:"httpClient"`
//		ServiceName string                         `yaml:"serviceName" json:"serviceName"`
//		Elastic     *elasticsearch.ElasticConfig   `yaml:"elastic" json:"elastic"`
//		RocketMq    *rocketmqConfig.RocketmqConfig `yaml:"rocketMq" json:"rocketMq"`
//		PxqUrl      string                         `yaml:"pxqUrl" json:"pxqUrl"`
//		FHLAdapter  *FHLAdapter                    `yaml:"fhlAdapter" json:"fhlAdapter"`
//	}
type FHLAdapter struct {
	AutoFetchVenue         bool             `yaml:"autoFetchVenue" json:"autoFetchVenue"`
	SmgRegionsConfig       []*RegionConfigs `yaml:"smgRegions" json:"smgRegions"`
	PostPicURL             string           `yaml:"postPicURL" json:"postPicURL"`
	SmgOpenAutoCreateVenue bool             `yaml:"SmgOpenAutoCreateVenue" json:"smgOpenAutoCreateVenue"`
}
type RegionConfigs struct {
	FloorId      int    `json:"floorId"`
	RegionId     string `json:"regionId"`
	XSize        int    `json:"XSize"`
	YSize        int    `json:"YSize"`
	RegionRow    int    `json:"regionRow"`
	RegionColumn int    `json:"regionColumn"`
}

func init() {
	flag.StringVar(&confPath, "Conf", "", "conf values")
	//Init()
}

func Init() (err error) {
	var (
		yamlFile string
	)
	if confPath != "" {
		yamlFile, err = filepath.Abs(confPath)
	} else {
		yamlFile, err = filepath.Abs(enum.GatewayYaml.String())
	}
	if err != nil {
		return
	}
	yamlRead, err := os.ReadFile(yamlFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(yamlRead, Conf)
	if err != nil {
		return
	}
	return
}

// ApolloInit 配置初始化，依赖本地环境变量
func ApolloInit() error {

	_ = os.Setenv(enum.ApolloNamespaces.String(), enum.ApolloAppName.String())
	_ = os.Setenv(enum.ApolloAppID.String(), enum.ServiceName.String())
	_ = os.Setenv(enum.ApolloCluster.String(), enum.ApolloClusterValue.String())
	_ = os.Setenv(enum.ApolloMetaAddr.String(), os.Getenv(enum.ApolloMetaAddrValue.String()))
	_ = os.Setenv(enum.ApolloCacheDir.String(), enum.ApolloCacheDirValue.String())
	if err := paladin.Init(apollo.PaladinDriverApollo); err != nil {
		panic(err)
	}
	if err := paladin.Get(enum.ApolloAppName.String()).UnmarshalJSON(&Conf); err != nil {
		panic(err)
	}
	if err := paladin.Watch(enum.ApolloAppName.String(), &M); err != nil {
		panic(err)
	}
	return nil
}

func CreateTracer(servieName string) (opentracing.Tracer, io.Closer, error) {
	var cfg = jaegercfg.Configuration{
		ServiceName: servieName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: enum.JaegerSampleConfigParam.Float64(),
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
			// 按实际情况替换你的 ip
			CollectorEndpoint: Conf.Jaeger,
		},
	}
	jLogger := jaegerlog.StdLogger
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
	)
	return tracer, closer, err
}
