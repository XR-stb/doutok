package auth

var defaultManager *JWTManager

// InitJWT 初始化全局 JWT 管理器
func InitJWT(secret, issuer string, expireHour int) {
	defaultManager = NewJWTManager(secret, issuer, expireHour)
}

// GenerateToken 使用全局管理器生成 token
func GenerateToken(userID int64, username, role string) (string, error) {
	if defaultManager == nil {
		defaultManager = NewJWTManager("default-secret", "doutok", 72)
	}
	return defaultManager.Generate(userID, username, role)
}

// ParseToken 使用全局管理器解析 token
func ParseToken(tokenStr string) (*Claims, error) {
	if defaultManager == nil {
		defaultManager = NewJWTManager("default-secret", "doutok", 72)
	}
	return defaultManager.Parse(tokenStr)
}
