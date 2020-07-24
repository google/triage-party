package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/constants"
	"github.com/google/triage-party/pkg/models"
	"golang.org/x/oauth2"
	"k8s.io/klog/v2"
	"net/http"
)

type GithubProvider struct {
	client *github.Client
}

func (p *GithubProvider) getListOptions(m models.ListOptions) github.ListOptions {
	return github.ListOptions{
		Page:    m.Page,
		PerPage: m.PerPage,
	}
}

func (p *GithubProvider) getIssues(i []*github.Issue) []*models.Issue {
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

func (p *GithubProvider) getRate(i *github.Rate) models.Rate {
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

func (p *GithubProvider) getResponse(i *github.Response) *models.Response {
	r := models.Response{
		NextPage:      i.NextPage,
		PrevPage:      i.PrevPage,
		FirstPage:     i.FirstPage,
		LastPage:      i.LastPage,
		NextPageToken: i.NextPageToken,
		Rate:          p.getRate(&(*i).Rate),
	}
	return &r
}

func (p *GithubProvider) getIssueListByRepoOptions(sp models.SearchParams) *github.IssueListByRepoOptions {
	return &github.IssueListByRepoOptions{
		ListOptions: p.getListOptions(sp.IssueListByRepoOptions.ListOptions),
		State:       sp.State,
	}
}

func (p *GithubProvider) IssuesListByRepo(sp models.SearchParams) (i []*models.Issue, r *models.Response, err error) {
	opt := p.getIssueListByRepoOptions(sp)
	gi, gr, err := p.client.Issues.ListByRepo(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, opt)
	i = p.getIssues(gi)
	r = p.getResponse(gr)
	return
}

func (p *GithubProvider) getIssuesListCommentsOptions(sp models.SearchParams) *github.IssueListCommentsOptions {
	return &github.IssueListCommentsOptions{
		ListOptions: p.getListOptions(sp.IssueListCommentsOptions.ListOptions),
	}
}

func (p *GithubProvider) getIssueComments(i []*github.IssueComment) []*models.IssueComment {
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
	opt := p.getIssuesListCommentsOptions(sp)
	gc, gr, err := p.client.Issues.ListComments(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, opt)
	i = p.getIssueComments(gc)
	r = p.getResponse(gr)
	return
}

func (p *GithubProvider) getIssuesListIssueTimelineOptions(sp models.SearchParams) *github.ListOptions {
	return &github.ListOptions{
		PerPage: sp.ListOptions.PerPage,
	}
}

func (p *GithubProvider) getIssueTimeline(i []*github.Timeline) []*models.Timeline {
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
	opt := p.getIssuesListIssueTimelineOptions(sp)
	it, ir, err := p.client.Issues.ListIssueTimeline(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, opt)
	i = p.getIssueTimeline(it)
	r = p.getResponse(ir)
	return
}

func (p *GithubProvider) getPullRequestsListOptions(sp models.SearchParams) *github.PullRequestListOptions {
	return &github.PullRequestListOptions{
		ListOptions: p.getListOptions(sp.IssueListCommentsOptions.ListOptions),
		State:       sp.PullRequestListOptions.State,
		Sort:        sp.PullRequestListOptions.Sort,
		Direction:   sp.PullRequestListOptions.Direction,
	}
}

func (p *GithubProvider) getPullRequestsList(i []*github.PullRequest) []*models.PullRequest {
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
	opt := p.getPullRequestsListOptions(sp)
	gpr, gr, err := p.client.PullRequests.List(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, opt)
	i = p.getPullRequestsList(gpr)
	r = p.getResponse(gr)
	return
}

func (p *GithubProvider) getPullRequest(i *github.PullRequest) *models.PullRequest {
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
	i = p.getPullRequest(pr)
	r = p.getResponse(gr)
	return
}

func (p *GithubProvider) getPullRequestListComments(i []*github.PullRequestComment) []*models.PullRequestComment {
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

func (p *GithubProvider) getPullRequestsListCommentsOptions(sp models.SearchParams) *github.PullRequestListCommentsOptions {
	return &github.PullRequestListCommentsOptions{
		ListOptions: p.getListOptions(sp.ListOptions),
	}
}

func (p *GithubProvider) PullRequestsListComments(sp models.SearchParams) (i []*models.PullRequestComment, r *models.Response, err error) {
	opt := p.getPullRequestsListCommentsOptions(sp)
	pr, gr, err := p.client.PullRequests.ListComments(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, opt)
	i = p.getPullRequestListComments(pr)
	r = p.getResponse(gr)
	return
}

func (p *GithubProvider) getPullRequestsListReviews(i []*github.PullRequestReview) []*models.PullRequestReview {
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
	opt := p.getListOptions(sp.ListOptions)
	pr, gr, err := p.client.PullRequests.ListReviews(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, &opt)
	i = p.getPullRequestsListReviews(pr)
	r = p.getResponse(gr)
	return
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
		&oauth2.Token{AccessToken: mustReadToken(*c.GithubTokenFile, constants.GithubTokenEnvVar)},
	)))
	githubProvider = &GithubProvider{
		client: cl,
	}
}
