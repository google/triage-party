package models

// Source represents a reference's source.
type Source struct {
	ID    *int64  `json:"id,omitempty"`
	URL   *string `json:"url,omitempty"`
	Actor *User   `json:"actor,omitempty"`
	Type  *string `json:"type,omitempty"`
	Issue *Issue  `json:"issue,omitempty"`
}

// GetIssue returns the Issue field.
func (s *Source) GetIssue() *Issue {
	if s == nil {
		return nil
	}
	return s.Issue
}
