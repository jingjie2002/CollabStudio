package database

import (
	"collab-server/config"
	"collab-server/models"
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	var err error
	dbName := config.GetEnv("DB_NAME", "collab.db")

	// 开启 LogMode(logger.Info) 可以看到 SQL 语句，调试很方便
	DB, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 🛠️ 更新：自动迁移 User, Document 和 Message
	err = DB.AutoMigrate(&models.User{}, &models.Document{}, &models.Message{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("✅ Database connected and migrated successfully!")
}
