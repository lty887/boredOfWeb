package api

import (
	"net/http"
	"strconv"

	"boring-websites-admin/internal/auth"
	"boring-websites-admin/internal/middleware"
	"boring-websites-admin/internal/models"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册API路由
func RegisterRoutes(r *gin.Engine) {
	// API路由组
	api := r.Group("/api")

	// 公开路由
	api.POST("/login", login)
	// 公开获取网页列表
	api.GET("/websites", getWebsites)

	// 需要认证的路由
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// 网页链接管理
		protected.GET("/websites/:id", getWebsite)
		protected.POST("/websites", createWebsite)
		protected.PUT("/websites/:id", updateWebsite)
		protected.DELETE("/websites/:id", deleteWebsite)

		// 网站设置管理
		protected.GET("/settings", getSettings)
		protected.GET("/settings/:key", getSetting)
		protected.POST("/settings", createSetting)
		protected.PUT("/settings/:key", updateSetting)
		protected.DELETE("/settings/:key", deleteSetting)
	}
}

// login 登录API
type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

func login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 获取管理员
	admin, err := models.GetAdminByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// 检查密码
	if !admin.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// 生成token
	token, err := auth.GenerateToken(admin.ID, admin.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 返回响应
	var resp loginResponse
	resp.Token = token
	resp.User.ID = admin.ID
	resp.User.Username = admin.Username

	c.JSON(http.StatusOK, resp)
}

// getWebsites 获取所有网页链接
func getWebsites(c *gin.Context) {
	websites, err := models.GetAllWebsites()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get websites"})
		return
	}

	c.JSON(http.StatusOK, websites)
}

// getWebsite 根据ID获取网页链接
func getWebsite(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}

	website, err := models.GetWebsiteByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website not found"})
		return
	}

	c.JSON(http.StatusOK, website)
}

// createWebsite 创建网页链接
type websiteRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	URL         string `json:"url" binding:"required"`
}

func createWebsite(c *gin.Context) {
	var req websiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	website := models.Website{
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
	}

	if err := models.CreateWebsite(&website); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create website"})
		return
	}

	c.JSON(http.StatusCreated, website)
}

// updateWebsite 更新网页链接
func updateWebsite(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}

	var req websiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	website, err := models.GetWebsiteByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website not found"})
		return
	}

	website.Title = req.Title
	website.Description = req.Description
	website.URL = req.URL

	if err := models.UpdateWebsite(&website); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update website"})
		return
	}

	c.JSON(http.StatusOK, website)
}

// deleteWebsite 删除网页链接
func deleteWebsite(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}

	if err := models.DeleteWebsite(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete website"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Website deleted successfully"})
}

// getSettings 获取所有设置
func getSettings(c *gin.Context) {
	settings, err := models.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// getSetting 根据键获取设置
func getSetting(c *gin.Context) {
	key := c.Param("key")
	setting, err := models.GetSettingByKey(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// createSetting 创建设置
type settingRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

func createSetting(c *gin.Context) {
	var req settingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	setting := models.Setting{
		Key:   req.Key,
		Value: req.Value,
	}

	if err := models.CreateSetting(&setting); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create setting"})
		return
	}

	c.JSON(http.StatusCreated, setting)
}

// updateSetting 更新设置
func updateSetting(c *gin.Context) {
	key := c.Param("key")
	var req settingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	setting, err := models.GetSettingByKey(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	setting.Value = req.Value

	if err := models.UpdateSetting(&setting); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// deleteSetting 删除设置
func deleteSetting(c *gin.Context) {
	key := c.Param("key")

	if err := models.DeleteSetting(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setting deleted successfully"})
}