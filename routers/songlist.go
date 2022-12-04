package routers

import (
	"context"
	"net/http"
	"songlist/managers"
	"songlist/models"
	"songlist/utils"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/wxnacy/wgo/arrays"
	"go.mongodb.org/mongo-driver/bson"
)

var songLog = logrus.WithField("fun", "songlistRouter")

const songParty = "/songlist"

func songlist() {
	http.Handle(songParty+"/get", cors(verify(http.HandlerFunc(getSongLists)), http.MethodGet))
	http.Handle(songParty+"/create", cors(verify(http.HandlerFunc(createSongList)), http.MethodPost))
	http.Handle(songParty+"/delete", cors(verify(http.HandlerFunc(deleteSongList)), http.MethodPost))
	http.Handle(songParty+"/getSongs", cors(verify(http.HandlerFunc(getSongsInSongList)), http.MethodPost))
	http.Handle(songParty+"/insertSongs", cors(verify(http.HandlerFunc(insertSongsInSongList)), http.MethodPost))
	http.Handle(songParty+"/deleteSongs", cors(verify(http.HandlerFunc(deleteSongsInSongList)), http.MethodPost))
}

func getSongLists(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	coll := managers.DB.Collection("user")
	var user models.User
	if err := coll.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user); err != nil {
		songLog.WithField("err", err).Error("Database error.")
		http.Error(w, "Database error.", http.StatusInternalServerError)
		return
	}

	if user.SongLists == nil {
		response(w, ResponseData{Message: "ok", Data: make([]interface{}, 0)})
	} else {
		response(w, ResponseData{Message: "ok", Data: user.SongLists})
	}
}

func createSongList(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	songmids := utils.ReadStringArray(r.PostFormValue("songmids"))

	if title == "" {
		songLog.Error("Need title.")
		http.Error(w, "Need title.", http.StatusBadRequest)
		return
	}

	songlist := models.SongList{Title: title, Songmids: songmids}

	// 获取对应的用户
	username := r.Context().Value("username").(string)
	coll := managers.DB.Collection("user")
	var user models.User
	if err := coll.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user); err != nil {
		songLog.WithField("err", err).Error("Database error.")
		http.Error(w, "Database error.", http.StatusInternalServerError)
		return
	}

	// 获取用户现有歌单
	songlists := user.SongLists

	// title 查重
	for _, info := range songlists {
		if info.Title == title {
			songLog.Error("Duplicate title.")
			http.Error(w, "Duplicate title.", http.StatusBadRequest)
			return
		}
	}

	// 更新歌单列表
	songlists = append(songlists, songlist)
	filter := bson.D{{Key: "username", Value: username}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "songlists", Value: songlists}}}}
	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		songLog.Error("Database error.")
		http.Error(w, "Database error.", http.StatusInternalServerError)
		return
	}

	response(w, ResponseData{Message: "ok", Data: songlists})
}

func deleteSongList(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	if title == "" {
		songLog.Error("Need title.")
		http.Error(w, "Need title.", http.StatusBadRequest)
		return
	}
	// 获取对应的用户
	username := r.Context().Value("username").(string)
	coll := managers.DB.Collection("user")
	var user models.User
	if err := coll.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user); err != nil {
		songLog.WithField("err", err).Error("Database error.")
		http.Error(w, "Database error.", http.StatusInternalServerError)
		return
	}

	// 获取用户现有歌单
	songlists := user.SongLists

	newSongLists := make([]models.SongList, 0)
	// title 对比
	removed := false
	for _, info := range songlists {
		if info.Title != title {
			newSongLists = append(newSongLists, info)
		} else {
			removed = true
		}
	}

	if removed {
		// 更新歌单
		filter := bson.D{{Key: "username", Value: username}}
		update := bson.D{{Key: "$set", Value: bson.D{{Key: "songlists", Value: newSongLists}}}}
		_, err := coll.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			songLog.Error("Database error.")
			http.Error(w, "Database error.", http.StatusInternalServerError)
			return
		}
		response(w, ResponseData{Message: "ok", Data: newSongLists})
	} else {
		songLog.Error("There is no corresponding songlist.")
		http.Error(w, "There is no corresponding songlist.", http.StatusBadRequest)
	}
}

func getSongsInSongList(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	if title == "" {
		songLog.Error("Need title.")
		http.Error(w, "Need title.", http.StatusBadRequest)
		return
	}

	// 获取对应的用户
	username := r.Context().Value("username").(string)
	coll := managers.DB.Collection("user")
	var user models.User
	if err := coll.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user); err != nil {
		songLog.WithField("err", err).Error("Database error.")
		http.Error(w, "Database error.", http.StatusInternalServerError)
		return
	}

	for _, info := range user.SongLists {
		if info.Title == title {
			response(w, ResponseData{Message: "ok", Data: info})
			return
		}
	}

	songLog.Error("The songlist doesn't exists.")
	http.Error(w, "The songlist doesn't exists.", http.StatusBadRequest)
}

func insertSongsInSongList(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	songmids := utils.ReadStringArray(r.PostFormValue("songmids"))
	if title == "" || songmids == nil {
		songLog.Error("Need title and songmids.")
		http.Error(w, "Need title and songmids.", http.StatusBadRequest)
		return
	}

	// 获取对应的用户
	username := r.Context().Value("username").(string)
	coll := managers.DB.Collection("user")
	var user models.User
	if err := coll.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user); err != nil {
		songLog.WithField("err", err).Error("Database error.")
		http.Error(w, "Database error.", http.StatusInternalServerError)
		return
	}

	newSongLists := make([]models.SongList, 0)
	inserted := false
	for _, info := range user.SongLists {
		if info.Title == title {
			currentSongmids := info.Songmids

			// 去重
			toInsertSongmids := make([]string, 0)
			for _, mid := range songmids {
				if arrays.ContainsString(currentSongmids, mid) == -1 && arrays.ContainsString(toInsertSongmids, mid) == -1 {
					toInsertSongmids = append(toInsertSongmids, mid)
				}
			}

			if len(toInsertSongmids) == 0 {
				songLog.Error("The songs already exist.")
				http.Error(w, "The songs already exist.", http.StatusBadRequest)
				return
			}

			inserted = true

			// 获取插入位置
			var index int
			var err error
			if r.PostFormValue("index") == "" {
				index = len(currentSongmids)
			} else {
				index, err = strconv.Atoi(r.PostFormValue("index"))
				if err != nil {
					songLog.WithField("err", err).Error("Index error.")
					http.Error(w, "Index error.", http.StatusBadRequest)
					return
				}
			}

			if index < 0 || index > len(currentSongmids) {
				songLog.Error("Index error.")
				http.Error(w, "Index error.", http.StatusBadRequest)
				return
			}

			// 插入数据
			newSongmids := make([]string, len(currentSongmids)+len(toInsertSongmids))
			copy(newSongmids[:index], currentSongmids[:index])
			copy(newSongmids[index:index+len(toInsertSongmids)], toInsertSongmids)
			copy(newSongmids[index+len(toInsertSongmids):], currentSongmids[index:])

			newSongLists = append(newSongLists, models.SongList{Title: title, Songmids: newSongmids})

		} else {
			newSongLists = append(newSongLists, info)
		}
	}

	if inserted {
		filter := bson.D{{Key: "username", Value: username}}
		update := bson.D{{Key: "$set", Value: bson.D{{Key: "songlists", Value: newSongLists}}}}
		_, err := coll.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			songLog.Error("Database error.")
			http.Error(w, "Database error.", http.StatusInternalServerError)
			return
		}
		response(w, ResponseData{Message: "ok", Data: newSongLists})
	} else {
		songLog.Error("There is no corresponding songlist.")
		http.Error(w, "There is no corresponding songlist.", http.StatusBadRequest)
		return
	}
}

func deleteSongsInSongList(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	songmids := utils.ReadStringArray(r.PostFormValue("songmids"))
	if title == "" || songmids == nil {
		songLog.Error("Need title and songmids.")
		http.Error(w, "Need title and songmids.", http.StatusBadRequest)
		return
	}

	// 获取对应的用户
	username := r.Context().Value("username").(string)
	coll := managers.DB.Collection("user")
	var user models.User
	if err := coll.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user); err != nil {
		songLog.WithField("err", err).Error("Database error.")
		http.Error(w, "Database error.", http.StatusInternalServerError)
		return
	}

	newSongLists := make([]models.SongList, 0)
	removed := false
	for _, info := range user.SongLists {
		if info.Title == title {
			newSongmids := make([]string, 0)

			for _, mid := range info.Songmids {
				if arrays.ContainsString(songmids, mid) == -1 {
					newSongmids = append(newSongmids, mid)
				} else {
					removed = true
				}
			}

			if len(newSongmids) >= len(info.Songmids) {
				songLog.Error("The songs don't exist.")
				http.Error(w, "The songs don't exist.", http.StatusBadRequest)
				return
			}

			newSongLists = append(newSongLists, models.SongList{Title: title, Songmids: newSongmids})

		} else {
			newSongLists = append(newSongLists, info)
		}
	}

	if removed {
		filter := bson.D{{Key: "username", Value: username}}
		update := bson.D{{Key: "$set", Value: bson.D{{Key: "songlists", Value: newSongLists}}}}
		_, err := coll.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			songLog.Error("Database error.")
			http.Error(w, "Database error.", http.StatusInternalServerError)
			return
		}
		response(w, ResponseData{Message: "ok", Data: newSongLists})
	} else {
		songLog.Error("There is no corresponding songlist.")
		http.Error(w, "There is no corresponding songlist.", http.StatusBadRequest)
		return
	}
}
