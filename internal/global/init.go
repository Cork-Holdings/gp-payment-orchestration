package global

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/casbin/casbin/v2"
	sio "github.com/doquangtan/socketio/v4"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type Model interface {
	Permissions() map[string][]string
	TableName() string
	AutoMigrate(*gorm.DB) error
}
type App struct {
	DB          *gorm.DB
	Mongo       *mongo.Client
	MongoDBName string
	Cache       *redis.Client
	Validator   *validator.Validate
	TaskQ       *asynq.Client
	Perms       *casbin.Enforcer
	Models      []Model
	IO          *sio.Io
	MQ          *rmq
}

var a *App
var o sync.Once

func New() *App {
	o.Do(func() {
		a = &App{
			DB: GetDB(),
			// Mongo:       GetMongo(),
			// MongoDBName: GetMongoDBName(),
			Validator: GetValidator(),
			// TaskQ:       GetTaskQueue(),
			Cache: GetCache(),
			IO:    GetSocketIO(),
			MQ:    GetMQ(),
		}
	})

	return a
}

func (a *App) Close() {
	if a.TaskQ != nil {
		a.TaskQ.Close()
	}
	if a.Cache != nil {
		a.Cache.Close()
	}
	closeDB()
	closeMongo()
	if io != nil {
		io.Close()
	}
}

func (a *App) Register(models ...Model) {
	a.Models = append(a.Models, models...)
	for _, m := range a.Models {
		fmt.Println(m.TableName())
		if err := m.AutoMigrate(a.DB); err != nil {
			log.Print(err)
		}
	}
}

func (a *App) GetPermissions(group string) map[string][]string {
	permissions := map[string][]string{}
	for _, mm := range a.Models {
		permissions[mm.TableName()] = mm.Permissions()[group]
	}

	j, _ := json.MarshalIndent(permissions, "", " ")
	fmt.Println(string(j))

	return permissions
}
