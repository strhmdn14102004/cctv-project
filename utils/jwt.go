package utils

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTUtil struct {
	secretKey  string
	expiration time.Duration
}

func NewJWTUtil(secretKey string, expiration time.Duration) *JWTUtil {
	return &JWTUtil{
		secretKey:  secretKey,
		expiration: expiration,
	}
}

type Claims struct {
	UserID int    `json:"userId"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

func (j *JWTUtil) GenerateToken(userID int, role string) (string, error) {
	expirationTime := time.Now().Add(j.expiration)

	claims := &Claims{
		UserID: userID,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "cctv-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTUtil) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, jwt.ErrSignatureInvalid
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				return nil, jwt.ErrSignatureInvalid
			}
		}
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
