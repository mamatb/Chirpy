package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mamatb/Chirpy/internal/database"
)

func respPlainOk(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(HeaderContentType, ContentTypePlain)
	body := []byte("OK")
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respPlainError(w http.ResponseWriter, _ *http.Request, message string) {
	w.WriteHeader(400)
	w.Header().Set(HeaderContentType, ContentTypePlain)
	body := []byte(message)
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respPlainForbidden(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(403)
	w.Header().Set(HeaderContentType, ContentTypePlain)
	body := []byte("FORBIDDEN")
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respPlainNotFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(404)
	w.Header().Set(HeaderContentType, ContentTypePlain)
	body := []byte("NOT FOUND")
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonUser(w http.ResponseWriter, _ *http.Request, user database.User,
	token string, refreshToken string) {
	w.Header().Set(HeaderContentType, ContentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(jsonUser{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonToken(w http.ResponseWriter, _ *http.Request, token string) {
	w.Header().Set(HeaderContentType, ContentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(jsonToken{
		Token: token,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonChirp(w http.ResponseWriter, _ *http.Request, chirp database.Chirp) {
	w.Header().Set(HeaderContentType, ContentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(jsonChirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonChirps(w http.ResponseWriter, _ *http.Request, chirps []database.Chirp) {
	w.Header().Set(HeaderContentType, ContentTypeJson)
	var err error
	var body []byte
	var chirpsJson []jsonChirp
	for _, chirp := range chirps {
		chirpsJson = append(chirpsJson, jsonChirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		})
	}
	if body, err = json.Marshal(chirpsJson); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonUserCreated(w http.ResponseWriter, r *http.Request, user database.User) {
	w.WriteHeader(201)
	respJsonUser(w, r, user, "", "")
}

func respJsonChirpCreated(w http.ResponseWriter, r *http.Request, chirp database.Chirp) {
	w.WriteHeader(201)
	respJsonChirp(w, r, chirp)
}

func respJsonError(w http.ResponseWriter, _ *http.Request, message string) {
	w.WriteHeader(400)
	w.Header().Set(HeaderContentType, ContentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(jsonError{
		Error: message,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonUnauthorized(w http.ResponseWriter, _ *http.Request, message string) {
	w.WriteHeader(401)
	w.Header().Set(HeaderContentType, ContentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(jsonError{
		Error: message,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}
