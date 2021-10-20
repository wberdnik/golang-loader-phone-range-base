package sqlstore

import (
	"bitbucket.org/proflead/golang/internal/app/model"
	"errors"
	"strconv"
)

//City Repositary
type KladrRepository struct {
	store IStore
}

// Constructor
func NewKladr(store IStore) *KladrRepository {
	return &KladrRepository{
		store: store,
	}
}

// поиск по имени. Возвращает id или ошибку
func (r *KladrRepository) FindByName(name string) (uint64, error) {
	var have uint64
	err := r.store.DB().QueryRow("SELECT `city_id` FROM `prof_city` WHERE `name` LIKE ?", name).Scan(&have)
	if err != nil || have == 0 {
		return 0, err
	}
	return have, nil
}

// Ищет строку по условию без конструкций IN()
func (r *KladrRepository) FindOne(condition map[string]string) (model.Kladr_Model, error) {
	have := model.Kladr_Model{}
	strCond := ""

	for field, val := range condition {
		if strCond != "" {
			strCond += " AND "
		}
		if _, err := strconv.Atoi(val); err == nil {
			strCond += field + " = \"" + val + "\""
		} else {
			strCond += field + " = " + val
		}
	}

	err := r.store.DB().QueryRow("SELECT `id`, `name`, `socr`, `code_kladr`, "+
		"`district_id`, `near_city_id`, `city_id`, `region_id`, `status`, "+
		"`Is_Capital` FROM `prof_kladr` WHERE "+strCond).Scan(&have.Id, &have.Name, &have.Socr, &have.Code_kladr,
		&have.District_id, &have.Near_city_id, &have.City_id, &have.Region_id, &have.Status, &have.Is_Capital)
	if err != nil {
		return have, err
	}
	if have.Id == 0 {
		return have, errors.New(" неверный kladr_id")
	}
	return have, nil
}
