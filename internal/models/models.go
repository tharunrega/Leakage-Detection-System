package models

type Token struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Owner string `json:"owner"`
}

type LeakResult struct {
	TokenType  string  `json:"token_type"`
	Repo       string  `json:"repo"`
	URL        string  `json:"url"`
	User       string  `json:"user"`
	Owner      string  `json:"owner,omitempty"`
	Path       string  `json:"path,omitempty"`
	Location   string  `json:"location"`
	Snippet    string  `json:"snippet"`
	Confidence float64 `json:"confidence"`
}
