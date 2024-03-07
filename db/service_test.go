package db

import (
	"cveHunter/entry"
	"testing"
)

func TestGetAliAVDDbServiceSingleton(t *testing.T) {
	c := GetAVDDbServiceSingleton()
	err := c.Inset([]entry.AVD{
		{Name: "1"},
		{Name: "2"},
		{Name: "3"},
		{Name: "4"},
	}...)
	if err != nil {
		t.Error(err)
		return
	}
}
