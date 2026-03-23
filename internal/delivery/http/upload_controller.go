package http

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UploadController interface {
	UploadImage(ctx fiber.Ctx) error
	GetFile(ctx fiber.Ctx) error
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
	allowedExt := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowedExt[ext] {
		c.Log.Warnf("Invalid file extension : %s", ext)
		return fiber.NewError(fiber.StatusBadRequest, "Format file tidak didukung. Gunakan: jpg, jpeg, png, webp")
	}

	maxSize := int64(5 * 1024 * 1024)
	if file.Size > maxSize {
		return fiber.NewError(fiber.StatusBadRequest, "Ukuran file maksimal 5MB")
	}

	src, err := file.Open()
	if err != nil {
		c.Log.Warnf("Failed to open file : %+v", err)
		return fiber.ErrInternalServerError
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		c.Log.Warnf("Failed to decode image : %+v", err)
		return fiber.NewError(fiber.StatusBadRequest, "File gambar rusak atau tidak valid")
	}

	subDir := time.Now().Format("2006/01")
	dir := filepath.Join(c.UploadDir, subDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.Log.Warnf("Failed create upload directory : %+v", err)
		return fiber.ErrInternalServerError
	}

	filename := fmt.Sprintf("%s.webp", uuid.New().String())
	filePath := filepath.Join(dir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		c.Log.Warnf("Failed to create destination file : %+v", err)
		return fiber.ErrInternalServerError
	}
	defer out.Close()

	if err := webp.Encode(out, img, &webp.Options{Lossless: false, Quality: 80}); err != nil {
		c.Log.Warnf("Failed to encode webp : %+v", err)
		return fiber.ErrInternalServerError
	}

	fileInfo, _ := out.Stat()
	url := fmt.Sprintf("/uploads/%s/%s", subDir, filename)

	res := model.WebResponse[any]{
		Data: fiber.Map{
			"url":      url,
			"filename": filename,
			"old_size": file.Size,
			"new_size": fileInfo.Size(),
		},
	}
	
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *uploadControllerImpl) GetFile(ctx fiber.Ctx) error {
	return ctx.SendFile("./uploads/"+ctx.Params("*"))
}