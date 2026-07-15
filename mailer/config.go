package mailer

// Config holds the SMTP server settings used by SMTPMailer.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// NewConfig builds an SMTP Config from its individual settings.
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
