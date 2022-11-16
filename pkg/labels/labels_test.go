package labels

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfigJson(t *testing.T) {
	validJsonStr := `{"services":[{"name":"remote","url":"http://bundle-server.nginx:80"}],"bundles":{"authz":{"service":"remote","resource":"authz.gz","persist":true,"polling":{"min_delay_seconds":30,"max_delay_seconds":60}}}}`
	invalidJsonStr := `sss`
	type JsonStr struct {
		str string
		res bool
	}
	data := []JsonStr{
		{
			str: validJsonStr,
			res: true,
		},
		{
			str: invalidJsonStr,
			res: false,
		},
	}
	for _, d := range data {
		res := json.Valid([]byte(d.str))
		if res != d.res {
			t.Errorf("expect %v,but get %v", d.res, res)
		}
	}
}

func TestUpdateConfig(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "b",
			Namespace: "a",
		},
	}

	jsonStr := `{"bundles":{"authz":{"persist":true,"polling":{"max_delay_seconds":60,"min_delay_seconds":30},"resource":"authz.gz","service":"remote"}},"labels":{"localtion":"beijing","aaa":"bbbb","tanjunchen":{"name":"tanjunchen","age":25}}}`
	jsonStrExpect := `{"bundles":{"authz":{"persist":true,"polling":{"max_delay_seconds":60,"min_delay_seconds":30},"resource":"authz.gz","service":"remote"}},"labels":{"pod_ns":"a","pod_name":"b","localtion":"beijing","aaa":"bbbb","tanjunchen":{"name":"tanjunchen","age":25}}}`
	withoutJsonStr := `{"bundles":{"authz":{"persist":true,"polling":{"max_delay_seconds":60,"min_delay_seconds":30},"resource":"authz.gz","service":"remote"}},"aaa":"aaa"}`
	withoutJsonStr2 := `{"bundles":{"authz":{"persist":true,"polling":{"max_delay_seconds":60,"min_delay_seconds":30},"resource":"authz.gz","service":"remote"}},"labels":{"pod_ns":"a","pod_name":"b"},"aaa":"aaa"}`
	type JsonStr struct {
		value  string
		expect string
	}

	data := []JsonStr{
		{
			value:  jsonStr,
			expect: jsonStrExpect,
		},
		{
			value:  withoutJsonStr,
			expect: withoutJsonStr2,
		},
	}
	s := NewService()
	for _, d := range data {
		updateValue, err := s.UpdateBundleConfig(pod, d.value)
		if err != nil {
			t.Errorf("updateBundleConfig err %v", err)
		}
		ok, err := assertEqualJSON(updateValue, d.expect)
		if !ok || err != nil {
			t.Errorf("expect %v,but get %v,err %v", d.expect, updateValue, err)
		}
	}
}

func assertEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}
