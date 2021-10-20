package model

import "database/sql"

// Lead
type Kladr_Model struct {
	Id            uint64
	Name          string        //`единственное название`
	Socr          string        //`Тип населенного пункта`
	Code_kladr    string        //`Длинный код кладр`
	District_id   sql.NullInt64 //`Принадлежит району kladr_id`
	Near_city_id  sql.NullInt64 //`Ближайший город kladr_id`
	City_id       sql.NullInt64 //`Принадлежит город city_id`
	Region_id     uint64        //`Принадлежит региону region_id`
	Status        uint8         //		Нет	0	0 -норма, 1 - стоп слово,2 - модерация
	Is_Capital    uint8         // 2 - это регион/район 1 - центр региона/района 0 - нет
	Settlement_id uint64        // Свободное поле
}
