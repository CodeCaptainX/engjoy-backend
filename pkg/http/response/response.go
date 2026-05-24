package response

import (
	"sentenceminer/pkg/i18n"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data"`
}

func NewResponse(c *fiber.Ctx, key string, statusCode int, data interface{}) Response {
	lang := c.Locals("lang").(string)
	return Response{
		Success:    true,
		Message:    i18n.Translate(key, lang),
		StatusCode: statusCode,
		Data:       data,
	}
}

func JSON(c *fiber.Ctx, statusCode int, key string, data interface{}) error {
	return c.Status(statusCode).JSON(NewResponse(c, key, statusCode, data))
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

func NewResponseWithPaging(c *fiber.Ctx, key string, statusCode int, data interface{}, page int, limit int, total int) ResponseWithPaging {
	lang := c.Locals("lang").(string)
	return ResponseWithPaging{
		Success:    true,
		Message:    i18n.Translate(key, lang),
		StatusCode: statusCode,
		Data:       data,
		Page:       page,
		Limit:      limit,
		Total:      total,
	}
}

func JSONWithPaging(c *fiber.Ctx, statusCode int, key string, data interface{}, page int, limit int, total int) error {
	return c.Status(statusCode).JSON(NewResponseWithPaging(c, key, statusCode, data, page, limit, total))
}
