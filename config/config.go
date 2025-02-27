package config

type Config struct {
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
	}
	JWT struct {
		Secret string
		TTL    int64
	}
	Email struct {
		SMTPHost     string
		SMTPPort     int
		SMTPUser     string
		SMTPPassword string
		FromEmail    string
	}
}

var AppConfig Config

func Init() {
	// TODO: 从环境变量或配置文件加载配置
	// 这里先使用硬编码的配置用于演示
	AppConfig.Database.Host = "localhost"
	AppConfig.Database.Port = 5432
	AppConfig.Database.User = "postgres"
	AppConfig.Database.DBName = "mdr_backend"
	AppConfig.Database.Password = "1234"
	AppConfig.JWT.Secret = "your-secret-key"
	AppConfig.JWT.TTL = 86400 // 24小时
}
