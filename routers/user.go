package routers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"songlist/managers"
	"songlist/models"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userLog = logrus.WithField("fun", "userRouter")

const userParty = "/user"

func user() {
	http.Handle(userParty+"/checkToken", cors(verify(http.HandlerFunc(checkToken)), http.MethodGet))
	http.Handle(userParty+"/register", cors(http.HandlerFunc(register), http.MethodPost))
	http.Handle(userParty+"/login", cors(http.HandlerFunc(login), http.MethodPost))
	http.Handle(userParty+"/logout", cors(http.HandlerFunc(logout), http.MethodGet))
}

func checkToken(w http.ResponseWriter, r *http.Request) {
	response(w, ResponseData{Message: "ok"})
}

func register(w http.ResponseWriter, r *http.Request) {
	var jsonData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&jsonData)
	if err != nil {
		userLog.Error("Get params failed.")
		http.Error(w, "Get params failed.", http.StatusBadRequest)
		return
	}

	// 检查参数
	username := jsonData["username"].(string)
	password := jsonData["password"].(string)
	if username == "" || password == "" {
		userLog.Error("Need username and password.")
		http.Error(w, "Need username and password.", http.StatusBadRequest)
		return
	}

	coll := managers.DB.Collection("user")

	// 判断用户名是否存在
	var result bson.M
	if err := coll.FindOne(context.TODO(), bson.M{"username": username}).Decode(&result); err != nil {
		if err != mongo.ErrNoDocuments {
			userLog.WithField("err", err).Error("Database error.")
			http.Error(w, "Database error.", http.StatusInternalServerError)
			return
		}

	} else {
		userLog.WithField("err", err).Error("Username has been registered")
		http.Error(w, "Username has been registered", http.StatusInternalServerError)
		return
	}

	// 创建用户
	user := models.User{UserName: username}
	_, err = models.Register(&user, password)
	if err != nil {
		userLog.WithField("err", err).Error("Database error.")
		http.Error(w, "Database error.", http.StatusInternalServerError)
		return
	}

	userLog.Info("The user: " + username + " is registered successfully.")
	response(w, ResponseData{Message: "ok"})
}

// POST login
// PARAMS username=string&password=string
// RETURN JSON{id: string, token: string, name: string, sex: enum?, face: string, desc: string}

func login(w http.ResponseWriter, r *http.Request) {
	var jsonData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&jsonData)
	if err != nil {
		userLog.Error("Login failed.")
		http.Error(w, "Login failed.", http.StatusBadRequest)
		return
	}

	username := jsonData["username"].(string)
	password := jsonData["password"].(string)
	if username == "" || password == "" {
		userLog.Error("Login failed.")
		http.Error(w, "Login failed.", http.StatusBadRequest)
		return
	}

	coll := managers.DB.Collection("user")
	//取出user的密文与盐，并验证用户和密码
	var user models.User
	if err := coll.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			userLog.WithField("err", err).Error("Username or password is wrong.")
			http.Error(w, "Username or password is wrong.", http.StatusBadRequest)
		} else {
			userLog.WithField("err", err).Error("Database error.")
			http.Error(w, "Database error.", http.StatusInternalServerError)
		}
		return
	}

	if !bytes.Equal(user.Password, models.PasswordMaker(password, user.Salt)) {
		userLog.Error("Username or password is wrong.")
		http.Error(w, "Username or password is wrong.", http.StatusBadRequest)
		return
	}

	if err := models.Login(w, username); err != nil {
		userLog.WithField("err", err).Error("Login failed.")
		http.Error(w, "Login failed.", http.StatusInternalServerError)
		return
	}

	userLog.Info("The user: " + username + " is logined successfully.")
	response(w, ResponseData{Message: "ok", Data: username})
}

func logout(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	if err != nil {
		userLog.Error("No Cookie: token")
		http.Error(w, "No Cookie: token", http.StatusBadRequest)
		return
	}
	if err = managers.Redis.Del(managers.TOKEN + token.Value).Err(); err != nil {
		userLog.WithField("err", err).Error("No token is stored.")
		http.Error(w, "No token is stored.", http.StatusUnauthorized)
		return
	}

	userLog.Info("Logout successfully.")
	response(w, ResponseData{Message: "ok"})
}
