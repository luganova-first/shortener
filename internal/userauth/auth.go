package userauth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"net/http"
	"time"
)

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

// UserItem представляет один элемент ответа
type UserItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Хранилище ID пользователей и их URLs
type Users struct {
	UserIDs  []int              // для отслеживания существующих пользователей
	UserURLs map[int][]UserItem // userID -> список URL
}

const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "supersecretkey"

func NewUsers() *Users {
	return &Users{
		UserIDs:  make([]int, 0),
		UserURLs: make(map[int][]UserItem),
	}
}

func (users *Users) MakeUserID() int {
	var max int

	if len(users.UserIDs) == 0 {
		max = 0
	} else {
		max = users.UserIDs[len(users.UserIDs)-1]
	}

	max++

	users.UserIDs = append(users.UserIDs, max)

	return max
}

func GetUserID(tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SECRET_KEY), nil
		})
	if err != nil {
		log.Println(err)
		return -1
	}

	if !token.Valid {
		log.Println("Token is not valid")
		return -1
	}

	return claims.UserID
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString(users *Users, userID int) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func SetUserCookie(h http.Handler, users *Users) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			if err == http.ErrNoCookie {
				userID := users.MakeUserID()

				tokenString, err := BuildJWTString(users, userID)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				cookie = &http.Cookie{
					Name:     "jwt",
					Value:    tokenString,
					Expires:  time.Now().Add(24 * time.Hour),
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
				}

				http.SetCookie(w, cookie)

				ctx := r.Context()
				ctx = context.WithValue(ctx, "userID", userID)
				r = r.WithContext(ctx)
			} else {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// передаём управление хендлеру
		h.ServeHTTP(w, r)
	})
}
