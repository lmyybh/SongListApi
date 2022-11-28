package managers

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbLog = logrus.WithField("fun", "db")

var DB *mongo.Client

func InitDatabase(wg *sync.WaitGroup) {
	// 连接到 MongoDB
	var err error
	DB, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(CONFIG.DB.URI))
	if err != nil {
		dbLog.WithError(err).Panic("Open mongodb failed")
	}

	dbLog.Info("Have connected to mongodb")
	wg.Done()
}
