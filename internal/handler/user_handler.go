package handler

import (
	"downloader_gochat/configs"
	"downloader_gochat/internal/service"
	"downloader_gochat/model"
	"downloader_gochat/pkg/response"
	"downloader_gochat/util"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type IUserHandler interface {
	RegisterUser(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	GetToken(c *fiber.Ctx) error
	LogOut(c *fiber.Ctx) error
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
//------------------------------------------

// RegisterUser godoc
//
//	@Summary		Register a new user
//	@Description	Register a new user with the provided credentials
//	@Description	Unlike the main server, this one doesn't handle ip detection and ip location
//	@Description	Also detect multiple login on same device as new device login, can be handled on client side with adding 'deviceInfo.fingerprint'
//	@Description	Also doesn't handle and send emails
//	@Tags			User
//	@Param			noCookie	query		bool					true	"return refreshToken in response body instead of saving in cookie"
//	@Param			user		body		model.RegisterViewModel	true	"User object"
//	@Success		200			{object}	model.UserViewModel
//	@Failure		400			{object}	response.ResponseErrorModel
//	@Router			/v1/user/signup [post]
func (h *UserHandler) RegisterUser(c *fiber.Ctx) error {
	var registerVM model.RegisterViewModel
	err := c.BodyParser(&registerVM)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	noCookie := c.QueryBool("noCookie", false)

	registerUserError := registerVM.Validate()
	if len(registerUserError) > 0 {
		return response.ResponseError(c, strings.Join(registerUserError, ", "), fiber.StatusBadRequest)
	}
	registerVM.Normalize()

	// ip := c.IP()
	// u, err := url.Parse(ip)

	result, err := h.userService.SignUp(&registerVM)
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

	if !noCookie {
		c.Cookie(&fiber.Cookie{
			Name:        "refreshToken",
			Value:       result.Token.RefreshToken,
			Path:        "/",
			Expires:     time.Now().Add(time.Duration(configs.GetConfigs().RefreshTokenExpireDay) * 24 * time.Hour),
			Secure:      true,
			HTTPOnly:    true,
			SameSite:    "none",
			SessionOnly: false,
		})
		result.Token.RefreshToken = ""
	}

	return response.ResponseCreated(c, result)
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Login with provided credentials
//	@Tags			User
//	@Param			noCookie	query		bool					true	"return refreshToken in response body instead of saving in cookie"
//	@Param			user	body		model.LoginViewModel	true	"User object"
//	@Success		200		{object}	model.UserViewModel
//	@Failure		400		{object}	response.ResponseErrorModel
//	@Router			/v1/user/login [post]
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var loginVM model.LoginViewModel
	err := c.BodyParser(&loginVM)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	noCookie := c.QueryBool("noCookie", false)

	loginUserError := loginVM.Validate()
	if len(loginUserError) > 0 {
		return response.ResponseError(c, strings.Join(loginUserError, ", "), fiber.StatusBadRequest)
	}
	loginVM.Normalize()

	// ip := c.IP()
	// u, err := url.Parse(ip)

	result, err := h.userService.LoginUser(&loginVM)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	if result == nil {
		return response.ResponseError(c, "Cannot find user", fiber.StatusNotFound)
	}

	if !noCookie {
		c.Cookie(&fiber.Cookie{
			Name:        "refreshToken",
			Value:       result.Token.RefreshToken,
			Path:        "/",
			Expires:     time.Now().Add(time.Duration(configs.GetConfigs().RefreshTokenExpireDay) * 24 * time.Hour),
			Secure:      true,
			HTTPOnly:    true,
			SameSite:    "none",
			SessionOnly: false,
		})
		result.Token.RefreshToken = ""
	}

	return response.ResponseOKWithData(c, result)
}

// GetToken godoc
//
//	@Summary		Get Token
//	@Description	Get new Tokens, also return `refreshToken`
//	@Tags			User
//	@Param			noCookie	query		bool					true	"return refreshToken in response body instead of saving in cookie"
//	@Param			profileImages	query		bool					true	"also return profile images, slower response"
//	@Param			user	body		model.DeviceInfo	true	"Device Info"
//	@Success		200		{object}	model.UserViewModel
//	@Failure		400,401		{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/getToken [put]
func (h *UserHandler) GetToken(c *fiber.Ctx) error {
	var deviceInfo model.DeviceInfo
	err := c.BodyParser(&deviceInfo)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	noCookie := c.QueryBool("noCookie", false)
	addProfileImages := c.QueryBool("profileImages", false)

	deviceInfoError := deviceInfo.Validate()
	if len(deviceInfoError) > 0 {
		return response.ResponseError(c, strings.Join(deviceInfoError, ", "), fiber.StatusBadRequest)
	}
	deviceInfo.Normalize()

	// ip := c.IP()
	// u, err := url.Parse(ip)

	refreshToken := c.Locals("refreshToken").(string)
	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	result, token, err := h.userService.GetToken(&deviceInfo, refreshToken, jwtUserData, addProfileImages)

	if !noCookie && token != nil {
		c.Cookie(&fiber.Cookie{
			Name:        "refreshToken",
			Value:       token.RefreshToken,
			Path:        "/",
			Expires:     time.Now().Add(time.Duration(configs.GetConfigs().RefreshTokenExpireDay) * 24 * time.Hour),
			Secure:      true,
			HTTPOnly:    true,
			SameSite:    "none",
			SessionOnly: false,
		})
		if result != nil {
			result.Token.RefreshToken = ""
		}
	}

	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	if result == nil {
		return response.ResponseError(c, "Cannot find device", fiber.StatusNotFound)
	}

	return response.ResponseOKWithData(c, result)
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