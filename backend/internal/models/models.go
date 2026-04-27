package models

import (
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB 全局数据库连接
var DB *gorm.DB

// Admin 管理员模型
type Admin struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Website 网页链接模型
type Website struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	URL         string    `json:"url" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Setting 网站设置模型
type Setting struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex;not null"`
	Value     string    `json:"value" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InitDB 初始化数据库连接
func InitDB() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "./database.db"
	}

	db, err := gorm.Open(sqlite.Open(dbURL), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(&Admin{}, &Website{}, &Setting{})
	if err != nil {
		return err
	}

	DB = db
	return nil
}

// InitAdmin 初始化管理员账号
func InitAdmin() error {
	// 检查是否已有管理员账号
	var count int64
	DB.Model(&Admin{}).Count(&count)
	if count > 0 {
		return nil // 已有管理员账号，无需初始化
	}

	// 获取管理员账号配置
	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")

	if username == "" {
		username = "admin"
	}

	if password == "" {
		password = "admin123"
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建管理员账号
	admin := Admin{
		Username: username,
		Password: string(hashedPassword),
	}

	return DB.Create(&admin).Error
}

// GetAllWebsites 获取所有网页链接
func GetAllWebsites() ([]Website, error) {
	var websites []Website
	err := DB.Find(&websites).Error
	return websites, err
}

// GetWebsiteByID 根据ID获取网页链接
func GetWebsiteByID(id uint) (Website, error) {
	var website Website
	err := DB.First(&website, id).Error
	return website, err
}

// CreateWebsite 创建网页链接
func CreateWebsite(website *Website) error {
	return DB.Create(website).Error
}

// UpdateWebsite 更新网页链接
func UpdateWebsite(website *Website) error {
	return DB.Save(website).Error
}

// DeleteWebsite 删除网页链接
func DeleteWebsite(id uint) error {
	return DB.Delete(&Website{}, id).Error
}

// GetAdminByUsername 根据用户名获取管理员
func GetAdminByUsername(username string) (Admin, error) {
	var admin Admin
	err := DB.Where("username = ?", username).First(&admin).Error
	return admin, err
}

// CheckPassword 检查密码是否正确
func (a *Admin) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}

// GetAllSettings 获取所有设置
func GetAllSettings() ([]Setting, error) {
	var settings []Setting
	err := DB.Find(&settings).Error
	return settings, err
}

// GetSettingByKey 根据键获取设置
func GetSettingByKey(key string) (Setting, error) {
	var setting Setting
	err := DB.Where("key = ?", key).First(&setting).Error
	return setting, err
}

// CreateSetting 创建设置
func CreateSetting(setting *Setting) error {
	return DB.Create(setting).Error
}

// UpdateSetting 更新设置
func UpdateSetting(setting *Setting) error {
	return DB.Save(setting).Error
}

// UpsertSetting 创建或更新设置
func UpsertSetting(key, value string) error {
	var setting Setting
	result := DB.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		// 如果不存在，创建新设置
		setting = Setting{
			Key:   key,
			Value: value,
		}
		return DB.Create(&setting).Error
	} else {
		// 如果存在，更新设置
		setting.Value = value
		return DB.Save(&setting).Error
	}
}

// DeleteSetting 删除设置
func DeleteSetting(key string) error {
	return DB.Where("key = ?", key).Delete(&Setting{}).Error
}