package env

import (
	"fmt"
	"moredoc/pkg/storage/gsp"
	"moredoc/pkg/storage/sql"
	"moredoc/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/namsral/flag"
)

/*
参数定义:
|Type	|Flag			|Prefix Environment	|File			|
|-------|---------------|-------------------|---------------|
|int	|-age 2			|VIID_AGE=2			|age 2			|
|bool	|-female		|VIID_FEMALE=true	|female true	|
|float	|-length 175.5	|VIID_LENGTH=175.5	|length 175.5	|
|string	|-name Gloria	|VIID_NAME=Gloria	|name Gloria	|
*/

const (
	envPrefix = "VIID"
)

var (
	UseBasicStorage = []Option{UseMYSQL, UseGSP}
)

// DevopsEnv  用来放需要使用的devops的环境变量参数
type DevopsEnv struct {
	InfraPasswd string // 密码
}

// SystemInfo 定义系统基础信息
type SystemInfo struct {
	ProductName    string `json:"ProductName"`
	ProductVersion string `json:"ProductVersion"`
	NginxBind      string `json:"NginxBind"`
	BaseConfigPath string `json:"BaseConfigPath"`
	DevopsPassword string `json:"InfraPasswd"`
}

// ProcessInfo 定义进程基础信息
type ProcessInfo struct {
	BindAddr string `json:"BindAddr"`
	IsDebug  bool   `json:"IsDebug"`
}

// BaseConfig 定义配置参数最小集, 可被子服务配置对象继承的
type BaseConfig struct {
	*flag.FlagSet               // 参数解析必须的变量
	DevopsEnvSet  *flag.FlagSet // 参数解析必须的变量

	// 下面为必选参数, 所有进程共有（目前为3个）
	*ProcessInfo             // 必选参数, 进程基础信息
	DevopsEnv    *DevopsEnv  // 必选参数, Devops密码
	SystemInfo   *SystemInfo // 必选参数, 系统基础信息

	// 下面为可选参数（目前为7个）,按照使用频率排序
	Mysql *sql.Config // 可选参数
	GSP   *gsp.Config // 可选参数
}

// NewNewBaseConfig 实例基础配置对象, 用来给其他服务继承用.
// ETCD, MYSQL, ES, GSP, KAFKA, REDIS, KVDB需要根据情况通过opts参数进行选配.
func (cf *BaseConfig) NewBaseConfig(srvName string, opts ...Option) *BaseConfig {
	cf.FlagSet = flag.NewFlagSetWithEnvPrefix(srvName, envPrefix, flag.ContinueOnError)
	cf.String(flag.DefaultConfigFlagname, "", "path to config file") // 有文件解析文件
	// 进程基本信息
	cf.ProcessInfo = new(ProcessInfo)
	cf.StringVar(&cf.BindAddr, "bind", "127.0.0.1:19090", "Listen address like 'ip:port'")
	cf.BoolVar(&cf.IsDebug, "debug", false, "turn debug mode on or off(default)")
	// 系统基本信息
	cf.SystemInfo = new(SystemInfo)
	cf.StringVar(&cf.SystemInfo.ProductName, "productName", "视图级联网关", "系统名称")
	cf.StringVar(&cf.SystemInfo.ProductVersion, "productVersion", "V1.2", "系统版本")
	cf.StringVar(&cf.SystemInfo.NginxBind, "nginx", "127.0.0.1:80", "Nginx address like 'ip:port'")
	cf.StringVar(&cf.SystemInfo.BaseConfigPath, "baseConfigPath", types.BaseConfigPath, "基础配置文件路径")
	// 解析需要用到的Devops 的环境变量(所以不能带"VIID"前缀)
	cf.DevopsEnvSet = flag.NewFlagSet(srvName+"-Devops", flag.ContinueOnError)
	cf.DevopsEnv = new(DevopsEnv)
	// 需要定义两遍, 先解析一遍默认配置
	cf.StringVar(&cf.DevopsEnv.InfraPasswd, "devops_infra_password", "", "[Devops]infra_password")
	cf.DevopsEnvSet.StringVar(&cf.DevopsEnv.InfraPasswd, "devops_infra_password",
		cf.DevopsEnv.InfraPasswd, "[Devops]infra_password") // 解析环境变量
	// 配置其他可选参数
	for _, opt := range opts {
		opt(cf)
	}
	return cf
}

type Option func(*BaseConfig)

func UseMYSQL(cf *BaseConfig) {
	cf.Mysql = new(sql.Config)
	cf.StringVar(&cf.Mysql.Username, "sqluser", "root", "数据库用户名")
	cf.StringVar(&cf.Mysql.Password, "sqlpassword", "", "数据库密码")
	cf.StringVar(&cf.Mysql.Host, "sqlhost", "127.0.0.1", "数据库IP地址")
	cf.IntVar(&cf.Mysql.Port, "sqlport", 3306, "数据库端口")
	cf.StringVar(&cf.Mysql.DB, "dbname", "moredoc", "数据库名称")
	cf.StringVar(&cf.Mysql.Prefix, "prefix", "mnt_", "数据表前缀")
	cf.IntVar(&cf.Mysql.MaxIdle, "maxidle", 32, "MaxIdle")
	cf.IntVar(&cf.Mysql.MaxOpen, "maxopen", 16, "MaxOpen")
	cf.BoolVar(&cf.Mysql.ShowSQL, "showsql", true, "ShowSQL")
}

func UseGSP(cf *BaseConfig) {
	cf.GSP = new(gsp.Config)
	cf.StringVar(&cf.GSP.AccessKey, "accessKey", gsp.DefaultAK, "accessKey of the GSP")
	cf.StringVar(&cf.GSP.SecretKey, "secretKey", gsp.DefaultSK, "secretKey of the GSP")
	cf.StringVar(&cf.GSP.Addrs, "gspAddr", "", "address of gsp")
	cf.StringVar(&cf.GSP.Regions, "gspRegions", "", "gsp Regions")
	cf.BoolVar(&cf.GSP.Asyn, "gspAsyn", false, "PutObject Asynchronous")
}

// GetConfigPath 获取configs目录(带/后缀).
func (cf *BaseConfig) GetConfigPath() string {
	if cf != nil && cf.SystemInfo != nil {
		return cf.SystemInfo.BaseConfigPath
	}
	return "./configs/"
}

// Paraser 定义对自服务配置的解析
func (cf *BaseConfig) ModuleParaser(args []string) error {
	if err := cf.FlagSet.Parse(args); err != nil {
		return err
	}
	// 解析Devops的环境变量
	if err := cf.DevopsEnvSet.Parse([]string{}); err != nil {
		return err
	}
	if cf.DevopsEnv != nil && cf.DevopsEnv.InfraPasswd != "" {
		cf.SystemInfo.DevopsPassword = cf.DevopsEnv.InfraPasswd
		if cf.Mysql != nil && cf.Mysql.Password == "" {
			cf.Mysql.Password = cf.DevopsEnv.InfraPasswd
		}
	}
	if cf.IsDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	cf.PrintParasedConfig()
	return nil
}

// PrintConfig 打印配置参数
func (cf *BaseConfig) PrintParasedConfig() {
	pFunc := func(f *flag.Flag) {
		// MARK: 辅助打印, 打印配置参数
		fmt.Printf("===> \t %s: %s\tDefault: %s\tUsage: %s\n", f.Name, f.Value.String(), f.DefValue, f.Usage)
	}
	if cf.SystemInfo != nil {
		fmt.Printf("**************************************** %s %s ****************************************\n",
			cf.SystemInfo.ProductName, cf.SystemInfo.ProductVersion)
	} else {
		fmt.Printf("**************************************** VIID(云图/视图级联网关) ****************************************\n")
	}
	cf.FlagSet.VisitAll(pFunc)
	cf.DevopsEnvSet.VisitAll(pFunc)
}

// //////////////////////////////////

// Args contains apps console arguments
type Args []string

// Args returns the command line arguments associated
func NewArgs(params []string) Args {
	return Args(params)
}

// Get returns the nth argument, or else a blank string
func (a Args) Get(n int) string {
	if len(a) > n {
		return a[n]
	}
	return ""
}

// First returns the first argument, or else a blank string
func (a Args) First() string {
	return a.Get(0)
}

// Tail returns the rest of the arguments (not the first one)
// or else an empty string slice
func (a Args) Tail() []string {
	if len(a) >= 2 {
		return []string(a)[1:]
	}
	return []string{}
}

// Present checks if there are any arguments present
func (a Args) Present() bool {
	return len(a) != 0
}
