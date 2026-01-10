package mailer

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewConfig(
	host string,
	port int,
	username string,
	password string,
	from string,
) *Config {
	return &Config{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}
