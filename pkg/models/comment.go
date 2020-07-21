package models

import (
	"github.com/google/triage-party/pkg/interfaces"
	"time"
)

// Comment is a loose internal comment structure matching issue, PR, code review comments
type Comment struct {
	User        *User
	Created     time.Time
	Updated     time.Time
	Body        string
	AuthorAssoc string
	Reactions   *Reactions
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

func (i *Comment) GetReactions() *Reactions {
	if i == nil {
		return nil
	}
	return i.Reactions
}

func (i *Comment) GetUpdatedAt() time.Time {
	if i == nil {
		return time.Now() // TODO need review
	}
	return i.Updated
}

func (i *Comment) GetCreatedAt() time.Time {
	if i == nil {
		return time.Now() // TODO need review
	}
	return i.Created
}

func (i *Comment) GetAuthorAssociation() string {
	if i == nil {
		return ""
	}
	return i.AuthorAssoc
}

func (i *Comment) GetBody() string {
	if i == nil {
		return ""
	}
	return i.Body
}

func (i *Comment) GetUser() *User {
	if i == nil {
		return nil
	}
	return i.User
}
