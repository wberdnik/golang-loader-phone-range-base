package main

import (
	"bitbucket.org/proflead/golang/internal/app/config"
	"bitbucket.org/proflead/golang/internal/app/csv_loader"
	"flag"
	"github.com/burntSushi/toml"
	"log"
)

var (
	configPath string
	сsvFile    string
)

func init() {
	flag.StringVar(&configPath, "config-path", "./configs/csv_loader.toml", "path to config file")
	flag.StringVar(&сsvFile, "csv-path", "C:\\Users\\___\\Desktop\\rangeBase.csv", "path to CSV file")
}

func main() {
	flag.Parse()
	cfg := config.NewConfig()

	_, err := toml.DecodeFile(configPath, cfg)

	if err != nil {
		log.Fatal(err)
	}

	if err := csv_loader.Start(cfg, сsvFile); err != nil {
		panic("Не удалось запустить программу " + err.Error())
	}
}
