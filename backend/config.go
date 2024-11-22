package main

type Config struct {
	OidcProvider        string `json:"oidcProvider"`
	OidcClientId        string `json:"oidcClientId"`
	DisableSecureCookie bool   `json:"disableSecureCookie"`
	MacKey              string `json:"macKey"`
}
