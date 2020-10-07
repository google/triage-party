package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/triage-party/pkg/constants"
	"github.com/xanzy/go-gitlab"
)

type GitlabProvider struct {
	client *gitlab.Client
}

func initGitlab(c Config) {
	token := os.Getenv(constants.GitlabTokenEnvVar)
	path := *c.GitlabTokenFile
	if (token == "") && (path == "") {
		return
	}
	cl, err := gitlab.NewClient(mustReadToken(path, token, constants.GitlabTokenEnvVar, constants.GitlabProviderName))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	gitlabProvider = &GitlabProvider{
		client: cl,
	}
}

func (p *GitlabProvider) getListProjectIssuesOptions(sp SearchParams) *gitlab.ListProjectIssuesOptions {
	var state *string
	if sp.IssueListByRepoOptions.State == constants.OpenState {
		s := constants.OpenedState
		state = &s
	}
	return &gitlab.ListProjectIssuesOptions{
		ListOptions:  p.getListOptions(sp.IssueListByRepoOptions.ListOptions),
		State:        state,
		CreatedAfter: &sp.IssueListByRepoOptions.Since,
	}
}

func (p *GitlabProvider) getListOptions(m ListOptions) gitlab.ListOptions {
	return gitlab.ListOptions{
		Page:    m.Page,
		PerPage: m.PerPage,
	}
}

func (p *GitlabProvider) getUserFromIssueAssignee(i *gitlab.IssueAssignee) *User {
	if i == nil {
		return nil
	}
	id := int64(i.ID)
	return &User{
		ID:        &id,
		Name:      &i.Name,
		Login:     &i.Username,
		AvatarURL: &i.AvatarURL,
		HTMLURL:   &i.WebURL,
	}
}

func (p *GitlabProvider) getUserFromIssueAuthor(i *gitlab.IssueAuthor) *User {
	id := int64(i.ID)
	return &User{
		ID:        &id,
		Name:      &i.Name,
		Login:     &i.Username,
		AvatarURL: &i.AvatarURL,
		HTMLURL:   &i.WebURL,
	}
}

func (p *GitlabProvider) getIssues(i []*gitlab.Issue) []*Issue {
	r := make([]*Issue, len(i))
	for k, v := range i {
		id := int64(v.ID)
		m := Issue{
			Assignee:  p.getUserFromIssueAssignee(v.Assignee),
			HTMLURL:   &v.WebURL,
			Title:     &v.Title,
			URL:       &v.WebURL,
			User:      p.getUserFromIssueAuthor(v.Author),
			UpdatedAt: v.UpdatedAt,
			State:     &v.State,
			ClosedAt:  v.ClosedAt,
			Number:    &v.IID,
			Milestone: p.getMilestone(v.Milestone),
			ID:        &id,
			CreatedAt: v.CreatedAt,
		}
		r[k] = &m
	}
	return r
}

func (p *GitlabProvider) getRate(i *gitlab.Response) Rate {
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
	return Rate{
		Limit:     l,
		Remaining: r,
		Reset:     Timestamp{tm},
	}
}

func (p *GitlabProvider) getResponse(i *gitlab.Response) *Response {
	r := Response{
		NextPage: i.NextPage,
		Rate:     p.getRate(i),
	}
	return &r
}

// https://docs.gitlab.com/ee/api/issues.html#list-project-issues
func (p *GitlabProvider) IssuesListByRepo(ctx context.Context, sp SearchParams) (i []*Issue, r *Response, err error) {
	opt := p.getListProjectIssuesOptions(sp)
	is, gr, err := p.client.Issues.ListProjectIssues(p.getProjectId(sp.Repo), opt)
	i = p.getIssues(is)
	r = p.getResponse(gr)
	return
}

func (p *GitlabProvider) getListIssueNotesOptions(sp SearchParams) *gitlab.ListIssueNotesOptions {
	return &gitlab.ListIssueNotesOptions{
		ListOptions: p.getListOptions(sp.IssueListCommentsOptions.ListOptions),
	}
}

func (p *GitlabProvider) getUserFromNote(i *gitlab.Note) *User {
	id := int64(i.ID)
	return &User{
		ID:        &id,
		Name:      &i.Author.Name,
		Login:     &i.Author.Username,
		Email:     &i.Author.Email,
		AvatarURL: &i.Author.AvatarURL,
		HTMLURL:   &i.Author.WebURL,
	}
}

func (p *GitlabProvider) getIssueComments(i []*gitlab.Note) []*IssueComment {
	r := make([]*IssueComment, len(i))
	for k, v := range i {
		m := &IssueComment{
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
func (p *GitlabProvider) IssuesListComments(ctx context.Context, sp SearchParams) (i []*IssueComment, r *Response, err error) {
	opt := p.getListIssueNotesOptions(sp)
	in, gr, err := p.client.Notes.ListIssueNotes(p.getProjectId(sp.Repo), sp.IssueNumber, opt)
	i = p.getIssueComments(in)
	r = p.getResponse(gr)
	return
}

func (p *GitlabProvider) IssuesListIssueTimeline(ctx context.Context, sp SearchParams) (i []*Timeline, r *Response, err error) {
	// TODO need discuss - gitlab dont provide events by issue number (Issues, Merge Requests)
	i = make([]*Timeline, 0)
	r = &Response{}
	err = nil
	fmt.Println("provider.IssuesListIssueTimeline method is not implemented for gitlab")
	return
}

func (p *GitlabProvider) getListProjectMergeRequestsOptions(sp SearchParams) *gitlab.ListProjectMergeRequestsOptions {
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
		State:       &sp.State,
	}
}

func (p *GitlabProvider) getUserFromBasicUser(i *gitlab.BasicUser, allowNil bool) *User {
	if allowNil {
		if i == nil {
			return nil
		}
	} else {
		if i == nil {
			panic("User should not be nil")
		}
	}

	id := int64(i.ID)
	return &User{
		ID:        &id,
		Name:      &i.Name,
		Login:     &i.Username,
		AvatarURL: &i.AvatarURL,
		HTMLURL:   &i.WebURL,
	}
}

func (p *GitlabProvider) getMilestone(i *gitlab.Milestone) *Milestone {
	if i == nil {
		return nil
	}

	id := int64(i.ID)

	var it *gitlab.ISOTime
	var dueDate *time.Time
	it = i.DueDate
	if it != nil {
		dd := time.Time(*it)
		dueDate = &dd
	}

	return &Milestone{
		ID:          &id,
		Number:      &i.IID,
		Title:       &i.Title,
		Description: &i.Description,
		DueOn:       dueDate,
		State:       &i.State,
		URL:         &i.WebURL,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}

func (p *GitlabProvider) getPullRequest(v *gitlab.MergeRequest) *PullRequest {
	id := int64(v.ID)
	m := &PullRequest{
		Assignee:  p.getUserFromBasicUser(v.Assignee, true),
		User:      p.getUserFromBasicUser(v.Author, false),
		Body:      &v.Description,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		ClosedAt:  v.ClosedAt,
		URL:       &v.WebURL,
		Title:     &v.Title,
		State:     &v.State,
		ID:        &id,
		Number:    &v.IID,
		Milestone: p.getMilestone(v.Milestone),
		HTMLURL:   &v.WebURL,
	}
	return m
}

func (p *GitlabProvider) getPullRequests(i []*gitlab.MergeRequest) []*PullRequest {
	r := make([]*PullRequest, len(i))
	for k, v := range i {
		m := p.getPullRequest(v)
		r[k] = m
	}
	return r
}

func (p *GitlabProvider) PullRequestsList(ctx context.Context, sp SearchParams) (i []*PullRequest, r *Response, err error) {
	opt := p.getListProjectMergeRequestsOptions(sp)
	in, gr, err := p.client.MergeRequests.ListProjectMergeRequests(p.getProjectId(sp.Repo), opt)
	i = p.getPullRequests(in)
	r = p.getResponse(gr)
	return
}

func (p *GitlabProvider) PullRequestsGet(ctx context.Context, sp SearchParams) (i *PullRequest, r *Response, err error) {
	opt := &gitlab.GetMergeRequestsOptions{}
	in, gr, err := p.client.MergeRequests.GetMergeRequest(p.getProjectId(sp.Repo), sp.IssueNumber, opt)
	i = p.getPullRequest(in)
	r = p.getResponse(gr)
	return
}

func (p *GitlabProvider) getPullRequestComments(i []*gitlab.Note) []*PullRequestComment {
	r := make([]*PullRequestComment, len(i))
	for k, v := range i {
		id := int64(v.ID)
		m := &PullRequestComment{
			ID:        &id,
			Body:      &v.Body,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			User:      p.getUserFromNote(v),
		}
		r[k] = m
	}
	return r
}

func (p *GitlabProvider) PullRequestsListComments(ctx context.Context, sp SearchParams) (i []*PullRequestComment, r *Response, err error) {
	opt := &gitlab.ListMergeRequestNotesOptions{
		ListOptions: p.getListOptions(sp.ListOptions),
	}
	in, gr, err := p.client.Notes.ListMergeRequestNotes(p.getProjectId(sp.Repo), sp.IssueNumber, opt)
	i = p.getPullRequestComments(in)
	r = p.getResponse(gr)
	return
}

func (p *GitlabProvider) getPullRequestReviews(i *gitlab.MergeRequestApprovals) []*PullRequestReview {
	r := make([]*PullRequestReview, len(i.ApprovedBy))
	state := "APPROVED"
	for k, v := range i.ApprovedBy {
		m := &PullRequestReview{
			User:  p.getUserFromBasicUser(v.User, false),
			State: &state,
		}
		r[k] = m
	}
	return r
}

func (p *GitlabProvider) PullRequestsListReviews(ctx context.Context, sp SearchParams) (i []*PullRequestReview, r *Response, err error) {
	// TODO need to clarify
	in, gr, err := p.client.MergeRequests.GetMergeRequestApprovals(p.getProjectId(sp.Repo), sp.IssueNumber)
	i = p.getPullRequestReviews(in)
	r = p.getResponse(gr)
	return
}

// https://gitlab.com/gitlab-org/gitlab-foss/-/issues/28342#note_23852124
func (p *GitlabProvider) getProjectId(repo Repo) string {
	var u string
	if repo.Group != "" {
		u = repo.Organization + "/" + repo.Group + "/" + repo.Project
	} else {
		u = repo.Organization + "/" + repo.Project
	}
	return u
}
