package logic

import (
	"context"
	"errors"
	"testing"

	"single-starter/internal/app/user/dto"
	"single-starter/internal/app/user/model"
)

type stubUserService struct {
	user *model.User
	err  error
}

func (s *stubUserService) Create(context.Context, *dto.CreateUserInput) (*model.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

func (s *stubUserService) GetByID(context.Context, string) (*model.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

type stubNotifier struct {
	called bool
	err    error
}

func (s *stubNotifier) NotifyUserCreated(context.Context, string) error {
	s.called = true
	return s.err
}

func TestCreateUser_NotifiesAfterCreate(t *testing.T) {
	n := &stubNotifier{}
	l := NewUserLogic(&stubUserService{user: &model.User{ID: "1", Name: "alice"}}, n)

	got, err := l.CreateUser(context.Background(), &dto.CreateUserInput{Name: "alice"})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if got.ID != "1" {
		t.Fatalf("CreateUser() id = %q, want 1", got.ID)
	}
	if !n.called {
		t.Fatal("notifier was not called")
	}
}

func TestCreateUser_NotifierFailureIsNotFatal(t *testing.T) {
	n := &stubNotifier{err: errors.New("channel down")}
	l := NewUserLogic(&stubUserService{user: &model.User{ID: "1"}}, n)

	if _, err := l.CreateUser(context.Background(), &dto.CreateUserInput{}); err != nil {
		t.Fatalf("CreateUser() should tolerate notifier failure, got %v", err)
	}
}

func TestCreateUser_ServiceFailureSkipsNotifier(t *testing.T) {
	n := &stubNotifier{}
	l := NewUserLogic(&stubUserService{err: errors.New("db down")}, n)

	if _, err := l.CreateUser(context.Background(), &dto.CreateUserInput{}); err == nil {
		t.Fatal("CreateUser() want error")
	}
	if n.called {
		t.Fatal("notifier should not be called when create fails")
	}
}
