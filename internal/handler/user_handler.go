package handler

import (
	"downloader_gochat/internal/service"
	"downloader_gochat/model"
	"downloader_gochat/pkg/response"
	"downloader_gochat/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type IUserHandler interface {
	RegisterUser(c *gin.Context)
	Login(c *gin.Context)
	LogOut(c *gin.Context)
	GetAllUser(c *gin.Context)
	GetDetailUser(c *gin.Context)
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

//todo : replace gin with fiber

func (h *UserHandler) RegisterUser(c *gin.Context) {
	var registerVM model.RegisterViewModel
	err := c.ShouldBindJSON(&registerVM)
	if err != nil {
		response.ResponseError(c, err.Error(), http.StatusInternalServerError)
		return
	}

	registerUserError := registerVM.Validate()
	if len(registerUserError) > 0 {
		response.ResponseCustomError(c, registerUserError, http.StatusBadRequest)
		return
	}

	result, err := h.userService.CreateUser(&registerVM)
	if err != nil {
		response.ResponseError(c, err.Error(), http.StatusInternalServerError)
		return
	}

	if result.ID == 0 {
		if result.Username == registerVM.Username {
			response.ResponseCustomError(c, "username already exist", http.StatusConflict)
		} else {
			response.ResponseCustomError(c, "email already exist", http.StatusConflict)
		}
		return
	}

	response.ResponseCreated(c, result)
}

func (h *UserHandler) Login(c *gin.Context) {
	var loginVM model.LoginViewModel
	err := c.ShouldBindJSON(&loginVM)
	if err != nil {
		response.ResponseError(c, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	validateUser, err := h.userService.LoginUser(&loginVM)
	if err != nil {
		response.ResponseError(c, err.Error(), http.StatusInternalServerError)
		return
	}

	if validateUser == nil {
		validateUser = &model.UserViewModel{}
	}

	// Generete JWT
	token, err := util.CreateJwtToken(validateUser.ID, validateUser.Username)
	if err != nil {
		response.ResponseError(c, err.Error(), http.StatusInternalServerError)
		return
	}

	c.SetCookie("jwt", token.AccessToken, 3600, "/", "localhost", false, true)

	userData := map[string]interface{}{
		"access_token": token.AccessToken,
		"expired":      token.ExpireAt,
		"id":           validateUser.ID,
		"username":     validateUser.Username,
	}

	response.ResponseOKWithData(c, userData)
}

func (h *UserHandler) LogOut(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *UserHandler) GetAllUser(c *gin.Context) {
	result, err := h.userService.GetListUser()
	if err != nil {
		response.ResponseError(c, err.Error(), http.StatusInternalServerError)
		return
	}

	if result == nil {
		result = &[]model.UserViewModel{}
	}

	response.ResponseOKWithData(c, result)
}

func (h *UserHandler) GetDetailUser(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		response.ResponseError(c, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userService.GetDetailUser(userId)
	if err != nil {
		response.ResponseError(c, err.Error(), http.StatusInternalServerError)
		return
	}

	if result == nil {
		result = &model.UserViewModel{}
	}

	response.ResponseOKWithData(c, result)
}
