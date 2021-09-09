package pitaya

import (
	"github.com/google/uuid"
	"github.com/topfreegames/pitaya/v2/acceptor"
	"github.com/topfreegames/pitaya/v2/agent"
	"github.com/topfreegames/pitaya/v2/cluster"
	"github.com/topfreegames/pitaya/v2/config"
	"github.com/topfreegames/pitaya/v2/conn/codec"
	"github.com/topfreegames/pitaya/v2/conn/message"
	"github.com/topfreegames/pitaya/v2/defaultpipelines"
	"github.com/topfreegames/pitaya/v2/groups"
	"github.com/topfreegames/pitaya/v2/logger"
	"github.com/topfreegames/pitaya/v2/metrics"
	"github.com/topfreegames/pitaya/v2/metrics/models"
	"github.com/topfreegames/pitaya/v2/pipeline"
	"github.com/topfreegames/pitaya/v2/router"
	"github.com/topfreegames/pitaya/v2/serialize"
	"github.com/topfreegames/pitaya/v2/serialize/json"
	"github.com/topfreegames/pitaya/v2/service"
	"github.com/topfreegames/pitaya/v2/session"
	"github.com/topfreegames/pitaya/v2/worker"
)

// Builder holds dependency instances for a pitaya App
type Builder struct {
	acceptors        []acceptor.Acceptor
	Config           config.BuilderConfig
	DieChan          chan bool
	PacketDecoder    codec.PacketDecoder
	PacketEncoder    codec.PacketEncoder
	MessageEncoder   *message.MessagesEncoder
	Serializer       serialize.Serializer
	Router           *router.Router
	RPCClient        cluster.RPCClient
	RPCServer        cluster.RPCServer
	MetricsReporters []metrics.Reporter
	Server           *cluster.Server
	ServerMode       ServerMode
	ServiceDiscovery cluster.ServiceDiscovery
	Groups           groups.GroupService
	SessionPool      session.SessionPool
	Worker           *worker.Worker
	HandlerHooks     *pipeline.HandlerHooks
}

// PitayaBuilder Builder interface
type PitayaBuilder interface {
	Build() Pitaya
}

// NewBuilderWithConfigs return a builder instance with default dependency instances for a pitaya App
// with configs defined by a config file (config.Config) and default paths (see documentation).
func NewBuilderWithConfigs(
	isFrontend bool,
	serverType string,
	serverMode ServerMode,
	serverMetadata map[string]string,
	conf *config.Config,
) *Builder {
	builderConfig := config.NewBuilderConfig(conf)
	customMetrics := config.NewCustomMetricsSpec(conf)
	prometheusConfig := config.NewPrometheusConfig(conf)
	statsdConfig := config.NewStatsdConfig(conf)
	etcdSDConfig := config.NewEtcdServiceDiscoveryConfig(conf)
	natsRPCServerConfig := config.NewNatsRPCServerConfig(conf)
	natsRPCClientConfig := config.NewNatsRPCClientConfig(conf)
	workerConfig := config.NewWorkerConfig(conf)
	enqueueOpts := config.NewEnqueueOpts(conf)
	groupServiceConfig := config.NewMemoryGroupConfig(conf)
	return NewBuilder(
		isFrontend,
		serverType,
		serverMode,
		serverMetadata,
		*builderConfig,
		*customMetrics,
		*prometheusConfig,
		*statsdConfig,
		*etcdSDConfig,
		*natsRPCServerConfig,
		*natsRPCClientConfig,
		*workerConfig,
		*enqueueOpts,
		*groupServiceConfig,
	)
}

// NewDefaultBuilder return a builder instance with default dependency instances for a pitaya App,
// with default configs
// 返回一个默认的building
// 可以根据默认的配置创建一个基本的app实例
func NewDefaultBuilder(isFrontend bool, serverType string, serverMode ServerMode, serverMetadata map[string]string, builderConfig config.BuilderConfig) *Builder {
	customMetrics := config.NewDefaultCustomMetricsSpec()
	prometheusConfig := config.NewDefaultPrometheusConfig()
	statsdConfig := config.NewDefaultStatsdConfig()
	etcdSDConfig := config.NewDefaultEtcdServiceDiscoveryConfig()
	natsRPCServerConfig := config.NewDefaultNatsRPCServerConfig()
	natsRPCClientConfig := config.NewDefaultNatsRPCClientConfig()
	workerConfig := config.NewDefaultWorkerConfig()
	enqueueOpts := config.NewDefaultEnqueueOpts()
	groupServiceConfig := config.NewDefaultMemoryGroupConfig()
	return NewBuilder(
		isFrontend,
		serverType,
		serverMode,
		serverMetadata,
		builderConfig,
		*customMetrics,
		*prometheusConfig,
		*statsdConfig,
		*etcdSDConfig,
		*natsRPCServerConfig,
		*natsRPCClientConfig,
		*workerConfig,
		*enqueueOpts,
		*groupServiceConfig,
	)
}

// NewBuilder return a builder instance with default dependency instances for a pitaya App,
// with configs explicitly defined
// 根据基本的以来关系返回一个构建器
func NewBuilder(isFrontend bool, //是否前台服务
	serverType string, //服务器类型
	serverMode ServerMode, //服务模式
	serverMetadata map[string]string, //元数据
	config config.BuilderConfig, //构建器配置
	customMetrics models.CustomMetricsSpec, //自定义数据规范
	prometheusConfig config.PrometheusConfig, //普罗米修斯配置
	statsdConfig config.StatsdConfig, //启动配置
	etcdSDConfig config.EtcdServiceDiscoveryConfig, //服务器发现配置
	natsRPCServerConfig config.NatsRPCServerConfig, //远程rpc服务
	natsRPCClientConfig config.NatsRPCClientConfig, //远程rpc客户端
	workerConfig config.WorkerConfig, //worker配置
	enqueueOpts config.EnqueueOpts, //配置队列设置
	groupServiceConfig config.MemoryGroupConfig, //存储配置
) *Builder {
	server := cluster.NewServer(uuid.New().String(), serverType, isFrontend, serverMetadata)
	dieChan := make(chan bool)

	metricsReporters := []metrics.Reporter{}
	if config.Metrics.Prometheus.Enabled {
		metricsReporters = addDefaultPrometheus(prometheusConfig, customMetrics, metricsReporters, serverType)
	}

	if config.Metrics.Statsd.Enabled {
		metricsReporters = addDefaultStatsd(statsdConfig, metricsReporters, serverType)
	}

	handlerHooks := pipeline.NewHandlerHooks()
	if config.DefaultPipelines.StructValidation.Enabled {
		configureDefaultPipelines(handlerHooks)
	}

	sessionPool := session.NewSessionPool()

	var serviceDiscovery cluster.ServiceDiscovery
	var rpcServer cluster.RPCServer
	var rpcClient cluster.RPCClient
	if serverMode == Cluster {
		var err error
		serviceDiscovery, err = cluster.NewEtcdServiceDiscovery(etcdSDConfig, server, dieChan)
		if err != nil {
			logger.Log.Fatalf("error creating default cluster service discovery component: %s", err.Error())
		}

		rpcServer, err = cluster.NewNatsRPCServer(natsRPCServerConfig, server, metricsReporters, dieChan, sessionPool)
		if err != nil {
			logger.Log.Fatalf("error setting default cluster rpc server component: %s", err.Error())
		}

		rpcClient, err = cluster.NewNatsRPCClient(natsRPCClientConfig, server, metricsReporters, dieChan)
		if err != nil {
			logger.Log.Fatalf("error setting default cluster rpc client component: %s", err.Error())
		}
	}

	worker, err := worker.NewWorker(workerConfig, enqueueOpts)
	if err != nil {
		logger.Log.Fatalf("error creating default worker: %s", err.Error())
	}

	gsi := groups.NewMemoryGroupService(groupServiceConfig)
	if err != nil {
		panic(err)
	}

	return &Builder{
		acceptors:        []acceptor.Acceptor{},                                                  //接收器组
		Config:           config,                                                                 //配置
		DieChan:          dieChan,                                                                //
		PacketDecoder:    codec.NewPomeloPacketDecoder(),                                         //反序列化器
		PacketEncoder:    codec.NewPomeloPacketEncoder(),                                         //序列化器
		MessageEncoder:   message.NewMessagesEncoder(config.Pitaya.Handler.Messages.Compression), //消息序列化器
		Serializer:       json.NewSerializer(),                                                   //json序列化器
		Router:           router.New(),                                                           //路由
		RPCClient:        rpcClient,                                                              //远程rpc客户端
		RPCServer:        rpcServer,                                                              //远程rpc服务器
		MetricsReporters: metricsReporters,                                                       //周期调用
		Server:           server,                                                                 //服务器
		ServerMode:       serverMode,                                                             //服务器模式
		Groups:           gsi,                                                                    //组
		HandlerHooks:     handlerHooks,                                                           //句柄狗子
		ServiceDiscovery: serviceDiscovery,                                                       //服务器发现
		SessionPool:      sessionPool,                                                            //会话对象池
		Worker:           worker,                                                                 //workder
	}
}

// AddAcceptor adds a new acceptor to app
func (builder *Builder) AddAcceptor(ac acceptor.Acceptor) {
	if !builder.Server.Frontend {
		logger.Log.Error("tried to add an acceptor to a backend server, skipping")
		return
	}
	builder.acceptors = append(builder.acceptors, ac)
}

// Build returns a valid App instance
// 返回一个有效的app实例
// 创建handlerPool
// 设置路由 router
// 创建remotserver 根据是否standalong
// 创建代理工厂
// 创建handlerservice 服务

func (builder *Builder) Build() Pitaya {
	handlerPool := service.NewHandlerPool()
	var remoteService *service.RemoteService
	if builder.ServerMode == Standalone {
		if builder.ServiceDiscovery != nil || builder.RPCClient != nil || builder.RPCServer != nil {
			panic("Standalone mode can't have RPC or service discovery instances")
		}
	} else {
		if !(builder.ServiceDiscovery != nil && builder.RPCClient != nil && builder.RPCServer != nil) {
			panic("Cluster mode must have RPC and service discovery instances")
		}

		builder.Router.SetServiceDiscovery(builder.ServiceDiscovery)

		remoteService = service.NewRemoteService(
			builder.RPCClient,
			builder.RPCServer,
			builder.ServiceDiscovery,
			builder.PacketEncoder,
			builder.Serializer,
			builder.Router,
			builder.MessageEncoder,
			builder.Server,
			builder.SessionPool,
			builder.HandlerHooks,
			handlerPool,
		)

		builder.RPCServer.SetPitayaServer(remoteService)
	}

	agentFactory := agent.NewAgentFactory(builder.DieChan,
		builder.PacketDecoder,
		builder.PacketEncoder,
		builder.Serializer,
		builder.Config.Pitaya.Heartbeat.Interval,
		builder.MessageEncoder,
		builder.Config.Pitaya.Buffer.Agent.Messages,
		builder.SessionPool,
		builder.MetricsReporters,
	)

	handlerService := service.NewHandlerService(
		builder.PacketDecoder,
		builder.Serializer,
		builder.Config.Pitaya.Buffer.Handler.LocalProcess,
		builder.Config.Pitaya.Buffer.Handler.RemoteProcess,
		builder.Server,
		remoteService,
		agentFactory,
		builder.MetricsReporters,
		builder.HandlerHooks,
		handlerPool,
	)

	return NewApp(
		builder.ServerMode,       //模式
		builder.Serializer,       //序列化器
		builder.acceptors,        //接收器
		builder.DieChan,          //通道
		builder.Router,           //路由
		builder.Server,           //服务
		builder.RPCClient,        //rpc客户端
		builder.RPCServer,        //rpc服务端
		builder.Worker,           //worker
		builder.ServiceDiscovery, //服务发现
		remoteService,            //远程服务
		handlerService,           //句柄服务
		builder.Groups,           //组
		builder.SessionPool,      //会话对象池
		builder.MetricsReporters, //周期调用
		builder.Config.Pitaya,    //pitaya配置
	)
}

// NewDefaultApp returns a default pitaya app instance
// 返回一个默认的app实例
func NewDefaultApp(isFrontend bool, serverType string, serverMode ServerMode, serverMetadata map[string]string, config config.BuilderConfig) Pitaya {
	builder := NewDefaultBuilder(isFrontend, serverType, serverMode, serverMetadata, config)
	return builder.Build()
}

//配置默认管线
func configureDefaultPipelines(handlerHooks *pipeline.HandlerHooks) {
	handlerHooks.BeforeHandler.PushBack(defaultpipelines.StructValidatorInstance.Validate)
}

// 添加默认普罗米修斯
func addDefaultPrometheus(config config.PrometheusConfig, customMetrics models.CustomMetricsSpec, reporters []metrics.Reporter, serverType string) []metrics.Reporter {
	prometheus, err := CreatePrometheusReporter(serverType, config, customMetrics)
	if err != nil {
		logger.Log.Errorf("failed to start prometheus metrics reporter, skipping %v", err)
	} else {
		reporters = append(reporters, prometheus)
	}
	return reporters
}

// 添加一个默认的规则
func addDefaultStatsd(config config.StatsdConfig, reporters []metrics.Reporter, serverType string) []metrics.Reporter {
	statsd, err := CreateStatsdReporter(serverType, config)
	if err != nil {
		logger.Log.Errorf("failed to start statsd metrics reporter, skipping %v", err)
	} else {
		reporters = append(reporters, statsd)
	}
	return reporters
}
