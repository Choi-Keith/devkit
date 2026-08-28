// Package ecode 定义本服务的业务错误码。
// 通用错误码(1000-17999)由 golang-common 内置,业务错误码从 20000 开始,按模块分文件注册。
package ecode

import (
	"net/http"

	"github.com/Choi-Keith/golang-common/ecode"
)

var (
	ErrUserNotFound  = ecode.New(20000, "user not found", ecode.WithHTTPStatus(http.StatusNotFound))
	ErrEmailExists   = ecode.New(20001, "email already exists", ecode.WithHTTPStatus(http.StatusConflict))
	ErrInvalidParams = ecode.New(20002, "invalid parameters", ecode.WithHTTPStatus(http.StatusBadRequest))
)
