package patch

import (
	"fmt"
	"testing"
)

func TestPrint(t *testing.T) {
	value := "--diagnostic-addr=0.0.0.0:8282"
	target := fmt.Sprintf("--diagnostic-addr=0.0.0.0:%v", 8282)
	if value != target {
		t.Errorf("expect value %v,but get %v", value, target)
	}
}
