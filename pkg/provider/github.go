package provider

import (
	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/models"
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

func getIssues(i []*github.Issue) (r []*models.Issue) {
	// TODO
	mi := models.Issue{}
	r = append(r, &mi)
	return
}

func getResponse(i *github.Response) (r *models.Response) {
	// TODO
	return &models.Response{}
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

func getIssueComments(i []*github.IssueComment) (r []*models.IssueComment) {
	// TODO
	ic := models.IssueComment{}
	r = append(r, &ic)
	return
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

func getIssueTimeline(i []*github.Timeline) (r []*models.Timeline) {
	// TODO
	it := models.Timeline{}
	r = append(r, &it)
	return
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

func getPullRequestsList(i []*github.PullRequest) (r []*models.PullRequest) {
	// TODO
	it := models.PullRequest{}
	r = append(r, &it)
	return
}

func (p *GithubProvider) PullRequestsList(sp models.SearchParams) (i []*models.PullRequest, r *models.Response, err error) {
	opt := getPullRequestsListOptions(sp)
	gpr, gr, err := p.client.PullRequests.List(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, opt)
	i = getPullRequestsList(gpr)
	r = getResponse(gr)
	return
}

func getPullRequest(i *github.PullRequest) (r *models.PullRequest) {
	// TODO
	return &models.PullRequest{}
}

func (p *GithubProvider) PullRequestsGet(sp models.SearchParams) (i *models.PullRequest, r *models.Response, err error) {
	pr, gr, err := p.client.PullRequests.Get(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber)
	i = getPullRequest(pr)
	r = getResponse(gr)
	return
}

func getPullRequestListComments(i []*github.PullRequestComment) (r []*models.PullRequestComment) {
	// TODO
	it := models.PullRequestComment{}
	r = append(r, &it)
	return
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

func getPullRequestsListReviews(i []*github.PullRequestReview) (r []*models.PullRequestReview) {
	// TODO
	it := models.PullRequestReview{}
	r = append(r, &it)
	return
}

func (p *GithubProvider) PullRequestsListReviews(sp models.SearchParams) (i []*models.PullRequestReview, r *models.Response, err error) {
	opt := getListOptions(sp.ListOptions)
	pr, gr, err := p.client.PullRequests.ListReviews(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, &opt)
	i = getPullRequestsListReviews(pr)
	r = getResponse(gr)
	return
}
