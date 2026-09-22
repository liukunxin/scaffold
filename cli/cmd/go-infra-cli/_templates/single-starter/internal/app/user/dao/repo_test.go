package dao

import (
	"context"
	"testing"

	kerr "github.com/liukunxin/go-infra/pkg/base/errors"
	"single-starter/internal/app/user/codes"
	"single-starter/internal/app/user/model"
)

func TestMemoryUserRepo(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepo()

	if _, err := repo.GetByID(ctx, "missing"); err == nil {
		t.Fatal("GetByID() on empty repo: want error, got nil")
	}

	want := &model.User{ID: "1", Name: "alice", Email: "alice@example.com"}
	if err := repo.Create(ctx, want); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(ctx, "1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got != want {
		t.Fatalf("GetByID() = %+v, want %+v", got, want)
	}
}

func TestMemoryUserRepo_NotFoundCarriesBusinessCode(t *testing.T) {
	_, err := NewUserRepo().GetByID(context.Background(), "1")
	if err == nil {
		t.Fatal("want error")
	}

	// dao 用 kerr.WrapError(StatusNotFound, codes.UserNotFound) 产出错误，
	// Controller 再借 kerr 自动映射 HTTP status，业务码原样透出。
	er := kerr.FromError(err)
	if er.Status != kerr.StatusNotFound {
		t.Fatalf("status = %v, want %v", er.Status, kerr.StatusNotFound)
	}
	if er.Code != codes.UserNotFound {
		t.Fatalf("code = %v, want %v", er.Code, codes.UserNotFound)
	}
}
