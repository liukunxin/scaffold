package convert

import (
	"testing"

	"single-starter/internal/app/user/model"
	"single-starter/internal/app/user/ro"
)

func TestCreateReqToDTO(t *testing.T) {
	cases := []struct {
		name  string
		req   ro.CreateUserReq
		wantN string
		wantE string
	}{
		{name: "normal", req: ro.CreateUserReq{Name: "alice", Email: "alice@example.com"}, wantN: "alice", wantE: "alice@example.com"},
		{name: "empty", req: ro.CreateUserReq{}, wantN: "", wantE: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CreateReqToDTO(&tc.req)
			if got.Name != tc.wantN || got.Email != tc.wantE {
				t.Fatalf("CreateReqToDTO() = %+v, want name=%q email=%q", got, tc.wantN, tc.wantE)
			}
		})
	}
}

func TestModelToVO(t *testing.T) {
	got := ModelToVO(&model.User{ID: "1", Name: "alice", Email: "alice@example.com"})
	if got.ID != "1" || got.Name != "alice" || got.Email != "alice@example.com" {
		t.Fatalf("ModelToVO() = %+v", got)
	}
}
