// Package response 提供统一的 API 响应格式与业务错误码。
package response

import (
	"encoding/json"
	"net/http"
)

// 业务错误码：0 表示成功，其余按类别分段。
const (
	CodeOK              = 0
	CodeInvalidParam    = 40001 // 参数错误
	CodeNotFound        = 40401 // 资源不存在
	CodeMethodNotAllow  = 40501 // 方法不允许
	CodeInternalError   = 50001 // 系统错误
)

// Body 统一响应体。
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// WriteJSON 将统一响应体写入响应。
func WriteJSON(w http.ResponseWriter, status int, body Body) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// OK 返回成功响应。
func OK(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusOK, Body{Code: CodeOK, Message: "success", Data: data})
}

// Fail 返回业务错误响应（不暴露内部错误细节）。
func Fail(w http.ResponseWriter, status, code int, message string) {
	WriteJSON(w, status, Body{Code: code, Message: message, Data: nil})
}

// InvalidParam 参数错误。
func InvalidParam(w http.ResponseWriter, message string) {
	if message == "" {
		message = "参数错误"
	}
	Fail(w, http.StatusBadRequest, CodeInvalidParam, message)
}

// NotFound 资源不存在。
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "资源不存在"
	}
	Fail(w, http.StatusNotFound, CodeNotFound, message)
}

// Internal 系统错误（客户端只看到通用提示）。
func Internal(w http.ResponseWriter) {
	Fail(w, http.StatusInternalServerError, CodeInternalError, "系统繁忙，请稍后重试")
}
