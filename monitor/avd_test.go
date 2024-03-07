package monitor

import (
	"cveHunter/proxy"
	"encoding/json"
	"testing"
)

func TestAliAVDMonitor_GetLatestAVDList(t *testing.T) {
	monitor := GetAVDMonitorSingleton()
	proxy.GetSingleton().Add(monitor)
	proxy.GetSingleton().SetProxy("http://127.0.0.1:8080")
	items, err := monitor.GetLatestAVDList()
	if err != nil {
		t.Error(err)
		return
	}
	for _, item := range items {
		t.Log(item)
	}
}

func TestAVDMonitor_GetAVDDetail(t *testing.T) {
	monitor := GetAVDMonitorSingleton()
	proxy.GetSingleton().Add(monitor)
	proxy.GetSingleton().SetProxy("http://127.0.0.1:8080")
	avd, err := monitor.GetAVDDetail("AVD-2024-27198")
	if err != nil {
		t.Error(err)
		return
	}
	marshal, err := json.Marshal(avd)
	if err != nil {
		return
	}
	t.Log(string(marshal))
}
