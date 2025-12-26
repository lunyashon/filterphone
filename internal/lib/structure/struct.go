package structure

type Config struct {
	TcpPort     string `env:"TCP_PORT" env-default:"60777"`
	HostDb      string `env:"HOST_DB" env-default:"127.0.0.1"`
	PortDb      string `env:"PORT_DB" env-default:"5432"`
	NameDb      string `env:"NAME_DB" env-default:"db"`
	LoginDb     string `env:"LOGIN_DB" env-default:"root"`
	PassDb      string `env:"PASS_DB" env-default:"root"`
	LogPath     string `env:"LOG_PATH" env-default:"./logs"`
	TokenSecret string `env:"TOKEN_SECRET" env-default:"secret"`
}

type Numbers struct {
	Code           int16  `json:"code" db:"code"`
	From           int    `json:"from" db:"from_n"`
	To             int    `json:"to" db:"to_n"`
	Capacity       int    `json:"capacity" db:"capacity"`
	Operator       string `json:"operator" db:"operator"`
	Region         string `json:"region" db:"region"`
	Territory      string `json:"territory" db:"territory"`
	INN            int64  `json:"inn" db:"inn"`
	MobileOperator string `json:"mobile_operator" db:"mobile_operator"`
}
