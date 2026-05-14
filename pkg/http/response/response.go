package response

import "github.com/gofiber/fiber/v2"

type Response struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data"`
}

func NewResponse(message string, statusCode int, data interface{}) Response {
	return Response{
		Success:    true,
		Message:    message,
		StatusCode: statusCode,
		Data:       data,
	}
}

func JSON(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(NewResponse(message, statusCode, data))
}

type ResponseWithPaging struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Total      int         `json:"total"`
}

func NewResponseWithPaging(message string, statusCode int, data interface{}, page int, limit int, total int) ResponseWithPaging {
	return ResponseWithPaging{
		Success:    true,
		Message:    message,
		StatusCode: statusCode,
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
	}
}

func JSONWithPaging(c *fiber.Ctx, statusCode int, message string, data interface{}, page int, limit int, total int) error {
	return c.Status(statusCode).JSON(NewResponseWithPaging(message, statusCode, data, page, limit, total))
}
