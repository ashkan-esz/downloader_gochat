package response

import (
	"github.com/gofiber/fiber/v2"
)

type ResponseOKWithDataModel struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type ResponseOKModel struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResponseErrorModel struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResponseErrorCustomModel struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
}

func ResponseOKWithData(c *fiber.Ctx, data interface{}) error {
	response := ResponseOKWithDataModel{
		Code:    200,
		Data:    data,
		Message: "OK",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func ResponseCreated(c *fiber.Ctx, data interface{}) error {
	response := ResponseOKWithDataModel{
		Code:    201,
		Data:    data,
		Message: "Created",
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func ResponseOK(c *fiber.Ctx, message string) error {
	response := ResponseOKModel{
		Code:    200,
		Message: message,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func ResponseError(c *fiber.Ctx, err string, code int) error {
	response := ResponseErrorModel{
		Code:    99,
		Message: err,
	}

	return c.Status(code).JSON(response)
}

func ResponseCustomError(c *fiber.Ctx, err interface{}, code int) error {
	response := ResponseErrorCustomModel{
		Code:    code,
		Message: err,
	}

	return c.Status(code).JSON(response)
}
