package queryModel

type ExecuteResult struct {
	Count         *int64       `json:"count,omitempty"`
	NextPageToken string       `json:"next_page_token,omitempty"`
	Data          []TableModel `json:"data"`
}
