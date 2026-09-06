package user

import (
	"context"
	"testing"

	"github.com/LalatinaHub/LatinaApi/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	getUserByIDFn            func(ctx context.Context, id int64) (*model.User, error)
	createUserWithDefaultsFn func(ctx context.Context, id int64) (*model.User, error)
}

func (m *mockUserRepo) GetActiveUsersGroupedByVPN(ctx context.Context) (map[string][]model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) DeductQuota(ctx context.Context, userID int64, usedBytes int64) (int64, bool, error) {
	return 0, false, nil
}
func (m *mockUserRepo) DeductQuotaBatch(ctx context.Context, usages map[int64]int64) ([]int64, error) {
	return nil, nil
}
func (m *mockUserRepo) CreateUser(ctx context.Context, u *model.User) (int64, error) {
	return 1, nil
}
func (m *mockUserRepo) GetUserByToken(ctx context.Context, token string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, id)
	}
	return nil, model.ErrUserNotFound
}
func (m *mockUserRepo) CreateUserWithDefaults(ctx context.Context, id int64) (*model.User, error) {
	if m.createUserWithDefaultsFn != nil {
		return m.createUserWithDefaultsFn(ctx, id)
	}
	return &model.User{ID: id, Token: "new-token"}, nil
}

type mockKVRepo struct {
	getValueByKeyFn func(ctx context.Context, key string) (string, error)
}

func (m *mockKVRepo) GetAll(ctx context.Context) (map[string]any, error) {
	return nil, nil
}
func (m *mockKVRepo) GetValueByKey(ctx context.Context, key string) (string, error) {
	if m.getValueByKeyFn != nil {
		return m.getValueByKeyFn(ctx, key)
	}
	return "", nil
}

func TestUserService_GetUser_Existing(t *testing.T) {
	userRepo := &mockUserRepo{
		getUserByIDFn: func(ctx context.Context, id int64) (*model.User, error) {
			return &model.User{ID: id, Token: "existing-token"}, nil
		},
	}
	kvRepo := &mockKVRepo{
		getValueByKeyFn: func(ctx context.Context, key string) (string, error) {
			return "valid-api-token", nil
		},
	}

	svc := NewUserService(userRepo, kvRepo)
	u, err := svc.GetUser(context.Background(), "valid-api-token", 10)
	require.NoError(t, err)
	assert.Equal(t, int64(10), u.ID)
	assert.Equal(t, "existing-token", u.Token)
}

func TestUserService_GetUser_AutoProvision(t *testing.T) {
	userRepo := &mockUserRepo{
		getUserByIDFn: func(ctx context.Context, id int64) (*model.User, error) {
			return nil, model.ErrUserNotFound
		},
		createUserWithDefaultsFn: func(ctx context.Context, id int64) (*model.User, error) {
			return &model.User{ID: id, Token: "provisioned-token"}, nil
		},
	}
	kvRepo := &mockKVRepo{
		getValueByKeyFn: func(ctx context.Context, key string) (string, error) {
			return "", nil // No token configured
		},
	}

	svc := NewUserService(userRepo, kvRepo)
	u, err := svc.GetUser(context.Background(), "any-token", 20)
	require.NoError(t, err)
	assert.Equal(t, int64(20), u.ID)
	assert.Equal(t, "provisioned-token", u.Token)
}

func TestUserService_GetUser_Unauthorized(t *testing.T) {
	userRepo := &mockUserRepo{}
	kvRepo := &mockKVRepo{
		getValueByKeyFn: func(ctx context.Context, key string) (string, error) {
			return "expected-token", nil
		},
	}

	svc := NewUserService(userRepo, kvRepo)
	_, err := svc.GetUser(context.Background(), "wrong-token", 10)
	assert.ErrorIs(t, err, model.ErrUnauthorized)
}
