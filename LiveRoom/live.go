package LiveRoom

import (
	"48go/config"
	"48go/request"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

type livesResp struct {
	Status  int    `json:"status"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Content Lives  `json:"content"`
}

type Lives struct {
	Next     string     `json:"next"`
	LiveList []LiveInfo `json:"liveList"`
}
type LiveInfo struct {
	LiveId                 LiveId `json:"liveId"`
	CoverPath              string `json:"coverPath"`
	Title                  string `json:"title"`
	LiveType               int    `json:"liveType"`
	Status                 int    `json:"status"`
	Ctime                  `json:"ctime"`
	UserInfo               UserInfo `json:"userInfo"`
	LiveMode               int      `json:"liveMode"`
	PictureOrientation     int      `json:"pictureOrientation"`
	IsCollection           int      `json:"isCollection"`
	InMicrophoneConnection bool     `json:"inMicrophoneConnection"`
	CoverWidth             int      `json:"coverWidth"`
	CoverHeight            int      `json:"coverHeight"`
}

type UserId string
type Ctime string
type LiveId string

type UserInfo struct {
	UserId     UserId        `json:"userId"`
	Nickname   string        `json:"nickname"`
	Avatar     string        `json:"avatar"`
	Badge      []interface{} `json:"badge"`
	Level      int           `json:"level"`
	IsStar     bool          `json:"isStar"`
	Friends    string        `json:"friends"`
	Followers  string        `json:"followers"`
	TeamLogo   string        `json:"teamLogo"`
	Signature  string        `json:"signature"`
	BgImg      string        `json:"bgImg"`
	Vip        bool          `json:"vip"`
	UserRole   int           `json:"userRole"`
	PfUrl      string        `json:"pfUrl"`
	EffectUser bool          `json:"effectUser"`
}

type liveResp struct {
	Status  int    `json:"status"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Content Live   `json:"content"`
}

type Live struct {
	LiveId             LiveId        `json:"liveId"`
	RoomId             string        `json:"roomId"`
	OnlineNum          int           `json:"onlineNum"`
	Type               int           `json:"type"`
	LiveType           int           `json:"liveType"`
	Review             bool          `json:"review"`
	NeedForward        bool          `json:"needForward"`
	SystemMsg          string        `json:"systemMsg"`
	MsgFilePath        string        `json:"msgFilePath"`
	PlayStreamPath     string        `json:"playStreamPath"`
	User               User          `json:"user"`
	Carousels          Carousels     `json:"carousels"`
	TopUser            []interface{} `json:"topUser"`
	Mute               bool          `json:"mute"`
	LiveMode           int           `json:"liveMode"`
	PictureOrientation int           `json:"pictureOrientation"`
	IsCollection       int           `json:"isCollection"`
	MergeStreamUrl     string        `json:"mergeStreamUrl"`
	Crm                string        `json:"crm"`
	CoverPath          string        `json:"coverPath"`
	Title              string        `json:"title"`
	Ctime              `json:"ctime"`
	Announcement       string        `json:"announcement"`
	SpecialBadge       []interface{} `json:"specialBadge"`
}
type User struct {
	UserId     UserId `json:"userId"`
	UserName   string `json:"userName"`
	UserAvatar string `json:"userAvatar"`
	Level      int    `json:"level"`
}
type Carousels struct {
	CarouselTime string   `json:"carouselTime"`
	Carousels    []string `json:"carousels"`
}

func (c Ctime) Int() int {
	i, err := strconv.Atoi(string(c))
	if err != nil {
		return 0
	}
	return i
}

func (LiveId LiveId) Int() int {
	i, err := strconv.Atoi(string(LiveId))
	if err != nil {
		return 0
	}
	return i
}

func (u UserId) Int() int {
	i, err := strconv.Atoi(string(u))
	if err != nil {
		return 0
	}
	return i
}

// Format example:2006-01-02 15:04:05
func (ts Ctime) Format(format string) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	t, err := strconv.Atoi(string(ts))
	if err != nil {
		return ""
	}
	return time.UnixMilli(int64(t)).Format(format)
}

// TypeString return LiveType string
func (l *Live) TypeString() string {
	return typeString(l.LiveType, l.LiveMode)
}

func (l *Live) videoName() string {
	nameBuf := strings.Builder{}
	nameBuf.WriteString(l.User.UserName)
	nameBuf.WriteString(" ")
	nameBuf.WriteString(l.Ctime.Format("2006-01-02 15.04.5"))
	nameBuf.WriteString(" ")
	nameBuf.WriteString(macFileNameReplace(l.Title))
	nameBuf.WriteString(l.TypeString())
	nameBuf.WriteString(".mp4")
	return nameBuf.String()
}

// TypeString return LiveType string
func (LiveInfo *LiveInfo) TypeString() string {
	return typeString(LiveInfo.LiveType, LiveInfo.LiveMode)
}

func typeString(liveType, liveMode int) string {
	if liveType == 1 && liveMode == 0 {
		return "直播"
	} else if liveType == 1 && liveMode == 1 {
		return "录屏"
	} else if liveType == 2 {
		return "电台"
	} else {
		return "未知"
	}
}

// GetLives body example: {"debug":true, "next":0, "record":true, "groupid":0}
func GetLives(url string, body io.Reader) (*Lives, error) {
	var livesR = livesResp{}

	resp, err := request.POST(url, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(respBody, &livesR)
	if err != nil {
		return nil, err
	}
	var lives = new(Lives)
	lives.Next = livesR.Content.Next
	for _, v := range livesR.Content.LiveList {
		lives.LiveList = append(lives.LiveList, v.Format())
	}
	return lives, nil
}

func (LiveInfo *LiveInfo) GetLive() (*Live, error) {
	return GetLive(config.Config.ApiUrls["LIVE_ONE_URL"], string(LiveInfo.LiveId))
}

func (LiveInfo *LiveInfo) ShowString() string {
	liveType := ""
	if LiveInfo.TypeString() == "电台" {
		liveType = color.RedString(LiveInfo.TypeString())
	} else {
		liveType = color.GreenString(LiveInfo.TypeString())
	}
	return fmt.Sprintf("%s %10s: %s %s %s", LiveInfo.Ctime.Format("15:04:05"),
		color.RedString(string(fmt.Sprintf("%10s", LiveInfo.UserInfo.UserId))),
		color.MagentaString(LiveInfo.UserInfo.Nickname), color.CyanString(string(LiveInfo.LiveId)), liveType)
}
func (LiveInfo *LiveInfo) Show() {
	fmt.Println(LiveInfo.ShowString())
}

func GetLive(url, liveId string) (*Live, error) {
	var liveR = liveResp{}
	type baser struct {
		LiveId string `json:"liveId"`
	}
	var base = baser{
		LiveId: liveId,
	}
	data, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}
	resp, err := request.POST(url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &liveR)
	if err != nil {
		return nil, err
	}
	d := liveR.Content.Format()
	return &d, nil
	//return &(liveR.Content.Format()), nil
}

type DownOption func(string, *Live) error

// Down download file
func (l *Live) Down(fp string, w io.Writer, options ...DownOption) {
	var verr error
	fp = filepath.Join(fp, l.User.UserName, l.Ctime.Format("2006-01-02 15.04.05"))
	os.MkdirAll(fp, fs.FileMode(0766))
	if verr = downVideo(filepath.Join(fp, l.videoName()), w, config.Config.Cover, l); verr != nil && w != nil {
		fmt.Fprintln(w, verr)
	}
	for _, option := range options {
		if err := option(fp, l); err != nil && w != nil {
			fmt.Fprintln(w, err)
		}
	}
	if verr == nil {
		audioName := filepath.Join(fp, l.videoName())
		audioName = string(audioName[:len(audioName)-3]) + "m4a"
		mp4ConterM4a(filepath.Join(fp, l.videoName()), audioName, filepath.Join(fp, "封面.jpg"), l.User.UserName, w)
	}
}

func (l *Live) DefualtDown(fp string, w io.Writer) {
	l.Down(fp, w, DownCoverOption, DownCarouselsOption, DownMsgOption)
}

func macFileNameReplace(s string) string {
	cmp := regexp.MustCompile("[:/']")
	return cmp.ReplaceAllString(s, "")
}

// Down video
// shell = "ffmpeg -rw_timeout 10000000 -i '{}' -c copy '{}.{}'"
func downVideo(out string, w io.Writer, cover bool, l *Live) error {

	timeout := strconv.Itoa(config.Config.FfmpegTimeout * 1000000)

	shell := []string{"-rw_timeout", timeout, "-i", l.PlayStreamPath, "-c",
		"copy"}
	if cover {
		shell = append(shell, "-y")
	} else {
		shell = append(shell, "-n")
	}
	shell = append(shell, "-f", "mp4", out)
	cmd := exec.Command("ffmpeg", shell...)
	cmd.Stderr = w
	cmd.Stdout = w
	return cmd.Run()
}

// f"ffmpeg -i '{name}.{self.videoFormat}' -i '{img}' -map 0:a -map 1:v -c copy -disposition:v:0 attached_pic -metadata artist='{artist}' -metadata album='口袋48' '{name}.m4a'"
func mp4ConterM4a(in, out, img, artist string, w io.Writer) {
	shell := []string{"-i", in, "-i", img, "-map", "0:a", "-map", "1:v", "-c", "copy", "-y", "-disposition:v:0",
		"attached_pic", "-metadata", fmt.Sprintf("artist=%s", artist), "-metadata", "album='口袋48'", out}
	cmd := exec.Command("ffmpeg", shell...)
	cmd.Stderr = w
	cmd.Stdout = w
	cmd.Run()
}

// DownMsgOption 下载弹幕文件
func DownMsgOption(fp string, l *Live) error {
	if l.MsgFilePath == "" {
		return nil
	}
	resp, err := request.GET(l.MsgFilePath, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	name := strings.Split(l.MsgFilePath, "/")
	buf := bufio.NewReader(resp.Body)
	return save(filepath.Join(fp, name[len(name)-1]), buf)
}

type IsNotRadio error

// DownCarouselsOption 下载电台的背景图片
func DownCarouselsOption(fp string, l *Live) error {
	if l.TypeString() != "电台" {
		return IsNotRadio(fmt.Errorf("%s %s %s %s no Background picture",
			l.User.UserName, l.User.UserId, l.LiveId, l.TypeString()))
	}
	for _, url := range l.Carousels.Carousels {
		resp, err := request.GET(url, nil)
		if err != nil {
			continue
		}
		name := strings.Split(url, "/")
		buf := bufio.NewReader(resp.Body)
		err = save(filepath.Join(fp, name[len(name)-1]), buf)
		resp.Body.Close()
		if err != nil {
			continue
		}
	}
	return nil
}

// DownCoverOption 直播封面图片
func DownCoverOption(fp string, l *Live) error {
	resp, err := request.GET(l.CoverPath, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf := bufio.NewReader(resp.Body)
	return save(filepath.Join(fp, "封面.jpg"), buf)
}

func save(file string, buf *bufio.Reader) error {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR, fs.FileMode(0666))
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(f)
	return err
}

func (LiveInfo LiveInfo) Format() LiveInfo {
	prefix := config.Config.SourceFileUrl
	LiveInfo.CoverPath = prefix + LiveInfo.CoverPath
	LiveInfo.UserInfo.Avatar = prefix + LiveInfo.UserInfo.Avatar
	LiveInfo.UserInfo.TeamLogo = prefix + LiveInfo.UserInfo.TeamLogo
	LiveInfo.UserInfo.BgImg = prefix + LiveInfo.UserInfo.BgImg
	LiveInfo.UserInfo.PfUrl = prefix + LiveInfo.UserInfo.PfUrl
	return LiveInfo
}

func (l Live) Format() Live {
	prefix := config.Config.SourceFileUrl
	l.User.UserAvatar = prefix + l.User.UserAvatar
	l.CoverPath = prefix + l.CoverPath
	for i, carousel := range l.Carousels.Carousels {
		l.Carousels.Carousels[i] = prefix + carousel
	}

	return l
}
