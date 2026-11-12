package http

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/repository"
)

type NotificationController struct {
	Repo *repository.NotificationRepository
}

func notificationMember(ctx fiber.Ctx) (int, error) {
	claims, ok := ctx.Locals("user").(*model.Claims)
	if !ok || claims == nil {
		return 0, fiber.ErrUnauthorized
	}
	if claims.Role != "member" {
		return 0, fiber.ErrForbidden
	}
	var id string
	switch d := claims.Details.(type) {
	case model.MemberDetails:
		id = d.MemberId
	case *model.MemberDetails:
		if d != nil {
			id = d.MemberId
		}
	case map[string]interface{}:
		id, _ = d["member_id"].(string)
	}
	member, err := strconv.Atoi(id)
	if err != nil || member <= 0 {
		return 0, fiber.ErrForbidden
	}
	return member, nil
}
func (c *NotificationController) List(ctx fiber.Ctx) error {
	member, err := notificationMember(ctx)
	if err != nil {
		return err
	}
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "20"))
	page, size = model.NormalizePaging(page, size)
	unread, err := strconv.ParseBool(ctx.Query("unread_only", "false"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	rows, total, err := c.Repo.List(ctx.Context(), member, page, size, unread)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return ctx.JSON(model.WebResponse[[]model.Notification]{Data: rows, Paging: model.NewPageMetaData(page, size, total)})
}
func (c *NotificationController) Count(ctx fiber.Ctx) error {
	member, err := notificationMember(ctx)
	if err != nil {
		return err
	}
	n, err := c.Repo.Count(ctx.Context(), member)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return ctx.JSON(model.WebResponse[map[string]int]{Data: map[string]int{"count": n}})
}
func (c *NotificationController) Read(ctx fiber.Ctx) error {
	member, err := notificationMember(ctx)
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(ctx.Params("notification_id"), 10, 64)
	if err != nil || id <= 0 {
		return fiber.ErrNotFound
	}
	n, err := c.Repo.Read(ctx.Context(), member, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return fiber.ErrNotFound
	}
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return ctx.JSON(model.WebResponse[model.Notification]{Data: n})
}
