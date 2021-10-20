package sqlstore

import "strings"

//Phone Operators
type OperatorsRepository struct {
	store    IStore
	memcache map[string]uint64
}

// Constructor
func NewOperators(store IStore) *OperatorsRepository {
	l := make(map[string]uint64, 1000)
	return &OperatorsRepository{
		store:    store,
		memcache: l,
	}
}

// Сохраняет
func (r *OperatorsRepository) save(name string) (uint64, error) {
	res, err := r.store.DB().Exec("INSERT INTO `phone_operators`( `name`) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	lid, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(lid), nil
}

// поиск по имени. Возвращает id или ошибку
func (r *OperatorsRepository) IdByName(name string) uint64 {
	name = strings.ReplaceAll(name, "\"\"\"", "@")
	name = strings.ReplaceAll(name, "\"\"", "@")
	name = strings.ReplaceAll(name, "\"", "")
	name = strings.ReplaceAll(name, "@", "\"")
	if r, ok := r.memcache[name]; ok {
		return r
	}
	var have uint64
	err := r.store.DB().QueryRow("SELECT `id` FROM `phone_operators` WHERE `name` LIKE ?", name).Scan(&have)
	if err != nil || have == 0 {
		have, err = r.save(name)
		if err != nil {
			r.store.Logger().Error("Фатальная ошибка добавления оператора ", err)
			panic("Фатальная ошибка добавления оператора")
		}
	}
	r.memcache[name] = have
	return have
}
