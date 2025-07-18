package server //修改密码和修改部分存在问题，记得要改，后面要给密码加入确认程序

import (
	"bufio"
	"fmt"
	"gochatroom/service"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Sender net.Conn
	Text   string
}

var (
	clients   = make(map[net.Conn]string) // 保存客户端连接与用户名
	userConns = make(map[string]net.Conn) // 用户名 → 连接

	mutex      = sync.Mutex{} // 用于并发安全
	broadcast  = make(chan Message)
	broadcast1 = make(chan Message)
)

// 启动 TCP 聊天服务器
func Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Println("服务器启动，监听:", addr)
	go handleBroadcast() //通过broadcast通道写入msg
	go handleBroadcast1()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("接收连接失败:", err)
			continue
		}
		go handleConnection(conn) //消息写入broadcast通道
	}
}
func handleBroadcast() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for conn := range clients { //通过这里实现消息共有
			if conn == msg.Sender {
				continue
			}
			fmt.Fprintln(conn, msg.Text) //将msg发送给conn
		}
		mutex.Unlock()
	}
}
func handleBroadcast1() {
	for {
		msg := <-broadcast1
		mutex.Lock()
		for conn := range clients { //通过这里实现消息共有
			if conn == msg.Sender {
				fmt.Fprintln(conn, msg.Text) //将msg发送给conn
			}
		}
		mutex.Unlock()
	}
}

// 处理客户端连接
func handleConnection(conn net.Conn) {
	defer func() {
		mutex.Lock()
		delete(clients, conn)
		delete(userConns, conn.RemoteAddr().String())
		mutex.Unlock()
		conn.Close()
	}()
	reader := bufio.NewReader(conn) //解析和存储

	writer := bufio.NewWriter(conn) //一个带缓冲写入器

	// 发送提示信息时立即刷新
	//通过tcp向客户端发送消息
	var username string
	var msg string
	var err error
	for {
		fmt.Fprintln(conn, "请输入用户名: ") // 自动加 \n
		writer.Flush()
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)

		fmt.Fprintln(conn, "请输入密码:")
		writer.Flush() //这样无需等待缓冲区满
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		fmt.Fprintln(conn, "登录或注册？输入 login 或 register: ")
		writer.Flush()
		action, _ := reader.ReadString('\n')
		action = strings.TrimSpace(strings.ToLower(action))
		if action == "register" {
			msg, err = service.Register(username, password)
			break
		} else if action == "login" {
			if _, exist := userConns[username]; exist {
				conn.Write([]byte("已在其它设备登录请勿重复登录\n"))
				continue
			} else {
				msg, err = service.Login(username, password)
			}
			break
		} else {
			msg = "操作失败，请在register和login中选择不要输入无效选项"
			conn.Write([]byte(msg + "\n"))
		}
	}

	if err != nil {
		conn.Write([]byte("错误：" + err.Error() + "\n"))
		return
	}
	mutex.Lock()
	clients[conn] = username
	userConns[username] = conn
	mutex.Unlock()
	defer delete(clients, conn)
	defer delete(userConns, username)
	conn.Write([]byte(msg + "当前在线人数" + strconv.Itoa(len(clients)) + "\n")) //播送消息
	//reader := bufio.NewReader(os.Stdin)
	broadcast <- Message{conn, fmt.Sprintf("用户%s加入聊天室,当前在线人数%d", username, len(clients))}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			broadcast <- Message{nil, fmt.Sprintf("用户%s离开聊天室,当前在线人数%d", username, len(clients)-1)}
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		} else if line == "*help" {
			fmt.Fprintln(conn, "-*help:查看命令帮助\n"+"*delete_me 注销账户（无法恢复）\n"+"-*change_password 修改密码\n"+"-*change_nickname 修改用户名\n"+"@某人（输入某人昵称） 进入私聊\n"+"-*exist 退出\n"+"-*sendfile 发送文件")
			continue
		} else if line == "*delete_me" {
			service.DeleteUser(username)
			fmt.Fprintln(conn, "-*delete_me 已执行，即将断开连接")
			broadcast <- Message{nil, fmt.Sprintf("用户%s离开聊天室,当前在线人数%d", username, len(clients)-1)}
			return
		} else if line == "*change_password" { //有问题部分
			var newPassword string
			fmt.Fprintln(conn, "请输入新密码:")
			writer.Flush()
			newPasswordLine, _ := reader.ReadString('\n')
			newPassword = strings.TrimSpace(newPasswordLine)
			if newPassword != "" {
				service.UpdatePassword(username, newPassword)
				fmt.Fprintln(conn, "-*change_password 操作成功")
			} else {
				fmt.Fprintln(conn, "-*change_password 操作失败请输入新的密码")
			}
			continue
		} else if line == "*change_nickname" {
			var newNickname string
			fmt.Fprintln(conn, "请输入新的用户名:")
			writer.Flush()
			newNicknameLine, _ := reader.ReadString('\n')
			newNickname = strings.TrimSpace(newNicknameLine)
			if newNickname != "" {
				err1 := service.UpdateUsername(username, newNickname)
				if err1 != nil {
					fmt.Fprintln(conn, "-*change_nickname 失败", err1.Error())
				} else {
					clients[conn] = newNickname
					delete(userConns, conn.RemoteAddr().String())
					username = newNickname
					userConns[newNickname] = conn
					fmt.Fprintln(conn, "-*change_nickname 操作成功")
				}
			} else {
				fmt.Fprintln(conn, "-*change_nickname 操作失败，请输入新的昵称")
			}
			continue
		} else if strings.HasPrefix(line, "@") {
			flag := 0
			for Conn := range clients {
				if clients[Conn] == strings.TrimSpace(strings.TrimPrefix(line, "@")) {
					flag = 1
					fmt.Fprintln(conn, "进入私聊模式，输入@返回公聊")
				}
			}
			if flag == 0 {
				fmt.Fprintln(conn, "@出错 没有该用户或用户未上线")
				continue
			}
			writer.Flush()
			for {
				Lmsg, _ := reader.ReadString('\n')
				receiver := strings.TrimSpace(strings.TrimPrefix(line, "@"))
				msg1 := strings.TrimSpace(Lmsg)
				if strings.HasPrefix(msg1, "@") {
					break
				}
				for Conn := range clients {
					if clients[Conn] == receiver {
						msg := Message{Conn, "私聊" + "[" + username + "]" + msg1}
						broadcast1 <- msg
					}
				}
			}
		} else if line == "*exist" {

			conn.Write([]byte("您已下线\n"))
			broadcast <- Message{nil, fmt.Sprintf("用户%s离开聊天室,当前在线人数%d", username, len(clients)-1)}
			break
		} else if strings.HasPrefix(line, "*upload|") {
			handleFileUpload(line, conn, reader, username)
		} else {
			msg := fmt.Sprintf("[%s]%s", username, line)
			broadcast <- Message{conn, msg}
		}
	}
}
func handleFileUpload(meta string, conn net.Conn, reader *bufio.Reader, username string) {
	meta = strings.TrimSpace(meta)
	if !strings.HasPrefix(meta, "*upload|") {
		conn.Write([]byte("❌ 上传格式错误\n"))
		return
	}

	parts := strings.Split(meta, "|")
	if len(parts) != 3 {
		conn.Write([]byte("文件信息解析失败\n"))
		return
	}
	time1 := int(time.Now().Unix())
	filename := strconv.Itoa(time1) + parts[1]
	filesize, _ := strconv.Atoi(parts[2])

	os.MkdirAll("server/uploads", os.ModePerm)
	savePath := filepath.Join("server/uploads", filename)
	file, err := os.Create(savePath)
	if err != nil {
		conn.Write([]byte("文件保存失败\n"))
		return
	}
	defer file.Close()

	buf := make([]byte, 1024)
	received := 0
	for received < filesize {
		n, err := reader.Read(buf)
		if err != nil {
			break
		}
		file.Write(buf[:n])
		received += n
	}
	conn.Write([]byte("文件上传完成\n"))
	broadcast <- Message{nil, fmt.Sprintf("用户%s已经完成文件上传", username)}
}
