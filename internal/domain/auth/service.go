package auth

import (
	"context"
	"errors"
	"time"

	"amigo-server/pkg/apperror"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Login(ctx context.Context, input LoginInput) (string, *LoginResponse, error)
	Logout() string
	Me(ctx context.Context, tokenString string) (*MeResponse, error)
	MeWithUUID(ctx context.Context, tokenString string) (string, *MeResponse, error)
	VerifyPassword(ctx context.Context, email, password string) (*User, error)
}

type service struct {
	repo        Repository
	jwtSecret   string
	tokenExpiry time.Duration
}

func NewService(repo Repository, jwtSecret string, tokenExpiry time.Duration) Service {
	return &service{
		repo:        repo,
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
	}
}

func (s *service) Login(ctx context.Context, input LoginInput) (string, *LoginResponse, error) {
	u, err := s.VerifyPassword(ctx, input.Email, input.Password)
	if err != nil {
		return "", nil, err
	}

	token, err := s.generateToken(u.ID)
	if err != nil {
		return "", nil, apperror.Wrap(err, 500, "failed to generate token")
	}

	return token, &LoginResponse{Name: u.Name}, nil
}

func (s *service) Logout() string {
	return ""
}

func (s *service) Me(ctx context.Context, tokenString string) (*MeResponse, error) {
	_, resp, err := s.MeWithUUID(ctx, tokenString)
	return resp, err
}

func (s *service) MeWithUUID(ctx context.Context, tokenString string) (string, *MeResponse, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return "", nil, apperror.Unauthorized("invalid token")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", nil, apperror.Unauthorized("invalid token claims")
	}

	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return "", nil, apperror.Wrap(err, 500, "failed to get user")
	}
	if u == nil {
		return "", nil, apperror.Unauthorized("user not found")
	}

	return userID, &MeResponse{Name: u.Name}, nil
}

func (s *service) VerifyPassword(ctx context.Context, email, password string) (*User, error) {
	u, passwordHash, err := s.repo.FindByEmailWithPassword(ctx, email)
	if err != nil {
		return nil, apperror.Wrap(err, 500, "failed to verify password")
	}
	if u == nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, apperror.Unauthorized("invalid credentials")
	}

	return &User{ID: u.ID, Name: u.Name}, nil
}

func (s *service) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.tokenExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *service) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}