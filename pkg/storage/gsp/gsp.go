package gsp

const (
	DefaultAK     = "LeBnhmXCSrYemEjv"
	DefaultSK     = "CGHdmxXRjIvjaUsZ"
	DefaultRegion = "defaultRegion"
	s3Error       = "s3 is not ok"
)

type Config struct {
	AccessKey string `json:"accessKey"` // 账号
	SecretKey string `json:"secretKey"` // 密钥
	Addrs     string `json:"address"`   // GSP S3服务的地址,多个情况下分号分割
	Regions   string `json:"regions"`   // 作用域
	Asyn      bool   `json:"asyn"`      // 是否异步存图
}
