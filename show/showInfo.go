package show

import (
	"48go/LiveRoom"
	"fmt"
	"log"
	"os"
	"sort"
	"time"
)

var livingUser = map[int]LiveRoom.LiveInfo{}

var logger = log.New(os.Stdout, "", log.Ltime)

// Liver 正在直播
type Liver []LiveRoom.LiveInfo

func (L Liver) Less(i, j int) bool {
	return L[i].Ctime.Int() < L[j].Ctime.Int()
}
func (L Liver) Len() int {
	return len(L)
}
func (L Liver) Swap(i, j int) {
	L[i], L[j] = L[j], L[i]
}
func (L Liver) Sort() {
	sort.Sort(L)
}

func (L Liver) Equal(L1 Liver) bool {
	if L.Len() != L1.Len() {
		return false
	}
	for i := 0; i < L.Len(); i++ {
		if L[i].LiveId != L1[i].LiveId {
			return false
		}
	}
	return true
}

func (L Liver) Show() {
	fmt.Println("")
	logger.Printf("当前正在直播:%d\n", L.Len())
	for _, v := range L {
		v.Show()
	}
	fmt.Println("")
}

func New(buf int) chan interface{} {
	c := make(chan interface{}, buf)
	go show(c)
	go cleanLiving(24 * time.Hour)
	return c
}

func show(info chan interface{}) {
	for i := range info {
		switch v := i.(type) {
		case LiveRoom.Live:
			logger.Println(v.User.UserName, v.TypeString(), v.Ctime.Format(""),
				v.LiveId, "视频地址:", v.PlayStreamPath)
		case LiveRoom.LiveInfo:
			if _, ok := livingUser[v.LiveId.Int()]; ok {
				continue
			}
			v.Show()
			livingUser[v.LiveId.Int()] = v

		case LiveRoom.Lives:
			for _, v := range v.LiveList {
				if _, ok := livingUser[v.LiveId.Int()]; ok {
					continue
				}
				v.Show()
				livingUser[v.LiveId.Int()] = v
			}

		case string:
			logger.Println(v)
		}
	}
}

// 清理超时的数据
func cleanLiving(duration time.Duration) {
	go func(duration2 time.Duration) {
		for {
			time.Sleep(1 * time.Hour)
			now := time.Now()
			for k, v := range livingUser {
				vTime := time.UnixMilli(int64(v.Ctime.Int()))
				if now.Sub(vTime) > duration2 {
					delete(livingUser, k)
				}
			}
		}
	}(duration)
}
