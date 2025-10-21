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
		Issuer:    jwtIssuer,
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
		jwt.WithIssuer(jwtIssuer),
	); err != nil {
		return uuid.Nil, err
	} else if id, err := uuid.Parse(claims.Subject); err != nil {
		return uuid.Nil, jwt.ErrTokenInvalidSubject
	} else {
		return id, nil
	}
}

func MakeRefreshToken() (string, error) {
	refreshToken := make([]byte, refreshTokenLength)
	if _, err := rand.Read(refreshToken); err != nil {
		return empty, err
	}
	return hex.EncodeToString(refreshToken), nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authorization := headers.Get(headerAuthorization)
	if len(authorization) == 0 {
		return empty, errors.New(errorMissingAuthBearer)
	}
	authorizationSplit := strings.Split(authorization, space)
	if len(authorizationSplit) != 2 || authorizationSplit[0] != authorizationBearer {
		return empty, errors.New(errorInvalidAuthBearer)
	}
	return authorizationSplit[1], nil
}

func GetApiKey(headers http.Header) (string, error) {
	authorization := headers.Get(headerAuthorization)
	if len(authorization) == 0 {
		return empty, errors.New(errorMissingAuthApiKey)
	}
	authorizationSplit := strings.Split(authorization, space)
	if len(authorizationSplit) != 2 || authorizationSplit[0] != authorizationApiKey {
		return empty, errors.New(errorInvalidAuthApiKey)
	}
	return authorizationSplit[1], nil
}
