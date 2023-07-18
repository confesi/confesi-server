package docs

type Created struct {
	Error *string `json:"error" example:"null"`
	Value *string `json:"value" example:"null"`
}

type ServerError struct {
	Error *string `json:"error" example:"server error"`
	Value *string `json:"value" example:"null"`
}
