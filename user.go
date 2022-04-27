package main

import "net"

type User struct {
	Name   string
	Addr   string
	C      chan string
	Conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		Conn:   conn,
		server: server,
	}

	//启动监听当前user cahnnel的go程
	go user.ListenMessage()

	return user
}

func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.BroadCast(this, "online")
}

func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "offline")
}

func (this *User) DoMessage(msg string) {
	this.server.BroadCast(this, msg)
}

//监听当前user channel的方法，一旦有消息，就直接发给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		//写给客户端
		this.Conn.Write([]byte(msg + "\n"))
	}

}
