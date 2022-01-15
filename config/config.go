package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

var Config = &Configures{}

type Configures struct {
	Cover         bool              `yaml:"cover"` // 当文件存在时是否覆盖
	UseLog        bool              `yaml:"useLog"`
	MaxThread     int               `yaml:"maxThread"`     // 同时录制的视频线程
	Timeout       int               `yaml:"timeout"`       // request timeout second
	FfmpegTimeout int               `yaml:"ffmpegTimeout"` // ffmpeg timeout second
	FetchTime     int               `yaml:"fetchTime"`     // spider interval time
	OutPath       string            `yaml:"outPath"`       // save file path
	SourceFileUrl string            `yaml:"sourceFileUrl"` // 资源链接前缀
	Follow        map[int]string    `yaml:"follow"`        // 关注的成员
	Headers       map[string]string `yaml:"Headers"`       // net header
	ApiUrls       map[string]string `yaml:"ApiUrls"`
}

func Read(file string, reload bool) {
	_, err := load(file)
	if err != nil {
		panic("load config file err")
	}
	var rl func()
	var modtime time.Time
	rl = func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
				rl()
			}
		}()
		for {
			time.Sleep(time.Duration(10) * time.Second)
			if stat, err := os.Stat(file); err != nil {
				log.Println(err)
				continue
			} else {
				if stat.ModTime() == modtime {
					continue
				} else {
					modtime = stat.ModTime()
					load(file)
				}
			}
		}
	}
	if reload {
		stat, _ := os.Stat(file)
		modtime = stat.ModTime()
		go rl()
	}
}

func load(file string) (*Configures, error) {
	var C = new(Configures)
	f, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(f, C)
	if err != nil {
		panic(err)
		// return nil, err
	}
	yaml.Unmarshal(f, Config)
	return Config, nil
}
