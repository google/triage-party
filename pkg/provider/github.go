package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/models"
	"golang.org/x/oauth2"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"strings"
)

type GithubProvider struct {
	client *github.Client
}

func getListOptions(m models.ListOptions) github.ListOptions {
	return github.ListOptions{
		Page:    m.Page,
		PerPage: m.PerPage,
	}
}

func getIssues(i []*github.Issue) []*models.Issue {
	r := make([]*models.Issue, len(i))
	for k, v := range i {
		m := models.Issue{}
		b, err := json.Marshal(v)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			fmt.Println(err)
		}
		r[k] = &m
	}
	return r
}

func getRate(i *github.Rate) models.Rate {
	r := models.Rate{}
	b, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(b, &r)
	if err != nil {
		fmt.Println(err)
	}
	return r
}

func getResponse(i *github.Response) *models.Response {
	r := models.Response{
		NextPage:      i.NextPage,
		PrevPage:      i.PrevPage,
		FirstPage:     i.FirstPage,
		LastPage:      i.LastPage,
		NextPageToken: i.NextPageToken,
		Rate:          getRate(&(*i).Rate),
	}
	return &r
}

func getIssueListByRepoOptions(sp models.SearchParams) *github.IssueListByRepoOptions {
	return &github.IssueListByRepoOptions{
		ListOptions: getListOptions(sp.IssueListByRepoOptions.ListOptions),
		State:       sp.State,
	}
}

func (p *GithubProvider) IssuesListByRepo(sp models.SearchParams) (i []*models.Issue, r *models.Response, err error) {
	opt := getIssueListByRepoOptions(sp)
	gi, gr, err := p.client.Issues.ListByRepo(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, opt)
	i = getIssues(gi)
	r = getResponse(gr)
	return
}

func getIssuesListCommentsOptions(sp models.SearchParams) *github.IssueListCommentsOptions {
	return &github.IssueListCommentsOptions{
		ListOptions: getListOptions(sp.IssueListCommentsOptions.ListOptions),
	}
}

func getIssueComments(i []*github.IssueComment) []*models.IssueComment {
	r := make([]*models.IssueComment, len(i))
	for k, v := range i {
		m := models.IssueComment{}
		b, err := json.Marshal(v)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			fmt.Println(err)
		}
		r[k] = &m
	}
	return r
}

func (p *GithubProvider) IssuesListComments(sp models.SearchParams) (i []*models.IssueComment, r *models.Response, err error) {
	opt := getIssuesListCommentsOptions(sp)
	gc, gr, err := p.client.Issues.ListComments(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, opt)
	i = getIssueComments(gc)
	r = getResponse(gr)
	return
}

func getIssuesListIssueTimelineOptions(sp models.SearchParams) *github.ListOptions {
	return &github.ListOptions{
		PerPage: sp.ListOptions.PerPage,
	}
}

func getIssueTimeline(i []*github.Timeline) []*models.Timeline {
	r := make([]*models.Timeline, len(i))
	for k, v := range i {
		m := models.Timeline{}
		b, err := json.Marshal(v)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			fmt.Println(err)
		}
		r[k] = &m
	}
	return r
}

func (p *GithubProvider) IssuesListIssueTimeline(sp models.SearchParams) (i []*models.Timeline, r *models.Response, err error) {
	opt := getIssuesListIssueTimelineOptions(sp)
	it, ir, err := p.client.Issues.ListIssueTimeline(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, opt)
	i = getIssueTimeline(it)
	r = getResponse(ir)
	return
}

func getPullRequestsListOptions(sp models.SearchParams) *github.PullRequestListOptions {
	return &github.PullRequestListOptions{
		ListOptions: getListOptions(sp.IssueListCommentsOptions.ListOptions),
		State:       sp.PullRequestListOptions.State,
		Sort:        sp.PullRequestListOptions.Sort,
		Direction:   sp.PullRequestListOptions.Direction,
	}
}

func getPullRequestsList(i []*github.PullRequest) []*models.PullRequest {
	r := make([]*models.PullRequest, len(i))
	for k, v := range i {
		m := models.PullRequest{}
		b, err := json.Marshal(v)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			fmt.Println(err)
		}
		r[k] = &m
	}
	return r
}

func (p *GithubProvider) PullRequestsList(sp models.SearchParams) (i []*models.PullRequest, r *models.Response, err error) {
	opt := getPullRequestsListOptions(sp)
	gpr, gr, err := p.client.PullRequests.List(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, opt)
	i = getPullRequestsList(gpr)
	r = getResponse(gr)
	return
}

func getPullRequest(i *github.PullRequest) *models.PullRequest {
	r := models.PullRequest{}
	b, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(b, &r)
	if err != nil {
		fmt.Println(err)
	}
	return &r
}

func (p *GithubProvider) PullRequestsGet(sp models.SearchParams) (i *models.PullRequest, r *models.Response, err error) {
	pr, gr, err := p.client.PullRequests.Get(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)
	i = getPullRequest(pr)
	r = getResponse(gr)
	return
}

func getPullRequestListComments(i []*github.PullRequestComment) []*models.PullRequestComment {
	r := make([]*models.PullRequestComment, len(i))
	for k, v := range i {
		m := models.PullRequestComment{}
		b, err := json.Marshal(v)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			fmt.Println(err)
		}
		r[k] = &m
	}
	return r
}

func getPullRequestsListCommentsOptions(sp models.SearchParams) *github.PullRequestListCommentsOptions {
	return &github.PullRequestListCommentsOptions{
		ListOptions: getListOptions(sp.ListOptions),
	}
}

func (p *GithubProvider) PullRequestsListComments(sp models.SearchParams) (i []*models.PullRequestComment, r *models.Response, err error) {
	opt := getPullRequestsListCommentsOptions(sp)
	pr, gr, err := p.client.PullRequests.ListComments(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, opt)
	i = getPullRequestListComments(pr)
	r = getResponse(gr)
	return
}

func getPullRequestsListReviews(i []*github.PullRequestReview) []*models.PullRequestReview {
	r := make([]*models.PullRequestReview, len(i))
	for k, v := range i {
		m := models.PullRequestReview{}
		b, err := json.Marshal(v)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(b, &m)
		if err != nil {
			fmt.Println(err)
		}
		r[k] = &m
	}
	return r
}

func (p *GithubProvider) PullRequestsListReviews(sp models.SearchParams) (i []*models.PullRequestReview, r *models.Response, err error) {
	opt := getListOptions(sp.ListOptions)
	pr, gr, err := p.client.PullRequests.ListReviews(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, &opt)
	i = getPullRequestsListReviews(pr)
	r = getResponse(gr)
	return
}

func MustReadToken(path string, env string) string {
	token := os.Getenv(env)
	if path != "" {
		t, err := ioutil.ReadFile(path)
		if err != nil {
			klog.Exitf("unable to read token file: %v", err)
		}
		token = string(t)
		klog.Infof("loaded %d byte github token from %s", len(token), path)
	} else {
		klog.Infof("loaded %d byte github token from %s", len(token), env)
	}

	token = strings.TrimSpace(token)
	if len(token) < 8 {
		klog.Exitf("github token impossibly small: %q", token)
	}
	return token
}

func MustCreateGithubClient(githubAPIRawURL string, httpClient *http.Client) *github.Client {
	if githubAPIRawURL != "" {
		client, err := github.NewEnterpriseClient(githubAPIRawURL, githubAPIRawURL, httpClient)
		if err != nil {
			klog.Exitf("unable to create GitHub client: %v", err)
		}
		return client
	}
	return github.NewClient(httpClient)
}

func initGithub(ctx context.Context, c Config) {
	cl := MustCreateGithubClient(*c.GithubAPIRawURL, oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: MustReadToken(*c.GithubTokenFile, "GITHUB_TOKEN")},
	)))
	githubProvider = &GithubProvider{
		client: cl,
	}
}
