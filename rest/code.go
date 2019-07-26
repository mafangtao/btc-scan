package rest

const (
	OK                  = 0     //成功
	BadRequest          = 40002 // 参数不合法，请检查参数
	InternalServerError = 50000
)

var statusText = map[int]string{
	OK:                  "Success",
	BadRequest:          "参数不合法，请检查参数",
	InternalServerError: "服务器内部错误",
}

func StatusText(code int) string {
	return statusText[code]
}

func HttpStatusCode(code int) int {
	return code / 100
}
