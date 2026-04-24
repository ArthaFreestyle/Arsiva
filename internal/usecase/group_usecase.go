package usecase

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
)

type GroupUseCase interface {
	// CRUD
	CreateGroup(ctx context.Context, req *model.GroupCreateRequest, userId string) (*model.GroupResponse, error)
	GetAllGroups(ctx context.Context, userId string, page int, size int, search string) ([]*model.GroupResponse, int, error)
	GetGroupDetail(ctx context.Context, groupId string, userId string) (*model.GroupDetailResponse, error)
	UpdateGroup(ctx context.Context, groupId string, req *model.GroupUpdateRequest, userId string) (*model.GroupResponse, error)
	DeleteGroup(ctx context.Context, groupId string, userId string) error

	// Member Management
	InviteMembersByEmail(ctx context.Context, groupId string, req *model.GroupInviteEmailRequest, userId string) error
	GenerateInviteLink(ctx context.Context, groupId string, userId string) (*model.GroupInviteResponse, error)
	JoinGroup(ctx context.Context, req *model.GroupJoinRequest, userId string) error
	RemoveMember(ctx context.Context, groupId string, memberId int, userId string) error
	GetGroupMembers(ctx context.Context, groupId string, userId string) ([]model.GroupMemberResponse, error)
}

type groupUseCaseImpl struct {
	GroupRepository repository.GroupRepository
	AssetRepository repository.AssetRepository
	Log             *logrus.Logger
	Validator       *validator.Validate
	JWTSecret       []byte
}

func NewGroupUseCase(groupRepo repository.GroupRepository, assetRepo repository.AssetRepository, log *logrus.Logger, validate *validator.Validate, secret []byte) GroupUseCase {
	return &groupUseCaseImpl{
		GroupRepository: groupRepo,
		AssetRepository: assetRepo,
		Log:             log,
		Validator:       validate,
		JWTSecret:       secret,
	}
}

func (u *groupUseCaseImpl) CreateGroup(ctx context.Context, req *model.GroupCreateRequest, userId string) (*model.GroupResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	guruId, err := u.GroupRepository.GetGuruIdByUserId(ctx, userId)
	if err != nil {
		u.Log.Warnf("User is not a guru or guru not found: %v", err)
		return nil, fiber.ErrForbidden
	}

	group := &entity.Group{
		GroupName: req.GroupName,
		CreatedBy: guruId,
	}

	createdGroup, err := u.GroupRepository.CreateGroup(ctx, group)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToGroupResponse(createdGroup), nil
}

func (u *groupUseCaseImpl) GetAllGroups(ctx context.Context, userId string, page int, size int, search string) ([]*model.GroupResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	guruId, err := u.GroupRepository.GetGuruIdByUserId(ctx, userId)
	if err != nil {
		u.Log.Warnf("User is not a guru or guru not found: %v", err)
		return nil, 0, fiber.ErrForbidden
	}

	groups, total, err := u.GroupRepository.GetAllGroupsByGuru(ctx, guruId, page, size, search)
	if err != nil {
		return nil, 0, fiber.ErrInternalServerError
	}

	return converter.ToGroupResponses(groups), total, nil
}

func (u *groupUseCaseImpl) GetGroupDetail(ctx context.Context, groupId string, userId string) (*model.GroupDetailResponse, error) {
	guruId, err := u.GroupRepository.GetGuruIdByUserId(ctx, userId)
	if err != nil {
		return nil, fiber.ErrForbidden
	}

	group, err := u.GroupRepository.GetGroupById(ctx, groupId)
	if err != nil {
		return nil, fiber.ErrNotFound
	}
	if group.CreatedBy != guruId {
		return nil, fiber.ErrForbidden
	}

	members, err := u.GroupRepository.GetGroupMembers(ctx, groupId)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToGroupDetailResponse(group, members), nil
}

func (u *groupUseCaseImpl) UpdateGroup(ctx context.Context, groupId string, req *model.GroupUpdateRequest, userId string) (*model.GroupResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		return nil, fiber.ErrBadRequest
	}

	guruId, err := u.GroupRepository.GetGuruIdByUserId(ctx, userId)
	if err != nil {
		return nil, fiber.ErrForbidden
	}

	group, err := u.GroupRepository.GetGroupById(ctx, groupId)
	if err != nil {
		return nil, fiber.ErrNotFound
	}
	if group.CreatedBy != guruId {
		return nil, fiber.ErrForbidden
	}

	group.GroupName = req.GroupName
	group.GroupThumbnailAssetId = req.ThumbnailAssetId

	updatedGroup, err := u.GroupRepository.UpdateGroup(ctx, group)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if req.ThumbnailAssetId != nil {
		u.AssetRepository.MarkAsUsed(ctx, []int{*req.ThumbnailAssetId})
	}

	return converter.ToGroupResponse(updatedGroup), nil
}

func (u *groupUseCaseImpl) DeleteGroup(ctx context.Context, groupId string, userId string) error {
	guruId, err := u.GroupRepository.GetGuruIdByUserId(ctx, userId)
	if err != nil {
		return fiber.ErrForbidden
	}

	group, err := u.GroupRepository.GetGroupById(ctx, groupId)
	if err != nil {
		return fiber.ErrNotFound
	}
	if group.CreatedBy != guruId {
		return fiber.ErrForbidden
	}

	err = u.GroupRepository.DeleteGroup(ctx, groupId)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return nil
}

func (u *groupUseCaseImpl) InviteMembersByEmail(ctx context.Context, groupId string, req *model.GroupInviteEmailRequest, userId string) error {
	if err := u.Validator.Struct(req); err != nil {
		return fiber.ErrBadRequest
	}

	guruId, err := u.GroupRepository.GetGuruIdByUserId(ctx, userId)
	if err != nil {
		return fiber.ErrForbidden
	}

	group, err := u.GroupRepository.GetGroupById(ctx, groupId)
	if err != nil {
		return fiber.ErrNotFound
	}
	if group.CreatedBy != guruId {
		return fiber.ErrForbidden
	}

	for _, email := range req.Emails {
		memberId, err := u.GroupRepository.GetMemberIdByEmail(ctx, email)
		if err == nil && memberId > 0 {
			_ = u.GroupRepository.AddMember(ctx, groupId, memberId)
		} else {
			u.Log.Warnf("Failed to find or add member with email %s: %v", email, err)
		}
	}
	return nil
}

func (u *groupUseCaseImpl) GenerateInviteLink(ctx context.Context, groupId string, userId string) (*model.GroupInviteResponse, error) {
	guruId, err := u.GroupRepository.GetGuruIdByUserId(ctx, userId)
	if err != nil {
		return nil, fiber.ErrForbidden
	}

	group, err := u.GroupRepository.GetGroupById(ctx, groupId)
	if err != nil {
		return nil, fiber.ErrNotFound
	}
	if group.CreatedBy != guruId {
		return nil, fiber.ErrForbidden
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	claims := jwt.MapClaims{
		"group_id": groupId,
		"guru_id":  guruId,
		"type":     "group_invite",
		"exp":      expiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(u.JWTSecret)
	if err != nil {
		u.Log.Errorf("Error signing group invite token: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://arsiva.app" // Fallback based on issue example
	}
	inviteLink := fmt.Sprintf("%s/v1/groups/join?token=%s", baseURL, tokenString)

	return &model.GroupInviteResponse{
		InviteToken: tokenString,
		InviteLink:  inviteLink,
		QRCodeData:  inviteLink,
		ExpiresAt:   expiresAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (u *groupUseCaseImpl) JoinGroup(ctx context.Context, req *model.GroupJoinRequest, userId string) error {
	if err := u.Validator.Struct(req); err != nil {
		return fiber.ErrBadRequest
	}

	token, err := jwt.Parse(req.InviteToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return u.JWTSecret, nil
	})

	if err != nil || !token.Valid {
		return fiber.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fiber.ErrUnauthorized
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "group_invite" {
		return fiber.ErrUnauthorized
	}

	groupId, ok := claims["group_id"].(string)
	if !ok {
		return fiber.ErrUnauthorized
	}

	_, err = u.GroupRepository.GetGroupById(ctx, groupId)
	if err != nil {
		return fiber.ErrNotFound
	}

	memberId, err := u.GroupRepository.GetMemberIdByUserId(ctx, userId)
	if err != nil {
		return fiber.ErrForbidden
	}

	isMember, _ := u.GroupRepository.IsMemberInGroup(ctx, groupId, memberId)
	if isMember {
		return fiber.NewError(fiber.StatusConflict, "Already a member of this group")
	}

	err = u.GroupRepository.AddMember(ctx, groupId, memberId)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return nil
}

func (u *groupUseCaseImpl) RemoveMember(ctx context.Context, groupId string, memberId int, userId string) error {
	guruId, err := u.GroupRepository.GetGuruIdByUserId(ctx, userId)
	if err != nil {
		return fiber.ErrForbidden
	}

	group, err := u.GroupRepository.GetGroupById(ctx, groupId)
	if err != nil {
		return fiber.ErrNotFound
	}
	if group.CreatedBy != guruId {
		return fiber.ErrForbidden
	}

	err = u.GroupRepository.RemoveMember(ctx, groupId, memberId)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return nil
}

func (u *groupUseCaseImpl) GetGroupMembers(ctx context.Context, groupId string, userId string) ([]model.GroupMemberResponse, error) {
	guruId, err := u.GroupRepository.GetGuruIdByUserId(ctx, userId)
	if err != nil {
		return nil, fiber.ErrForbidden
	}

	group, err := u.GroupRepository.GetGroupById(ctx, groupId)
	if err != nil {
		return nil, fiber.ErrNotFound
	}
	if group.CreatedBy != guruId {
		return nil, fiber.ErrForbidden
	}

	members, err := u.GroupRepository.GetGroupMembers(ctx, groupId)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToGroupMemberResponses(members), nil
}
