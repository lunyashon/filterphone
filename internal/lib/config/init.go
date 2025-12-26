package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

func GetInstance(path string) {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Config file not found: ", path, "; Use default config")
		path = getConfig()
	}

	if err := godotenv.Load(fmt.Sprintf("%s/.env", path)); err != nil {
		panic(err)
	}
}

func getConfig() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}
