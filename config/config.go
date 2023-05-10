package config

type Config struct {
	Database `yaml:"db"`
	Telegram `yaml:"tg"`
}

type Database struct {
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Telegram struct {
	Token string `yaml:"token"`
}
