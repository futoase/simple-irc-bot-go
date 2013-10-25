package main

import (
	"encoding/json"
	irc "github.com/fluffle/goirc/client"
	"github.com/howeyc/fsnotify"
	"io/ioutil"
	"time"
)

type IRCSetting struct {
	NickName string
	RealName string
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
	c := irc.SimpleClient(setting.NickName, "", setting.RealName)
	c.SSL = setting.SSL
	c.AddHandler(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) { JoinChannel(conn, line, setting) })
	c.AddHandler("join",
		func(conn *irc.Conn, line *irc.Line) { WelcomeToUnderground(conn, line, setting) })

	go FileWatch(c, setting, "test.txt")

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

func FileWatch(conn *irc.Conn, setting IRCSetting, filePath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if ev.IsModify() {
					conn.Notice(setting.Channel, "File update!: "+filePath)
				}
			}
		}
	}()

	err = watcher.Watch(filePath)
	if err != nil {
		panic(err)
	}

	<-done

	watcher.Close()
}
