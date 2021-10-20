package sqlstore

type RegionInfo struct {
	City_id_c    uint64
	City_id_fake uint64
	Name         string
	Region_id    uint64
}

//Region repository
type RegionRepository struct {
	store    IStore
	memcache map[uint64]RegionInfo
}

// Constructor
func NewRegion(store IStore) *RegionRepository {
	mk := make(map[uint64]RegionInfo)
	return &RegionRepository{
		store:    store,
		memcache: mk,
	}
}

// поиск по имени. Возвращает id или ошибку
func (r *RegionRepository) FindByName(name string) (uint64, error) {
	var have uint64
	err := r.store.DB().QueryRow("SELECT `region_id` FROM `prof_region` WHERE country_id = 3159 AND `name` LIKE ?", name).Scan(&have)
	if err != nil || have == 0 {
		return 0, err
	}
	return have, nil
}

// поиск regionInfo по region_id
func (r *RegionRepository) GetInfo(region_id uint64) (RegionInfo, error) {
	info, ok := r.memcache[region_id]
	if ok {
		return info, nil
	}
	have := RegionInfo{
		0, 0, "", region_id}
	err := r.store.DB().QueryRow("SELECT `city_id`, `name`, `city_id_c` FROM `prof_region` WHERE `region_id` = ?", region_id).
		Scan(&have.City_id_fake, &have.Name, &have.City_id_c)

	if err == nil {
		r.memcache[region_id] = have
	}
	return have, err
}

//
//func (r *RegionRepository) FindAllBySql(sql string) ([]*RegionInfo, error) {
//
//	rows, err := r.store.DB().Query(sql)
//	if err != nil {
//		return nil, err
//	}
//
//	defer func() { _ = rows.Close() }()
//
//	res := make([]*RegionInfo, 0, 100) // вряд ли у партнера больше промо, поэтому переинициализация среза маловероятна
//	for rows.Next() {
//		tmp := new(sql.NullString) // костыль
//		pm := new(model.Promo)
//		err := rows.Scan(
//			&pm.Name,
//			&pm.Origin,
//			&pm.Token,
//			tmp,
//			&pm.Email_LeadBack)
//		if err != nil {
//			r.store.Logger().Fatal(err)
//		}
//
//		ans := new(RegionInfo)
//		//ans.Email = ""
//		if pm.Email_LeadBack.Valid {
//			ans.Email = pm.Email_LeadBack.String
//		}
//
//		ans.Name = pm.Name
//		ans.Site = pm.Origin
//		ans.Token = pm.Token
//
//		if !tmp.Valid {
//			ans.RulesCount = 0
//		} else {
//			pm.LeadBack = tmp.String
//			pack, err := pm.Tornados() // для этого метада структура должна быть полностью заполнена
//			if err != nil {
//				ans.RulesCount = 0
//			} else {
//				ans.RulesCount = len(pack.Tornados)
//			}
//		}
//		res = append(res, ans)
//	}
//	if err = rows.Err(); err != nil {
//		r.store.Logger().Fatal(err)
//	}
//
//	return res, nil
//}
