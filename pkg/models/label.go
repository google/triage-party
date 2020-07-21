package models

// Label represents a GitHub label on an Issue
type Label struct {
	ID          *int64  `json:"id,omitempty"`
	URL         *string `json:"url,omitempty"`
	Name        *string `json:"name,omitempty"`
	Color       *string `json:"color,omitempty"`
	Description *string `json:"description,omitempty"`
	Default     *bool   `json:"default,omitempty"`
	NodeID      *string `json:"node_id,omitempty"`
}

func (a *Label) GetName() string {
	if a == nil || a.Name == nil {
		return ""
	}
	return *a.Name
}
