package repository

import (
	"fmt"
	"os"

	conf "{{.PkgName}}/config"

	"entgo.io/ent/dialect/sql"
	"{{.PkgName}}/repository/ent"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/shenjing023/llog"
	"golang.org/x/net/context"
)

var (
	redisClient *redis.Client
	entClient   *ent.Client
)

// Init init mysql and redis orm
func Init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.ServerCfg.Redis.Host, conf.ServerCfg.Redis.Port),
		Password: conf.ServerCfg.Redis.Password,
		DB:       0,
	})
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Error("connect to redis error: ", err)
		os.Exit(1)
	}

	dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.ServerCfg.DB.User, conf.ServerCfg.DB.Password, conf.ServerCfg.DB.Host, conf.ServerCfg.DB.Port, conf.ServerCfg.DB.Dbname)

	drv, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Error("mysql connection error: ", err)
		os.Exit(1)
	}
	// 获取数据库驱动中的sql.DB对象。
	db := drv.DB()
	if conf.ServerCfg.DB.MaxIdle > 0 {
		db.SetMaxIdleConns(conf.ServerCfg.DB.MaxIdle)
	}
	if conf.ServerCfg.DB.MaxOpen > 0 {
		db.SetMaxOpenConns(conf.ServerCfg.DB.MaxOpen)
	}
	entClient = ent.NewClient(ent.Driver(drv))
}

// Close close db connection
func Close() {
	entClient.Close()
	redisClient.Close()
}
