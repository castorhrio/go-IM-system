package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP   string
	Port int

	//在线用户列表
	OnLineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnLineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

//广播消息
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	server.Message <- sendMsg
}

func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message

		server.mapLock.Lock()
		for _, cli := range server.OnLineMap {
			cli.chanl <- msg
		}

		server.mapLock.Unlock()
	}
}

func (server *Server) Handle(conn net.Conn) {
	user := NewUser(conn, server)
	user.Online()

	//监听用户是否活跃
	isLive := make(chan bool)

	//接收客户端消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("conn read err:", err)
				return
			}

			//提取用户消息去除'\n'
			msg := string(buf[:n-1])
			//将消息进行广播
			user.SendMsg(msg)

			isLive <- true
		}
	}()

	for {
		select {
		case <-isLive:
		//不做任何操作更新定时器
		case <-time.After(time.Second * 300):
			user.DoMsg("你已下线")
			close(user.chanl)
			conn.Close()
			return
		}
	}
}

func (server *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port))
	if err != nil {
		fmt.Println("net listen error:", err)
		return
	}

	defer listener.Close()

	//监听message
	go server.ListenMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listen accept error:", err)
			continue
		}

		go server.Handle(conn)
	}
}
