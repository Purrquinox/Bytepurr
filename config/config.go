package config

type Config struct {
	Server  Server  `yaml:"server" validate:"required"`
	Storage Storage `yaml:"storage" validate:"required"`
	AWS     AWS     `yaml:"aws_s3" validate:"required"`
}

type Server struct {
	Port string `yaml:"port" comment:"Server Port" validate:"required"`
	Env  string `yaml:"env" comment:"Server Environment" validate:"required"`
}

type Storage struct {
	DatabaseURL string `yaml:"database_url" comment:"Database URL" validate:"required"`
	RedisURL    string `yaml:"redis_url" comment:"Redis URL" validate:"required"`
}

type AWS struct {
	Key      string `yaml:"key" comment:"AWS Access Key" validate:"required"`
	Secret   string `yaml:"secret" comment:"AWS Secret Key" validate:"required"`
	Endpoint string `yaml:"endpoint" comment:"AWS Endpoint" validate:"required"`
}
