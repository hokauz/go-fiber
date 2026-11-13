package hello

// Response is the payload for the hello endpoint.
type Response struct {
	RequestID string `json:"request_id"`
}

// PrivateResponse is the payload for the private hello endpoint.
type PrivateResponse struct {
	RequestID string `json:"request_id"`
	UserID    string `json:"user_id"`
}
