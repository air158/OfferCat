package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

var jwtSecret = []byte("我爱玩元梦之星王者荣耀真好玩") // jwt 加密密钥

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	// StandardClaims 已经弃用，使用 RegisteredClaims
	jwt.RegisteredClaims
}

// 生成 JWT Token
func GenerateToken(userID uint, username, role string) (string, error) {
	now := time.Now()
	expireTime := now.Add(24 * 7 * time.Hour) // Token 有效期 一周

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime), // 转换为 *NumericDate
			IssuedAt:  jwt.NewNumericDate(now),        // 转换为 *NumericDate
			Issuer:    "offercat",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 验证 JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
