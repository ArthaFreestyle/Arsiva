package http

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UploadController interface {
	UploadImage(ctx fiber.Ctx) error
}

type uploadControllerImpl struct {
	Log       *logrus.Logger
	UploadDir string
}

func NewUploadController(log *logrus.Logger, uploadDir string) UploadController {
	return &uploadControllerImpl{
		Log:       log,
		UploadDir: uploadDir,
	}
}

func (c *uploadControllerImpl) UploadImage(ctx fiber.Ctx) error {
	file, err := ctx.FormFile("image")
	if err != nil {
		c.Log.Warnf("Failed get file from form : %+v", err)
		return fiber.ErrBadRequest
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExt := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
	if !allowedExt[ext] {
		c.Log.Warnf("Invalid file extension : %s", ext)
		return fiber.NewError(fiber.StatusBadRequest, "Format file tidak didukung. Gunakan: jpg, jpeg, png, webp, gif")
	}

	maxSize := int64(5 * 1024 * 1024) // 5MB
	if file.Size > maxSize {
		return fiber.NewError(fiber.StatusBadRequest, "Ukuran file maksimal 5MB")
	}

	subDir := time.Now().Format("2006/01")
	dir := filepath.Join(c.UploadDir, subDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.Log.Warnf("Failed create upload directory : %+v", err)
		return fiber.ErrInternalServerError
	}

	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(dir, filename)

	if err := ctx.SaveFile(file, filePath); err != nil {
		c.Log.Warnf("Failed save file : %+v", err)
		return fiber.ErrInternalServerError
	}

	url := fmt.Sprintf("/uploads/%s/%s", subDir, filename)

	res := model.WebResponse[	any]{
		Data: fiber.Map{
			"url":      url,
			"filename": file.Filename,
			"size":     file.Size,
		},
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}
