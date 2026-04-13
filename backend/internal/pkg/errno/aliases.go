package errno

// 别名 - 让 handler 代码更直观
var (
	ErrInvalidParam = ErrParam
	ErrInternal     = ErrServer
	ErrAuthFailed   = ErrAuth
	ErrForbidden    = ErrPermission
)
