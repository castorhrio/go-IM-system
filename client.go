package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(ip string, port int) *Client {
	client := &Client{
		ServerIp:   ip,
		ServerPort: port,
		flag:       9999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))

	if err != nil {
		fmt.Println("net.Dial error", err)
		return nil
	}

	client.conn = conn

	return client
}

//接收服务器消息
func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn)
	// for {
	// 	buf := make()
	// 	client.conn.Read(buf)
	// 	fmt.Println(buf)
	// }
}

var serverIP string
var serverPort int

func init() {
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "设施服务器ip地址（默认127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设施服务器ip地址（默认8888）")
}

func (client *Client) menu() bool {
	fmt.Println("1.群聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更改用户名")
	fmt.Println("0.退出")

	var flag int
	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("请输入正确的操作编号")
		return false
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}

		switch client.flag {
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			fmt.Println("更改用户名...")
			break
		}
	}
}

//群聊模式
func (client *Client) PublicChat() {
	fmt.Println("请输入聊天内容，exit退出")
	var chatmsg string
	fmt.Scanln(&chatmsg)

	for chatmsg != "exit" {
		if len(chatmsg) != 0 {
			sendmsg := chatmsg + "\n"
			_, err := client.conn.Write([]byte(sendmsg))
			if err != nil {
				fmt.Println("conn write err:", err)
				break
			}
		}

		chatmsg = ""
		fmt.Println("请输入聊天内容，exit退出")
		fmt.Scanln(&chatmsg)
	}
}

func (client *Client) FindUsers() {
	sendmsg := "who\n"
	_, err := client.conn.Write([]byte(sendmsg))
	if err != nil {
		fmt.Println("")
		return
	}
}

//私聊模式
func (client *Client) PrivateChat() {
	var remoteUser string
	var chatMsg string

	client.FindUsers()
	fmt.Println("请选择输入聊天对象用户名，exit退出:")
	fmt.Scanln(&remoteUser)

	for remoteUser != "exit" {
		fmt.Println("请输入消息内容,exit退出")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteUser + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("chat msg send faild,", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println("请输入消息内容,exit退出")
			fmt.Scanln(&chatMsg)
		}

		client.FindUsers()
		fmt.Println("请选择输入聊天对象用户名，exit退出:")
		fmt.Scanln(&remoteUser)
	}
}

//更改用户名
func (client *Client) UpdateName() bool {
	fmt.Println("请输入用户名")
	fmt.Scanln(&client.Name)

	msg := "rename|" + client.Name
	_, err := client.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn.write err:", err)
		return false

	}

	return true
}

func main() {
	flag.Parse()

	client := NewClient(serverIP, serverPort)
	if client == nil {
		fmt.Println("连接服务器失败.....")
		return
	}

	go client.DealResponse()

	fmt.Println("连接服务器成功....")
	client.Run()
}
