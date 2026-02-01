package snyk

// PaginatedLinks represents links on a collection document.
//
// See: https://jsonapi.org/format/#fetching-pagination
type PaginatedLinks struct {
	Self    string `json:"self,omitempty"`
	Related string `json:"related,omitempty"`
	First   string `json:"first,omitempty"`
	Last    string `json:"last,omitempty"`
	Prev    string `json:"prev,omitempty"`
	Next    string `json:"next,omitempty"`
}

func (l PaginatedLinks) String() string { return Stringify(l) }
