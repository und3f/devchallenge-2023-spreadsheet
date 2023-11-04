package entities

type CellResponse struct {
	Value  string `json:"value"`
	Result string `json:"result"`

	// Removed for API format compliance
	Error *string `json:"-"`
}
