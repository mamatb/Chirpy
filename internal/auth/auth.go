package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	HeaderAuthorization = "Authorization"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPasswordHash(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(id uuid.UUID, secret string, expiry time.Duration) (string, error) {
	start := jwt.NumericDate{Time: time.Now()}
	end := jwt.NumericDate{Time: start.Add(expiry)}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		Subject:   id.String(),
		ExpiresAt: &end,
		IssuedAt:  &start,
	})
	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString string, secret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	if _, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer("chirpy"),
	); err != nil {
		return uuid.Nil, err
	} else if id, err := uuid.Parse(claims.Subject); err != nil {
		return uuid.Nil, jwt.ErrTokenInvalidSubject
	} else {
		return id, nil
	}
}

func GetBearerToken(headers http.Header) (string, error) {
	authorization := headers.Get(HeaderAuthorization)
	if len(authorization) == 0 {
		return "", errors.New("empty Authorization Bearer header")
	}
	authorizationSplit := strings.Split(authorization, " ")
	if len(authorizationSplit) != 2 || authorizationSplit[0] != "Bearer" {
		return "", errors.New("invalid Authorization Bearer header")
	}
	return authorizationSplit[1], nil
}
