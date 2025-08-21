package config

import "time"

var HTTPServer struct {
	Addr     string
	FileRoot string
	KeyRoot  string
}

var Kubectl struct {
	Config []string
}

var Slack struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

var Postgres struct {
	DSN string
}

var Workflow struct {
	Namespace       string
	QueueName       string
	Server          string
	ActivityTimeout time.Duration
}

var Gitlab struct {
	Token string
}
