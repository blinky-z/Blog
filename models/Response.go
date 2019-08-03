package models

// Response - struct for sending payload from server and more info about occurred error
// It behaves like Either Monad: 'Error' field is set if error occurred, otherwise 'Body' contains payload
type Response struct {
	Error interface{} `json:"error"`
	Body  interface{} `json:"body"`
}
