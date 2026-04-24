package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

func TestNewCeritaUseCase(t *testing.T) {
	uc := NewCeritaUseCase(nil, nil, nil, validator.New())
	if uc == nil {
		t.Fatal("expected usecase instance")
	}
}

type ceritaRepositoryStub struct {
	createFn      func(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error)
	updateFn      func(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error)
	createSceneFn func(ctx context.Context, ceritaId int, scene *entity.Scene) (*entity.Scene, error)
	updateSceneFn func(ctx context.Context, ceritaId int, sceneId int, scene *entity.Scene) (*entity.Scene, error)
	deleteSceneFn func(ctx context.Context, ceritaId int, sceneId int) error
}

func (r *ceritaRepositoryStub) FindAll(ctx context.Context, page int, size int, search string) ([]*entity.CeritaInteraktif, int, error) {
	return nil, 0, nil
}

func (r *ceritaRepositoryStub) FindById(ctx context.Context, ceritaId int) (*entity.CeritaInteraktif, error) {
	return &entity.CeritaInteraktif{CeritaId: ceritaId}, nil
}

func (r *ceritaRepositoryStub) FindAllManage(ctx context.Context, page int, size int, search string, userId string, role string) ([]*entity.CeritaInteraktif, int, error) {
	return nil, 0, nil
}

func (r *ceritaRepositoryStub) FindByIdManage(ctx context.Context, ceritaId int, userId string, role string) (*entity.CeritaInteraktif, error) {
	return &entity.CeritaInteraktif{CeritaId: ceritaId}, nil
}

func (r *ceritaRepositoryStub) Create(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error) {
	return r.createFn(ctx, cerita)
}

func (r *ceritaRepositoryStub) Update(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error) {
	return r.updateFn(ctx, cerita)
}

func (r *ceritaRepositoryStub) CreateScene(ctx context.Context, ceritaId int, scene *entity.Scene) (*entity.Scene, error) {
	return r.createSceneFn(ctx, ceritaId, scene)
}

func (r *ceritaRepositoryStub) UpdateScene(ctx context.Context, ceritaId int, sceneId int, scene *entity.Scene) (*entity.Scene, error) {
	return r.updateSceneFn(ctx, ceritaId, sceneId, scene)
}

func (r *ceritaRepositoryStub) DeleteScene(ctx context.Context, ceritaId int, sceneId int) error {
	return r.deleteSceneFn(ctx, ceritaId, sceneId)
}

func (r *ceritaRepositoryStub) Delete(ctx context.Context, ceritaId int) error {
	return nil
}

func TestCreateCerita_IgnoresScenesAndCreatesDraft(t *testing.T) {
	t.Parallel()

	repo := &ceritaRepositoryStub{
		createFn: func(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error) {
			if cerita.IsPublished {
				t.Fatalf("expected new story to be draft")
			}
			if len(cerita.Scenes) != 0 {
				t.Fatalf("expected create to not include scenes, got %d", len(cerita.Scenes))
			}
			cerita.CeritaId = 10
			return cerita, nil
		},
		updateFn: func(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error) {
			return cerita, nil
		},
		createSceneFn: func(ctx context.Context, ceritaId int, scene *entity.Scene) (*entity.Scene, error) { return scene, nil },
		updateSceneFn: func(ctx context.Context, ceritaId int, sceneId int, scene *entity.Scene) (*entity.Scene, error) {
			return scene, nil
		},
		deleteSceneFn: func(ctx context.Context, ceritaId int, sceneId int) error { return nil },
	}

	uc := NewCeritaUseCase(repo, nil, logrus.New(), validator.New())
	_, err := uc.CreateCerita(context.Background(), &model.CeritaRequest{
		Judul:       "Cerita Baru",
		KategoriId:  1,
		XpReward:    50,
		IsPublished: true,
	}, "1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSceneCRUDMethods(t *testing.T) {
	t.Parallel()

	repo := &ceritaRepositoryStub{
		createFn: func(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error) {
			return cerita, nil
		},
		updateFn: func(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error) {
			return cerita, nil
		},
		createSceneFn: func(ctx context.Context, ceritaId int, scene *entity.Scene) (*entity.Scene, error) {
			if ceritaId != 2 {
				t.Fatalf("expected cerita id 2, got %d", ceritaId)
			}
			scene.SceneId = 11
			scene.CeritaId = ceritaId
			return scene, nil
		},
		updateSceneFn: func(ctx context.Context, ceritaId int, sceneId int, scene *entity.Scene) (*entity.Scene, error) {
			if ceritaId != 2 || sceneId != 11 {
				t.Fatalf("unexpected ids: cerita=%d scene=%d", ceritaId, sceneId)
			}
			scene.SceneId = sceneId
			scene.CeritaId = ceritaId
			return scene, nil
		},
		deleteSceneFn: func(ctx context.Context, ceritaId int, sceneId int) error {
			if ceritaId != 2 || sceneId != 11 {
				t.Fatalf("unexpected ids for delete: cerita=%d scene=%d", ceritaId, sceneId)
			}
			return nil
		},
	}

	uc := NewCeritaUseCase(repo, nil, logrus.New(), validator.New())
	sceneRequest := &model.SceneRequest{
		SceneKey:  "scene_awal",
		SceneText: "awal",
	}

	created, err := uc.CreateScene(context.Background(), 2, sceneRequest)
	if err != nil {
		t.Fatalf("expected no error on create scene, got %v", err)
	}
	if created.SceneId != 11 {
		t.Fatalf("expected created scene id 11, got %d", created.SceneId)
	}

	updated, err := uc.UpdateScene(context.Background(), 2, 11, sceneRequest)
	if err != nil {
		t.Fatalf("expected no error on update scene, got %v", err)
	}
	if updated.SceneId != 11 {
		t.Fatalf("expected updated scene id 11, got %d", updated.SceneId)
	}

	if err := uc.DeleteScene(context.Background(), 2, 11); err != nil {
		t.Fatalf("expected no error on delete scene, got %v", err)
	}
}
