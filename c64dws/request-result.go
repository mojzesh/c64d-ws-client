package c64dws

type Result map[string]any

// Request result base
type RequestResultBase struct {
	Status int `json:"status"`
}

// Request result
type RequestResult struct {
	RequestResultBase
	Token      string  `json:"token"`
	Result     *Result `json:"result"`
	BinaryData []byte  `json:"-"`
}

// Request error
type RequestError struct {
	RequestResultBase
	Error string `json:"error"`
}

// Helper functions
func GetResultOrError(msg any) (*RequestResult, *RequestError) {
	switch v := msg.(type) {
	case RequestResult:
		return &v, nil
	case RequestError:
		return nil, &v
	default:
		return nil, nil
	}
}
