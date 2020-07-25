package provider

import (
	"encoding/json"
	"fmt"
	"github.com/google/triage-party/pkg/constants"
	"github.com/google/triage-party/pkg/models"
	"github.com/xanzy/go-gitlab"
	"log"
	"strconv"
	"time"
)

type GitlabProvider struct {
	client *gitlab.Client
}

func initGitlab(c Config) {
	cl, err := gitlab.NewClient(mustReadToken(*c.GithubTokenFile, constants.GitlabTokenEnvVar))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	gitlabProvider = &GitlabProvider{
		client: cl,
	}
}

func (p *GitlabProvider) getListProjectIssuesOptions(sp models.SearchParams) *gitlab.ListProjectIssuesOptions {
	return &gitlab.ListProjectIssuesOptions{
		ListOptions:  p.getListOptions(sp.IssueListByRepoOptions.ListOptions),
		State:        &sp.IssueListByRepoOptions.State,
		CreatedAfter: &sp.IssueListByRepoOptions.Since,
	}
}

func (p *GitlabProvider) getListOptions(m models.ListOptions) gitlab.ListOptions {
	return gitlab.ListOptions{
		Page:    m.Page,
		PerPage: m.PerPage,
	}
}

func (p *GitlabProvider) getIssues(i []*gitlab.Issue) []*models.Issue {
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

func (p *GitlabProvider) getRate(i *gitlab.Response) models.Rate {
	l, err := strconv.Atoi(i.Header.Get(constants.GitlabRateLimitHeader))
	if err != nil {
		fmt.Println(err)
	}
	r, err := strconv.Atoi(i.Header.Get(constants.GitlabRateLimitRemainingHeader))
	if err != nil {
		fmt.Println(err)
	}
	rs, err := strconv.Atoi(i.Header.Get(constants.GitlabRateLimitResetHeader))
	if err != nil {
		fmt.Println(err)
	}
	tm := time.Unix(int64(rs), 0)
	return models.Rate{
		Limit:     l,
		Remaining: r,
		Reset:     models.Timestamp{tm},
	}
}

func (p *GitlabProvider) getResponse(i *gitlab.Response) *models.Response {
	r := models.Response{
		NextPage: i.NextPage,
		Rate:     p.getRate(i),
	}
	return &r
}

// https://docs.gitlab.com/ee/api/issues.html#list-project-issues
func (p *GitlabProvider) IssuesListByRepo(sp models.SearchParams) (i []*models.Issue, r *models.Response, err error) {
	opt := p.getListProjectIssuesOptions(sp)
	is, gr, err := p.client.Issues.ListProjectIssues(sp.Repo.Project, opt)
	i = p.getIssues(is)
	r = p.getResponse(gr)
	return
}

func (p *GitlabProvider) getListIssueNotesOptions(sp models.SearchParams) *gitlab.ListIssueNotesOptions {
	return &gitlab.ListIssueNotesOptions{
		ListOptions: p.getListOptions(sp.IssueListCommentsOptions.ListOptions),
	}
}

func (p *GitlabProvider) getUserFromNote(i *gitlab.Note) *models.User {
	id := int64(i.ID)
	return &models.User{
		ID:        &id,
		Name:      &i.Author.Name,
		Login:     &i.Author.Username, // TODO need to clarify
		Email:     &i.Author.Email,
		AvatarURL: &i.Author.AvatarURL,
		HTMLURL:   &i.Author.WebURL, // TODO need to clarify
	}
}

func (p *GitlabProvider) getIssueComments(i []*gitlab.Note) []*models.IssueComment {
	r := make([]*models.IssueComment, len(i))
	for k, v := range i {
		m := &models.IssueComment{
			User:      p.getUserFromNote(v),
			Body:      &v.Body,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		}
		r[k] = m
	}
	return r
}

// https://docs.gitlab.com/ce/api/notes.html#list-project-issue-notes
func (p *GitlabProvider) IssuesListComments(sp models.SearchParams) (i []*models.IssueComment, r *models.Response, err error) {
	opt := p.getListIssueNotesOptions(sp)
	in, gr, err := p.client.Notes.ListIssueNotes(sp.Repo.Project, sp.IssueNumber, opt)
	i = p.getIssueComments(in)
	r = p.getResponse(gr)
	return
}

func (p *GitlabProvider) IssuesListIssueTimeline(sp models.SearchParams) (i []*models.Timeline, r *models.Response, err error) {
	// TODO need discuss - gitlab dont provide events by issue number (Issues, Merge Requests)
	return
}

func (p *GitlabProvider) getListProjectMergeRequestsOptions(sp models.SearchParams) *gitlab.ListProjectMergeRequestsOptions {
	var orderBy string
	if sp.PullRequestListOptions.Sort == constants.UpdatedSortOption {
		orderBy = constants.UpdatedAtSortOption
	} else {
		orderBy = constants.CreatedAtSortOption
	}
	return &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: p.getListOptions(sp.PullRequestListOptions.ListOptions),
		Sort:        &sp.PullRequestListOptions.Direction,
		OrderBy:     &orderBy,
	}
}

func (p *GitlabProvider) getUserFromBasicUser(i *gitlab.BasicUser) *models.User {
	id := int64(i.ID)
	return &models.User{
		ID:        &id,
		Name:      &i.Name,
		Login:     &i.Username, // TODO need to clarify
		AvatarURL: &i.AvatarURL,
		HTMLURL:   &i.WebURL, // TODO need to clarify
	}
}

func (p *GitlabProvider) getMilestone(i *gitlab.Milestone) *models.Milestone {
	id := int64(i.ID)
	dd := time.Time(*i.DueDate)
	return &models.Milestone{
		ID:          &id,
		Number:      &i.IID,
		Title:       &i.Title,
		Description: &i.Description,
		DueOn:       &dd,
		State:       &i.State,
		URL:         &i.WebURL, // TODO need to clarify
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}

func (p *GitlabProvider) getPullRequest(v *gitlab.MergeRequest) *models.PullRequest {
	id := int64(v.ID)
	m := &models.PullRequest{
		Assignee:  p.getUserFromBasicUser(v.Assignee),
		User:      p.getUserFromBasicUser(v.Author),
		Body:      &v.Description, // TODO need to clarify
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		ClosedAt:  v.ClosedAt,
		URL:       &v.WebURL,
		Title:     &v.Title,
		State:     &v.State,
		ID:        &id,
		Number:    &v.IID,
		Milestone: p.getMilestone(v.Milestone),
	}
}

func (p *GitlabProvider) getPullRequests(i []*gitlab.MergeRequest) []*models.PullRequest {
	r := make([]*models.PullRequest, len(i))
	for k, v := range i {
		m := p.getPullRequest(v)
		r[k] = m
	}
	return r
}

func (p *GitlabProvider) PullRequestsList(sp models.SearchParams) (i []*models.PullRequest, r *models.Response, err error) {
	opt := p.getListProjectMergeRequestsOptions(sp)
	in, gr, err := p.client.MergeRequests.ListProjectMergeRequests(sp.Repo.Project, opt)
	i = p.getPullRequests(in)
	r = p.getResponse(gr)
	return
}

func (p *GitlabProvider) PullRequestsGet(sp models.SearchParams) (i *models.PullRequest, r *models.Response, err error) {
	opt := &gitlab.GetMergeRequestsOptions{}
	in, gr, err := p.client.MergeRequests.GetMergeRequest(sp.Repo.Project, sp.IssueNumber, opt)
	i = p.getPullRequest(in)
	r = p.getResponse(gr)
	return
}

func (p *GitlabProvider) PullRequestsListComments(sp models.SearchParams) ([]*models.PullRequestComment, *models.Response, error) {

}

func (p *GitlabProvider) PullRequestsListReviews(sp models.SearchParams) ([]*models.PullRequestReview, *models.Response, error) {

}
