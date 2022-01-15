package main

import (
	"48go/LiveRoom"
	"48go/config"
	"48go/request"
	"48go/show"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

var configFile = "/48LiveGo/config.yaml"

// var configFile = "config.yaml"

var Conf = config.Config
var logger *log.Logger
var pool *LiveRoom.Pool
var c = show.New(10)

func init() {
	config.Read(configFile, true)
	request.Headers = Conf.Headers
	request.Timeout = time.Second * time.Duration(Conf.Timeout)
	pool = LiveRoom.New(config.Config.MaxThread)
	go pool.ShowDownloading(c)
	logger = log.New(os.Stdout, "", log.LstdFlags)
}

func main() {

	// l, _ := LiveRoom.GetLive(config.Config.ApiUrls["LIVE_ONE_URL"], "397539447385427968")
	// l.DefualtDown(Conf.OutPath, os.Stdout)
	fmt.Println("start")
	Follow()
	fmt.Println("end")
}

func base(Next int, Record bool, Groupid int) io.Reader {
	type baser struct {
		Debug   bool `json:"debug"`
		Next    int  `json:"next"`
		Record  bool `json:"record"`
		Groupid int  `json:"groupid"`
	}
	var base = baser{
		Debug:   true,
		Next:    Next,
		Record:  Record,
		Groupid: Groupid,
	}
	data, err := json.Marshal(base)
	if err != nil {
		return nil
	}
	return bytes.NewBuffer(data)
}

func Follow() {
	var Next string
	var Liver1 show.Liver
	var Liver2 show.Liver
	retry := 0
	for {
		var getLives = func() error {
			nextInt, _ := strconv.Atoi(Next)
			lives, err := LiveRoom.GetLives(Conf.ApiUrls["LIVE_LIST_URL"], base(nextInt, false, 0))
			if err != nil {
				logger.Println(err)
				if retry < 6 {
					retry++
				}
				time.Sleep((1 << retry) * time.Second)
				return err
				// return err
			} else {
				retry = 0
			}
			Next = lives.Next
			Liver1 = append(Liver1, lives.LiveList...)
			return err
		}
		for err := getLives(); Next != "0" || err != nil; {
			err = getLives()
		}

		for _, liveInfo := range Liver1 {
			if _, ok := Conf.Follow[liveInfo.UserInfo.UserId.Int()]; !ok {
				continue
			}
			procesLive(liveInfo)
		}
		// 比较两个直播切片，不相等则打印内容
		Liver1.Sort()
		if !Liver1.Equal(Liver2) {
			Liver1.Show()
		}
		Liver2 = Liver1
		Liver1 = Liver1[0:0]

		time.Sleep(time.Duration(Conf.FetchTime) * time.Second)
	}
	//pool.Wait()
}

func procesLive(liveInfo LiveRoom.LiveInfo) {
	if live, err := liveInfo.GetLive(); err == nil {
		err := pool.Add(live)
		if _, ok := err.(LiveRoom.Exist); ok {
			return
		} else if _, ok := err.(LiveRoom.Pull); ok {
			return
		}
		go func(live *LiveRoom.Live) {
			c <- fmt.Sprintf("下载: %s %s %s 视频地址: %s", live.User.UserName, live.TypeString(),
				live.Ctime.Format(""), live.PlayStreamPath)
			live.Down(Conf.OutPath, nil, LiveRoom.DownCoverOption, LiveRoom.DownCarouselsOption,
				LiveRoom.DownMsgOption)
			c <- fmt.Sprintf("%s 下载结束\n", live.User.UserName)
			time.Sleep(time.Duration(2) * time.Second)
			pool.Done(live.LiveId.Int())
		}(live)
	} else {
		logger.Println(err)
	}

}
