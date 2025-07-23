package main

import (
	"fmt"
	"net/http"
)

// 创建处理器函数
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World", r.URL.Path)
}
func main() {
	http.HandleFunc("/", handler) //映射地址，处理请求
	http.ListenAndServe(":8080", nil)
}
