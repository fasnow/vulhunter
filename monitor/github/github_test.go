package github

import (
	"cveHunter/proxy"
	"encoding/json"
	"testing"
)

func TestGithubMonitor_SearchCVEAll(t *testing.T) {
	monitor := GetSingleton()
	proxy.GetSingleton().Add(monitor)
	proxy.GetSingleton().SetProxy("http://username:password@127.0.0.1:8001")
	_, items, err := monitor.SearchCVEAll()
	if err != nil {
		t.Error(err)
		return
	}
	marshal, err := json.Marshal(items)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(marshal))
}
