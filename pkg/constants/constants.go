package constants

const (
	// CsmOpaInjectionNamespace namespace 级别 sidecar 注入标签
	CsmOpaInjectionNamespace = "csmopa-injection"

	// CsmOpaInjectionAnnotation 工作负载是否注入 sidecar 注解
	CsmOpaInjectionAnnotation = "sidecar.csmopa.io/csmopa-injection"

	CsmOpaInjectionAnnotationFalse = "false"

	// CsmOpaInjectionWorkLoad workload 工作负载级别 sidecar 注入标签
	CsmOpaInjectionWorkLoad = "csmopa-injection/workload"

	// CsmOpaInjectionValue Pod 注入 opa sidecar 使用标签的 value 值
	CsmOpaInjectionValue = "enabled"

	// CsmOpaInjectionStatus Pod 注入 opa sidecar 的状态
	CsmOpaInjectionStatus = "csmopa-injection/status"

	// CsmOpaInjectValue Pod 注入 opa sidecar 状态的 value 值
	CsmOpaInjectValue = "injected"

	// Configmap type configmap
	Configmap = "configmap"

	// Bundle type bundle
	Bundle = "bundle"

	// OpaImageDefaultAddr opa 默认镜像地址
	OpaImageDefaultAddr = "openpolicyagent/opa:latest-istio"

	// OpaImageAddr opa 镜像地址标签
	OpaImageAddr = "sidecar.csmopa.io/opa_image_url"

	// RegoSourceType opa-sidecar 策略类型 configmap/bundle
	RegoSourceType = "sidecar.csmopa.io/rego_source_type"

	// RegoConfigMapName opa-sidecar 策略类型如果是 configmap，则使用的 configmap 名称
	RegoConfigMapName = "sidecar.csmopa.io/rego_source_cm"

	// RegoConfigMapNameKey opa-sidecar 策略类型如果是 configmap，则使用的 configmap 中的 key 名称
	RegoConfigMapNameKey = "sidecar.csmopa.io/rego_source_cm_key"

	// RegoConfigYamlGrpcPolicyPath opa-sidecar opa 认证的路径
	RegoConfigYamlGrpcPolicyPath = "sidecar.csmopa.io/grpc_policy_path"

	// RegoConfigYamlGrpcAddr opa-sidecar optional, default is ":9898"
	RegoConfigYamlGrpcAddr = "sidecar.csmopa.io/grpc_addr"

	// GrpcPolicyAddr policy address, default ":9898"
	GrpcPolicyAddr = ":9898"

	// RegoAddrPort optional, default is "9899"
	RegoAddrPort = "sidecar.csmopa.io/addr_port"

	// AddrPortDefault value
	AddrPortDefault = 9899

	// RegoDiagnosticAddrPort optional, default is "9900"
	RegoDiagnosticAddrPort = "sidecar.csmopa.io/diagnostic_addr_port"

	// DiagnosticAddrPortDefault value
	DiagnosticAddrPortDefault = 9900

	// RegoSourceBundleServiceUrl for bundle url, required
	RegoSourceBundleServiceUrl = "sidecar.csmopa.io/rego_source_bundle_service_url"

	// RegoSourceBundleResourcePath for bundle resource path, required
	RegoSourceBundleResourcePath = "sidecar.csmopa.io/rego_source_bundle_resource_path"

	// RegoSourceBundleServiceBearerToken for bundle BearerToken, when bundle_server need authentication,supply token by this, optional
	RegoSourceBundleServiceBearerToken = "sidecar.csmopa.io/rego_source_bundle_service_bearer_token"

	// RegoSourceBundleServiceSecretName for bundle service secret name
	RegoSourceBundleServiceSecretName = "sidecar.csmopa.io/rego_source_bundle_service_secret_name"

	// RegoSourceBundleServiceCertKeyNameInSecret for bundle service secret cert key name
	RegoSourceBundleServiceCertKeyNameInSecret = "sidecar.csmopa.io/rego_source_bundle_service_cert_key_name_in_secret"

	// RegoSourceBundleServicePrivateKeyKeyNameInSecret for bundle service secret private key name
	RegoSourceBundleServicePrivateKeyKeyNameInSecret = "sidecar.csmopa.io/rego_source_bundle_service_cert_key_name_in_secret"

	// RegoSourceBundlePollingMinDelaySeconds optional, default is 30
	RegoSourceBundlePollingMinDelaySeconds = "sidecar.csmopa.io/rego_source_bundle_polling_min_delay_seconds"

	// RegoSourceBundlePollingMinDelayDefaultSeconds optional, default is 30
	RegoSourceBundlePollingMinDelayDefaultSeconds = 30

	// RegoSourceBundlePollingMaxDelaySeconds  optional, default is 60
	RegoSourceBundlePollingMaxDelaySeconds = "sidecar.csmopa.io/rego_source_bundle_polling_max_delay_seconds"

	// RegoSourceBundlePollingMaxDelayDefaultSeconds optional, default is 60
	RegoSourceBundlePollingMaxDelayDefaultSeconds = 60

	// SidecarCsmOpaIoRegoSourceBundleConfig 必选值,opa 启动时的 config-file 文件内容
	SidecarCsmOpaIoRegoSourceBundleConfig = "sidecar.csmopa.io/rego_source_bundle_config"

	// SidecarCsmOpaIoRegoSourceBundleFiles 可选值，认证配置文件内容
	SidecarCsmOpaIoRegoSourceBundleFiles = "sidecar.csmopa.io/rego_source_bundle_files"
)

const (
	Labels  = "labels"
	PodNs   = "pod_ns"
	PodName = "pod_name"
)

var IgnoredNamespaces = []string{
	"kube-system",
	"kube-public",
}
