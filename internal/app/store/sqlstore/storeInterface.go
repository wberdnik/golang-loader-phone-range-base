package sqlstore

import (
	"bitbucket.org/proflead/golang/internal/app/model"
	"database/sql"
	"github.com/sirupsen/logrus"
)

type IStore interface {
	DB() *sql.DB
	Logger() *logrus.Logger
	GetOperatorsRep() *OperatorsRepository
	SaveResult(region_id, city_id, start_number, end_number, prefix, operator_id int64, region_name string)
	GetGeoInfo(adr string) model.Geo
}
