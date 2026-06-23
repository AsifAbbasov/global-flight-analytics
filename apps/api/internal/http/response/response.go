package response

import "github.com/gofiber/fiber/v2"

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OK(c *fiber.Ctx, data interface{}) error {
	return c.JSON(SuccessResponse{
		Success: true,
		Data:    data,
	})
}

func Error(c *fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}
