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

// RegisterUser godoc
//
//	@Summary		Register a new user
//	@Description	Register a new user with the provided credentials
//	@Tags			User
//	@Param			user	body		model.RegisterViewModel	true	"User object"
//	@Success		201		{object}	model.UserViewModel
//	@Failure		400		{object}	response.ResponseErrorModel
//	@Router			/v1/user/signup [post]
func (h *UserHandler) RegisterUser(c *fiber.Ctx) error {
	var registerVM model.RegisterViewModel
	err := c.BodyParser(&registerVM)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	registerUserError := registerVM.Validate()
	if len(registerUserError) > 0 {
		return response.ResponseError(c, registerUserError, fiber.StatusBadRequest)
	}

	result, err := h.userService.CreateUser(&registerVM)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	if result.UserId == 0 {
		if result.Username == registerVM.Username {
			return response.ResponseError(c, "username already exist", fiber.StatusConflict)
		} else {
			return response.ResponseError(c, "email already exist", fiber.StatusConflict)
		}
	}

	return response.ResponseCreated(c, result)
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Login with provided credentials
//	@Tags			User
//	@Param			user	body		model.LoginViewModel	true	"User object"
//	@Success		200		{object}	model.UserViewModel
//	@Failure		400		{object}	response.ResponseErrorModel
//	@Router			/v1/user/login [post]
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
	token, err := util.CreateJwtToken(validateUser.UserId, validateUser.Username)
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
		"userid":       validateUser.UserId,
		"username":     validateUser.Username,
	}

	return response.ResponseOKWithData(c, userData)
}

// LogOut godoc
//
//	@Summary		Logout
//	@Description	Logout the currently logged in user
//	@Tags			User
//	@Success		200	{object}	model.UserViewModel
//	@Failure		401	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/logout [get]
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

// GetAllUser godoc
//
//	@Summary		Get all users
//	@Description	Get a list of all users
//	@Tags			User
//	@Success		200	{object}	model.UserViewModel
//	@Failure		401	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/ [get]
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

// GetDetailUser godoc
//
//	@Summary				Get user details
//	@Description			Get details of a specific user
//	@Tags					User
//	@Security				BearerAuth
//	@Param					user_id			path		string	true	"User UserId"
//	@Success				200				{object}	[]model.UserViewModel
//	@Failure				400,401,403		{object}	response.ResponseErrorModel
//	@Router					/v1/user/{user_id} [get]
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
