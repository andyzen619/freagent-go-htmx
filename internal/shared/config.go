package config

import "os"

var (
	ClientID     = os.Getenv("YAHOO_CLIENT_ID")
	ClientSecret = os.Getenv("YAHOO_CLIENT_SECRET")
	RedirectURI  = os.Getenv("YAHOO_REDIRECT_URI")
)
