package snyk

// Organization represents a Snyk organization.
type Organization struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
