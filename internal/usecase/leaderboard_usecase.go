package usecase

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
)

type LeaderboardUseCase interface {
	GetPublic(ctx context.Context, req *model.PublicLeaderboardRequest) (*model.PublicLeaderboardResponse, error)
	GetGroup(ctx context.Context, groupId string, req *model.GroupLeaderboardRequest, claims *model.Claims) (*model.GroupLeaderboardResponse, error)
}

type leaderboardUseCaseImpl struct {
	Repo      repository.LeaderboardRepository
	GroupRepo repository.GroupRepository
	Log       *logrus.Logger
}

func NewLeaderboardUseCase(
	repo repository.LeaderboardRepository,
	groupRepo repository.GroupRepository,
	log *logrus.Logger,
) LeaderboardUseCase {
	return &leaderboardUseCaseImpl{
		Repo:      repo,
		GroupRepo: groupRepo,
		Log:       log,
	}
}

func (u *leaderboardUseCaseImpl) GetPublic(ctx context.Context, req *model.PublicLeaderboardRequest) (*model.PublicLeaderboardResponse, error) {
	// Validate and apply defaults.
	if req.Page < 1 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "page harus >= 1")
	}
	if req.Size < 1 || req.Size > 100 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "size harus antara 1 dan 100")
	}

	period := repository.LeaderboardPeriod(req.Period)
	if req.Period == "" {
		period = repository.LeaderboardPeriodAlltime
	}
	if period != repository.LeaderboardPeriodAlltime && period != repository.LeaderboardPeriodMonthly {
		return nil, fiber.NewError(fiber.StatusBadRequest, "period tidak valid, gunakan 'alltime' atau 'monthly'")
	}

	offset := (req.Page - 1) * req.Size

	entries, total, periodStart, err := u.Repo.FetchPublic(ctx, period, req.SekolahId, req.Size, offset)
	if err != nil {
		u.Log.Warnf("GetPublic FetchPublic: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	var periodStartStr *string
	if period == repository.LeaderboardPeriodMonthly && !periodStart.IsZero() {
		// Format with explicit Jakarta offset (+07:00) as required.
		s := periodStart.In(time.FixedZone("WIB", 7*3600)).Format(time.RFC3339)
		periodStartStr = &s
	}

	items := converter.ToPublicLeaderboardItems(entries)
	if items == nil {
		items = []model.PublicLeaderboardItem{}
	}

	return &model.PublicLeaderboardResponse{
		Period:      string(period),
		PeriodStart: periodStartStr,
		Page:        req.Page,
		Size:        req.Size,
		Total:       total,
		Items:       items,
	}, nil
}

func (u *leaderboardUseCaseImpl) GetGroup(ctx context.Context, groupId string, req *model.GroupLeaderboardRequest, claims *model.Claims) (*model.GroupLeaderboardResponse, error) {
	if req.Page < 1 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "page harus >= 1")
	}
	if req.Size < 1 || req.Size > 100 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "size harus antara 1 dan 100")
	}

	// Step 1: check group exists (404 before 403).
	groupName, groupThumbnail, found, err := u.Repo.GetGroupHeader(ctx, groupId)
	if err != nil {
		u.Log.Warnf("GetGroup GetGroupHeader: %v", err)
		return nil, fiber.ErrInternalServerError
	}
	if !found {
		return nil, fiber.ErrNotFound
	}

	// Step 2: verify access — mirrors GetGroupContents in group_usecase.go.
	if err := u.checkGroupAccess(ctx, groupId, claims); err != nil {
		return nil, err
	}

	offset := (req.Page - 1) * req.Size

	entries, total, err := u.Repo.FetchGroup(ctx, groupId, req.Size, offset)
	if err != nil {
		u.Log.Warnf("GetGroup FetchGroup: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	items := converter.ToGroupLeaderboardItems(entries)
	if items == nil {
		items = []model.GroupLeaderboardItem{}
	}

	return &model.GroupLeaderboardResponse{
		Group: &model.GroupLeaderboardHeader{
			GroupId:        groupId,
			GroupName:      groupName,
			GroupThumbnail: groupThumbnail,
		},
		Page:  req.Page,
		Size:  req.Size,
		Total: total,
		Items: items,
	}, nil
}

func (u *leaderboardUseCaseImpl) checkGroupAccess(ctx context.Context, groupId string, claims *model.Claims) error {
	switch claims.Role {
	case "super_admin":
		return nil
	case "guru":
		guruId, err := u.GroupRepo.GetGuruIdByUserId(ctx, claims.UserId)
		if err != nil {
			return fiber.ErrForbidden
		}
		group, err := u.GroupRepo.GetGroupById(ctx, groupId)
		if err != nil {
			return fiber.ErrForbidden
		}
		if group.CreatedBy != guruId {
			return fiber.ErrForbidden
		}
		return nil
	case "member":
		memberId, err := u.GroupRepo.GetMemberIdByUserId(ctx, claims.UserId)
		if err != nil {
			return fiber.ErrForbidden
		}
		isMember, err := u.GroupRepo.IsMemberInGroup(ctx, groupId, memberId)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		if !isMember {
			return fiber.ErrForbidden
		}
		return nil
	default:
		return fiber.ErrForbidden
	}
}
