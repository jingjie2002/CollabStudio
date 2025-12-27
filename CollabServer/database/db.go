package database

import (
	"collab-server/config"
	"collab-server/models"
	"fmt"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	var err error
	dbName := config.GetEnv("DB_NAME", "collab.db")

	// 🟢 权限修复：确保数据库文件可读写 (0644)
	// 如果文件存在但权限不对，尝试修复
	if _, statErr := os.Stat(dbName); statErr == nil {
		if chmodErr := os.Chmod(dbName, 0644); chmodErr != nil {
			log.Printf("⚠️ 无法修改数据库权限: %v", chmodErr)
		} else {
			log.Printf("✅ 数据库文件权限已设置为 0644: %s", dbName)
		}
	}

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

// =============================================================================
// Close 优雅停机专用：安全关闭数据库连接
// =============================================================================
// GORM 底层使用 database/sql 的连接池。
// 直接退出程序不会自动关闭连接，可能导致：
// 1. 未提交的事务丢失
// 2. SQLite 的 WAL 日志未刷入主文件
// 3. 连接资源泄漏
//
// 调用 Close() 会：
// 1. 等待所有活跃查询完成
// 2. 关闭连接池中的所有连接
// 3. 刷新 SQLite 的 WAL 日志
// =============================================================================
func Close() {
	if DB == nil {
		return
	}

	// GORM v2 需要通过 DB.DB() 获取底层的 *sql.DB 对象
	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("⚠️ 获取数据库连接失败: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("⚠️ 关闭数据库连接失败: %v", err)
	} else {
		fmt.Println("✅ 数据库连接已安全关闭")
	}
}
