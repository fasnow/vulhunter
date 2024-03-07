package main

import (
	"cveHunter/config"
	"cveHunter/db"
	"cveHunter/entry"
	"cveHunter/logger"
	"cveHunter/monitor"
	"cveHunter/push"
	"cveHunter/utils"
	"fmt"
	"strings"
	"time"
)

func RunGithubMonitor() {
	githubMonitor := monitor.GetGithubSingleton()
	_, items, err := githubMonitor.SearchCVEAll()
	if err != nil {
		logger.Info(err.Error())
		return
	}
	var pushCVEList []db.GithubCVE
	for _, item := range items {
		authors := db.GetServiceSingleton().FindAuthorsByCVE(item.Name)
		authorRecorded := false
		for _, author := range authors {
			if author == item.Owner.Login {
				authorRecorded = true
				break
			}
		}
		if len(authors) < config.GetSingleton().Github.MaxAuthorNumPerCve || authorRecorded {
			cves, err := db.GetServiceSingleton().FindAllByAuthorAndCVE(item.Owner.Login, item.Name)
			if err != nil {
				logger.Info(err.Error())
				continue
			}

			//对于同一作者
			if len(cves) >= config.GetSingleton().Github.MaxRecordNumPerAuthor {
				//此纪录对于同一作者的收录达到上限
				logger.Info(fmt.Sprintf("for %s,the number of %s's records is at its maximum, ignore this record", item.Name, item.Owner.Login))
				continue
			} else {
				recorded := false
				for _, cve := range cves {
					if cve.HtmlUrl == item.HTMLURL {
						recorded = true
						break
					}
				}
				if recorded {
					//已收录则跳过
					continue
				}
				logger.Info(fmt.Sprintf("insert record [%s/%s] %s", item.Owner.Login, item.Name, item.HTMLURL))
				t := db.GithubCVE{
					Name:        item.Name,
					HtmlUrl:     item.HTMLURL,
					Description: item.Description,
					Author:      item.Owner.Login,
				}
				if err = db.GetServiceSingleton().Insert(t); err != nil {
					logger.Info(fmt.Sprintf("insert record [%s/%s] %s error: %s", item.Owner.Login, item.Name, item.HTMLURL, err.Error()))
				}
				pushCVEList = append(pushCVEList, t)
			}
		}
	}

	if len(pushCVEList) == 0 {
		return
	}
	title := "Github漏洞推送"
	if config.GetSingleton().DingTalk.Enable {
		if config.GetSingleton().DingTalk.WebHookAccessToken == "" || config.GetSingleton().DingTalk.WebHookSecret == "" {
			logger.Info("didn't configure dingtalk webhook_access_token or webhook_secret, skip push")
			return
		}
		go func() {
			retry := 5
			var is []string
			for _, cve := range pushCVEList {
				is = append(is,
					fmt.Sprintf("**漏洞编号**:  %s  \n**项目地址**:  %s  \n**漏洞描述**:  %s  \n", cve.Name, cve.HtmlUrl, cve.Description))
			}
			for {
				if retry <= 0 {
					break
				}
				code, err := push.GetDingTalkSingleton().Push(title, strings.Join(is, "  \n------  \n"))
				if code < 0 {
					logger.Info(err.Error())
					retry--
					time.Sleep(3 * time.Second)
					continue
				}
				logger.Info("push to dingtalk successfully")
				break
			}
		}()
	}

	if config.GetSingleton().LarkBot.Enable {
		if config.GetSingleton().LarkBot.WebHookAccessToken == "" || config.GetSingleton().LarkBot.WebHookSecret == "" {
			logger.Info("didn't configure lark_bot webhook_access_token or webhook_secret, skip push")
			return
		}
		go func() {
			retry := 5
			var is []string
			for _, cve := range pushCVEList {
				is = append(is, fmt.Sprintf("**漏洞编号**:%s  \n**地址**:%s  \n**描述**:%s  \n", cve.Name, cve.HtmlUrl, cve.Description))
			}
			msg := strings.Join(is, "---\n")
			for {
				if retry <= 0 {
					break
				}
				code, err := push.GetLarkSingleton().Push(msg)
				if code < 0 {
					logger.Info(err.Error())
					retry--
					time.Sleep(3 * time.Second)
					continue
				}
				logger.Info("push to lark chat group successfully")
				break
			}
		}()
	}

	if config.GetSingleton().LarkAssistant.Enable {
		if config.GetSingleton().LarkAssistant.WebHookAccessToken == "" {
			logger.Info("didn't configure lark_assistant webhook_access_token, skip push")
			return
		}
		go func() {
			retry := 5
			var is []string
			for _, cve := range pushCVEList {
				msg := fmt.Sprintf("**漏洞编号**:%s  \n**地址**:%s  \n**描述**:", cve.Name, cve.HtmlUrl)
				if cve.Description != "" {
					msg = fmt.Sprintf("%s<font color='grey'>%s</font>", msg, cve.Description)
				}
				is = append(is, msg)
			}
			msg := strings.Join(is, "\n\n")
			for {
				if retry <= 0 {
					break
				}
				code, err := push.GetLarkAssistantLarkSingleton().Push(title, msg)
				if code < 0 {
					logger.Info(err.Error())
					retry--
					time.Sleep(3 * time.Second)
					continue
				}
				logger.Info("push to lark assistant successfully")
				break
			}
		}()
	}
}

func RunAVDMonitor() {
	m := monitor.GetAVDMonitorSingleton()
	items, err := m.GetLatestAVDList()
	if err != nil {
		logger.Info(err.Error())
		return
	}
	var avds []entry.AVD
	for _, item := range items {
		avd := db.GetAVDDbServiceSingleton().GetByAVD(item.Number)
		if avd.Number != "" && avd.Number == item.Number {
			continue
		}
		detail, err := m.GetAVDDetail(item.Number)
		if err != nil {
			continue
		}
		item.Description = detail.Description
		item.ImpactVersion = detail.ImpactVersion
		item.Reference = detail.Reference
		avds = append(avds, item)
	}
	if len(avds) == 0 {
		return
	}
	var numbers []string
	for _, avd := range avds {
		numbers = append(numbers, avd.Number)
	}
	logger.Info(fmt.Sprintf("inset AVDS [ %s ]", strings.Join(numbers, ", ")))
	err = db.GetAVDDbServiceSingleton().Inset(avds...)
	if err != nil {
		logger.Info(fmt.Sprintf("inset AVDS [ %s ] err: %s", strings.Join(numbers, ", "), err.Error()))
	}
	title := "阿里云漏洞推送"
	if config.GetSingleton().DingTalk.Enable {
		if config.GetSingleton().DingTalk.WebHookAccessToken == "" || config.GetSingleton().DingTalk.WebHookSecret == "" {
			logger.Info("didn't configure dingtalk webhook_access_token or webhook_secret, skip push")
			return
		}
		go func() {
			//避免body过大，分段发送
			var is []string
			for i, avd := range avds {
				t := utils.ListTrimSpace(strings.Split(avd.ImpactVersion, ","))
				var tt string
				if len(t) > 0 {
					tt = strings.Join(t, "  \n") + "  \n"
				}
				is = append(is,
					fmt.Sprintf("**漏洞来源**:  %s  \n**漏洞编号**:  %s  \n**漏洞名称**:  %s  \n**披露日期**:  %s  \n**漏洞状态**:  %s  \n**漏洞描述**:  %s  \n**影响版本**:  \n%s**参考链接**:  %s  \n",
						"阿里云漏洞库",
						avd.Number,
						avd.Name,
						avd.DisclosureDate,
						avd.POC,
						avd.Description,
						tt,
						avd.Reference,
					))
				msg := strings.Join(is, "  \n------  \n")
				if (i+1)%10 == 0 || i == len(avds)-1 {
					retry := 5
					for {
						if retry <= 0 {
							break
						}
						code, err := push.GetDingTalkSingleton().Push(title, msg)
						if code < 0 {
							logger.Info(err.Error())
							retry--
							time.Sleep(3 * time.Second)
							continue
						}
						logger.Info("push to dingtalk successfully")
						time.Sleep(3 * time.Second)
						break
					}
					is = []string{}
				}
			}

		}()
	}

	if config.GetSingleton().LarkBot.Enable {
		if config.GetSingleton().LarkBot.WebHookAccessToken == "" || config.GetSingleton().LarkBot.WebHookSecret == "" {
			logger.Info("didn't configure lark_bot webhook_access_token or webhook_secret, skip push")
			return
		}
		go func() {
			var is []string
			for i, avd := range avds {
				t := utils.ListTrimSpace(strings.Split(avd.ImpactVersion, ","))
				var tt string
				if len(t) > 0 {
					tt = strings.Join(t, "\n") + "\n"
				}
				is = append(is, fmt.Sprintf("**漏洞来源**:  %s  \n**漏洞编号**:  %s  \n**漏洞名称**:  %s  \n**披露日期**:  %s  \n**漏洞状态**:  %s  \n**漏洞描述**:  %s  \n**影响版本**:  \n%s**参考链接**:  \n%s  \n",
					"阿里云漏洞库",
					avd.Number,
					avd.Name,
					avd.DisclosureDate,
					avd.POC,
					avd.Description,
					tt,
					avd.Reference,
				))
				msg := strings.Join(is, "---\n")
				if (i+1)%10 == 0 || i == len(avds)-1 {
					retry := 5
					for {
						if retry <= 0 {
							break
						}
						code, err := push.GetLarkSingleton().Push(msg)
						if code < 0 {
							logger.Info(err.Error())
							retry--
							time.Sleep(3 * time.Second)
							continue
						}
						logger.Info("push to lark chat group successfully")
						time.Sleep(3 * time.Second)
						break
					}
					is = []string{}
				}
			}
		}()
	}

	if config.GetSingleton().LarkAssistant.Enable {
		if config.GetSingleton().LarkAssistant.WebHookAccessToken == "" {
			logger.Info("didn't configure lark_assistant webhook_access_token, skip push")
			return
		}
		go func() {
			var is []string
			for i, avd := range avds {
				t := utils.ListTrimSpace(strings.Split(avd.ImpactVersion, ","))
				var tt string
				if len(t) > 0 {
					tt = strings.Join(t, "\n") + "\n"
				}
				is = append(is, fmt.Sprintf("**漏洞来源**:  %s  \n**漏洞编号**:  %s  \n**漏洞名称**:  %s  \n**披露日期**:  %s  \n**漏洞状态**:  %s  \n**漏洞描述**:  %s  \n**影响版本**:  \n%s**参考链接**:  \n%s  \n",
					"阿里云漏洞库",
					avd.Number,
					avd.Name,
					avd.DisclosureDate,
					avd.POC,
					avd.Description,
					tt,
					avd.Reference,
				))
				msg := strings.Join(is, "\n\n")
				if (i+1)%10 == 0 || i == len(avds)-1 {
					retry := 5
					for {
						if retry <= 0 {
							break
						}
						code, err := push.GetLarkAssistantLarkSingleton().Push(title, msg)
						if code < 0 {
							logger.Info(err.Error())
							retry--
							time.Sleep(3 * time.Second)
							continue
						}
						logger.Info("push to lark assistant successfully")
						time.Sleep(3 * time.Second)
						break
					}
					is = []string{}
				}
			}
		}()
	}
}
