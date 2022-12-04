package routers

import (
	"context"
	"encoding/json"
	"net/http"
	"songlist/managers"
	"strings"

	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
)

var indexLog = logrus.WithField("fun", "index")

type ResponseData struct {
	Message string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
}

func Index() {
	http.Handle("/ping", http.HandlerFunc(ping))
	user()
	songlist()
}

func ping(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Pong!"))
}

func cors(next http.Handler, methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
			w.Header().Set("Access-Control-Allow-Headers",
				"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		next.ServeHTTP(w, r)
	})
}

func verify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("token")
		if err != nil {
			// 没有携带Token
			indexLog.WithError(err).Error("No Cookie: token.")
			http.Error(w, "No Cookie: token.", http.StatusUnauthorized)
			return
		}

		// 判断Token是否正常
		username, err := managers.Redis.HGet(managers.TOKEN+token.Value, "username").Result()
		if err != nil {
			if err == redis.Nil {
				// 没有找到该Token
				http.Error(w, "Not login", http.StatusUnauthorized)
			} else {
				// 缓存错误
				indexLog.WithError(err).Error("Verify token failed.")
				http.Error(w, "Cache error.", http.StatusInternalServerError)
			}
			return
		}
		c := context.WithValue(r.Context(), "username", username)
		next.ServeHTTP(w, r.WithContext(c))
	})
}

func response(w http.ResponseWriter, resData ResponseData) {
	// 返回信息
	if err := json.NewEncoder(w).Encode(resData); err != nil {
		indexLog.WithField("err", err).Error("Return data failed.")
		http.Error(w, "Return data Failed.", http.StatusInternalServerError)
		return
	}
}
