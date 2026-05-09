package usecase

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
)

type MemberUseCase interface {
	Create(ctx context.Context, req *model.MemberCreateRequest) (*model.MemberResponse, error)
	FindById(ctx context.Context, memberId string, claims *model.Claims) (*model.MemberDetailResponse, error)
	FindAll(ctx context.Context, search string, page int, size int) ([]*model.MemberResponse, int, error)
	Update(ctx context.Context, memberId string, req *model.MemberUpdateProfileRequest, claims *model.Claims) (*model.MemberResponse, error)
	Delete(ctx context.Context, memberId string) error
	GetMe(ctx context.Context, claims *model.Claims) (*model.MemberDetailResponse, error)
	UpdateMe(ctx context.Context, req *model.MemberUpdateProfileRequest, claims *model.Claims) (*model.MemberResponse, error)
}

type memberUseCaseImpl struct {
	MemberRepository repository.MemberRepository
	Log              *logrus.Logger
	Validator        *validator.Validate
}

func NewMemberUseCase(memberRepo repository.MemberRepository, log *logrus.Logger, validate *validator.Validate) MemberUseCase {
	return &memberUseCaseImpl{
		MemberRepository: memberRepo,
		Log:              log,
		Validator:        validate,
	}
}

func (u *memberUseCaseImpl) Create(ctx context.Context, req *model.MemberCreateRequest) (*model.MemberResponse, error) {
	if err := u.Validator.Struct(req); err != nil {
		u.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	member := &entity.Member{
		UserId:    req.UserId,
		SekolahId: req.SekolahId,
		NIS:       req.NIS,
	}

	result, err := u.MemberRepository.Create(ctx, member)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, fiber.NewError(fiber.StatusConflict, "user_id sudah terdaftar sebagai member")
		}
		u.Log.Warnf("Failed create member: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToMemberResponse(result), nil
}

func (u *memberUseCaseImpl) FindById(ctx context.Context, memberId string, claims *model.Claims) (*model.MemberDetailResponse, error) {
	if claims.Role == "member" {
		claimsMemberId := extractMemberIdFromClaims(claims)
		if claimsMemberId == "" || claimsMemberId != memberId {
			return nil, fiber.ErrForbidden
		}
	}

	member, err := u.MemberRepository.FindById(ctx, memberId)
	if err != nil {
		u.Log.Warnf("Member not found: %v", err)
		return nil, fiber.ErrNotFound
	}

	var sekolah *entity.Sekolah
	if member.SekolahId != "" {
		sekolah, _ = u.MemberRepository.FindSekolahByMemberId(ctx, memberId)
	}

	return converter.ToMemberDetailResponse(member, sekolah), nil
}

func (u *memberUseCaseImpl) FindAll(ctx context.Context, search string, page int, size int) ([]*model.MemberResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	offset := (page - 1) * size

	members, total, err := u.MemberRepository.FindAll(ctx, search, size, offset)
	if err != nil {
		u.Log.Warnf("Failed get all member: %v", err)
		return nil, 0, fiber.ErrInternalServerError
	}

	return converter.ToMemberResponses(members), total, nil
}

func (u *memberUseCaseImpl) Update(ctx context.Context, memberId string, req *model.MemberUpdateProfileRequest, claims *model.Claims) (*model.MemberResponse, error) {
	if claims.Role == "member" {
		claimsMemberId := extractMemberIdFromClaims(claims)
		if claimsMemberId == "" || claimsMemberId != memberId {
			return nil, fiber.ErrForbidden
		}
	}

	if req.JenisKelamin != "" && req.JenisKelamin != "L" && req.JenisKelamin != "P" && req.JenisKelamin != "Lainnya" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "jenis_kelamin harus 'L', 'P', atau 'Lainnya'")
	}

	_, err := u.MemberRepository.FindById(ctx, memberId)
	if err != nil {
		u.Log.Warnf("Member not found: %v", err)
		return nil, fiber.ErrNotFound
	}

	member := &entity.Member{
		MemberId:     memberId,
		SekolahId:    req.SekolahId,
		NIS:          req.NIS,
		FotoProfil:   req.FotoProfil,
		Bio:          req.Bio,
		TanggalLahir: req.TanggalLahir,
		JenisKelamin: entity.JenisKelamin(req.JenisKelamin),
		Minat:        req.Minat,
	}

	result, err := u.MemberRepository.Update(ctx, member)
	if err != nil {
		u.Log.Warnf("Failed update member: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ToMemberResponse(result), nil
}

func (u *memberUseCaseImpl) Delete(ctx context.Context, memberId string) error {
	_, err := u.MemberRepository.FindById(ctx, memberId)
	if err != nil {
		u.Log.Warnf("Member not found: %v", err)
		return fiber.ErrNotFound
	}

	err = u.MemberRepository.Delete(ctx, memberId)
	if err != nil {
		u.Log.Warnf("Failed delete member: %v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func (u *memberUseCaseImpl) GetMe(ctx context.Context, claims *model.Claims) (*model.MemberDetailResponse, error) {
	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}
	return u.FindById(ctx, memberId, claims)
}

func (u *memberUseCaseImpl) UpdateMe(ctx context.Context, req *model.MemberUpdateProfileRequest, claims *model.Claims) (*model.MemberResponse, error) {
	memberId := extractMemberIdFromClaims(claims)
	if memberId == "" {
		return nil, fiber.ErrForbidden
	}
	return u.Update(ctx, memberId, req, claims)
}

func extractMemberIdFromClaims(claims *model.Claims) string {
	if claims.Details == nil {
		return ""
	}
	detailsMap, ok := claims.Details.(map[string]interface{})
	if !ok {
		return ""
	}
	memberId, _ := detailsMap["member_id"].(string)
	return memberId
}
