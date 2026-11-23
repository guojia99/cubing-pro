package configs

type GlobalConfig struct {
	Debug           bool        `yaml:"debug"`
	Dev             bool        `yaml:"dev"`
	BaseHost        string      `yaml:"baseHost"`
	ImageTempPath   string      `yaml:"imageTempPath"`
	BaseFontTTf     string      `yaml:"baseFontTTf"`
	DB              DBConfig    `yaml:"db"`
	EmailConfig     EmailConfig `yaml:"emailConfig"`
	Scramble        Scramble    `yaml:"scramble"`
	WcaDB           WcaDB       `yaml:"wcaDB"`
	AlgTrainersPath string      `yaml:"algTrainersPath"`
	WcaAuth2        WcaAuth2    `yaml:"wcaAuth2"`
}

type WcaAuth2 struct {
	AppID        string   `yaml:"appId"`
	AppSecret    string   `yaml:"appSecret"`
	RedirectBase string   `yaml:"redirectBase"` // OAuth 回调根地址，如 https://cubing.pro 或 http://localhost:8000
	FrontendBase string   `yaml:"frontendBase"` // 登录成功后跳回的前端域名，如 https://cubing.pro
	RedirectURLs []string `yaml:"redirectURLs"`
	AuthsPath    []string `yaml:"authsPath"`
	Auths        []string `yaml:"auths"` // scopes: public, dob, email, openid, profile
}

type WcaDB struct {
	MysqlUrl string `yaml:"mysqlUrl"`
	DbPath   string `yaml:"dbPath"`
	SyncPath string `yaml:"syncPath"`
}

type Scramble struct {
	Type     string `yaml:"type"` // tnoodle, rust_twisty；空值默认 tnoodle
	EndPoint string `yaml:"endpoint"`

	ScrambleDrawType string `yaml:"scrambleDrawType"` // 2mf8
	ScrambleUrl      string `yaml:"scramble"`
}

type DBConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type EmailConfig struct {
	SmtpHost string `yaml:"smtpHost"`
	SmtpPort int    `yaml:"smtpPort"`
	From     string `yaml:"from"`
	FromName string `yaml:"fromName"`
	Password string `yaml:"password"`
}
