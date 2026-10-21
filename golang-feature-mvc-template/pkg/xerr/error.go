// Package xerr 集中定义全部业务错误码(类型为 golang-common/ecode 的
// *ecode.Error),按 feature 分区。service 业务失败直接返回这里的错误,
// handler 将错误原样交给 response.Error(内部 ecode.FromError 识别
// *ecode.Error 并保留 code/HTTP status)。
//
// 号段约定:通用码 1000-17999 由 golang-common 内置;业务码 20000 起,
// 每个 feature 占 100 个号段。新 feature 在本文末追加自己的分区。
// ecode.New 对重复 code 会 panic,code 全局唯一。
package xerr

import (
	neth "net/http"

	"github.com/Choi-Keith/golang-common/ecode"
)

// ---- user 模块(20000-20099)----
var (
	ErrUserNotFound        = ecode.New(20000, "用户不存在", ecode.WithHTTPStatus(neth.StatusNotFound))
	ErrEmailTaken          = ecode.New(20001, "邮箱已被注册", ecode.WithHTTPStatus(neth.StatusConflict))
	ErrInvalidCredential   = ecode.New(20002, "邮箱或密码错误", ecode.WithHTTPStatus(neth.StatusUnauthorized))
	ErrRefreshTokenInvalid = ecode.New(20003, "refresh token 无效", ecode.WithHTTPStatus(neth.StatusUnauthorized))
	ErrRefreshTokenRevoked = ecode.New(20004, "refresh token 已失效,请重新登录", ecode.WithHTTPStatus(neth.StatusUnauthorized))
)

// ---- order 模块(20100-20199)----
var (
	ErrOrderNotFound       = ecode.New(20100, "订单不存在", ecode.WithHTTPStatus(neth.StatusNotFound))
	ErrOrderNotCancellable = ecode.New(20101, "订单状态不允许取消", ecode.WithHTTPStatus(neth.StatusConflict))
	ErrOrderUserNotFound   = ecode.New(20102, "下单用户不存在", ecode.WithHTTPStatus(neth.StatusBadRequest))
)
