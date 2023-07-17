package config

type Created struct {
	Error *string `json:"error" example:"woah"`
	Value *string `json:"value" extensions:"x-nullable"`
}
