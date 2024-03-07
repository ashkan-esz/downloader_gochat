package handler

import (
	"downloader_gochat/configs"
	"downloader_gochat/internal/service"
	"downloader_gochat/model"
	"downloader_gochat/pkg/response"
	"downloader_gochat/util"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type IUserHandler interface {
	RegisterUser(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	GetToken(c *fiber.Ctx) error
	LogOut(c *fiber.Ctx) error
	SetNotifToken(c *fiber.Ctx) error
	FollowUser(c *fiber.Ctx) error
	UnFollowUser(c *fiber.Ctx) error
	GetUserFollowers(c *fiber.Ctx) error
	GetUserFollowings(c *fiber.Ctx) error
	GetUserNotifications(c *fiber.Ctx) error
	GetUserSettings(c *fiber.Ctx) error
	UpdateUserSettings(c *fiber.Ctx) error
	UpdateUserFavoriteGenres(c *fiber.Ctx) error
	GetActiveSessions(c *fiber.Ctx) error
	GetUserProfile(c *fiber.Ctx) error
	EditUserProfile(c *fiber.Ctx) error
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
//	@Description	Device detection can be improved on client side with adding 'deviceInfo.fingerprint'
//	@Tags			User-Auth
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

	ip := c.IP()
	ips := c.IPs()
	if len(ips) > 0 {
		ip = ips[len(ips)-1]
	}

	result, err := h.userService.SignUp(&registerVM, ip)
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
//	@Tags			User-Auth
//	@Param			noCookie	query		bool					true	"return refreshToken in response body instead of saving in cookie"
//	@Param			user		body		model.LoginViewModel	true	"User object"
//	@Success		200			{object}	model.UserViewModel
//	@Failure		400			{object}	response.ResponseErrorModel
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

	ip := c.IP()
	ips := c.IPs()
	if len(ips) > 0 {
		ip = ips[len(ips)-1]
	}

	result, err := h.userService.LoginUser(&loginVM, ip)
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
//	@Tags			User-Auth
//	@Param			noCookie		query		bool				true	"return refreshToken in response body instead of saving in cookie"
//	@Param			profileImages	query		bool				true	"also return profile images, slower response"
//	@Param			user			body		model.DeviceInfo	true	"Device Info"
//	@Success		200				{object}	model.UserViewModel
//	@Failure		400,401			{object}	response.ResponseErrorModel
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

	ip := c.IP()
	ips := c.IPs()
	if len(ips) > 0 {
		ip = ips[len(ips)-1]
	}

	refreshToken := c.Locals("refreshToken").(string)
	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	result, token, err := h.userService.GetToken(&deviceInfo, refreshToken, jwtUserData, addProfileImages, ip)

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
//	@Description	Logout user, return accessToken as empty string and also reset/remove refreshToken cookie if use in browser
//	@Description	.in other environments reset refreshToken from client after successful logout.
//	@Tags			User-Auth
//	@Success		200
//	@Failure		401,403	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/logout [put]
func (h *UserHandler) LogOut(c *fiber.Ctx) error {
	refreshToken := c.Locals("refreshToken").(string)
	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	err := h.userService.LogOut(c, jwtUserData, refreshToken)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.ResponseError(c, "Cannot find device", fiber.StatusNotFound)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	c.Cookie(&fiber.Cookie{
		Name:        "refreshToken",
		Value:       "",
		Path:        "/",
		MaxAge:      -1,
		Expires:     time.Now(),
		Secure:      true,
		HTTPOnly:    true,
		SameSite:    "none",
		SessionOnly: false,
	})

	return response.ResponseOK(c, "")
}

// SetNotifToken godoc
//
//	@Summary		Notification Token
//	@Description	send device token as Notification token
//	@Tags			User-Notifications
//	@Param			notifToken	path	string	true	"notifToken"
//	@Success		200
//	@Failure		401,403,404	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/setNotifToken [put]
func (h *UserHandler) SetNotifToken(c *fiber.Ctx) error {
	refreshToken := c.Locals("refreshToken").(string)
	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	notifToken := c.Params("notifToken", "")
	err := h.userService.SetNotifToken(jwtUserData, refreshToken, notifToken)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.ResponseError(c, "Cannot find device", fiber.StatusNotFound)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return response.ResponseOK(c, "")
}

//------------------------------------------
//------------------------------------------

// FollowUser godoc
//
//	@Summary		Follow User
//	@Description	Add followId user to your following list
//	@Tags			User-Follow
//	@Param			followId		path		integer	true	"id on the user want to follow"
//	@Success		200				{object}	response.ResponseOKModel
//	@Failure		400,404,409,500	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/follow/:followId [post]
func (h *UserHandler) FollowUser(c *fiber.Ctx) error {
	followId, err := c.ParamsInt("followId", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	err = h.userService.FollowUser(jwtUserData, int64(followId))
	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return response.ResponseError(c, response.UserNotFound, fiber.StatusNotFound)
		} else if err.Error() == "duplicated key not allowed" {
			return response.ResponseError(c, response.AlreadyFollowed, fiber.StatusConflict)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return response.ResponseOK(c, "")
}

// UnFollowUser godoc
//
//	@Summary		UnFollow User
//	@Description	Remove followId user from users following list
//	@Tags			User-Follow
//	@Param			followId	path		integer	true	"id on the user want to unfollow"
//	@Success		200			{object}	response.ResponseOKModel
//	@Failure		400,404,500	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/unfollow/:followId [delete]
func (h *UserHandler) UnFollowUser(c *fiber.Ctx) error {
	followId, err := c.ParamsInt("followId", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	err = h.userService.UnFollowUser(jwtUserData, int64(followId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.ResponseError(c, response.UserNotFound, fiber.StatusNotFound)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return response.ResponseOK(c, "")
}

// GetUserFollowers godoc
//
//	@Summary		Followers
//	@Description	get user followers
//	@Tags			User-Follow
//	@Param			userId	path		integer	true	"id of user"
//	@Param			skip	path		integer	true	"skip"
//	@Param			limit	path		integer	true	"limit"
//	@Success		200		{object}	model.FollowUserDataModel
//	@Failure		400,500	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/followers/:userId/:skip/:limit [get]
func (h *UserHandler) GetUserFollowers(c *fiber.Ctx) error {
	userId, err := c.ParamsInt("userId", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}
	skip, err := c.ParamsInt("skip", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}
	limit, err := c.ParamsInt("limit", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}

	result, err := h.userService.GetUserFollowers(int64(userId), skip, limit)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, result)
}

// GetUserFollowings godoc
//
//	@Summary		Followings
//	@Description	get user followings
//	@Tags			User-Follow
//	@Param			userId	path		integer	true	"id of user"
//	@Param			skip	path		integer	true	"skip"
//	@Param			limit	path		integer	true	"limit"
//	@Success		200		{object}	model.FollowUserDataModel
//	@Failure		400,500	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/followings/:userId/:skip/:limit [get]
func (h *UserHandler) GetUserFollowings(c *fiber.Ctx) error {
	userId, err := c.ParamsInt("userId", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}
	skip, err := c.ParamsInt("skip", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}
	limit, err := c.ParamsInt("limit", 0)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}

	result, err := h.userService.GetUserFollowings(int64(userId), skip, limit)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, result)
}

//------------------------------------------
//------------------------------------------

// GetUserSettings godoc
//
//	@Summary		Get User Settings
//	@Description	Returns user settings for movies, downloadLinks and notifications.
//	@Tags			User-Setting
//	@Param			settingName	path		model.SettingName	true	"name of setting"
//	@Success		200			{object}	model.UserSettingsRes
//	@Failure		400,404,500	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/userSettings/:settingName [get]
func (h *UserHandler) GetUserSettings(c *fiber.Ctx) error {
	settingName := c.Params("settingName", "all")

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	result, err := h.userService.GetUserSettings(jwtUserData.UserId, model.SettingName(settingName))
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, result)
}

// UpdateUserSettings godoc
//
//	@Summary		Change user settings based on settingName.
//	@Description	Change user settings based on settingName.
//	@Tags			User-Setting
//	@Param			settingName				path		model.SettingName			true	"name of setting"
//	@Param			downloadLinksSettings	body		model.DownloadLinksSettings	false	"new setting values"
//	@Param			notificationSettings	body		model.NotificationSettings	false	"new setting values"
//	@Param			movieSettings			body		model.MovieSettings			false	"new setting values"
//	@Success		200						{object}	response.ResponseOKModel
//	@Failure		400,404,500				{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/updateUserSettings/:settingName [put]
func (h *UserHandler) UpdateUserSettings(c *fiber.Ctx) error {
	settingName := c.Params("settingName", "")

	settings := model.UserSettingsRes{
		DownloadLinksSettings: nil,
		NotificationSettings:  nil,
		MovieSettings:         nil,
	}

	err := c.BodyParser(&settings)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	switch settingName {
	case string(model.NotificationSettingsName):
		if settings.NotificationSettings == nil {
			return response.ResponseError(c, "Invalid fields for notificationSettings", fiber.StatusBadRequest)
		}
	case string(model.DownloadSettingsName):
		if settings.DownloadLinksSettings == nil {
			return response.ResponseError(c, "Invalid fields for downloadLinksSettings", fiber.StatusBadRequest)
		}
	case string(model.MovieSettingsName):
		if settings.MovieSettings == nil {
			return response.ResponseError(c, "Invalid fields for movieSettings", fiber.StatusBadRequest)
		}
	default:
		return response.ResponseError(c, "Invalid settingName", fiber.StatusBadRequest)
	}

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	err = h.userService.UpdateUserSettings(jwtUserData.UserId, model.SettingName(settingName), &settings)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOK(c, "")
}

//------------------------------------------
//------------------------------------------

// UpdateUserFavoriteGenres godoc
//
//	@Summary		Update Favorite Genres
//	@Description	maximum number of genres is 6, (error code 409).
//	@Tags			User
//	@Param			genres			path		string	true	"Array of String joined by '-', Example: action-sci_fi-drama"
//	@Success		200				{object}	response.ResponseOKModel
//	@Failure		400,404,409,500	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/UpdateUserFavoriteGenres/:genres [put]
func (h *UserHandler) UpdateUserFavoriteGenres(c *fiber.Ctx) error {
	genres := c.Params("genres", "")
	if strings.ToLower(genres) == ":genres" {
		return response.ResponseError(c, "Invalid value for genres", fiber.StatusBadRequest)
	}

	genresArray := strings.Split(genres, "-")
	for i := range genresArray {
		genresArray[i] = strings.ReplaceAll(genresArray[i], "_", "-")
		genresArray[i] = strings.TrimSpace(strings.ToLower(genresArray[i]))
	}
	genresArray = slices.Compact(genresArray)

	if len(genresArray) > 6 {
		return response.ResponseError(c, response.ExceedGenres, fiber.StatusConflict)
	}

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	err := h.userService.UpdateUserFavoriteGenres(jwtUserData.UserId, genresArray)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.ResponseError(c, response.UserNotFound, fiber.StatusNotFound)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOK(c, "")
}

// GetActiveSessions godoc
//
//	@Summary		Active Sessions
//	@Description	Return users current session and other active sections.
//	@Tags			User
//	@Success		200		{object}	model.ActiveSessionRes
//	@Failure		404,500	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/activeSessions [get]
func (h *UserHandler) GetActiveSessions(c *fiber.Ctx) error {
	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	refreshToken := c.Locals("refreshToken").(string)
	result, err := h.userService.GetActiveSessions(jwtUserData.UserId, refreshToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.ResponseError(c, response.SessionNotFound, fiber.StatusNotFound)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, result)
}

//------------------------------------------
//------------------------------------------

// GetUserProfile godoc
//
//	@Summary		Profile Data
//	@Description	Return users profile data. if dont provide userId, return current user profile
//	@Tags			User
//	@Param			userId						query		integer	false	"userId"
//	@Param			loadSettings				query		bool	false	"loadSettings"
//	@Param			loadFollowersCount			query		bool	false	"loadFollowersCount"
//	@Param			loadDevice					query		bool	false	"loadDevice"
//	@Param			loadProfileImages			query		bool	false	"loadProfileImages"
//	@Param			loadComputedFavoriteGenres	query		bool	false	"loadComputedFavoriteGenres"
//	@Success		200							{object}	model.UserProfileRes
//	@Failure		404,500						{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/profile [get]
func (h *UserHandler) GetUserProfile(c *fiber.Ctx) error {
	userId := int64(c.QueryInt("userId", 0))
	loadSettings := c.QueryBool("loadSettings", false)
	loadFollowersCount := c.QueryBool("loadFollowersCount", false)
	loadDevice := c.QueryBool("loadDevice", false)
	loadProfileImages := c.QueryBool("loadProfileImages", false)
	loadComputedFavoriteGenres := c.QueryBool("loadComputedFavoriteGenres", false)

	isSelfProfile := false
	refreshToken := ""
	if userId <= 0 {
		jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
		userId = jwtUserData.UserId
		isSelfProfile = true
		if loadDevice {
			refreshToken = c.Locals("refreshToken").(string)
		}
	}

	requestParams := model.UserProfileReq{
		UserId:                     userId,
		IsSelfProfile:              isSelfProfile,
		LoadSettings:               loadSettings,
		LoadFollowersCount:         loadFollowersCount,
		LoadProfileImages:          loadProfileImages,
		LoadComputedFavoriteGenres: loadComputedFavoriteGenres,
		RefreshToken:               refreshToken,
	}
	result, err := h.userService.GetUserProfile(&requestParams)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.ResponseError(c, response.SessionNotFound, fiber.StatusNotFound)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, result)
}

// EditUserProfile godoc
//
//	@Summary		Edit Profile
//	@Description	Edit profile data.
//	@Tags			User
//	@Param			user				body		model.EditProfileReq	true	"update fields"
//	@Success		200					{object}	response.ResponseOKModel
//	@Failure		400,401,404,409,500	{object}	response.ResponseErrorModel
//	@Router			/v1/user/editUserProfile [post]
func (h *UserHandler) EditUserProfile(c *fiber.Ctx) error {
	var editFields model.EditProfileReq
	err := c.BodyParser(&editFields)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	result, err := h.userService.EditUserProfile(jwtUserData.UserId, &editFields)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	if result == nil {
		return response.ResponseError(c, response.UserNotFound, fiber.StatusNotFound)
	}
	if result.UserId == 0 {
		if result.Username == strings.ToLower(editFields.Username) {
			return response.ResponseError(c, response.UsernameAlreadyExist, fiber.StatusConflict)
		} else {
			return response.ResponseError(c, response.EmailAlreadyExist, fiber.StatusConflict)
		}
	}

	return response.ResponseOK(c, "")
}
