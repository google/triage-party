package provider

import (
	"time"
)

// IComment is an interface that matches both GitHub issues and pull review comments
type IComment interface {
	GetAuthorAssociation() string
	GetBody() string
	GetCreatedAt() time.Time
	GetReactions() *Reactions
	GetHTMLURL() string
	GetID() int64
	GetURL() string
	GetUpdatedAt() time.Time
	GetUser() *User
	String() string
}

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

func NewComment(g IComment) *Comment {
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
		return time.Now()
	}
	return i.Updated
}

func (i *Comment) GetCreatedAt() time.Time {
	if i == nil {
		return time.Now()
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
