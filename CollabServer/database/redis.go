package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"collab-server/config"
	"github.com/redis/go-redis/v9"
)

// RDB 是全局的 Redis 客户端
var RDB *redis.Client
var ctx = context.Background()

// InitRedis 初始化 Redis 连接
func InitRedis() {
	// 从环境变量获取 Redis 配置
	addr := config.GetEnv("REDIS_ADDR", "127.0.0.1:6379")
	password := config.GetEnv("REDIS_PASSWORD", "")
	db := 0 // 默认使用 0 号数据库

	// 🟢 创建 Redis 客户端
	// 这里配置了连接池和超时时间，确保在高并发下的稳定性
	RDB = redis.NewClient(&redis.Options{
		Addr:         addr,            // Redis 地址
		Password:     password,        // 密码
		DB:           db,              // 数据库编号
		PoolSize:     10,              // 连接池大小：最多同时保持 10 个连接
		MinIdleConns: 5,               // 最小空闲连接数：保持 5 个连接随时可用
		DialTimeout:  5 * time.Second, // 建立连接超时：5秒
		ReadTimeout:  3 * time.Second, // 读取数据超时：3秒
		WriteTimeout: 3 * time.Second, // 写入数据超时：3秒
		PoolTimeout:  4 * time.Second, // 等待可用连接超时：4秒
	})

	// 🟢 报错防范：通过 Ping 测试连接是否成功
	// 我们使用 5 秒超时的上下文来测试连接
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := RDB.Ping(testCtx).Result()
	if err != nil {
		// 如果连接失败，我们打印错误但不要让程序 Fatal 闪退
		// 这样可以让程序在 Redis 挂掉时依然能运行（虽然部分功能受限）
		log.Printf("⚠️ Redis 连接失败: %v. 部分实时功能可能受限。", err)
		// 如果你希望 Redis 必须在线，可以使用 log.Fatal(err)
	} else {
		fmt.Println("✅ Redis connected successfully!")
	}
}
