package managers

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbLog = logrus.WithField("fun", "db")

var DB *mongo.Database

func InitDatabase(wg *sync.WaitGroup) {
	// 连接到 MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(CONFIG.DB.URI))
	if err != nil {
		dbLog.WithError(err).Panic("Open mongodb failed")
	}

	// 检查连接
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		dbLog.WithError(err).Panic("Ping mongodb failed")
	}

	DB = client.Database(CONFIG.DB.Database)

	dbLog.Info("Have connected to mongodb")
	wg.Done()
}
