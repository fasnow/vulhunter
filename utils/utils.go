package utils

import (
	"fmt"
	"strings"
	"time"
)

func GetTimestamp() string {
	//golang规定 必须是2006-01-02 15:04
	return fmt.Sprintf("[%s]", time.Now().Format("2006-01-02 15:04"))
}

func ListTrimSpace(slice []string) []string {
	var newSlice = make([]string, 0)
	for _, s := range slice {
		if strings.TrimSpace(s) != "" {
			newSlice = append(newSlice, s)
		}
	}
	return newSlice
}
