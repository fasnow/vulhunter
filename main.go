package main

import (
	"fmt"
	"strconv"
)

func main() {
	// HTML 中的 Unicode 字符
	htmlString := `Teamcity \u003c 2023.11.4`

	// 解码 HTML 中的转义字符
	unquotedString, err := strconv.Unquote(`"` + htmlString + `"`)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Unquoted string:", unquotedString)
}
