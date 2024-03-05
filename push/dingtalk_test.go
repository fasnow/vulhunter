package push

import (
	"cveHunter/db"
	proxy2 "cveHunter/proxy"
	"testing"
)

func TestGetSingleton(t *testing.T) {
	c := GetDingTalkSingleton()
	proxy2.GetSingleton().Add(c)
	proxy2.GetSingleton().SetProxy("http://127.0.0.1:8080")
	c.Push([]db.GithubCVE{
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
		{
			Name:        "CVE-2024-22752",
			Author:      "111",
			HtmlUrl:     "https://github.com/hacker625/CVE-2024-22752\n",
			Description: "Security Vulnerabilities of Software Programs and Web Applications",
		},
	}...)
}
