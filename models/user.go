package models

import (
	"context"
	"crypto/sha256"
	"net/http"
	"songlist/managers"
	"songlist/utils"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xdg-go/pbkdf2"
	"go.mongodb.org/mongo-driver/mongo"
)

var userLog = logrus.WithField("fun", "userModel")

type User struct {
	UserName    string   `bson:"username"`
	Salt        []byte   `bson:"salt,omitempty"`
	Password    []byte   `bson:"password"`
	PlayingList []string `bson:"playinglist,omitempty"`
}

func PasswordMaker(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, 4096, 128, sha256.New)
}

// 生成登录令牌
func TokenMaker() string {
	return utils.RandomURLBase64(24)
}

func Register(user *User, password string) (*mongo.InsertOneResult, error) {
	user.Salt = utils.RandomBytes(32)
	user.Password = PasswordMaker(password, user.Salt)
	//向数据库中储存
	coll := managers.DB.Collection("user")
	return coll.InsertOne(context.TODO(), user)
}

func Login(w http.ResponseWriter, username string) error {
	// 记录登录信息
	token := TokenMaker()
	if err := managers.HMSetAndExpire(managers.TOKEN+token, map[string]interface{}{"username": username, "loginTime": time.Now().UnixNano() / 1e6}, managers.UserCacheLife); err != nil {
		userLog.WithField("err", err).Error("Set user cache failed.")
		return err
	}
	//v := http.Cookie{Name: "token", Value: token, Path: "/", SameSite: 4, Secure: true}
	v := http.Cookie{Name: "token", Value: token, Path: "/"}
	http.SetCookie(w, &v)

	return nil
}
