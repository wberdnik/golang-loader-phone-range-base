package sqlstore

//City Repositary
type CityRepository struct {
	store    IStore
	memcache map[uint64]uint64
}

// Constructor
func NewCity(store IStore) *CityRepository {
	mk := make(map[uint64]uint64)
	return &CityRepository{
		store:    store,
		memcache: mk,
	}
}

// поиск по имени. Возвращает id или ошибку
func (r *CityRepository) FindByName(name string) (uint64, error) {
	var have uint64
	err := r.store.DB().QueryRow("SELECT `city_id` FROM `prof_city` WHERE `name` LIKE ?", name).Scan(&have)
	if err != nil || have == 0 {
		return 0, err
	}
	return have, nil
}

// поиск region_id по city_id
func (r *CityRepository) City2region(city_id uint64) (uint64, error) {
	region, ok := r.memcache[city_id]
	if ok {
		return region, nil
	}
	var have uint64
	err := r.store.DB().QueryRow("SELECT `region_id` FROM `prof_city` WHERE `city_id` = ?", city_id).Scan(&have)
	if err != nil || have == 0 {
		return 0, err
	}
	r.memcache[city_id] = have

	return have, nil
}

type CityInfo struct {
	Name       string
	Region_id  uint64
	RegionInfo RegionInfo
}

// поиск  по city_id
func (r *CityRepository) City2Info(city_id uint64) (CityInfo, error) {

	var have CityInfo
	err := r.store.DB().QueryRow("SELECT ct.`region_id`, ct.`name`, "+
		"IF(rn.`city_id` IS NULL, 0, rn.`city_id`), "+
		"IF(rn.`name` IS NULL, '', rn.`name`), IF(rn.`city_id_c` IS NULL, 0, rn.`city_id_c`) "+
		"FROM `prof_city` ct LEFT JOIN `prof_city` as rn ON  rn.`region_id` = ct.`region_id` "+
		"WHERE ct.`city_id` = ?", city_id).
		Scan(&have.Region_id, &have.Region_id, &have.RegionInfo.City_id_fake, &have.RegionInfo.Name, &have.RegionInfo.City_id_c)
	if err != nil || have.Region_id == 0 {
		return CityInfo{}, err
	}

	return have, nil
}
