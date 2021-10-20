package csv_loader

import (
	"bitbucket.org/proflead/golang/internal/app/model"
	"bitbucket.org/proflead/golang/internal/app/store"
	"database/sql"
	"os"
	"testing"
)

var (
	databaseDSN string
)

func TestMain(m *testing.M) {
	databaseDSN = os.Getenv("DATABASE_DSN")
	//	token = "hIzM5-4jXE-3lw_BjP-oBkRyr2Ey"
	//	user_id = 490
	if databaseDSN == "" {
		databaseDSN = "root:@/csv_loader"
	}
	os.Exit(m.Run())
}

type testPromo struct {
	u int
}

func (_l *testPromo) Save(*model.Promo, uint64) error {
	return nil
}
func (_l *testPromo) FindByToken(string, uint64) (*model.Promo, error) {
	pm := new(model.Promo)
	pm.Token = "1"
	pm.Name = "Первый"
	pm.Email_LeadBack = sql.NullString{
		Valid:  true,
		String: "t@t.ru",
	}
	pm.Origin = "Сайт"

	return pm, nil
}
func (_l *testPromo) FindAllByUser(_ uint64) (res []*model.Promo, err error) {
	err = nil
	pm := new(model.Promo)
	pm.Token = "1"
	pm.Name = "Первый"
	pm.Email_LeadBack = sql.NullString{
		Valid:  true,
		String: "t@t.ru",
	}
	pm.Origin = "Сайт"

	res = append(res, pm)

	pm = new(model.Promo)
	pm.Token = "2"
	pm.Name = "Второй"
	pm.Email_LeadBack = sql.NullString{
		Valid:  true,
		String: "t2@t.ru",
	}
	pm.Origin = "Сайт"

	res = append(res, pm)
	return
}

type testStore struct {
}

func (_l *testStore) Promo() store.IGeoRepository {
	return nil //new(testPromo)
}
func (_l *testStore) DB() *sql.DB {
	return nil
}
