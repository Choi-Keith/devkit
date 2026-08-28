package orderhttp

import (
	neth "net/http"

	"github.com/Choi-Keith/golang-common/ecode"
)

// Business error codes of the order module (20100+). Codes are
// process-global and registered exactly once here, at the transport
// boundary.
var (
	ErrOrderNotFound       = ecode.New(20100, "订单不存在", ecode.WithHTTPStatus(neth.StatusNotFound))
	ErrOrderNotCancellable = ecode.New(20101, "订单状态不允许取消", ecode.WithHTTPStatus(neth.StatusConflict))
)
