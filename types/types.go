package types

type Metadata struct {
	Key      string `json:"key"`
	UserID   string `json:"userID"`
	Platform string `json:"platform"`
	FileType string `json:"fileType"`
	FileSize int64  `json:"fileSize"`
}

type Response struct {
	Message string `json:"message" description:"Message of the response"`
}

type ApiError struct {
	Context map[string]string `json:"context,omitempty" description:"Context of the error. Usually used for validation error contexts"`
	Message string            `json:"message" description:"Message of the error"`
}
