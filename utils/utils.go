package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"time"
)

func GetTimestamp() string {
	//golang规定 必须是2006-01-02 15:04
	return fmt.Sprintf("[%s]", time.Now().Format("2006-01-02 15:04"))
}

func HmacSHA256(secret, message string) string {
	h := hmac.New(sha256.New, []byte(secret))
	if _, err := io.WriteString(h, message); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
