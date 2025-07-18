package service

import (
	"database/sql"
	"errors"
	"fmt"
	"gochatroom/db"
	"log"
	"strings"
)

func Register(username, password string) (string, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		return "", errors.New("用户名和密码不能为空")
	}

	var id int
	err := db.DB.QueryRow("SELECT id FROM user WHERE username = ?", username).Scan(&id)
	if err != sql.ErrNoRows {
		return "", errors.New("用户名已存在或查询失败")
	}

	_, err = db.DB.Exec("INSERT INTO user (username, password) VALUES (?, ?)", username, password)
	if err != nil {
		return "", errors.New("注册失败：" + err.Error())
	}
	return "注册成功，欢迎你，" + username + "！", nil
}

func Login(username, password string) (string, error) {
	log.Println("开始查询数据库...") // 调试
	log.Println("查询完成")       // 调试
	var storedPwd string
	err := db.DB.QueryRow("SELECT password FROM user WHERE username = ?", username).Scan(&storedPwd)
	if err == sql.ErrNoRows {
		return "", errors.New("用户名不存在")
	}

	if err != nil {
		return "", err
	}
	if storedPwd != password {
		return "", errors.New("密码错误")
	}
	return "欢迎回来，" + username + "！", nil
}

func DeleteUser(username string) error {
	_, err := db.DB.Exec("DELETE FROM user WHERE username = ?", username)
	return err
}
func UpdatePassword(username, password string) error {
	_, err := db.DB.Exec("UPDATE user SET password = ? WHERE username = ?", password, username)
	return err
}
func UpdateUsername(oldUsername, newUsername string) error {
	// 检查 newUsername 是否已存在
	var exists int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM user WHERE username = ?", newUsername).Scan(&exists)
	if err != nil {
		return fmt.Errorf("查询数据库出错: %v", err)
	}
	if exists > 0 {
		return fmt.Errorf("用户名 %s 已被占用", newUsername)
	}

	// 开始更新用户名
	res, err := db.DB.Exec("UPDATE user SET username = ? WHERE username = ?", newUsername, oldUsername)
	if err != nil {
		return fmt.Errorf("更新用户名失败: %v", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("未找到原用户名 %s，更新失败", oldUsername)
	}

	return nil
}
