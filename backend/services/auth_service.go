package services

import (
	"context"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/adrianmcmains/integrated-site/models"
	"github.com/adrianmcmains/integrated-site/repositories"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Role:         req.Role,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.TokenResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         *user,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(viper.GetString("auth.jwt_secret")), nil
	})

	if err != nil {
		return nil, err
	}

	// Validate claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Extract user ID from claims
		userID, err := uuid.Parse(claims["user_id"].(string))
		if err != nil {
			return nil, ErrInvalidToken
		}

		return &models.JWTClaims{
			UserID: userID,
			Email:  claims["email"].(string),
			Role:   claims["role"].(string),
		}, nil
	}

	return nil, ErrInvalidToken
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenResponse, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user by ID
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidToken
	}

	// Generate new tokens
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, _, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		User:         *user,
	}, nil
}

func (s *AuthService) generateToken(user *models.User) (string, time.Time, error) {
	// Set expiration time
	expiryDuration, err := time.ParseDuration(viper.GetString("auth.token_expiry"))
	if err != nil {
		expiryDuration = 24 * time.Hour // Default to 24 hours
	}
	expiresAt := time.Now().Add(expiryDuration)

	// Create claims
	claims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"role":       user.Role,
		"exp":        expiresAt.Unix(),
		"issued_at":  time.Now().Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(viper.GetString("auth.jwt_secret")))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (s *AuthService) generateRefreshToken(user *models.User) (string, time.Time, error) {
	// Set expiration time
	expiryDuration, err := time.ParseDuration(viper.GetString("auth.refresh_token_expiry"))
	if err != nil {
		expiryDuration = 7 * 24 * time.Hour // Default to 7 days
	}
	expiresAt := time.Now().Add(expiryDuration)

	// Create claims
	claims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"role":       user.Role,
		"exp":        expiresAt.Unix(),
		"issued_at":  time.Now().Unix(),
		"is_refresh": true,
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(viper.GetString("auth.jwt_secret")))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}