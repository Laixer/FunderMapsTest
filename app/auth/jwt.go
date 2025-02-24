package auth

import "github.com/golang-jwt/jwt/v5"

// TODO: Get from config
const SecretKey = "euwbghfliwsgdlyfihaerslfhlsefhlnlrelgnrlej"

func VerifyJWT(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func GenerateJWT(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(SecretKey))
}
