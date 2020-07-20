package models

import (
	"context"
	"github.com/google/triage-party/pkg/hubbub"
	"time"
)

type Issue struct {
}

type Response struct {
}

type Repo struct {
	Organization string
	Project      string
	Host         string
}

type SearchParams struct {
	Repo      Repo
	Filters   []hubbub.Filter
	Ctx       context.Context
	NewerThan time.Time
	Hidden    bool
	State     string
	UpdateAge time.Duration
	SearchKey string
}
