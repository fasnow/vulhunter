package monitor

import (
	"cveHunter/entry"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type AVDMonitor struct {
	ProxyClient *http.Client
}

var (
	aliAVDMonitor     *AVDMonitor
	aliAVDMonitorOnce sync.Once
)

func GetAVDMonitorSingleton() *AVDMonitor {
	aliAVDMonitorOnce.Do(func() {
		aliAVDMonitor = &AVDMonitor{
			ProxyClient: &http.Client{},
		}
	})
	return aliAVDMonitor
}

func (r *AVDMonitor) GetLatestAVDList() ([]entry.AVD, error) {
	request, err := http.NewRequest("GET", "https://avd.aliyun.com/high-risk/list", nil)
	if err != nil {
		return nil, err
	}
	response, err := r.ProxyClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(string(bytes))
	}

	return getLatestAVDList(response.Body)
}

func (r *AVDMonitor) GetAVDDetail(id string) (*entry.AVD, error) {
	request, err := http.NewRequest("GET", "https://avd.aliyun.com/detail?id="+id, nil)
	if err != nil {
		return nil, err
	}
	response, err := r.ProxyClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(string(bytes))
	}
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}
	avd := &entry.AVD{}

	div := doc.Find("div.text-detail.pt-2.pb-4").First()
	t := div.Find("div")
	description, offset := getDetailDescription(t)
	avd.Description = description
	for i := offset; i < t.Size(); i++ {
		startLabel, text, endOffset := getDetailItem(i, t)
		i = endOffset - 1
		switch startLabel {
		case "影响范围", "影响版本":
			avd.ImpactVersion = strings.Join(text, ",")
		case "安全版本":
			break
		case "参考链接":
			break
		}
	}

	avd.Reference = strings.Join(getDetailReference(doc), "\n")
	return avd, nil
}

// getDetailItem 提取两个标签之间的值
func getDetailItem(start int, t *goquery.Selection) (startLabel string, text []string, endOffset int) {
	for i := start; i < t.Size(); i++ {
		tt, _ := strconv.Unquote(`"` + t.Eq(i).Text() + `"`)
		switch tt {
		case "漏洞描述", "影响范围", "安全版本", "参考链接", "影响版本":
			if startLabel == "" {
				startLabel = tt
				continue
			}
			if endOffset != 0 {
				continue
			}
			return startLabel, text, i
		default:
			if startLabel != "" {
				text = append(text, tt)
			}
		}
		if i == t.Size()-1 {
			endOffset = t.Size()
		}
	}
	return startLabel, text, endOffset
}

func getDetailDescription(t *goquery.Selection) (text string, endOffset int) {
	for i := 0; i < t.Length(); i++ {
		tt, _ := strconv.Unquote(`"` + t.Eq(i).Text() + `"`)
		switch tt {
		case "漏洞描述", "影响范围", "安全版本", "参考链接", "影响版本":
			return text, i
		default:
			text += tt
		}
		if i == t.Length()-1 {
			endOffset = i + 1
		}
	}
	return text, endOffset
}

func getLatestAVDList(body io.Reader) ([]entry.AVD, error) {

	// 使用 goquery 解析 HTML 字符串
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return nil, err
	}

	// 定义 AVD 结构体列表
	avdList := make([]entry.AVD, 0)

	// 遍历 <tr> 标签
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		var avd entry.AVD

		// 获取 <td> 标签的文本内容
		s.Find("td").Each(func(j int, td *goquery.Selection) {
			text := strings.TrimSpace(td.Text())
			switch j {
			case 0:
				avd.Number = text
			case 1:
				avd.Name = text
			case 2:
				avd.VulType = text
			case 3:
				avd.DisclosureDate = text
			}
		})

		// 获取 <button> 标签的 title 属性
		cveTitle := s.Find("td:last-child button:nth-child(1)").AttrOr("title", "")
		pocTitle := s.Find("td:last-child button:nth-child(2)").AttrOr("title", "")
		avd.CVE = cveTitle
		avd.POC = pocTitle

		avdList = append(avdList, avd)
	})
	return avdList, nil
}

func getDetailReference(doc *goquery.Document) []string {
	var list []string

	// 选择第一个 <table> 元素
	table := doc.Find("table.table.table-sm.table-responsive").First()

	tbody := table.Find("tbody")
	trList := tbody.Find("tr")
	trList.Each(func(i int, tr *goquery.Selection) {
		// 遍历每个 <a> 标签并输出其 href 属性的值
		tr.Find("a").Each(func(i int, link *goquery.Selection) {
			href, exists := link.Attr("href")
			if exists {
				list = append(list, strings.TrimSpace(href))
			}
		})
	})
	return list
}
