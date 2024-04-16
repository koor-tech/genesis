package config

import (
	"path"
)

type Config struct {
	Listen string `default:":8000" yaml:"listen"`
	Mode   string `default:"debug" yaml:"mode"`

	Directories   Directories   `yaml:"directories"`
	Database      Database      `yaml:"database"`
	RabbitMQ      RabbitMQ      `yaml:"rabbitmq"`
	Notifications Notifications `yaml:"notifications"`
	CloudProvider CloudProvider `yaml:"cloudProvider"`
	Secret        string        `default:"abc" yaml:"secret"`
}

type Directories struct {
	DataDir string `default:"/koor/clients" yaml:"dataDir"`
}

func (c *Directories) TemplatesDir() string {
	return path.Join(c.DataDir, "templates")
}

func (c *Directories) ClientsDir() string {
	return path.Join(c.DataDir, "clients")
}

func (c *Directories) ChartsDir() string {
	return path.Join(c.DataDir, "charts")
}

type Database struct {
	Host       string `default:"localhost" yaml:"host"`
	Port       int    `default:"5432" yaml:"port"`
	User       string `default:"genesis" yaml:"user"`
	Password   string `yaml:"password"`
	Name       string `default:"genesis" yaml:"name"`
	SSLEnabled bool   `default:"false" yaml:"sslEnabled"`
}

type RabbitMQ struct {
	Host     string `default:"localhost" yaml:"host"`
	Port     int    `default:"4369" yaml:"port"`
	User     string `default:"genesis" yaml:"user"`
	Password string `yaml:"password"`
}

type NotificationType string

const (
	NotificationTypeNoop  NotificationType = "noop"
	NotificationTypeEmail NotificationType = "email"
)

type Notifications struct {
	Type NotificationType `default:"noop" yaml:"type"`

	Email EmailNotifications `yaml:"email"`
}

type EmailNotifications struct {
	Token string `yaml:"token"`

	From    string `yaml:"from"`
	ReplyTo string `yaml:"replyTo"`
}

type CloudProvider struct {
	Hetzner Hetzner `yaml:"hetzner"`
}

type Hetzner struct {
	Token string `yaml:"token"`
}
