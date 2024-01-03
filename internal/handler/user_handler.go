package handler

import (
	"downloader_gochat/internal/service"
	"downloader_gochat/model"
	"downloader_gochat/pkg/response"
	"downloader_gochat/util"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type IUserHandler interface {
	RegisterUser(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	LogOut(c *fiber.Ctx) error
	GetAllUser(c *fiber.Ctx) error
	GetDetailUser(c *fiber.Ctx) error
}

type UserHandler struct {
	userService service.IUserService
}

func NewUserHandler(userService service.IUserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

//------------------------------------------

func (h *UserHandler) RegisterUser(c *fiber.Ctx) error {
	var registerVM model.RegisterViewModel
	err := c.BodyParser(&registerVM)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	registerUserError := registerVM.Validate()
	if len(registerUserError) > 0 {
		return response.ResponseCustomError(c, registerUserError, fiber.StatusBadRequest)
	}

	result, err := h.userService.CreateUser(&registerVM)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	if result.ID == 0 {
		if result.Username == registerVM.Username {
			return response.ResponseCustomError(c, "username already exist", fiber.StatusConflict)
		} else {
			return response.ResponseCustomError(c, "email already exist", fiber.StatusConflict)
		}
	}

	return response.ResponseCreated(c, result)
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var loginVM model.LoginViewModel
	err := c.BodyParser(&loginVM)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusUnprocessableEntity)
	}

	validateUser, err := h.userService.LoginUser(&loginVM)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	if validateUser == nil {
		validateUser = &model.UserViewModel{}
	}

	// Generete JWT
	token, err := util.CreateJwtToken(validateUser.ID, validateUser.Username)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	c.Cookie(&fiber.Cookie{
		Name:        "jwt",
		Value:       token.AccessToken,
		Path:        "/",
		Domain:      "localhost",
		MaxAge:      3600,
		Expires:     time.Time{},
		Secure:      false,
		HTTPOnly:    true,
		SameSite:    "",
		SessionOnly: false,
	})

	userData := map[string]interface{}{
		"access_token": token.AccessToken,
		"expired":      token.ExpireAt,
		"id":           validateUser.ID,
		"username":     validateUser.Username,
	}

	return response.ResponseOKWithData(c, userData)
}

func (h *UserHandler) LogOut(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:        "jwt",
		Value:       "",
		Path:        "",
		Domain:      "",
		MaxAge:      -1,
		Expires:     time.Time{},
		Secure:      false,
		HTTPOnly:    true,
		SameSite:    "",
		SessionOnly: false,
	})
	return c.Status(fiber.StatusOK).JSON(map[string]string{"message": "logout successful"})
}

func (h *UserHandler) GetAllUser(c *fiber.Ctx) error {
	result, err := h.userService.GetListUser()
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	if result == nil {
		result = &[]model.UserViewModel{}
	}

	return response.ResponseOKWithData(c, result)
}

func (h *UserHandler) GetDetailUser(c *fiber.Ctx) error {
	userId, err := strconv.Atoi(c.Params("user_id"))
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}

	result, err := h.userService.GetDetailUser(userId)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	if result == nil {
		result = &model.UserViewModel{}
	}

	return response.ResponseOKWithData(c, result)
}
