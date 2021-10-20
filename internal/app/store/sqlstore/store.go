package sqlstore

import (
	"bitbucket.org/proflead/golang/internal/app/model"
	"database/sql"
	"github.com/sirupsen/logrus"
)

// Store
type Store struct {
	db        *sql.DB
	logger    *logrus.Logger
	operator  *OperatorsRepository
	cacheAddr map[string]model.Geo

	region_id, city_id, end_number, operator_id, id, prefix int64
}

// Static Constructor
func New(db *sql.DB, logger *logrus.Logger) IStore {
	return &Store{
		db:        db,
		logger:    logger,
		cacheAddr: make(map[string]model.Geo, 10000),
	}
}

/**
getter db
*/
func (s *Store) DB() *sql.DB {
	return s.db
}

/**
getter Logger
*/
func (s *Store) Logger() *logrus.Logger {
	return s.logger
}

// getter Operators repository
func (s *Store) GetOperatorsRep() *OperatorsRepository {
	if s.operator == nil {
		s.operator = NewOperators(s)
	}
	return s.operator
}

// Запись итоговых значений
func (s *Store) SaveResult(region_id, city_id, start_number, end_number, prefix, operator_id int64, region_name string) {
	if s.id > 0 {
		if region_id == s.region_id && city_id == s.city_id && s.prefix == prefix && operator_id == s.operator_id && s.end_number+1 == start_number {
			_, err := s.db.Exec("UPDATE `tel2region` SET `end_number`= ? WHERE `id` = ?", end_number, s.id)
			if err != nil {
				panic("Ошибка обновления таблицы tel2region")
			}
			s.end_number = end_number
			return
		}
	}
	res, err := s.db.Exec("INSERT INTO `tel2region`( `region_id`, `start_number`, `end_number`, `region_name`, `prefix`, `city_id`, `operator_id`) "+
		"VALUES (?,?,?,?,?,?,?)", region_id, start_number, end_number, region_name, prefix, city_id, operator_id)
	if err != nil {
		panic("Ошибка записи таблицу tel2region")
	}
	lid, err := res.LastInsertId()
	if err != nil {
		panic("Ошибка получения id tel2region")
	}

	s.region_id = region_id
	s.city_id = city_id
	s.operator_id = operator_id
	s.end_number = end_number
	s.id = lid
	s.prefix = prefix
}

// get Geo info by string
func (s *Store) GetGeoInfo(adr string) model.Geo {
	if v, ok := s.cacheAddr[adr]; ok {
		return v
	}

	have, err := GeoLocation(adr, s)
	if err == nil {
		s.cacheAddr[adr] = have
		return have
	}
	panic("Ошибка определения адреса " + adr)
}
