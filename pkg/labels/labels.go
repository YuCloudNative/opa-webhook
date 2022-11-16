package labels

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"

	"opa-webhook/pkg/constants"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

type Value struct {
	Labels map[string]interface{}
}

func (s *Service) CheckBundleConfigJson(sourceBundleConfig string) (bool, error) {
	glog.Infof("check if %s is valid json", sourceBundleConfig)
	res := json.Valid([]byte(sourceBundleConfig))
	if !res {
		return false, fmt.Errorf("invalid json %s", sourceBundleConfig)
	}
	glog.Infof("check json successful")
	return true, nil
}

func (s *Service) UpdateBundleConfig(pod *corev1.Pod, sourceBundleConfig string) (string, error) {
	ok, err := s.CheckBundleConfigJson(sourceBundleConfig)
	if !ok || err != nil {
		return "", err
	}

	var obj map[string]interface{}
	err = json.Unmarshal([]byte(sourceBundleConfig), &obj)
	if err != nil {
		return "", err
	}
	podNameSpace := pod.GetNamespace()
	podName := pod.GetName()
	glog.Infof("namespace ===>%s<=== name ===>%s<=== of pod", podNameSpace, podName)

	valueLabels := Value{}
	err = json.Unmarshal([]byte(sourceBundleConfig), &valueLabels)
	if err != nil {
		return "", err
	}
	if valueLabels.Labels == nil {
		valueLabels.Labels = make(map[string]interface{})
	}

	valueLabels.Labels[constants.PodNs] = podNameSpace
	valueLabels.Labels[constants.PodName] = podName
	obj[constants.Labels] = valueLabels.Labels

	jsonStr, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	glog.Infof("UpdateConfig namespace ===>%s<=== name ===>%s<=== of pod,source labels %v", podNameSpace, podName, string(jsonStr))
	return string(jsonStr), nil
}
