package errno

import "fmt"

type Errno struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Errno) Error() string {
	return fmt.Sprintf("code=%d, msg=%s", e.Code, e.Msg)
}

func New(code int, msg string) *Errno {
	return &Errno{Code: code, Msg: msg}
}

var (
	OK                 = New(0, "success")
	ErrServer          = New(10001, "server internal error")
	ErrParam           = New(10002, "invalid parameter")
	ErrAuth            = New(10003, "authentication failed")
	ErrTokenExpired    = New(10004, "token expired")
	ErrPermission      = New(10005, "permission denied")
	ErrNotFound        = New(10006, "resource not found")
	ErrTooManyRequests = New(10007, "too many requests")
	ErrUserExists      = New(20001, "user already exists")
	ErrUserNotFound    = New(20002, "user not found")
	ErrPasswordWrong   = New(20003, "incorrect password")
	ErrVideoNotFound   = New(30001, "video not found")
	ErrVideoProcessing = New(30002, "video is processing")
	ErrUploadFailed    = New(30003, "upload failed")
	ErrCommentNotFound = New(40001, "comment not found")
	ErrLiveNotFound    = New(50001, "live room not found")
	ErrLiveEnded       = New(50002, "live has ended")
	ErrChatRoomFull    = New(60001, "chat room is full")
)
