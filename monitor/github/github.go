package github

import (
	"cveHunter/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// https://docs.github.com/en/rest/search/search?apiVersion=2022-11-28#search-repositories
// ?q=tetris+language:assembly&sort=updated
const searchUrl = "https://api.github.com/search/repositories"

type Monitor struct {
	ProxyClient *http.Client
}

var (
	instance *Monitor
	once     sync.Once
)

func GetSingleton() *Monitor {
	once.Do(func() {
		instance = &Monitor{
			ProxyClient: &http.Client{},
		}
	})
	return instance
}

type Item struct {
	//ID       int    `json:"id"`
	//NodeID   string `json:"node_id"`
	Name string `json:"name"`
	//FullName string `json:"full_name"`
	Owner struct {
		Login string `json:"login"`
		//ID                int    `json:"id"`
		//NodeID            string `json:"node_id"`
		//AvatarURL         string `json:"avatar_url"`
		//GravatarID        string `json:"gravatar_id"`
		//URL               string `json:"url"`
		//ReceivedEventsURL string `json:"received_events_url"`
		//Type              string `json:"type"`
		//HTMLURL           string `json:"html_url"`
		//FollowersURL      string `json:"followers_url"`
		//FollowingURL      string `json:"following_url"`
		//GistsURL          string `json:"gists_url"`
		//StarredURL        string `json:"starred_url"`
		//SubscriptionsURL  string `json:"subscriptions_url"`
		//OrganizationsURL  string `json:"organizations_url"`
		//ReposURL          string `json:"repos_url"`
		//EventsURL         string `json:"events_url"`
		//SiteAdmin         bool   `json:"site_admin"`
	} `json:"owner"`
	//Private          bool      `json:"private"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
	//Fork             bool      `json:"fork"`
	//URL              string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	PushedAt  time.Time `json:"pushed_at"`
	//Homepage         string    `json:"homepage"`
	//Size             int       `json:"size"`
	//StargazersCount  int       `json:"stargazers_count"`
	//WatchersCount    int       `json:"watchers_count"`
	//Language         string    `json:"language"`
	//ForksCount       int       `json:"forks_count"`
	//OpenIssuesCount  int       `json:"open_issues_count"`
	//MasterBranch     string    `json:"master_branch"`
	//DefaultBranch    string    `json:"default_branch"`
	//Score            float64   `json:"score"`
	//ArchiveURL       string    `json:"archive_url"`
	//AssigneesURL     string    `json:"assignees_url"`
	//BlobsURL         string    `json:"blobs_url"`
	//BranchesURL      string    `json:"branches_url"`
	//CollaboratorsURL string    `json:"collaborators_url"`
	//CommentsURL      string    `json:"comments_url"`
	//CommitsURL       string    `json:"commits_url"`
	//CompareURL       string    `json:"compare_url"`
	//ContentsURL      string    `json:"contents_url"`
	//ContributorsURL  string    `json:"contributors_url"`
	//DeploymentsURL   string    `json:"deployments_url"`
	//DownloadsURL     string    `json:"downloads_url"`
	//EventsURL        string    `json:"events_url"`
	//ForksURL         string    `json:"forks_url"`
	//GitCommitsURL    string    `json:"git_commits_url"`
	//GitRefsURL       string    `json:"git_refs_url"`
	//GitTagsURL       string    `json:"git_tags_url"`
	//GitURL           string    `json:"git_url"`
	//IssueCommentURL  string    `json:"issue_comment_url"`
	//IssueEventsURL   string    `json:"issue_events_url"`
	//IssuesURL        string    `json:"issues_url"`
	//KeysURL          string    `json:"keys_url"`
	//LabelsURL        string    `json:"labels_url"`
	//LanguagesURL     string    `json:"languages_url"`
	//MergesURL        string    `json:"merges_url"`
	//MilestonesURL    string    `json:"milestones_url"`
	//NotificationsURL string    `json:"notifications_url"`
	//PullsURL         string    `json:"pulls_url"`
	//ReleasesURL      string    `json:"releases_url"`
	//SSHURL           string    `json:"ssh_url"`
	//StargazersURL    string    `json:"stargazers_url"`
	//StatusesURL      string    `json:"statuses_url"`
	//SubscribersURL   string    `json:"subscribers_url"`
	//SubscriptionURL  string    `json:"subscription_url"`
	//TagsURL          string    `json:"tags_url"`
	//TeamsURL         string    `json:"teams_url"`
	//TreesURL         string    `json:"trees_url"`
	//CloneURL         string    `json:"clone_url"`
	//MirrorURL        string    `json:"mirror_url"`
	//HooksURL         string    `json:"hooks_url"`
	//SvnURL           string    `json:"svn_url"`
	//Forks            int       `json:"forks"`
	//OpenIssues       int       `json:"open_issues"`
	//Watchers         int       `json:"watchers"`
	//HasIssues        bool      `json:"has_issues"`
	//HasProjects      bool      `json:"has_projects"`
	//HasPages         bool      `json:"has_pages"`
	//HasWiki          bool      `json:"has_wiki"`
	//HasDownloads     bool      `json:"has_downloads"`
	//Archived         bool      `json:"archived"`
	//Disabled         bool      `json:"disabled"`
	//Visibility       string    `json:"visibility"`
	//License          struct {
	//	Key     string `json:"key"`
	//	Name    string `json:"name"`
	//	URL     string `json:"url"`
	//	SpdxID  string `json:"spdx_id"`
	//	NodeID  string `json:"node_id"`
	//	HTMLURL string `json:"html_url"`
	//} `json:"license"`
}

func (r *Monitor) SearchCVEAll() (int, []Item, error) {
	params := url.Values{
		"q":        []string{"CVE"},
		"sort":     []string{"updated"},
		"page":     []string{"1"},
		"per_page": []string{"100"},
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", searchUrl, params.Encode()), nil)
	if err != nil {
		return -1, nil, err
	}
	request.Header.Add("Accept", "application/vnd.github+json")
	request.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Add("Authorization", config.GetSingleton().Github.GithubToken)

	response, err := r.ProxyClient.Do(request)
	if err != nil {
		return -1, nil, err
	}
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return -1, nil, err
	}
	type RespStructOfStatusCode200 struct {
		TotalCount        int  `json:"total_count"`
		IncompleteResults bool `json:"incomplete_results"`
		Items             Item `json:"items"`
	}
	if response.StatusCode != 200 {
		return response.StatusCode, nil, fmt.Errorf(string(bytes))
	}
	var t struct {
		TotalCount        int    `json:"total_count"`
		IncompleteResults bool   `json:"incomplete_results"`
		Items             []Item `json:"items"`
	}
	var items = make([]Item, 0)
	var tt = make([]Item, 0)
	if err := json.Unmarshal(bytes, &t); err != nil {
		fmt.Println(err)
		return 0, nil, fmt.Errorf("can't unmarshal github search result")
	}
	tt = append(tt, t.Items...)
	for _, item := range tt {
		name := item.Name
		re := regexp.MustCompile(`^(?i)CVE[-_—]?\d+[-_—]?\d+`)
		matches := re.FindStringSubmatch(name)
		if len(matches) > 0 {
			item.Name = matches[0]
			item.Description = strings.TrimSpace(item.Description)
			items = append(items, item)
		}
	}
	return 200, items, nil
}
