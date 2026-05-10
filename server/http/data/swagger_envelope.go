package data

// StandardResponse is the common JSON envelope { code, message, data } used by HTTP handlers.
type StandardResponse struct {
	Code    int         `json:"code" example:"0"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data"`
}
