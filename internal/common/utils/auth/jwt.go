package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSecretKey JWT 签名密钥
var jwtSecretKey = []byte("edu-evaluation-secret-key")

// GenerateToken 生成 JWT token
//
// 参数:
//   - claims map[string]interface{} 自定义载荷，包含 userId、username/studentNo、role 等
//
// 返回值:
//   - string 生成的 JWT token 字符串
//   - error 签名失败时返回错误
func GenerateToken(claims map[string]interface{}) (string, error) {
	now := time.Now()
	mapClaims := jwt.MapClaims{
		"exp": now.Add(24 * time.Hour).Unix(),
		"nbf": now.Unix(),
		"iat": now.Unix(),
	}
	for k, v := range claims {
		mapClaims[k] = v
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	return token.SignedString(jwtSecretKey)
}

// ParseToken 解析并验证 JWT token
//
// 参数:
//   - tokenString string 待解析的 JWT token 字符串
//
// 返回值:
//   - jwt.MapClaims 解析后的 token 载荷
//   - error token 无效或过期时返回错误
func ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
