package logger

import (
	"cveHunter/utils"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Info(msg string) {
	logDir, _ := filepath.Abs(filepath.Join("log"))
	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".txt")

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		//先创建log目录
		_ = os.MkdirAll(logDir, 0644)
	}

	// 如果不存在则创建，如果存在则追加写入
	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err == nil {
		defer file.Close()
	}
	msg = fmt.Sprintf("%s %s\n", utils.GetTimestamp(), msg)
	_, err = file.WriteString(msg)
	fmt.Print(msg)
}
