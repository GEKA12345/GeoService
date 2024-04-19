package responder

// swagger:model tokenResponse
type TokenResponse struct {
	// access token
	//
	// example: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ
	AccessToken string `json:"access_token"`
}

// swagger:model searchResponse
type SearchResponse struct {
	// list of searched address
	Addresses []*Address `json:"addresses"`
}

// swagger:model geocodeResponse
type GeocodeResponse struct {
	// list of searched address
	Addresses []*Address `json:"addresses"`
}

type Address struct {
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// swagger:model errorResponse
type ErrorResponse struct {
	// required: true
	Message string `json:"error"`
}
