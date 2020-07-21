package provider

import (
	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/hubbub"
	"github.com/google/triage-party/pkg/interfaces"
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

func getIssues(i []*github.Issue) (r []hubbub.Item) {
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

func (p *GithubProvider) IssuesListByRepo(sp models.SearchParams) (i []hubbub.Item, r *models.Response, err error) {
	opt := getIssueListByRepoOptions(sp)
	gi, gr, err := p.client.Issues.ListByRepo(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, opt)
	i = getIssues(gi)
	r = getResponse(gr)
	return
}

func getIssuesListCommentsOptions(sp models.SearchParams) *github.IssueListCommentsOptions {
	return &github.IssueListCommentsOptions{
		ListOptions: getListOptions(sp.IssueListCommentsOptions.ListOptions),
		State:       sp.State,
	}
}

func (p *GithubProvider) IssuesListComments(sp models.SearchParams) (i []interfaces.IIssueComment, r *models.Response, err error) {
	opt := getIssuesListCommentsOptions(sp)
	gI, gr, err := p.client.Issues.ListComments(sp.Ctx, sp.Repo.Organization, sp.Repo.Project, sp.IssueNumber, opt)
}
