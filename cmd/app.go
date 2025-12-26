package main

import (
	"os"
	"sync"

	"github.com/lunyashon/filterphone/internal/app"
	"github.com/lunyashon/filterphone/internal/database"
	"github.com/lunyashon/filterphone/internal/lib/config"
	"github.com/lunyashon/filterphone/internal/lib/logger"
	"github.com/lunyashon/filterphone/internal/lib/structure"
)

func main() {
	cfg := getConfig()
	log := logger.ExecLog(cfg.LogPath)

	var (
		once sync.Once
		db   *database.Database
		err  error
	)

	once.Do(func() {
		db, err = database.GetInstance(log, cfg)
		if err != nil {
			panic(err)
		}
	})

	defer db.Base.Close()

	app.Run(cfg, log, db)
}

func getConfig() *structure.Config {
	config.GetInstance("")
	return &structure.Config{
		TcpPort:     os.Getenv("TCP_PORT"),
		HostDb:      os.Getenv("HOST_DB"),
		PortDb:      os.Getenv("PORT_DB"),
		NameDb:      os.Getenv("NAME_DB"),
		LoginDb:     os.Getenv("LOGIN_DB"),
		PassDb:      os.Getenv("PASS_DB"),
		LogPath:     os.Getenv("LOG_PATH"),
		TokenSecret: os.Getenv("TOKEN_SECRET"),
	}
}
