package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func ValidateHash(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(id uuid.UUID, secret string, expiration time.Duration) (string, error) {
	start := jwt.NumericDate{Time: time.Now()}
	end := jwt.NumericDate{Time: start.Add(expiration)}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    JwtIssuer,
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
		jwt.WithIssuer(JwtIssuer),
	); err != nil {
		return uuid.Nil, err
	} else if id, err := uuid.Parse(claims.Subject); err != nil {
		return uuid.Nil, jwt.ErrTokenInvalidSubject
	} else {
		return id, nil
	}
}

func MakeRefreshToken() (string, error) {
	refreshToken := make([]byte, 32)
	if _, err := rand.Read(refreshToken); err != nil {
		return "", err
	}
	return hex.EncodeToString(refreshToken), nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authorization := headers.Get(HeaderAuthorization)
	if len(authorization) == 0 {
		return "", errors.New(ErrorMissingAuthBearer)
	}
	authorizationSplit := strings.Split(authorization, " ")
	if len(authorizationSplit) != 2 || authorizationSplit[0] != "Bearer" {
		return "", errors.New(ErrorInvalidAuthBearer)
	}
	return authorizationSplit[1], nil
}

func GetApiKey(headers http.Header) (string, error) {
	authorization := headers.Get(HeaderAuthorization)
	if len(authorization) == 0 {
		return "", errors.New(ErrorMissingAuthApiKey)
	}
	authorizationSplit := strings.Split(authorization, " ")
	if len(authorizationSplit) != 2 || authorizationSplit[0] != "ApiKey" {
		return "", errors.New(ErrorInvalidAuthApiKey)
	}
	return authorizationSplit[1], nil
}
