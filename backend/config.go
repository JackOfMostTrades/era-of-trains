package main

type DatabaseConfig struct {
	Bootstrap     bool   `json:"bootstrap"`
	SqlitePath    string `json:"sqlitePath"`
	MysqlHostname string `json:"mysqlHostname"`
	MysqlDatabase string `json:"mysqlDatabase"`
	MysqlUsername string `json:"mysqlUsername"`
	MysqlPassword string `json:"mysqlPassword"`
}

type EmailConfig struct {
	Disabled     bool   `json:"disabled"`
	SmtpServer   string `json:"smtpServer"`
	SmtpPort     int    `json:"smtpPort"`
	SmtpUsername string `json:"smtpUsername"`
	SmtpPassword string `json:"smtpPassword"`
}

type Config struct {
	OidcProvider        string          `json:"oidcProvider"`
	OidcClientId        string          `json:"oidcClientId"`
	DisableSecureCookie bool            `json:"disableSecureCookie"`
	MacKey              string          `json:"macKey"`
	WorkingDirectory    string          `json:"workingDirectory"`
	CgiMode             bool            `json:"cgiMode"`
	Email               *EmailConfig    `json:"email"`
	Database            *DatabaseConfig `json:"database"`
}
