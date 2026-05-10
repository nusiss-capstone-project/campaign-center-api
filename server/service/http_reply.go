package service

// HTTPReply is a transport-agnostic envelope mapped by HTTP handlers to JSON {code,message,data}.
type HTTPReply struct {
	HTTPStatus int
	Code       int
	Message    string
	Data       any
}
