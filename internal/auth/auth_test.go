package auth

import (
	"errors"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []string{
		"123456",
		"qwerty",
		"secret",
	}
	want := regexp.MustCompile(`^\$2a\$10\$[./0-9A-Za-z]{53}$`)
	for _, input := range tests {
		if output, err := HashPassword(input); err != nil || !want.MatchString(output) {
			t.Errorf(
				"HashPassword(\"%s\") = (\"%s\", %v), want (%#q, nil)",
				input, output, err, want,
			)
		}
	}
}

func TestCheckPasswordHash(t *testing.T) {
	testsOk, testsErr := map[string]string{}, map[string]string{}
	tests := []string{
		"123456",
		"qwerty",
		"secret",
	}
	for _, input := range tests {
		output, _ := bcrypt.GenerateFromPassword([]byte(input), bcrypt.DefaultCost)
		testsOk[input] = string(output)
		testsErr[input] = input
	}
	for input, output := range testsOk {
		if err := CheckPasswordHash(input, output); err != nil {
			t.Errorf(
				"CheckPassword(\"%s\", \"%s\") = (%v), want (nil)",
				input, output, err,
			)
		}
	}
	for input, output := range testsErr {
		if err := CheckPasswordHash(input, output); err == nil {
			t.Errorf(
				"CheckPassword(\"%s\", \"%s\") = (%v), want (error)",
				input, output, nil,
			)
		}
	}
}

func TestMakeJWT(t *testing.T) {
	tests := []string{
		"123456",
		"qwerty",
		"secret",
	}
	want := regexp.MustCompile(`^(eyJ[-_0-9A-Za-z]+\.){2}[-_0-9A-Za-z]+$`)
	for _, inputSecret := range tests {
		inputId, inputExpiry := uuid.New(), time.Second*time.Duration(60)
		if output, err := MakeJWT(inputId, inputSecret, inputExpiry); err != nil ||
			!want.MatchString(output) {
			t.Errorf(
				"MakeJWT(%s, \"%s\", %.2fs) = (\"%s\", %v), want (%#q, nil)",
				inputId, inputSecret, inputExpiry.Seconds(), output, err, want,
			)
		}
	}
}

func TestValidateJWT(t *testing.T) {
	tests, inputToken, inputSecret := map[string]error{}, "", "secret"
	issuer, start := "chirpy", jwt.NumericDate{Time: time.Now()}
	end := jwt.NumericDate{Time: start.Add(time.Second * time.Duration(60))}

	inputToken, _ = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  &start,
		ExpiresAt: &end,
		Subject:   uuid.New().String(),
	}).SignedString([]byte(inputSecret))
	tests[inputToken] = nil

	inputToken, _ = jwt.NewWithClaims(jwt.SigningMethodHS512,
		jwt.RegisteredClaims{}).SignedString([]byte(inputSecret))
	tests[inputToken] = jwt.ErrTokenSignatureInvalid

	inputToken, _ = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{}).SignedString([]byte("error"))
	tests[inputToken] = jwt.ErrTokenSignatureInvalid

	inputToken, _ = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{}).SignedString([]byte(inputSecret))
	tests[inputToken] = jwt.ErrTokenInvalidClaims

	inputToken, _ = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  &start,
		ExpiresAt: &start,
		Subject:   uuid.New().String(),
	}).SignedString([]byte(inputSecret))
	tests[inputToken] = jwt.ErrTokenExpired

	inputToken, _ = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "error",
		IssuedAt:  &start,
		ExpiresAt: &end,
		Subject:   uuid.New().String(),
	}).SignedString([]byte(inputSecret))
	tests[inputToken] = jwt.ErrTokenInvalidIssuer

	inputToken, _ = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  &start,
		ExpiresAt: &end,
		Subject:   "error",
	}).SignedString([]byte(inputSecret))
	tests[inputToken] = jwt.ErrTokenInvalidSubject

	for inputToken, want := range tests {
		if output, err := ValidateJWT(inputToken, inputSecret); !errors.Is(err, want) {
			t.Errorf(
				"ValidateJWT(\"%s\", \"%s\") = (%s, %v), want (%s, %v)",
				inputToken, inputSecret, output, err, output, want,
			)
		}
	}
}

func TestGetBearerToken(t *testing.T) {
	testsOk := []http.Header{
		{HeaderAuthorization: []string{"Bearer 123456"}},
		{HeaderAuthorization: []string{"Bearer qwerty"}},
		{HeaderAuthorization: []string{"Bearer secret"}},
	}
	testsErr := []http.Header{
		{},
		{HeaderAuthorization: []string{}},
		{HeaderAuthorization: []string{"Bearer"}},
	}
	for _, input := range testsOk {
		if output, err := GetBearerToken(input); err != nil {
			t.Errorf(
				"GetBearerToken(\"%s\") = (\"%s\", %v), want (token, nil)",
				input, output, err,
			)
		}
	}
	for _, input := range testsErr {
		if output, err := GetBearerToken(input); err == nil {
			t.Errorf(
				"GetBearerToken(\"%s\") = (\"%s\", %v), want (\"\", error)",
				input, output, nil,
			)
		}
	}
}
