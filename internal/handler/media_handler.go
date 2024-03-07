package handler

import (
	"downloader_gochat/configs"
	"downloader_gochat/internal/service"
	"downloader_gochat/model"
	"downloader_gochat/pkg/response"
	"downloader_gochat/util"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type IMediaHandler interface {
	UploadFile(c *fiber.Ctx) error
}

type MediaHandler struct {
	mediaService service.IMediaService
}

func NewMediaHandler(mediaService service.IMediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

//------------------------------------------
//------------------------------------------

// UploadFile godoc
//
//	@Summary		Upload File
//	@Description	upload and share media files in chats
//	@Tags			User-Chat
//	@Param			user		body		model.UploadMediaReq	true	"upload file data"
//	@Success		200			{object}	model.MediaFile
//	@Failure		400,401,404	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/user/media/upload [post]
func (m *MediaHandler) UploadFile(c *fiber.Ctx) error {
	var req model.UploadMediaReq
	err := c.BodyParser(&req)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	file, err := c.FormFile("mediaFile")
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
	}

	contentType := file.Header["Content-Type"][0]

	//---------------------------------------------
	dbconfig := configs.GetDbConfigs()
	if file.Size == 0 {
		return response.ResponseError(c, "File is empty", fiber.StatusBadRequest)
	}

	if file.Size > dbconfig.MediaFileSizeLimit*1024*1024 {
		return response.ResponseError(c, fmt.Sprintf("File size exceeds the limit (%vmb)", dbconfig.MediaFileSizeLimit), fiber.StatusBadRequest)
	}

	allowedExts := strings.Split(dbconfig.MediaFileExtensionLimit, ",")
	ext := filepath.Ext(file.Filename)
	validExtension := false
	for _, allowedExt := range allowedExts {
		if ext == "."+strings.TrimSpace(allowedExt) {
			validExtension = true
			break
		}
	}
	if !validExtension {
		return response.ResponseError(c, "Invalid file extension", fiber.StatusBadRequest)
	}
	ext = strings.Split(contentType, "/")[1]
	validExtension = false
	for _, allowedExt := range allowedExts {
		if ext == strings.TrimSpace(allowedExt) {
			validExtension = true
			break
		}
	}
	if !validExtension {
		return response.ResponseError(c, "Invalid file extension", fiber.StatusBadRequest)
	}
	//---------------------------------------------

	buffer, err := file.Open()
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	defer buffer.Close()

	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	result, err := m.mediaService.UploadFile(jwtUserData.UserId, &req, contentType, file.Size, file.Filename, buffer)
	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return response.ResponseError(c, "Receiver User Not Found", fiber.StatusNotFound)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, result)
}
