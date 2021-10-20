package csv_loader

import (
	"bitbucket.org/proflead/golang/internal/app/config"
	"bitbucket.org/proflead/golang/internal/app/store/sqlstore"
	"bufio"
	"os"
	"strconv"
	"strings"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"time"
)

// Start
func Start(config *config.Config, csvPath string) error {
	logger := logrus.New()
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		panic(err)
	}
	logger.SetLevel(level)
	logger.Info(`Start CSV Loader `)
	defer logger.Info(`Stopped CSV Loader`)

	file, err := os.OpenFile(csvPath, os.O_RDONLY, 0111)
	if err != nil {
		logger.Error("Не найден файл CSV", err)
		return err
	}
	defer file.Close()

	db, err := connectDB(config.DatabaseDSN)
	if err != nil {
		logger.Error("Ошибка подключения к БД", err)
		return err
	}
	defer func() { _ = db.Close() }()

	s := sqlstore.New(db, logger)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ",")
		if fields[1] == "От" {
			continue
		}
		start := fields[0] + fields[1]
		if len(fields[1]) != 7 {
			start = fields[0] + strings.Repeat("0", 7-len(fields[1])) + fields[1]
		}
		stop := fields[0] + fields[2]
		if len(fields[2]) != 7 {
			stop = fields[0] + strings.Repeat("0", 7-len(fields[2])) + fields[2]
		}
		operator := fields[4]
		geo := fields[5]
		registerPhone(fields[0], start, stop, operator, geo, s)
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Ошибка чтения файла", err)
		return err
	}

	return nil
}

func connectDB(DSN string) (*sql.DB, error) {
	db, err := sql.Open("mysql", DSN)
	if err != nil {
		return nil, err
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// регистрация строки CSV
func registerPhone(pref, start, stop, operator, geo string, store sqlstore.IStore) {
	operatorId := int64(store.GetOperatorsRep().IdByName(operator))
	startN, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		panic("Ошибка приведения к числу Start " + start)
	}
	stopN, err := strconv.ParseInt(stop, 10, 64)
	if err != nil {
		panic("Ошибка приведения к числу Stop " + stop)
	}
	prefN, err := strconv.Atoi(pref)
	if err != nil {
		panic("Ошибка приведения к числу Prefix " + pref)
	}
	geoInfo := store.GetGeoInfo(geo)

	store.SaveResult(int64(geoInfo.Region_id), int64(geoInfo.City_id), startN, stopN, int64(prefN), operatorId, geoInfo.Region_name)
}
