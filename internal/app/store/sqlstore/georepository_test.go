package sqlstore_test

import (
	"bitbucket.org/proflead/golang/internal/app/store/sqlstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	databaseDSN string
	token       string
	user_id     uint64
)

func TestMain(m *testing.M) {
	databaseDSN = os.Getenv("DATABASE_DSN")
	if databaseDSN == "" {
		databaseDSN = "root:@/csv_loader"
	}
	os.Exit(m.Run())
}

func TestGeoLocation(t *testing.T) {
	db, _ := sqlstore.TestDB(t, databaseDSN)
	s := sqlstore.New(db, logrus.New())
	Rep, err := sqlstore.GeoLocation("г. Улан-Удэ|Республика Бурятия", s)
	assert.NoError(t, err)
	assert.Nil(t, Rep)
}
