package errorx

const (
	defaultCode       = 1001
	InvalidUrl        = 1002
	IsAlreadyShortUrl = 1003
	IsAlreadyConvert  = 1004
	PageNotFound      = 1005
)

type CodeError struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Datas any    `json:"data"`
}

type CodeErrorResponse struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Datas any    `json:"data"`
}

func NewCodeError(code int, msg string, data any) error {
	return &CodeError{Code: code, Msg: msg, Datas: data}
}

func NewDefaultError(msg string, data any) error {
	return NewCodeError(defaultCode, msg, data)
}

func (e *CodeError) Error() string {
	return e.Msg
}

func (e *CodeError) Data() *CodeErrorResponse {
	return &CodeErrorResponse{
		Code:  e.Code,
		Msg:   e.Msg,
		Datas: e.Datas,
	}
}
