package interfaces

import (
	"github.com/google/triage-party/pkg/models"
	"time"
)

type IIssueComment interface {
	GetAuthorAssociation() string
	GetBody() string
	GetCreatedAt() time.Time
	GetReactions() *models.Reactions
	GetHTMLURL() string
	GetID() int64
	GetURL() string
	GetUpdatedAt() time.Time
	GetUser() *models.User
	String() string
}

// Item is an interface that matches both Issues and PullRequests
type IItem interface {
	// github specific
	GetAssignee() *models.User
	GetAuthorAssociation() string
	GetBody() string
	GetComments() int
	GetHTMLURL() string
	GetCreatedAt() time.Time
	GetID() int64
	GetMilestone() *models.Milestone
	GetNumber() int
	GetClosedAt() time.Time
	GetState() string
	GetTitle() string
	GetURL() string
	GetUpdatedAt() time.Time
	GetUser() *models.User
	String() string
	IsPullRequest() bool
}

// interface that matches both GitHub issues and pull review comments
type IComment interface {
	GetAuthorAssociation() string
	GetBody() string
	GetCreatedAt() time.Time
	GetReactions() *models.Reactions
	GetHTMLURL() string
	GetID() int64
	GetURL() string
	GetUpdatedAt() time.Time
	GetUser() *models.User
	String() string
}
