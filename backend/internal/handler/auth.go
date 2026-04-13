package handler

import (

	"github.com/gin-gonic/gin"
	"github.com/xiaoran/doutok/internal/model"
	"github.com/xiaoran/doutok/internal/pkg/auth"
	"github.com/xiaoran/doutok/internal/pkg/errno"
	"github.com/xiaoran/doutok/internal/pkg/logger"
	"github.com/xiaoran/doutok/internal/pkg/response"
	"github.com/xiaoran/doutok/internal/pkg/snowflake"
	"github.com/xiaoran/doutok/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var userRepo = repository.NewUserRepo()

type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname" binding:"required,min=1,max=64"`
	Phone    string `json:"phone"`
}

func Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, err.Error())
		return
	}

	// 检查用户名是否已存在
	existing, _ := userRepo.GetByUsername(c.Request.Context(), req.Username)
	if existing != nil {
		response.ErrorWithMsg(c, 200, errno.ErrUserExists, "用户名已存在")
		return
	}

	// bcrypt 加密密码
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("bcrypt hash failed", "err", err)
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}

	user := &model.User{
		ID:       snowflake.GenID(),
		Username: req.Username,
		Password: string(hashedPwd),
		Nickname: req.Nickname,
		Phone:    req.Phone,
		Status:   1,
		Role:     "user",
	}

	if err := userRepo.Create(c.Request.Context(), user); err != nil {
		logger.Error("create user failed", "err", err)
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "注册失败")
		return
	}

	// 生成 JWT
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		logger.Error("generate token failed", "err", err)
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}

	logger.Info("user registered", "user_id", user.ID, "username", user.Username)
	response.Success(c, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"token":    token,
	})
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInvalidParam, err.Error())
		return
	}

	user, err := userRepo.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		logger.Error("query user failed", "err", err)
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	if user == nil {
		response.ErrorWithMsg(c, 200, errno.ErrAuthFailed, "用户名或密码错误")
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrAuthFailed, "用户名或密码错误")
		return
	}

	if user.Status == 2 {
		response.ErrorWithMsg(c, 200, errno.ErrForbidden, "账号已被封禁")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		logger.Error("generate token failed", "err", err)
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}

	logger.Info("user logged in", "user_id", user.ID, "username", user.Username)
	response.Success(c, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"token":    token,
	})
}

func RefreshToken(c *gin.Context) {
	userID := c.GetInt64("user_id")
	username := c.GetString("username")
	role := c.GetString("role")

	token, err := auth.GenerateToken(userID, username, role)
	if err != nil {
		response.ErrorWithMsg(c, 200, errno.ErrInternal, "")
		return
	}
	response.Success(c, gin.H{"token": token})
}
