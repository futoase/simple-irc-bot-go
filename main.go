package main

import (
	"encoding/json"
	irc "github.com/fluffle/goirc/client"
	"io/ioutil"
	"time"
)

type IRCSetting struct {
	NickName string
	Server   string
	PassWord string
	SSL      bool
	Channel  string
}

func main() {
	r, err := ioutil.ReadFile("setting.json")
	if err != nil {
		panic(err)
	}
	setting := IRCSetting{}
	if err := json.Unmarshal([]byte(r), &setting); err != nil {
		panic(err)
	}
	c := irc.SimpleClient(setting.NickName)
	c.SSL = setting.SSL
	c.AddHandler(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) { JoinChannel(conn, line, setting) })
	c.AddHandler("join",
		func(conn *irc.Conn, line *irc.Line) { WelcomeToUnderground(conn, line, setting) })

	quit := make(chan bool)
	c.AddHandler(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })
	if err := c.Connect(setting.Server, setting.PassWord); err != nil {
		panic(err)
	}

	<-quit
}

func JoinChannel(conn *irc.Conn, line *irc.Line, setting IRCSetting) {
	conn.Join(setting.Channel)
}

func WelcomeToUnderground(conn *irc.Conn, line *irc.Line, setting IRCSetting) {
	if setting.NickName != line.Nick {
		time.Sleep(3000 * time.Millisecond)
		conn.Notice(setting.Channel, line.Nick+", Welcome to underground...")
	}
}
