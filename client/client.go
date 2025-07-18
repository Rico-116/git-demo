package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 客户端主函数
func main() {
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		fmt.Println("连接服务器失败:", err)
		return
	}
	defer conn.Close()

	fmt.Println("✅ 成功连接到服务器")
	// 先启动消息接收协程
	go receiveMessages(conn)

	// 主循环：非阻塞读取输入（需配合终端库如 termbox）
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "*sendfile" {
			handleUpload(conn)
		} else {
			conn.Write([]byte(text + "\n"))
		}
	}
}

// 接收来自服务端的广播消息
func receiveMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("❌ 与服务器断开连接。")
			os.Exit(0)
		}
		fmt.Print(msg)
	}
}
func handleUpload(conn net.Conn) {
	fmt.Print("输入需要上传的文件的绝对路径 ")
	reader := bufio.NewReader(os.Stdin)
	localPath, _ := reader.ReadString('\n')
	localPath = strings.TrimSpace(localPath)

	file, err := os.Open(localPath)
	if err != nil {
		fmt.Println("无法打开文件:", err)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		fmt.Println("获取文件信息失败:", err)
		return
	}

	filename := filepath.Base(localPath)
	filesize := int(info.Size())

	// 发送上传请求头
	conn.Write([]byte("*upload|" + filename + "|" + strconv.Itoa(filesize) + "\n"))
	// 发送文件数据
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err != nil {
			break
		}
		conn.Write(buf[:n])
	}
	fmt.Println("文件上传完成")
}
