package repository

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
)

func TestMemberAccountListIntegration(t *testing.T) {
	db := completionDB(t) // Isolated database using the real schema migrations.
	ctx := context.Background()
	_, err := db.Exec(ctx, `
		INSERT INTO users(user_id,username,email,password_hash,role,is_active,is_verified) VALUES
		(1,'ProfileStudent','one@example.com','hash','member',true,true),
		(2,'NewStudent','two@example.com','hash','member',true,false),
		(3,'InactiveStudent','three@example.com','hash','member',false,true),
		(4,'Teacher','four@example.com','hash','guru',true,true),
		(5,'Admin','five@example.com','hash','super_admin',true,true),
		(6,'OtherStudent','six@example.com','hash','member',true,true),
		(7,'InactiveNoProfile','seven@example.com','hash','member',false,false);
		INSERT INTO members(member_id,user_id,nis) VALUES (10,1,'NIS001'),(11,3,'NIS003'),(12,4,'NIS004'),(13,5,'NIS005'),(14,6,NULL),(15,NULL,'orphan');`)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewMemberRepository(db, logrus.New())
	list := func(search, status string, limit, offset int, ids []string, wantTotal int) []*entity.MemberAccount {
		t.Helper()
		rows, total, err := repo.FindAll(ctx, search, status, limit, offset)
		if err != nil {
			t.Fatal(err)
		}
		got := make([]string, 0, len(rows))
		for _, row := range rows {
			got = append(got, row.UserId)
		}
		if total != wantTotal || !reflect.DeepEqual(got, ids) {
			t.Fatalf("search=%q status=%s: ids=%v total=%d; want %v/%d", search, status, got, total, ids, wantTotal)
		}
		return rows
	}
	all := list("", "all", 10, 0, []string{"1", "2", "6"}, 3)
	if !all[0].HasProfile || all[0].MemberId == nil || *all[0].MemberId != "10" || all[0].TotalXP == nil || *all[0].TotalXP != 0 {
		t.Fatalf("profile: %+v", all[0])
	}
	missing := all[1]
	if missing.HasProfile || missing.MemberId != nil || missing.SekolahId != nil || missing.NIS != nil || missing.TotalXP != nil || missing.Level != nil || missing.FotoProfil != nil || missing.Bio != nil || missing.TanggalLahir != nil || missing.JenisKelamin != nil || missing.Minat != nil {
		t.Fatalf("missing profile must stay null: %+v", missing)
	}
	if !all[2].HasProfile || all[2].NIS != nil {
		t.Fatal("profile status must depend on row existence, not field completion")
	}
	counts, err := NewAnalyticsRepository(db, logrus.New()).CountUsersByRole(ctx)
	if err != nil || counts.Member != len(all) {
		t.Fatalf("summary=%+v err=%v", counts, err)
	}
	with := list("", "with_profile", 10, 0, []string{"1", "6"}, 2)
	without := list("", "without_profile", 10, 0, []string{"2"}, 1)
	if len(with)+len(without) != len(all) {
		t.Fatal("profile partitions differ from all")
	}
	list("NEWstudent", "all", 10, 0, []string{"2"}, 1)
	list("new", "with_profile", 10, 0, []string{}, 0)
	list("new", "without_profile", 10, 0, []string{"2"}, 1)
	list("nis001", "with_profile", 10, 0, []string{"1"}, 1)
	list("nis001", "without_profile", 10, 0, []string{}, 0)
	list("' OR TRUE --", "all", 10, 0, []string{}, 0)
	for i, id := range []string{"1", "2", "6"} {
		list("", "all", 1, i, []string{id}, 3)
	}
	list("", "all", 1, 3, []string{}, 3)
	_, err = db.Exec(ctx, `INSERT INTO members(user_id) VALUES (1)`)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" || pgErr.ConstraintName != "members_user_id_key" {
		t.Fatalf("duplicate profile must be rejected: %v", err)
	}
	if _, err := repo.Create(ctx, &entity.Member{UserId: "2", NIS: "NIS002"}); err != nil {
		t.Fatal(err)
	}
	list("", "all", 10, 0, []string{"1", "2", "6"}, 3)
	list("", "with_profile", 10, 0, []string{"1", "2", "6"}, 3)
	list("", "without_profile", 10, 0, []string{}, 0)
	counts, err = NewAnalyticsRepository(db, logrus.New()).CountUsersByRole(ctx)
	if err != nil || counts.Member != 3 {
		t.Fatalf("summary changed on profile creation: %+v %v", counts, err)
	}
}
