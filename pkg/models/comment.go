package models

import (
	"github.com/google/go-github/v31/github"
	"github.com/google/triage-party/pkg/interfaces"
	"time"
)

// Comment is a loose internal comment structure matching issue, PR, code review comments
type Comment struct {
	User        *github.User
	Created     time.Time
	Updated     time.Time
	Body        string
	AuthorAssoc string
	Reactions   *github.Reactions
	ReviewID    int64
}

func NewComment(g interfaces.IComment) *Comment {
	return &Comment{
		User:        g.GetUser(),
		Body:        g.GetBody(),
		AuthorAssoc: g.GetAuthorAssociation(),
		Created:     g.GetCreatedAt(),
		Updated:     g.GetUpdatedAt(),
		Reactions:   g.GetReactions(),
	}
}
