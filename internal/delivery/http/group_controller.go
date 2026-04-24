package http

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type GroupController interface {
	CreateGroup(ctx fiber.Ctx) error
	GetAllGroups(ctx fiber.Ctx) error
	GetGroupDetail(ctx fiber.Ctx) error
	UpdateGroup(ctx fiber.Ctx) error
	DeleteGroup(ctx fiber.Ctx) error
	InviteMembersByEmail(ctx fiber.Ctx) error
	GenerateInviteLink(ctx fiber.Ctx) error
	JoinGroup(ctx fiber.Ctx) error
	RemoveMember(ctx fiber.Ctx) error
	GetGroupMembers(ctx fiber.Ctx) error
}

type groupControllerImpl struct {
	GroupUseCase usecase.GroupUseCase
	Log          *logrus.Logger
}

func NewGroupController(groupUseCase usecase.GroupUseCase, log *logrus.Logger) GroupController {
	return &groupControllerImpl{
		GroupUseCase: groupUseCase,
		Log:          log,
	}
}

func (c *groupControllerImpl) CreateGroup(ctx fiber.Ctx) error {
	req := new(model.GroupCreateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	userId := ctx.Locals("userId").(string)
	group, err := c.GroupUseCase.CreateGroup(ctx.Context(), req, userId)
	if err != nil {
		c.Log.Warnf("Failed create group: %+v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.GroupResponse]{Data: group})
}

func (c *groupControllerImpl) GetAllGroups(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	userId := ctx.Locals("userId").(string)
	groups, _, err := c.GroupUseCase.GetAllGroups(ctx.Context(), userId, page, size, search)
	if err != nil {
		c.Log.Warnf("Failed get all groups: %v", err)
		return err
	}

	paging := &model.PageMetaData{
		Page: page,
		Size: size,
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[[]*model.GroupResponse]{
		Data:   groups,
		Paging: paging,
	})
}

func (c *groupControllerImpl) GetGroupDetail(ctx fiber.Ctx) error {
	groupId := ctx.Params("id")
	userId := ctx.Locals("userId").(string)

	group, err := c.GroupUseCase.GetGroupDetail(ctx.Context(), groupId, userId)
	if err != nil {
		c.Log.Warnf("Failed get group detail: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.GroupDetailResponse]{Data: group})
}

func (c *groupControllerImpl) UpdateGroup(ctx fiber.Ctx) error {
	req := new(model.GroupUpdateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	groupId := ctx.Params("id")
	userId := ctx.Locals("userId").(string)

	group, err := c.GroupUseCase.UpdateGroup(ctx.Context(), groupId, req, userId)
	if err != nil {
		c.Log.Warnf("Failed update group: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.GroupResponse]{Data: group})
}

func (c *groupControllerImpl) DeleteGroup(ctx fiber.Ctx) error {
	groupId := ctx.Params("id")
	userId := ctx.Locals("userId").(string)

	err := c.GroupUseCase.DeleteGroup(ctx.Context(), groupId, userId)
	if err != nil {
		c.Log.Warnf("Failed delete group: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[string]{Data: "Group deleted successfully"})
}

func (c *groupControllerImpl) InviteMembersByEmail(ctx fiber.Ctx) error {
	req := new(model.GroupInviteEmailRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	groupId := ctx.Params("id")
	userId := ctx.Locals("userId").(string)

	err := c.GroupUseCase.InviteMembersByEmail(ctx.Context(), groupId, req, userId)
	if err != nil {
		c.Log.Warnf("Failed invite members by email: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[string]{Data: "Members invited successfully"})
}

func (c *groupControllerImpl) GenerateInviteLink(ctx fiber.Ctx) error {
	groupId := ctx.Params("id")
	userId := ctx.Locals("userId").(string)

	inviteInfo, err := c.GroupUseCase.GenerateInviteLink(ctx.Context(), groupId, userId)
	if err != nil {
		c.Log.Warnf("Failed generate invite link: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.GroupInviteResponse]{Data: inviteInfo})
}

func (c *groupControllerImpl) JoinGroup(ctx fiber.Ctx) error {
	req := new(model.GroupJoinRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	userId := ctx.Locals("userId").(string)

	err := c.GroupUseCase.JoinGroup(ctx.Context(), req, userId)
	if err != nil {
		c.Log.Warnf("Failed join group: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[string]{Data: "Successfully joined group"})
}

func (c *groupControllerImpl) RemoveMember(ctx fiber.Ctx) error {
	groupId := ctx.Params("id")
	memberIdStr := ctx.Params("member_id")
	memberId, err := strconv.Atoi(memberIdStr)
	if err != nil {
		return fiber.ErrBadRequest
	}
	userId := ctx.Locals("userId").(string)

	err = c.GroupUseCase.RemoveMember(ctx.Context(), groupId, memberId, userId)
	if err != nil {
		c.Log.Warnf("Failed remove member: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[string]{Data: "Member removed successfully"})
}

func (c *groupControllerImpl) GetGroupMembers(ctx fiber.Ctx) error {
	groupId := ctx.Params("id")
	userId := ctx.Locals("userId").(string)

	members, err := c.GroupUseCase.GetGroupMembers(ctx.Context(), groupId, userId)
	if err != nil {
		c.Log.Warnf("Failed get group members: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[[]model.GroupMemberResponse]{Data: members})
}
