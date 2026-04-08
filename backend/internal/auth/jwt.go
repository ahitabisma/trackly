package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtService struct {
	secret   []byte
	issuer   string
	audience string
}

type JwtClaims struct {
	Sub   uint   `json:"sub"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func NewJwtService(secret, issuer, audience string) *JwtService {
	return &JwtService{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
	}
}

func (s *JwtService) GenerateToken(userID uint, email, role string) (string, error) {
	userClaims := JwtClaims{
		Sub:   userID,
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	return token.SignedString(s.secret)
}

func (s *JwtService) ValidateToken(tokenString string) (*JwtClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return s.secret, nil
		},
	)

	if err != nil {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*JwtClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
