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
