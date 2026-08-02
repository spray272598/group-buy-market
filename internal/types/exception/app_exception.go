package exception

import "group-buy-market/internal/types/enums"

// AppException 业务异常
type AppException struct {
	Code string
	Info string
}

func (e *AppException) Error() string {
	return e.Code + ": " + e.Info
}

func NewAppException(rc enums.ResponseCode) *AppException {
	return &AppException{Code: rc.Code, Info: rc.Info}
}

func NewAppExceptionMsg(code, info string) *AppException {
	return &AppException{Code: code, Info: info}
}

// AsAppException 尝试转换为业务异常
func AsAppException(err error) (*AppException, bool) {
	if err == nil {
		return nil, false
	}
	if e, ok := err.(*AppException); ok {
		return e, true
	}
	return nil, false
}
