package main

import (
	"cveHunter/config"
	"cveHunter/db"
	"cveHunter/logger"
	"cveHunter/monitor/github"
	"cveHunter/push"
	"fmt"
	"time"
)

func Run() {
	githubMonitor := github.GetSingleton()
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
	if len(pushCVEList) > 0 {
		go func() {
			retry := 5
			for {
				if retry <= 0 {
					break
				}
				code, err := push.GetDingTalkSingleton().Push(pushCVEList...)
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
}
