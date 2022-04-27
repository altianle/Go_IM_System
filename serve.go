package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	//读写锁
	mapLock sync.RWMutex

	//消息广播的channel
	Message chan string
}

//创建一个serve接口
func NewServe(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

//监听Message广播消息channel的goroutine，一旦有消息就发送给全部在线的User
func (this *Server) ListenMassager() {
	for {
		//监听
		msg := <-this.Message

		//将msg发送给全部在线的User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息
func (this *Server) BroadCast(user *User, msg string) {
	sendmsg := "[" + user.Addr + "]" + user.Name + ": " + msg
	this.Message <- sendmsg
}

func (this *Server) Handler(con net.Conn) {
	//...当前链接的业务
	//fmt.Println("链接建立成功")
	user := NewUser(con, this)

	//用户上线，将用户加入到online中
	user.Online()

	//接收用户消息

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := con.Read(buf)
			if n == 0 { //当前用户关闭连接
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err: ", err)
				return
			}

			//提取用户消息，去除\n
			msg := string(buf[: n-1])

			//将得到的消息进行广播
			user.DoMessage(msg)
		}
	}()



	//当前handle阻塞，防止go程死亡
	select {}
}

//启动服务器的接口
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		println("net.Listen err: ", err)
		return
	}

	//启动监听massage
	go this.ListenMassager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			println("listener.Accept: ", err)
			continue
		}

		//do handle
		go this.Handler(conn)
	}

	//close listen socket
	defer listener.Close()
}
