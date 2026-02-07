package config

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/task-monitor/api-server/internal/model"
)

// InitDB 初始化数据库连接
func InitDB(cfg *DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	// 配置GORM日志
	gormLogger := logger.Default.LogMode(logger.Info)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取底层的sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")
	return db, nil
}

// AutoMigrateAndSeed 自动建表并创建默认用户
func AutoMigrateAndSeed(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return fmt.Errorf("failed to migrate users table: %w", err)
	}

	var count int64
	db.Model(&model.User{}).Count(&count)
	if count == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash default password: %w", err)
		}
		defaultAdmin := model.User{
			Username: "admin",
			Password: string(hashedPassword),
		}
		if err := db.Create(&defaultAdmin).Error; err != nil {
			return fmt.Errorf("failed to create default admin: %w", err)
		}
		log.Println("Default admin user created (admin/admin123)")
	}
	return nil
}
