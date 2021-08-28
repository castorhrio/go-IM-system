package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	chanl  chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:  userAddr,
		Addr:  userAddr,
		chanl: make(chan string),
		conn:  conn,

		server: server,
	}

	go user.ListenMessage()
	return user
}

//监听当前用户
func (user *User) ListenMessage() {
	for {
		msg := <-user.chanl

		user.conn.Write([]byte(msg + "\n"))
	}
}

//用户上线
func (user *User) Online() {
	//将用户加入到onlineMap中
	user.server.mapLock.Lock()
	user.server.OnLineMap[user.Name] = user
	user.server.mapLock.Unlock()

	//广播用户上线消息
	user.server.BroadCast(user, "已上线")
}

//用户下线
func (user *User) Offline() {
	//将用户加入到onlineMap中
	user.server.mapLock.Lock()
	delete(user.server.OnLineMap, user.Name)
	user.server.mapLock.Unlock()

	//广播用户上线消息
	user.server.BroadCast(user, "下线")
}

//发送消息
func (user *User) DoMsg(msg string) {
	user.conn.Write([]byte(msg))
}

//消息业务
func (user *User) SendMsg(msg string) {
	if msg == "who" {
		//查询当前在线用户
		user.server.mapLock.Lock()
		for _, u := range user.server.OnLineMap {
			onlinemsg := u.Name + ":" + "在线...\n"
			user.DoMsg(onlinemsg)
		}
		user.server.mapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		new_name := strings.Split(msg, "|")[1]

		_, ok := user.server.OnLineMap[new_name]
		if ok {
			user.DoMsg("用户名已存在\n")
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnLineMap, user.Name)
			user.server.OnLineMap[new_name] = user
			user.server.mapLock.Unlock()

			user.Name = new_name
			user.DoMsg("用户名已更新:" + user.Name + "\n")
		}
	} else if len(msg) > 3 && msg[:3] == "to|" {
		remote_name := strings.Split(msg, "|")[1]
		if remote_name == "" {
			user.DoMsg("消息格式不正确，请使用\"to|用户名|消息内容\"格式\n")
			return
		}

		remote_user, ok := user.server.OnLineMap[remote_name]
		if !ok {
			user.DoMsg("用户不存在")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.DoMsg("消息内容不能为空")
			return
		}

		remote_user.DoMsg("来自" + user.Name + "的消息：" + content)

	} else {
		user.server.BroadCast(user, msg)
	}
}
