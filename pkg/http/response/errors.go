package response

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type Error struct {
	Errors map[string]interface{} `json:"errors"`
}

type APIErrorResponse struct {
	Success    bool      `json:"success"`
	Message    string    `json:"message"`
	StatusCode int       `json:"status_code"`
	Data       ErrorData `json:"data"`
}
type ErrorData struct {
	Error string `json:"error"`
}

type APIResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data"`
}

func APIResponseData(
	success bool,
	message string,
	statusCode int,
	data interface{},

) *APIResponse {
	return &APIResponse{
		Success:    success,
		Message:    message,
		StatusCode: statusCode,
		Data:       data,
	}
}

type ErrorResponse struct {
	MessageID  string
	Err        error
	StatusCode int
}

func (e *ErrorResponse) ErrorString() string {
	return fmt.Sprintf("MessageID: %s, Error: %v",
		e.MessageID, e.Err)
}
func (e *ErrorResponse) NewErrorResponse(messageID string, err error, StatusCode int) *ErrorResponse {
	return &ErrorResponse{
		MessageID:  messageID,
		Err:        err,
		StatusCode: StatusCode,
	}
}

func ApiResponseError(success bool, message string, statusCode int, err error) APIErrorResponse {
	return APIErrorResponse{
		Success:    success,
		Message:    message,
		StatusCode: statusCode,
		Data: ErrorData{
			Error: err.Error(),
		},
	}
}

func JSONError(c *fiber.Ctx, statusCode int, message string, err error) error {
	return c.Status(statusCode).JSON(ApiResponseError(false, message, statusCode, err))
}
