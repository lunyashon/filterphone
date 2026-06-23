package structure

import (
	"database/sql"
	"time"
)

type Config struct {
	TcpPort     string `env:"TCP_PORT" env-default:"60777"`
	HostDb      string `env:"HOST_DB" env-default:"127.0.0.1"`
	PortDb      string `env:"PORT_DB" env-default:"5432"`
	NameDb      string `env:"NAME_DB" env-default:"db"`
	LoginDb     string `env:"LOGIN_DB" env-default:"root"`
	PassDb      string `env:"PASS_DB" env-default:"root"`
	LogPath     string `env:"LOG_PATH" env-default:"./logs"`
	TokenSecret string `env:"TOKEN_SECRET" env-default:"secret"`
	B24ZaryaUrl string `env:"B24_ZARYA_URL" env-default:""`
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

type Retarget struct {
	NameArray      string       `db:"name_array"`
	FinalLink      string       `db:"final_link"`
	Operator       string       `db:"operator"`
	Phone          string       `db:"phone"`
	Territory      string       `db:"territory"`
	UniqueToken    string       `db:"unique_token"`
	TargetUrl      string       `db:"target_url"`
	UtmCompaign    string       `db:"utm_compaign"`
	UtmSource      string       `db:"utm_source"`
	UtmContent     string       `db:"utm_content"`
	UtmMedium      string       `db:"utm_medium"`
	UtmTerm        string       `db:"utm_term"`
	CreatedAt      time.Time    `db:"created_at"`
	TTL            time.Time    `db:"ttl"`
	LastFollowLink sql.NullTime `db:"last_follow_link"`
	FollowLink     int          `db:"follow_link"`
	UniqueIDArray  string       `db:"unique_id_array"`
}
