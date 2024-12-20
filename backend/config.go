package main

type DatabaseConfig struct {
	Bootstrap     bool   `json:"bootstrap"`
	SqlitePath    string `json:"sqlitePath"`
	MysqlHostname string `json:"mysqlHostname"`
	MysqlDatabase string `json:"mysqlDatabase"`
	MysqlUsername string `json:"mysqlUsername"`
	MysqlPassword string `json:"mysqlPassword"`
}

type Config struct {
	OidcProvider        string          `json:"oidcProvider"`
	OidcClientId        string          `json:"oidcClientId"`
	DisableSecureCookie bool            `json:"disableSecureCookie"`
	MacKey              string          `json:"macKey"`
	WorkingDirectory    string          `json:"workingDirectory"`
	CgiMode             bool            `json:"cgiMode"`
	Database            *DatabaseConfig `json:"database"`
}
