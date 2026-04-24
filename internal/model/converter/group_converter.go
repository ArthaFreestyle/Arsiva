package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"strconv"
)

func ToGroupResponse(group *entity.Group) *model.GroupResponse {
	if group == nil {
		return nil
	}

	createdAt := ""
	if group.CreatedAt != nil {
		createdAt = group.CreatedAt.Format("2006-01-02 15:04:05")
	}

	updatedAt := ""
	if group.UpdatedAt != nil {
		updatedAt = group.UpdatedAt.Format("2006-01-02 15:04:05")
	}

	thumbnail := ""
	if group.GroupThumbnail != nil {
		thumbnail = *group.GroupThumbnail
	}

	guruIdStr := group.Guru.GuruId
	guruIdInt, _ := strconv.Atoi(guruIdStr)
	// fallback if GuruId struct value is an int string and createdby is just int,
	// if the issue required created_by to be int, it matches. 
	if guruIdInt == 0 && group.CreatedBy != 0 {
	    guruIdInt = group.CreatedBy
	}

	return &model.GroupResponse{
		GroupId:   group.GroupId,
		GroupName: group.GroupName,
		Thumbnail: thumbnail,
		CreatedBy: model.GroupGuruResponse{
			GuruId:     guruIdInt,
			NIP:        group.Guru.NIP,
			BidangAjar: group.Guru.BidangAjar,
			Username:   group.Guru.Username,
		},
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		MemberCount: group.MemberCount,
	}
}

func ToGroupResponses(groups []*entity.Group) []*model.GroupResponse {
	var responses []*model.GroupResponse
	for _, group := range groups {
		responses = append(responses, ToGroupResponse(group))
	}
	return responses
}

func ToGroupMemberResponse(member *entity.GroupMember) model.GroupMemberResponse {
	if member == nil {
		return model.GroupMemberResponse{}
	}

	tanggalBergabung := ""
	if member.TanggalBergabung != nil {
		tanggalBergabung = member.TanggalBergabung.Format("2006-01-02 15:04:05")
	}

	fotoProfil := ""
	if member.FotoProfil != nil {
		fotoProfil = *member.FotoProfil
	}

	return model.GroupMemberResponse{
		MemberId:         member.MemberId,
		Username:         member.Username,
		Email:            member.Email,
		NIS:              member.NIS,
		FotoProfil:       fotoProfil,
		TanggalBergabung: tanggalBergabung,
	}
}

func ToGroupMemberResponses(members []*entity.GroupMember) []model.GroupMemberResponse {
	var responses []model.GroupMemberResponse
	for _, member := range members {
		responses = append(responses, ToGroupMemberResponse(member))
	}
	return responses
}

func ToGroupDetailResponse(group *entity.Group, members []*entity.GroupMember) *model.GroupDetailResponse {
	if group == nil {
		return nil
	}

	resp := ToGroupResponse(group)

	memberResps := ToGroupMemberResponses(members)
    if memberResps == nil {
        memberResps = []model.GroupMemberResponse{}
    }

	return &model.GroupDetailResponse{
		GroupId:     resp.GroupId,
		GroupName:   resp.GroupName,
		Thumbnail:   resp.Thumbnail,
		CreatedBy:   resp.CreatedBy,
		CreatedAt:   resp.CreatedAt,
		UpdatedAt:   resp.UpdatedAt,
		MemberCount: resp.MemberCount,
		Members:     memberResps,
	}
}
