package base

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pelletier/go-toml"
)

const (
	LinkModeTUN = "tun"
	LinkModeTAP = "tap"
)

var (
	Cfg = &ServerConfig{}
)

// # ReKey time (in seconds)
// rekey-time = 172800
// # ReKey method
// # Valid options: ssl, new-tunnel
// #  ssl: Will perform an efficient rehandshake on the channel allowing
// #       a seamless connection during rekey.
// #  new-tunnel: Will instruct the client to discard and re-establish the channel.
// #       Use this option only if the connecting clients have issues with the ssl
// #       option.
// rekey-method = ssl

type ServerConfig struct {
	LinkAddr      string `toml:"link_addr" info:"vpn服务对外地址"`
	ServerAddr    string `toml:"server_addr" info:"前台服务监听地址"`
	AdminAddr     string `toml:"admin_addr" info:"后台服务监听地址"`
	ProxyProtocol bool   `toml:"proxy_protocol" info:"TCP代理协议"`
	DbFile        string `toml:"db_file" info:"数据库地址"`
	CertFile      string `toml:"cert_file" info:"证书文件"`
	CertKey       string `toml:"cert_key" info:"证书密钥"`
	UiPath        string `toml:"ui_path" info:"ui文件路径"`
	FilesPath     string `toml:"files_path" info:"外部下载文件路径"`
	LogPath       string `toml:"log_path" info:"日志文件路径"`
	LogLevel      string `toml:"log_level" info:"日志等级"`
	Issuer        string `toml:"issuer" info:"系统名称"`
	AdminUser     string `toml:"admin_user" info:"管理用户名"`
	AdminPass     string `toml:"admin_pass" info:"管理用户密码"`
	JwtSecret     string `toml:"jwt_secret" info:"JWT密钥"`

	LinkMode    string   `toml:"link_mode" info:"虚拟网络类型"` // tun tap
	Ipv4CIDR    string   `toml:"ipv4_cidr" info:"ip地址网段"` // 192.168.1.0/24
	Ipv4Gateway string   `toml:"ipv4_gateway" info:"ipv4_gateway"`
	Ipv4Pool    []string `toml:"ipv4_pool" info:"IPV4起止地址池"` // Pool[0]=192.168.1.100 Pool[1]=192.168.1.200
	IpLease     int      `toml:"ip_lease"  info:"IP租期(秒)"`

	MaxClient       int    `toml:"max_client" info:"最大用户连接"`
	MaxUserClient   int    `toml:"max_user_client" info:"最大单用户连接"`
	DefaultGroup    string `toml:"default_group" info:"默认用户组"`
	CstpKeepalive   int    `toml:"cstp_keepalive" info:"keepalive时间(秒)"` // in seconds
	CstpDpd         int    `toml:"cstp_dpd" info:"死链接检测时间(秒)"`           // Dead peer detection in seconds
	MobileKeepalive int    `toml:"mobile_keepalive" info:"移动端keepalive接检测时间(秒)"`
	MobileDpd       int    `toml:"mobile_dpd" info:"移动端死链接检测时间(秒)"`

	SessionTimeout int `toml:"session_timeout" info:"session过期时间(秒)"` // in seconds
	AuthTimeout    int `toml:"auth_timeout" info:"auth_timeout"`      // in seconds
}

func initServerCfg() {
	b, err := ioutil.ReadFile(serverFile)
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(b, Cfg)
	if err != nil {
		panic(err)
	}

	sf, _ := filepath.Abs(serverFile)
	base := filepath.Dir(sf)

	// 转换成绝对路径
	Cfg.DbFile = getAbsPath(base, Cfg.DbFile)
	Cfg.CertFile = getAbsPath(base, Cfg.CertFile)
	Cfg.CertKey = getAbsPath(base, Cfg.CertKey)
	Cfg.UiPath = getAbsPath(base, Cfg.UiPath)
	Cfg.FilesPath = getAbsPath(base, Cfg.FilesPath)
	Cfg.LogPath = getAbsPath(base, Cfg.LogPath)

	if len(Cfg.JwtSecret) < 20 {
		fmt.Println("请设置 jwt_secret 长度20位以上")
		os.Exit(0)
	}

	fmt.Printf("ServerCfg: %+v \n", Cfg)
}

func getAbsPath(base, cfile string) string {
	if cfile == "" {
		return ""
	}

	abs := filepath.IsAbs(cfile)
	if abs {
		return cfile
	}
	return filepath.Join(base, cfile)
}

type SCfg struct {
	Name string      `json:"name"`
	Info string      `json:"info"`
	Data interface{} `json:"data"`
}

func ServerCfg2Slice() []SCfg {
	ref := reflect.ValueOf(Cfg)
	s := ref.Elem()

	var datas []SCfg

	typ := s.Type()
	numFields := s.NumField()
	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		value := s.Field(i)
		tag := field.Tag.Get("toml")
		tags := strings.Split(tag, ",")
		info := field.Tag.Get("info")

		datas = append(datas, SCfg{Name: tags[0], Info: info, Data: value.Interface()})
	}

	return datas
}
