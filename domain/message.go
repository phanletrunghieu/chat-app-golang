package domain

type Message struct {
	Type byte        `json:"type"` //0: system, 1: text message
	From *User       `json:"from,omitempty"`
	To   *User       `json:"to,omitempty"`
	Data interface{} `json:"data,omitempty"`
}
