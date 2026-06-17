package grpc

import (
	"context"
	"time"

	auth "github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/auth"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/proto/authpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	app *global.App
}

func NewAuthServer(app *global.App) *AuthServer {
	return &AuthServer{app: app}
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	var user auth.User
	err := s.app.DB.Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}

	if user.Password != req.Password {
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}

	accessToken := "access_" + user.Username + "_" + string(user.Role) + "_" + time.Now().Format(time.RFC3339)
	refreshToken := "refresh_" + user.Username + "_" + time.Now().Format(time.RFC3339)

	err = s.app.Cache.Set(ctx, refreshToken, user.Username, 24*time.Hour).Err()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cache session: %v", err)
	}

	return &authpb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *authpb.RefreshRequest) (*authpb.RefreshResponse, error) {
	username, err := s.app.Cache.Get(ctx, req.RefreshToken).Result()
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired refresh token")
	}

	var user auth.User
	err = s.app.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "user not found")
	}

	s.app.Cache.Del(ctx, req.RefreshToken)

	newAccessToken := "access_" + user.Username + "_" + string(user.Role) + "_" + time.Now().Format(time.RFC3339)
	newRefreshToken := "refresh_" + user.Username + "_" + time.Now().Format(time.RFC3339)

	err = s.app.Cache.Set(ctx, newRefreshToken, user.Username, 24*time.Hour).Err()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cache session: %v", err)
	}

	return &authpb.RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthServer) GetKeys(ctx context.Context, req *authpb.KeysRequest) (*authpb.KeysResponse, error) {
	return &authpb.KeysResponse{
		PublicKeys: []string{"public_key_mock_1", "public_key_mock_2"},
	}, nil
}

func (s *AuthServer) GetRoles(ctx context.Context, req *authpb.RolesRequest) (*authpb.RolesResponse, error) {
	return &authpb.RolesResponse{
		Roles: []string{"admin", "user"},
	}, nil
}

func (s *AuthServer) GetPermissions(ctx context.Context, req *authpb.PermissionsRequest) (*authpb.PermissionsResponse, error) {
	user := auth.User{}
	perms := user.Permissions()
	rolePerms, ok := perms[req.Role]
	if !ok {
		return &authpb.PermissionsResponse{Permissions: []string{}}, nil
	}
	return &authpb.PermissionsResponse{
		Permissions: rolePerms,
	}, nil
}
