package entity

import "time"

type GroupMember struct {
	GroupId          string     `db:"group_id"`
	MemberId         int        `db:"member_id"`
	TanggalBergabung *time.Time `db:"tanggal_bergabung"`
	Username         string     `db:"username"`
	Email            string     `db:"email"`
	NIS              string     `db:"nis"`
	FotoProfil       *string    `db:"foto_profil"`
}
