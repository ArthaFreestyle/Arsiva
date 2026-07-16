package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/mailer"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
)

// groupInviteExpiry is how long an emailed group-invite token stays valid.
const groupInviteExpiry = 7 * 24 * time.Hour

type GroupUseCase interface {
	// CRUD
	CreateGroup(ctx context.Context, req *model.GroupCreateRequest, userId string) (*model.GroupResponse, error)
	GetAllGroups(ctx context.Context, userId string, role string, page int, size int, search string) ([]*model.GroupResponse, int, error)
	GetGroupDetail(ctx context.Context, groupId string, userId string) (*model.GroupDetailResponse, error)
	UpdateGroup(ctx context.Context, groupId string, req *model.GroupUpdateRequest, userId string) (*model.GroupResponse, error)
	DeleteGroup(ctx context.Context, groupId string, userId string) error

	// Member Management
	InviteMembersByEmail(ctx context.Context, groupId string, req *model.GroupInviteEmailRequest, userId string) (*model.GroupInviteEmailResponse, error)
	GenerateInviteLink(ctx context.Context, groupId string, userId string) (*model.GroupInviteResponse, error)
	JoinGroup(ctx context.Context, req *model.GroupJoinRequest, userId string) error
	RemoveMember(ctx context.Context, groupId string, memberId int, userId string) error
	GetGroupMembers(ctx context.Context, groupId string, userId string) ([]model.GroupMemberResponse, error)

	// Group Contents
	AddContentToGroup(ctx context.Context, groupId string, req *model.GroupContentCreateRequest, userId string) (*model.GroupContentResponse, error)
	GetGroupContents(ctx context.Context, groupId string, contentType string, userId string, role string) ([]model.GroupContentResponse, error)
	RemoveContentFromGroup(ctx context.Context, groupId string, groupContentId int, userId string) error
}

type groupUseCaseImpl struct {
	GroupRepository repository.GroupRepository
	AssetRepository repository.AssetRepository
	Log             *logrus.Logger
	Validator       *validator.Validate
	JWTSecret       []byte
	Mailer          mailer.Mailer
	// InviteBaseURL is the frontend page a group invite link points at, e.g.
	// "https://arsiva.id/join-group". The token is appended as a query param.
	InviteBaseURL string
}

func NewGroupUseCase(groupRepo repository.GroupRepository, assetRepo repository.AssetRepository, log *logrus.Logger, validate *validator.Validate, secret []byte, mail mailer.Mailer, inviteBaseURL string) GroupUseCase {
	return &groupUseCaseImpl{
		GroupRepository: groupRepo,
		AssetRepository: assetRepo,
		Log:             log,
		Validator:       validate,
		JWTSecret:       secret,
		Mailer:          mail,
		InviteBaseURL:   inviteBaseURL,
	}
}

// buildGroupInviteToken signs a group-scoped invite JWT (type "group_invite",
// consumed by JoinGroup) and returns it with its expiry. Shared by the emailed
// invite flow and the shareable/QR invite link.
func (u *groupUseCaseImpl) buildGroupInviteToken(groupId string, guruId int) (string, time.Time, error) {
	expiresAt := time.Now().Add(groupInviteExpiry)
	claims := jwt.MapClaims{
		"group_id": groupId,
		"guru_id":  guruId,
		"type":     "group_invite",
		"exp":      expiresAt.Unix(),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(u.JWTSecret)
	return signed, expiresAt, err
}

// buildGroupInviteURL appends the invite token to the configured frontend join page,
// e.g. https://arsiva.id/join-group?token=<t>.
func (u *groupUseCaseImpl) buildGroupInviteURL(token string) string {
	sep := "?"
	if strings.Contains(u.InviteBaseURL, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%stoken=%s", u.InviteBaseURL, sep, url.QueryEscape(token))
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

func (u *groupUseCaseImpl) GetAllGroups(ctx context.Context, userId string, role string, page int, size int, search string) ([]*model.GroupResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	if role == "member" {
		memberId, err := u.GroupRepository.GetMemberIdByUserId(ctx, userId)
		if err != nil {
			u.Log.Warnf("Member not found for userId %s: %v", userId, err)
			return nil, 0, fiber.ErrForbidden
		}
		groups, total, err := u.GroupRepository.GetAllGroupsByMember(ctx, memberId, page, size, search)
		if err != nil {
			return nil, 0, fiber.ErrInternalServerError
		}
		return converter.ToGroupResponses(groups), total, nil
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

// InviteMembersByEmail emails a group-invitation link to each address. It does NOT
// auto-add anyone: recipients join by clicking the link (→ login/register → JoinGroup),
// which works for existing members and brand-new students alike. Returns a per-email
// confirmation so the FE can report which invites failed. The email body is rendered
// once and reused for every recipient — the invite token is group-scoped, not
// per-recipient.
func (u *groupUseCaseImpl) InviteMembersByEmail(ctx context.Context, groupId string, req *model.GroupInviteEmailRequest, userId string) (*model.GroupInviteEmailResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("InviteMembersByEmail: invalid request: %+v", err)
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

	token, _, err := u.buildGroupInviteToken(groupId, guruId)
	if err != nil {
		u.Log.Errorf("InviteMembersByEmail: error signing invite token: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	inviterName := group.Guru.Username
	if inviterName == "" {
		inviterName = "Guru Arsiva"
	}

	data := mailer.GroupInviteEmail{
		GroupName:    group.GroupName,
		InviterName:  inviterName,
		PersonalNote: strings.TrimSpace(req.Message),
		ButtonLabel:  "Gabung Grup",
		InviteURL:    u.buildGroupInviteURL(token),
		ExpiryDays:   int(groupInviteExpiry.Hours() / 24),
		SecurityNote: "Kalau kamu tidak mengenal pengundang atau tidak ingin bergabung, kamu bisa mengabaikan email ini dengan aman.",
		Preheader:    fmt.Sprintf("%s mengundangmu ke grup %s di Arsiva.", inviterName, group.GroupName),
	}
	subject := fmt.Sprintf("Undangan Bergabung ke Grup %s di Arsiva", group.GroupName)

	textBody := mailer.RenderGroupInviteText(data)
	htmlBody, err := mailer.RenderGroupInviteHTML(data)
	if err != nil {
		u.Log.Warnf("InviteMembersByEmail: failed to render HTML email, falling back to text: %v", err)
		htmlBody = textBody
	}

	resp := &model.GroupInviteEmailResponse{}
	seen := make(map[string]struct{}, len(req.Emails))
	for _, raw := range req.Emails {
		email := strings.ToLower(strings.TrimSpace(raw))
		if email == "" {
			continue
		}
		if _, dup := seen[email]; dup {
			continue // skip duplicate addresses so nobody is emailed twice
		}
		seen[email] = struct{}{}
		resp.Total++

		if u.Mailer == nil {
			u.Log.Warnf("InviteMembersByEmail: mailer not configured; cannot invite %s", email)
			resp.Failed++
			resp.Results = append(resp.Results, model.GroupInviteEmailResult{Email: email, Status: "failed"})
			continue
		}
		if err := u.Mailer.SendHTML(email, subject, htmlBody, textBody); err != nil {
			u.Log.Warnf("InviteMembersByEmail: failed to send invite to %s: %v", email, err)
			resp.Failed++
			resp.Results = append(resp.Results, model.GroupInviteEmailResult{Email: email, Status: "failed"})
			continue
		}
		resp.Sent++
		resp.Results = append(resp.Results, model.GroupInviteEmailResult{Email: email, Status: "sent"})
	}

	return resp, nil
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

	tokenString, expiresAt, err := u.buildGroupInviteToken(groupId, guruId)
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

func (u *groupUseCaseImpl) AddContentToGroup(ctx context.Context, groupId string, req *model.GroupContentCreateRequest, userId string) (*model.GroupContentResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Invalid request body AddContentToGroup: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	guruId, err := u.GroupRepository.GetGuruIdByUserId(ctx, userId)
	if err != nil {
		u.Log.Warnf("User is not a guru or guru not found: %v", err)
		return nil, fiber.ErrForbidden
	}

	group, err := u.GroupRepository.GetGroupById(ctx, groupId)
	if err != nil {
		return nil, fiber.ErrNotFound
	}
	if group.CreatedBy != guruId {
		return nil, fiber.ErrForbidden
	}

	contentExists, err := u.GroupRepository.ContentExists(ctx, req.ContentType, req.ContentId)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}
	if !contentExists {
		return nil, fiber.NewError(fiber.StatusUnprocessableEntity, "Content not found or not published")
	}

	alreadyAssigned, err := u.GroupRepository.IsContentAlreadyAssigned(ctx, groupId, req.ContentType, req.ContentId)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}
	if alreadyAssigned {
		return nil, fiber.NewError(fiber.StatusConflict, "Content already assigned to this group")
	}

	content, err := u.GroupRepository.AddContent(ctx, groupId, req.ContentType, req.ContentId)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	res := converter.ToGroupContentResponse(content)
	return &res, nil
}

func (u *groupUseCaseImpl) GetGroupContents(ctx context.Context, groupId string, contentType string, userId string, role string) ([]model.GroupContentResponse, error) {
	group, err := u.GroupRepository.GetGroupById(ctx, groupId)
	if err != nil {
		return nil, fiber.ErrNotFound
	}

	switch role {
	case "guru":
		guruId, err := u.GroupRepository.GetGuruIdByUserId(ctx, userId)
		if err != nil {
			return nil, fiber.ErrForbidden
		}
		if group.CreatedBy != guruId {
			return nil, fiber.ErrForbidden
		}
	case "member":
		memberId, err := u.GroupRepository.GetMemberIdByUserId(ctx, userId)
		if err != nil {
			return nil, fiber.ErrForbidden
		}
		isMember, err := u.GroupRepository.IsMemberInGroup(ctx, groupId, memberId)
		if err != nil {
			return nil, fiber.ErrInternalServerError
		}
		if !isMember {
			return nil, fiber.ErrForbidden
		}
	case "super_admin":
		// super_admin can access any group's contents
	default:
		return nil, fiber.ErrForbidden
	}

	contents, err := u.GroupRepository.GetContentsByGroupId(ctx, groupId, contentType)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToGroupContentResponses(contents), nil
}

func (u *groupUseCaseImpl) RemoveContentFromGroup(ctx context.Context, groupId string, groupContentId int, userId string) error {
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

	err = u.GroupRepository.RemoveContent(ctx, groupContentId, groupId)
	if err != nil {
		if errors.Is(err, repository.ErrGroupContentNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.ErrInternalServerError
	}

	return nil
}
