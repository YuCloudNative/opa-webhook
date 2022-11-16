package patch

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"opa-webhook/pkg/labels"

	"opa-webhook/pkg/constants"
)

const (
	containerName      = "csmopainjected-opa-istio"
	volumeName         = "csmopainjected-"
	imagePullPolicy    = "Always"
	policyBasePath     = "/policy"
	livenessProbePath  = "/health?plugins"
	readinessProbePath = "/health?plugins"
	command            = "/app/run.sh"
)

const (
	grpcAddrKey           = "grpc_addr"
	grpcPolicyPathKey     = "grpc_policy_path"
	addrPortKey           = "addr_port"
	diagnosticAddrPortKey = "diagnostic_addr_port"
	configKey             = "config"
	filesKey              = "files"
)

const (
	podNamespace = "POD_NAMESPACE"
	podName      = "POD_NAME"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// CreatePatch create mutation patch for resource
func CreatePatch(pod *corev1.Pod) ([]byte, error) {
	regoSourceType, _ := pod.Annotations[constants.RegoSourceType]
	if regoSourceType == constants.Bundle {
		return getPatchFromBundle(pod)
	} else if regoSourceType == constants.Configmap {
		return getPatchFromConfigMap(pod)
	}
	return nil, errors.New(fmt.Sprintf("%v %v is not supported", constants.RegoSourceType, regoSourceType))
}

func getPatchFromBundle(pod *corev1.Pod) ([]byte, error) {
	var patch []patchOperation
	var args []string

	opaImageAddr := pod.Annotations[constants.OpaImageAddr]

	regoConfigYamlGrpcPolicyPath := pod.Annotations[constants.RegoConfigYamlGrpcPolicyPath]

	regoConfigYamlGrpcAddr := pod.Annotations[constants.RegoConfigYamlGrpcAddr]
	if regoConfigYamlGrpcAddr == "" {
		regoConfigYamlGrpcAddr = constants.GrpcPolicyAddr
	}

	addrPort := pod.Annotations[constants.RegoAddrPort]
	if addrPort == "" {
		addrPort = strconv.Itoa(constants.AddrPortDefault)
	}

	diagnosticAddrPort := pod.Annotations[constants.RegoDiagnosticAddrPort]
	if diagnosticAddrPort == "" {
		diagnosticAddrPort = strconv.Itoa(constants.DiagnosticAddrPortDefault)
	}

	// 必选值
	configContent := pod.Annotations[constants.SidecarCsmOpaIoRegoSourceBundleConfig]
	service := labels.NewService()
	ok, err := service.CheckBundleConfigJson(configContent)
	if !ok || err != nil {
		return nil, err
	}

	// base64 encode
	encodedConfig := base64.StdEncoding.EncodeToString([]byte(configContent))

	// 给 bundle 增加 labels 标签
	//service := labels.NewService()
	//config, err := service.UpdateBundleConfig(pod, regoSourceBundleConfig)
	//if config == "" || err != nil {
	//	return nil, err
	//}

	// 可选值
	regoSourceBundleFile := pod.Annotations[constants.SidecarCsmOpaIoRegoSourceBundleFiles]
	// base64 encode
	encodedregoSourceBundleFile := base64.StdEncoding.EncodeToString([]byte(regoSourceBundleFile))

	glog.Infof("the bundle args are %s=%v,%s=%v,%s=%v,%s=%v,%s=%v,%s=%v,%s=%v", constants.OpaImageAddr, opaImageAddr, constants.RegoConfigYamlGrpcPolicyPath, regoConfigYamlGrpcPolicyPath,
		constants.RegoConfigYamlGrpcAddr, regoConfigYamlGrpcAddr, constants.RegoAddrPort, addrPort, constants.RegoDiagnosticAddrPort, diagnosticAddrPort,
		constants.SidecarCsmOpaIoRegoSourceBundleConfig, encodedConfig, constants.SidecarCsmOpaIoRegoSourceBundleFiles, encodedregoSourceBundleFile)

	envs := []corev1.EnvVar{
		{
			Name: podName,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: podNamespace,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
	}

	args = append(args, fmt.Sprintf("--%s", grpcAddrKey))
	args = append(args, fmt.Sprintf("%s", regoConfigYamlGrpcAddr))
	args = append(args, fmt.Sprintf("--%s", grpcPolicyPathKey))
	args = append(args, fmt.Sprintf("%s", regoConfigYamlGrpcPolicyPath))
	args = append(args, fmt.Sprintf("--%s", addrPortKey))
	args = append(args, fmt.Sprintf("%s", addrPort))
	args = append(args, fmt.Sprintf("--%s", diagnosticAddrPortKey))
	args = append(args, fmt.Sprintf("%s", diagnosticAddrPort))
	args = append(args, fmt.Sprintf("--%s", configKey))
	args = append(args, fmt.Sprintf("%s", encodedConfig))
	if len(encodedregoSourceBundleFile) > 0 {
		args = append(args, fmt.Sprintf("--%s", filesKey))
		args = append(args, fmt.Sprintf("%s", encodedregoSourceBundleFile))
	}

	cmd := []string{command}

	livenessProbe := corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: livenessProbePath,
				Port: intstr.Parse(diagnosticAddrPort),
			},
		},
	}

	readinessProbe := corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: readinessProbePath,
				Port: intstr.Parse(diagnosticAddrPort),
			},
		},
	}

	container := corev1.Container{
		Name:            containerName,
		Image:           opaImageAddr,
		Command:         cmd,
		Args:            args,
		Env:             envs,
		LivenessProbe:   &livenessProbe,
		ReadinessProbe:  &readinessProbe,
		ImagePullPolicy: imagePullPolicy,
	}
	containers := make([]corev1.Container, 0)
	containers = append(containers, container)
	patch = append(patch, addContainer(pod.Spec.Containers, containers, "/spec/containers")...)
	return json.Marshal(patch)
}

func getPatchFromConfigMap(pod *corev1.Pod) ([]byte, error) {
	var patch []patchOperation
	var args []string

	opaImageAddr := pod.Annotations[constants.OpaImageAddr]

	regoConfigMapName := pod.Annotations[constants.RegoConfigMapName]

	regoConfigMapNameKey := pod.Annotations[constants.RegoConfigMapNameKey]

	regoConfigYamlGrpcPolicyPath := pod.Annotations[constants.RegoConfigYamlGrpcPolicyPath]

	regoConfigYamlGrpcAddr := pod.Annotations[constants.RegoConfigYamlGrpcAddr]
	if regoConfigYamlGrpcAddr == "" {
		regoConfigYamlGrpcAddr = constants.GrpcPolicyAddr
	}

	addrPort := pod.Annotations[constants.RegoAddrPort]
	if addrPort == "" {
		addrPort = strconv.Itoa(constants.AddrPortDefault)
	}

	diagnosticAddrPort := pod.Annotations[constants.RegoDiagnosticAddrPort]
	if diagnosticAddrPort == "" {
		diagnosticAddrPort = strconv.Itoa(constants.DiagnosticAddrPortDefault)
	}

	glog.Infof("the configmap args are %s=%v,%s=%v,%s=%v,%s=%v,%s=%v,%s=%v", constants.OpaImageAddr, opaImageAddr, constants.RegoConfigMapName, regoConfigMapName,
		constants.RegoConfigYamlGrpcPolicyPath, regoConfigYamlGrpcPolicyPath, constants.RegoConfigYamlGrpcAddr, regoConfigYamlGrpcAddr,
		constants.RegoAddrPort, addrPort, constants.RegoDiagnosticAddrPort, diagnosticAddrPort)

	args = append(args, "run", "--server")
	args = append(args, "--set", fmt.Sprintf("plugins.envoy_ext_authz_grpc.addr=%v", regoConfigYamlGrpcAddr))
	args = append(args, "--set", fmt.Sprintf("plugins.envoy_ext_authz_grpc.path=%v", regoConfigYamlGrpcPolicyPath))
	args = append(args, "--set", "decision_logs.console=true")
	args = append(args, fmt.Sprintf("--addr=localhost:%s", addrPort))
	args = append(args, fmt.Sprintf("--diagnostic-addr=0.0.0.0:%s", diagnosticAddrPort))
	args = append(args, fmt.Sprintf("%v/%v", policyBasePath, regoConfigMapNameKey))

	livenessProbe := corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: livenessProbePath,
				Port: intstr.Parse(diagnosticAddrPort),
			},
		},
	}

	readinessProbe := corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: readinessProbePath,
				Port: intstr.Parse(diagnosticAddrPort),
			},
		},
	}

	volumeNameCm := volumeName + regoConfigMapName

	volumeMount := corev1.VolumeMount{
		Name:      volumeNameCm,
		MountPath: policyBasePath,
	}
	volumeMounts := make([]corev1.VolumeMount, 0)
	volumeMounts = append(volumeMounts, volumeMount)

	container := corev1.Container{
		Name:            containerName,
		Image:           opaImageAddr,
		Args:            args,
		VolumeMounts:    volumeMounts,
		LivenessProbe:   &livenessProbe,
		ReadinessProbe:  &readinessProbe,
		ImagePullPolicy: imagePullPolicy,
	}
	containers := make([]corev1.Container, 0)
	containers = append(containers, container)

	volume := corev1.Volume{
		Name: volumeNameCm,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: regoConfigMapName,
				},
			},
		},
	}

	volumes := make([]corev1.Volume, 0)
	volumes = append(volumes, volume)

	patch = append(patch, addContainer(pod.Spec.Containers, containers, "/spec/containers")...)
	patch = append(patch, addVolume(pod.Spec.Volumes, volumes, "/spec/volumes")...)

	return json.Marshal(patch)
}

func addContainer(target, added []corev1.Container, basePath string) (patch []patchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.Container{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}

func addVolume(target, added []corev1.Volume, basePath string) (patch []patchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.Volume{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}

func updateAnnotation(target map[string]string, added map[string]string, basePath string) (patch []patchOperation) {
	for key, value := range added {
		if target == nil || target[key] == "" {
			target = map[string]string{}
			patch = append(patch, patchOperation{
				Op:   "add",
				Path: basePath,
				Value: map[string]string{
					key: value,
				},
			})
		} else {
			patch = append(patch, patchOperation{
				Op:    "replace",
				Path:  basePath + key,
				Value: value,
			})
		}
	}
	return patch
}
