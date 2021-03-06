package main

import (
	"encoding/json"
	"github.com/ActiveState/tail"
	irc "github.com/fluffle/goirc/client"
	"github.com/howeyc/fsnotify"
	"io/ioutil"
	"regexp"
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

	go MonitorFiles(c, setting, "monitored-file-path.json")
	go WatchFile(c, setting, "test.txt")

	quit := make(chan bool)
	c.AddHandler(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })
	if err := c.Connect(setting.Server, setting.PassWord); err != nil {
		panic(err)
	}

	<-quit
}

func MonitorFiles(conn *irc.Conn, setting IRCSetting, monitorFilePath string) {
	r, err := ioutil.ReadFile(monitorFilePath)
	if err != nil {
		panic(err)
	}

	var monitorFiles []string
	err = json.Unmarshal([]byte(r), &monitorFiles)
	if err != nil {
		panic(err)
	}

	for _, v := range monitorFiles {
		go TailFile(conn, setting, v)
	}
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

func TailFile(conn *irc.Conn, setting IRCSetting, filePath string) {

	// Matching of 2013-10-25 10:00:00 +0900 [info]: hogehoge.
	r, err := regexp.Compile(`^\d{4}-\d\d-\d\d \d\d:\d\d:\d\d \+\d{4} \[info\]:(.*)$`)
	if err != nil {
		panic(err)
	}

	t, err := tail.TailFile(filePath, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		panic(err)
	}
	for line := range t.Lines {
		matches := r.FindStringSubmatch(line.Text)
		if len(matches) == 2 {
			conn.Notice(setting.Channel, filePath+": "+matches[1])
		}
	}
}

func WatchFile(conn *irc.Conn, setting IRCSetting, filePath string) {
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
					conn.Notice(setting.Channel, "Update File!: "+filePath)
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
