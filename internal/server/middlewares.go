package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/vook88/go-url-shortener/internal/authn"
	"github.com/vook88/go-url-shortener/internal/contextkeys"
	"github.com/vook88/go-url-shortener/internal/logger"
	"github.com/vook88/go-url-shortener/internal/storage"
)

var CookieAuthName = "auth-token"

func checkSupportedContentType(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	supportedContentTypes := []string{"application/json", "text/html"}

	for _, t := range supportedContentTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip && checkSupportedContentType(r) {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, r)
	})
}

func AuthMiddlewareCheckAndCreate(storage storage.URLStorage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID int
			log := logger.GetLogger()

			cookie, err := r.Cookie(CookieAuthName)
			if err != nil {
				log.Error().Msgf("Error when parsing Cookie: %s", err.Error())
				if !errors.Is(err, http.ErrNoCookie) {
					log.Debug().Msg(err.Error())
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			} else {
				userID, err = authn.GetUserID(cookie.Value)
				log.Debug().Msgf("User ID: %d", userID)
				if err == nil {
					ctx2 := context.WithValue(r.Context(), contextkeys.UserIDKey, userID)
					next.ServeHTTP(w, r.WithContext(ctx2))
					return
				}
				log.Error().Msg(err.Error())
				if !errors.Is(err, authn.ErrTokenIsNotValid) {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
			}
			userID, err = storage.GenerateUserID(r.Context())
			if err != nil {
				log.Debug().Msg(err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			encodedValue, err2 := authn.BuildJWTString(userID)
			if err2 != nil {
				log.Debug().Msg(err2.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     CookieAuthName,
				Value:    encodedValue,
				Path:     "/",
				HttpOnly: true,
			})

			ctx2 := context.WithValue(r.Context(), contextkeys.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx2))
		})
	}
}

func AuthMiddlewareCheckOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger()
		cookie, err := r.Cookie(CookieAuthName)
		if err != nil {
			log.Error().Msgf("Error when parsing Cookie: %s", err.Error())
			if !errors.Is(err, http.ErrNoCookie) {
				log.Debug().Msg(err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID, err := authn.GetUserID(cookie.Value)
		if err != nil {
			log.Error().Msg(err.Error())
			if !errors.Is(err, authn.ErrTokenIsNotValid) {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx2 := context.WithValue(r.Context(), contextkeys.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx2))
	})
}
