package userhttp

import (
	neth "net/http"

	"github.com/Choi-Keith/golang-common/ecode"
)

// Business error codes of this service (20000+). Codes are process-global
// and registered exactly once here, at the transport boundary.
var (
	ErrUserNotFound = ecode.New(20000, "用户不存在", ecode.WithHTTPStatus(neth.StatusNotFound))
	ErrEmailTaken   = ecode.New(20001, "邮箱已被注册", ecode.WithHTTPStatus(neth.StatusConflict))
)
