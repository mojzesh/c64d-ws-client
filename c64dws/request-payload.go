package c64dws

// Request Parameters
type Params map[string]any

// Request Payload Base
type RequestPayloadBase struct {
	Fn     string  `json:"fn"`
	Params *Params `json:"params,omitempty"`
}

// Request Payload With Token
type RequestPayloadWithToken struct {
	RequestPayloadBase
	Token string `json:"token"`
}
