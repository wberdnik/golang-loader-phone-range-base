package sqlstore

import (
	"bitbucket.org/proflead/golang/internal/app/model"
	"bitbucket.org/proflead/golang/internal/pkg"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// Trust class
type Trust struct {
	collection      []model.Kladr_Model
	t_city          map[uint64]uint16
	t_region        map[uint64]uint16
	t_settlement_id map[uint64]uint16
	t_district_id   map[uint64]uint16
	store           IStore
}

// Constructor class Trust
func NewTrust(store IStore) *Trust {
	return &Trust{
		store:           store,
		collection:      make([]model.Kladr_Model, 0, 100),
		t_city:          make(map[uint64]uint16, 20),
		t_region:        make(map[uint64]uint16, 20),
		t_settlement_id: make(map[uint64]uint16, 20),
		t_district_id:   make(map[uint64]uint16, 20),
	}
}

// обрезание морфологии слова (окончаний)
func clearEnds(word string) string {
	word = strings.Trim(word, " \t\n\r")
	if word == "крым" {
		return word
	}

	ln := len(word)
	if ln < 4 {
		return word
	}

	word_ends := map[int][]string{
		2: {"ый", "ий", "ая", "яя", "ое", "ее", "ые", "ие", "ой", "ей", "ых", "их", "ым", "им", "ую", "юю", "ом", "ем"},
		3: {"ого", "его", "ому", "ему", "ыми", "ими"},
	}

	if pkg.StrArrayContains(word_ends[2], word[ln-2:]) {
		return word[:ln-2]
	}
	if ln < 5 {
		return word
	}

	if pkg.StrArrayContains(word_ends[3], word[ln-3:]) {
		return word[:ln-3]
	}
	return word
}

//Item of syntax rules
type ruleSyntax struct {
	in  []string
	out string
}

//Cжатие различия знаков препинания, скобок, удаление множественных пробелов
func SyntaxClear(str string) string {
	if strings.Trim(str, " \t\n\r") == "" {
		return str
	}
	l_str := " " + strings.ToLower(str) + " "

	syntax_rules := []ruleSyntax{
		{[]string{",", ";", ":", "@", "|", "/", "\\\\", "#", "*"}, " , "}, // что на что
		{[]string{"{", "{", "<<", "«"}, " ( "},
		{[]string{"."}, ". "},
		{[]string{"}", "]", ">>", "»"}, " ) "},
		{[]string{"ё"}, "е"},
		{[]string{"a"}, "а"}, // латынь - кирилица
		{[]string{"e"}, "е"},
		{[]string{"x"}, "х"},
		{[]string{"p"}, "р"},
		{[]string{"o"}, "о"},
		{[]string{"c"}, "с"},
		{[]string{"     ", "    ", "   ", "  ", " "}, " "},
	}

	for _, rule := range syntax_rules {
		for _, substr := range rule.in {
			l_str = strings.ReplaceAll(l_str, substr, rule.out)
		}
	}
	return strings.Trim(l_str, " \t\n\r")
}

//Item of terminals rules
type ruleTerminals struct {
	in     []string
	marker string
	extra  string
	newIn  []string
	prior  int
}

// Single responsibility структуры правил маркеров гео объектов
func getTerminalRules() map[string]ruleTerminals {
	terminals := map[string]ruleTerminals{

		"region": {
			in: []string{"обл", "обл.", "область", "об", "области",
				"респ", "респ.", "республика", "респ-ка",
				"край", "кр.", "кр-й", "края",
				"АО", "автономная обл", "авт.обл", "авт.обл.", "авт. обл", "авт. обл.", "а. обл.", "а-обл.", "а. обл", "а-обл", "а обл", "а.обл",
				"автономная область", "автономной области",
				"АОбл", "автоном. окр", "авт.окр", "авт.окр.", "авт. окр", "авт. окр.", "а. окр.", "а-окр.", "а. окр", "а-о", "а о", "а.окр",
				"автоном", "автономный округ", "автономного округа",
			},
			marker: " ~!@n#region ",
			extra:  " ",
			newIn:  []string{"обл", "Респ", "край", "АО", "Аобл"},
			prior:  2}, // ~!@r# место маркера {n - неважно, r - справо, l- слева}

		"district": {
			in:     []string{"рн", "рн.", "р-н", "р-он", "район", "района", "р-н."},
			marker: " ~!@r#district ",
			extra:  " ",
			newIn:  []string{"р-н"},
			prior:  2},

		"city": {
			in:     []string{"г", "г.", "гор", "гор.", "город", "городо", "города"},
			marker: " ~!@l#city ",
			extra:  " ",
			newIn:  []string{"г", "г.", "городо"},
			prior:  2},

		"settlement": {
			in: []string{"х", "х.", "хутор", "п", "п.", "пос.", "поселок", "с", "с.", "сел", "сел.", "село",
				"д.", "дер", "дер.", "деревня", "ст-ца", "станица", "с-ца", "снт", "снт.", "сл.", "слабода", "квартал", "кварт",
				"дп", "аул", "пгт", "пгт.", "ст", "тер", "нп", "рзд", "д", "рп", "казарм", "мкр", "с/с",
				"промзо", "у", "п/ст", "м", "заимка", "остров", "с/п", "кв-л", "аал", "ж/д_рз",
				"ж/д_ст", "кп", "автодо", "кордон", "высел", "почино", "арбан", "жилрай", "с/мо",
				"ж/д_ка", "массив", "п/о", "гп", "с/а", "ж/д_бу", "ж/д_по", "ж/д_пл", "погост", "волост",
				"сл", "ж/д_оп", "п/р", "жилзон", "г.о.", "с/о", "лпх"},
			marker: " ~!@n#settlement ",
			extra:  " ",
			newIn: []string{"х", "п", "с", "ст-ца", "снт", "дп", "аул", "пгт", "ст", "тер", "нп", "рзд", "д",
				"рп", "казарм", "мкр", "с/с", "промзо", "у", "п/ст", "м", "заимка", "остров", "с/п", "кв-л", "аал", "ж/д_рз",
				"ж/д_ст", "кп", "автодо", "кордон", "высел", "почино", "арбан", "жилрай", "с/мо",
				"ж/д_ка", "массив", "п/о", "гп", "с/а", "ж/д_бу", "ж/д_по", "ж/д_пл", "погост", "волост",
				"сл", "ж/д_оп", "п/р", "жилзон", "г.о.", "с/о", "лпх",
			},
			prior: 2},
	}
	return terminals
}

// Работа с маркерами географических объектов
func ClearMarkers(str, group string, markers ...string) string {
	marker := ""
	if len(markers) > 0 {
		marker = markers[0]
	}

	terminals := getTerminalRules()
	ruleSyntax := terminals[group]

	l_str := " " + str + " "
	for _, value := range ruleSyntax.in {
		l_str = strings.ReplaceAll(l_str, " "+value+" ", marker)
		l_str = strings.ReplaceAll(l_str, "  ", " ")
	}

	l_str = strings.ReplaceAll(l_str, "  ", " ")
	return strings.Trim(l_str, " \t\n\r")
}

//Оставить только значимые слова без маркеров вида "г", "р-н", запятых, знаков препинания и т.п. в массив и без окончаний
func Str2ImportantWords(str string) []string {
	if strings.Trim(str, " \t\n\r") == "" {
		return []string{}
	}

	lStr := strings.Trim(SyntaxClear(str), " \t\n\r")
	lStr = " " + strings.ToLower(lStr) + " "
	terminals := getTerminalRules()

	for _, r := range terminals {
		for _, value := range r.in {
			lStr = strings.ReplaceAll(lStr, " "+value+" ", " ")
			lStr = strings.ReplaceAll(lStr, "  ", " ")
		}
	}

	syntax_rules2 := []ruleSyntax{
		{[]string{",", ";", ":", "@", "|", "/", "#", "*", "{", "{", "<<", "\\\\", //   """,
			"«", "(", "}", "}", ">>", "»", ")", "+", "-", "_", "`", "~", "!", "№", "$$",
			"; ", "%", "^", ":", "&", " = ", ">", "<"}, " "}, // что на что
		{[]string{"."}, ". "},
		{[]string{"ё"}, "е"},
		{[]string{"a"}, "а"}, // латынь - кирилица
		{[]string{"e"}, "е"},
		{[]string{"x"}, "х"},
		{[]string{"p"}, "р"},
		{[]string{"o"}, "о"},
		{[]string{"c"}, "с"},
		{[]string{"     ", "    ", "   ", "  ", " ", "  "}, " "},
	}

	for _, rule := range syntax_rules2 {
		for _, substr := range rule.in {
			lStr = strings.ReplaceAll(lStr, substr, " ")
		}
	}

	tmp := strings.Fields(strings.Trim(lStr, " \t\n\r"))
	res := make([]string, 0, len(tmp))
	for _, value := range tmp {
		word := strings.Trim(value, " \t\n\r")
		if len(word) > 2 {
			res = append(res, word)
		}
	}
	return res
}

// ответ Recovery()
type GeoKeyValue map[string]uint64

// Дополняет Гео информацию по входным значениям
func (_l *Trust) Recovery(fieldName, value string) (GeoKeyValue, error) {
	rt := make(GeoKeyValue, 10)
	switch fieldName {
	case "code_kladr":

		kladrRep := NewKladr(_l.store)
		model, err := kladrRep.FindOne(map[string]string{
			"code_kladr": value,
		})
		if err != nil {
			return rt, err
		}
		rt["region_id"] = model.Region_id
		rt["country_id"] = 3159
		if model.City_id.Valid && model.City_id.Int64 > 0 {
			rt["city_id"] = uint64(model.City_id.Int64)
		} else {
			if model.Is_Capital == 0 {
				rt["settlement_id"] = model.Id
			}
		}
		if model.District_id.Valid && model.District_id.Int64 > 0 {
			rt["district_id"] = uint64(model.District_id.Int64)
		}
		return rt, nil

	case "region_id":

		regionRep := NewRegion(_l.store)
		reg_id, err := strconv.Atoi(value)
		if err != nil {
			return rt, err
		}
		_, err = regionRep.GetInfo(uint64(reg_id))
		if err != nil {
			return rt, err
		}
		rt["country_id"] = 3159
		rt["region_id"] = uint64(reg_id)

		return rt, nil

	case "district_id":
		d_id, err := strconv.Atoi(value)
		if err != nil {
			return rt, err
		}

		kladrRep := NewKladr(_l.store)
		model, err := kladrRep.FindOne(map[string]string{
			"id": fmt.Sprint(d_id),
		})
		if err != nil {
			return rt, err
		}
		rt["region_id"] = model.Region_id
		rt["district_id"] = uint64(d_id)
		rt["country_id"] = 3159
		if model.City_id.Valid && model.City_id.Int64 > 0 {
			rt["city_id"] = uint64(model.City_id.Int64)
		}
		return rt, nil
	case "city_id":
		city_id, err := strconv.Atoi(value)
		if err != nil {
			return rt, err
		}

		cityRep := NewCity(_l.store)
		reg_id, err := cityRep.City2region(uint64(city_id))
		if err != nil {
			return rt, err
		}
		rt["city_id"] = uint64(city_id)
		rt["region_id"] = reg_id
		rt["country_id"] = 3159
		return rt, nil
	}

	what_id, err := strconv.Atoi(value)
	if err != nil {
		return rt, err
	}

	rt[fieldName] = uint64(what_id)
	return rt, nil
}

type Fact struct {
	field string
	val   string
}

// регистрирует запись в матрице трастовости по принципу аддитивности, т.е многословные названия
func (_l *Trust) RegisterFacts(mass_variant []Fact) {
	if len(mass_variant) == 0 {
		return
	}

	local := map[string]map[uint64]uint16{
		"city_id":       make(map[uint64]uint16, 20),
		"district_id":   make(map[uint64]uint16, 30),
		"region_id":     make(map[uint64]uint16, 10),
		"settlement_id": make(map[uint64]uint16, 100),
	}

	for _, rul := range mass_variant {

		recover, err := _l.Recovery(rul.field, rul.val)
		if err != nil {
			continue
		}
		for field, val := range recover {
			if _, ok := local[field]; ok {
				local[field][val] = 1 // это свертка  в пределах одной транзакции
			}
		}

		equal := false // есть ли полное соответствие

		if len(_l.collection) > 0 {
			for _, kl := range _l.collection {
				equal = true
				for r_key, r_val := range recover {
					switch r_key {
					case "city_id":
						if !kl.City_id.Valid && r_val != 0 {
							equal = false
							break
						}

						if kl.City_id.Valid && uint64(kl.City_id.Int64) != r_val {
							equal = false
							break
						}

					//case "country_id": не может не совпасть
					//country_id[val] = 1

					case "district_id":
						if !kl.District_id.Valid && r_val != 0 {
							equal = false
							break
						}

						if kl.District_id.Valid && uint64(kl.District_id.Int64) != r_val {
							equal = false
							break
						}
					case "region_id":
						if kl.Region_id != r_val {
							equal = false
							break
						}
					case "settlement_id":
						if kl.Settlement_id != r_val {
							equal = false
							break
						}
					}
				} // for recover

				if equal {
					break // collection ActiveRecord of Kladr Нашли первую запись
				}
			} // for collection
		}

		if !equal { // Если не найден - добавляем
			nKl := model.Kladr_Model{
				Id:         99,
				Name:       "",
				Socr:       "",
				Code_kladr: "",
				District_id: sql.NullInt64{
					Int64: 0,
					Valid: false,
				},
				Near_city_id: sql.NullInt64{
					Int64: 0,
					Valid: false,
				},
				City_id: sql.NullInt64{
					Int64: 0,
					Valid: false,
				},
				Region_id:     0,
				Status:        0,
				Is_Capital:    0,
				Settlement_id: 0,
			}

			for r_key, r_val := range recover {
				switch r_key {
				case "city_id":
					nKl.City_id.Valid = true
					nKl.City_id.Int64 = int64(r_val)

				case "district_id":

					nKl.District_id.Valid = true
					nKl.District_id.Int64 = int64(r_val)
				case "region_id":
					nKl.Region_id = r_val

				case "settlement_id":
					nKl.Settlement_id = r_val
				}
			} // for recover
			_l.collection = append(_l.collection, nKl)

		}
	} // mass_variant

	// фиксируем упоминания в общей матрице

	for key := range local["city_id"] {
		if v, ok := _l.t_city[key]; ok {
			_l.t_city[key] = v + 1
			continue
		}
		_l.t_city[key] = 1
	}
	for key := range local["district_id"] {
		if v, ok := _l.t_district_id[key]; ok {
			_l.t_district_id[key] = v + 1
			continue
		}
		_l.t_district_id[key] = 1
	}

	for key := range local["region_id"] {

		if v, ok := _l.t_region[key]; ok {
			_l.t_region[key] = v + 1
			continue
		}
		_l.t_region[key] = 1
	}

	for key := range local["settlement_id"] {
		if v, ok := _l.t_settlement_id[key]; ok {
			_l.t_settlement_id[key] = v + 1
		}
		_l.t_settlement_id[key] = 1
	}
}

// Регистрирует запись в матрице трастовости по принципу аддитивности, т.е многословные названия
// Только в составе RefisterFacts
func (_l *Trust) RegisterFact(rul Fact) {
	mass_variant := make([]Fact, 0, 1)
	mass_variant = append(mass_variant, rul)
	_l.RegisterFacts(mass_variant)
}

// Расчет взвешенного значения по kladr модели
func (_l *Trust) getVes(kladr model.Kladr_Model) uint64 {
	vesCity := uint16(0)
	vesSettlement := uint16(0)

	vesRegion_c := uint16(0)
	vesRegion_s := uint16(0)

	kladrRep := NewKladr(_l.store)

	if kladr.City_id.Valid && kladr.City_id.Int64 > 0 {
		vesRegion_c = 0
		cityRep := NewCity(_l.store)
		cityId := uint64(kladr.City_id.Int64)
		if Region_id, err := cityRep.City2region(cityId); err == nil {
			if t, ok := _l.t_region[Region_id]; ok {
				vesRegion_c = t
				vesCity = _l.t_city[cityId]
			}
		}
	}

	if kladr.Settlement_id > 0 {
		if model, err := kladrRep.FindOne(map[string]string{"id": fmt.Sprint(kladr.Settlement_id)}); err == nil {
			vesSettlement = _l.t_settlement_id[kladr.Settlement_id]
			vesRegion_s = _l.t_region[model.Region_id]
		}
	}

	// развесовка регион - треть значимой части (города, села, района). Страна идет мелким довеском, но больше чем индексатор уникальности в массиве
	if vesCity > vesSettlement {
		return uint64(vesCity)*1000000 + uint64(vesRegion_c)*300000
	}
	if vesCity < vesSettlement {
		return uint64(vesSettlement)*1000000 + uint64(vesRegion_s)*300000
	}

	if vesCity > 0 && vesSettlement > 0 { // есть деревня и есть город
		if uint64(vesRegion_c)*300000 > uint64(vesRegion_s)*300000 {
			return uint64(vesCity)*1000000 + uint64(vesRegion_c)*300000
		} else {
			return uint64(vesSettlement)*1000000 + uint64(vesRegion_s)*300000
		}
	}

	// не определен город или деревня

	if kladr.District_id.Valid && kladr.District_id.Int64 > 0 {

		if model, err := kladrRep.FindOne(map[string]string{"id": fmt.Sprint(kladr.District_id)}); err == nil {
			vesRegion_s = _l.t_region[model.Region_id]
			district := _l.t_district_id[uint64(kladr.District_id.Int64)]
			return uint64(district*10000 + vesRegion_s*3000)
		}
	}

	if kladr.Region_id > 0 {
		regionRep := NewRegion(_l.store)
		if _, err := regionRep.GetInfo(kladr.Region_id); err == nil {
			vesRegion_c = _l.t_region[kladr.Region_id]
			return uint64(vesRegion_c * 1000)
		}
	}
	return uint64(0)
}

type prioritets struct {
	ves   uint64
	model model.Kladr_Model
}

// Возвращает упорядоченный по убыванию трастовости ассоциативный массив "значение траста" => \app\models\Kladr
/*func (_l *Trust) trust_all() []prioritets {
	if len(_l.collection) == 0 {
		return make([]prioritets, 0, 0)
	}

	r := make(map[uint64]model.Kladr_Model, 100)
	vesa := make([]uint64, 0, 100)
	for _, value := range _l.collection {
		ves := _l.getVes(value)
		_, ok := r[ves]
		for ok {
			ves++
			_, ok = r[ves]
		}
		vesa = append(vesa, ves)
		r[ves] = value
	}
	sort.Slice(vesa, func(i, j int) bool { return vesa[i] > vesa[j] })

	a := make([]prioritets, 10)
	n := 0
	for _, ves := range vesa {
		if n > 10 {
			break
		}
		n++
		a = append(a, prioritets{ves: ves, model: r[ves]})
	}
	return a
}*/

// Найти единственный лучший вариант
func (_l *Trust) trust_one() prioritets {
	have := prioritets{
		uint64(0),
		model.Kladr_Model{},
	}
	for _, value := range _l.collection {
		ves := _l.getVes(value)
		if ves > have.ves {
			have.ves = ves
			have.model = value
		}
	}
	return have
}

type AssArray map[string]string

// Находит варианты адреса по заданной строке
func findByAddress(addressParam AssArray, store IStore, answers ...*Trust) (*Trust, error) {
	var answer *Trust
	// default params

	if len(answers) != 0 {
		answer = answers[0]
	} else {
		answer = NewTrust(store)
	}
	var err error

	pref := "prof_"
	fields := map[string]int{"city": 0, "region": 0}

	lDb := store.DB()

	// входной фильтр
	address_free := ""

	addressMap := make(AssArray, 100)

	for name, val := range addressParam {
		if val == "" {
			continue
		}

		if _, ok := fields[name]; ok {
			addressMap[name] = SyntaxClear(val)
		} else {
			address_free += " " + val
		}
	}
	//fmt.Println("Address free ", address_free)
	//////////////////////////////////////////////////////////////////
	if val, ok := addressMap["region"]; ok {
		lStr := ClearMarkers(val, "region")

		if lStr != "" {

			rows, err := lDb.Query("SELECT DISTINCT `region_id` "+
				"FROM `"+pref+"region` "+
				"WHERE `name`= ? "+
				"UNION "+
				"SELECT DISTINCT `region_id` FROM `"+pref+"kladr` WHERE `Is_Capital` =2 "+
				"AND `district_id` IS NULL AND `name` = ? LIMIT 100", lStr, lStr)

			facts := make([]Fact, 0, 100)
			if err == nil {
				defer func() { _ = rows.Close() }()
				x := uint64(0)
				for rows.Next() {
					if rows.Scan(&x) == nil {
						facts = append(facts, Fact{"region_id", fmt.Sprint(x)})
					}
				}
			}

			if len(facts) > 0 {
				answer.RegisterFacts(facts)

			} else { // пословный поиск региона

				arr_words := Str2ImportantWords(addressMap["region"])
				str4rekursive := ""
				for _, word := range arr_words {
					word = clearEnds(word) // морфология

					rows, err := lDb.Query("SELECT DISTINCT `region_id` "+
						"FROM `"+pref+"region` "+
						"WHERE `name` LIKE ? "+
						"UNION "+
						"SELECT DISTINCT  `region_id` "+
						"FROM `"+pref+"kladr` WHERE "+
						"`Is_Capital` =2 AND `district_id` IS NULL "+
						"AND `name` LIKE ? LIMIT 100", "%"+word+"%", "%"+word+"%")
					if err == nil {
						defer func() { _ = rows.Close() }()

						for rows.Next() {
							x := uint64(0)
							if rows.Scan(&x) == nil {
								facts = append(facts, Fact{"region_id", fmt.Sprint(x)})
							}
						}
					}

					if len(facts) > 0 {
						answer.RegisterFacts(facts)
					} else {
						str4rekursive += word + " "
					}
				} //foreach ($arr_words as $word)

				if str4rekursive != "" { // доопределяем значимые слова
					if answer, err = findByAddress(AssArray{"f": str4rekursive}, store, answer); err != nil {
						return nil, err
					}
				}
			}
		} //if ($lStr !== "")
	} //ok := addressMap["region"]

	//////////////////////////////////////////////////////////////////

	if lStr, ok := addressMap["city"]; ok {
		lStr = ClearMarkers(lStr, "city")
		if lStr != "" {
			rule := getTerminalRules()["city"]
			facts := make([]Fact, 0, 100)

			args := []interface{}{lStr}
			for _, v := range rule.newIn {
				args = append(args, v)
			}
			args = append(args, lStr)

			rows, err := lDb.Query("SELECT DISTINCT "+
				"`city_id` as city_id, `region_id` as region_id "+
				"FROM `"+pref+"city` WHERE `name`= ? "+
				"UNION "+
				"SELECT DISTINCT `city_id`, `region_id` "+
				"FROM `"+pref+"kladr` WHERE "+
				"`socr` IN (?"+strings.Repeat(",?", len(rule.newIn)-1)+") AND `city_id` IS NOT NULL "+
				"AND `name` = ? LIMIT 100", args...)

			if err == nil {
				defer func() { _ = rows.Close() }()

				for rows.Next() {
					x := uint64(0)
					y := uint64(0)
					if rows.Scan(&x, &y) == nil {
						facts = append(facts, Fact{"city_id", fmt.Sprint(x)})
					}
				}
			}

			if len(facts) > 0 {
				answer.RegisterFacts(facts)
			} else { // пословный поиск региона
				arr_words := Str2ImportantWords(addressMap["city"])
				str4rekursive := ""
				for _, word := range arr_words {
					word = clearEnds(word) // морфология

					facts := make([]Fact, 0, 100)

					args := []interface{}{"%" + word + "%"}
					for _, v := range rule.newIn {
						args = append(args, v)
					}
					args = append(args, "%"+word+"%")

					rows, err := lDb.Query("SELECT DISTINCT `city_id` as city_id, "+
						"`region_id` as region_id FROM `"+pref+"city` WHERE `name`LIKE ? "+
						"UNION "+
						"SELECT DISTINCT `city_id`, `region_id` FROM `"+pref+"kladr` "+
						"WHERE `socr` IN (?"+strings.Repeat(",?", len(rule.newIn)-1)+") AND "+
						"`city_id` IS NOT NULL AND `name` LIKE ? LIMIT 100",
						args...)

					if err == nil {
						defer func() { _ = rows.Close() }()

						for rows.Next() {
							x := uint64(0)
							y := uint64(0)
							if rows.Scan(&x, &y) == nil {
								facts = append(facts, Fact{"city_id", fmt.Sprint(x)})
							}
						}
					}

					if len(facts) > 0 {
						answer.RegisterFacts(facts)
					} else {
						str4rekursive += word + " "
					}
				} //foreach ($arr_words as $word)
				if str4rekursive != "" { // доопределяем значимые слова
					if answer, err = findByAddress(AssArray{"f": str4rekursive}, store, answer); err != nil {
						return nil, err
					}
				}
			}
		} //if ($lStr !== "")
	}
	//////////////////////////////////////////////////////////////////

	if address_free != "" { //работаем с голой строкой
		address_free = SyntaxClear(address_free)           // нормируем знаки препинания
		lStr := ClearMarkers(address_free, "city", "@#c ") // Знака @ быть не может - о выщещен пред. функцией	)
		lStr = ClearMarkers(lStr, "region", "@#r ")
		lStr = ClearMarkers(lStr, "district", "@#d ")
		lStr = ClearMarkers(lStr, "settlement", "@#s ")
		lStr = strings.ReplaceAll(lStr, ",", "@")
		blocks := strings.Split(lStr, "@")

		if len(blocks) < 2 { // нет маркеров
			arr_words := Str2ImportantWords(address_free)
			for _, word := range arr_words {
				word = clearEnds(word) // морфология
				switch word {
				case "россия":
					continue
				case "москва":
					answer.RegisterFact(Fact{"city_id", "4400"})
					continue
				case "башкортостан":
					answer.RegisterFact(Fact{"region_id", "3296"})
					continue
				case "татарстан":
					answer.RegisterFact(Fact{"region_id", "5246"})
					continue
				case "ялта":
					answer.RegisterFact(Fact{"city_id", "200000007"})
					continue
				}

				kladrRep := NewKladr(store)
				model, err := kladrRep.FindOne(AssArray{
					"Is_Capital": "1",
					"name":       word,
				})

				if err != nil { // названия регионов имеют преимущество при немаркированном вводе
					if model.City_id.Valid && model.City_id.Int64 > 0 {
						answer.RegisterFact(Fact{"city_id", fmt.Sprint(model.City_id.Int64)})
						continue
					}
				}
				//////////////////////////

				name := "%" + word + "%"
				rows, err := lDb.Query("SELECT DISTINCT  1 as id, 0 as code_kladr, "+
					"`city_id` as city_id, 0 as region_id, 0 as country_id  "+
					"FROM `"+pref+"city` WHERE `name` LIKE ? "+
					"UNION "+
					"SELECT DISTINCT  2 as id, 0 as code_kladr, 0 as city_id, "+
					"`region_id`, 0 as country_id  FROM `"+pref+"region` "+
					"WHERE `name` LIKE ? "+
					"UNION "+
					"SELECT DISTINCT "+
					" 3 as id, 0 as code_kladr, 0 city_id, 0 as region_id, `country_id` "+
					"FROM `"+pref+"country` "+
					"WHERE `name`LIKE ? "+
					"UNION "+
					"SELECT DISTINCT  "+
					"4 as id,`code_kladr`,`city_id`, 0,0 "+
					"FROM `"+pref+"kladr` "+
					"WHERE `name` LIKE ? LIMIT 100", name, name, name, name)

				facts := make([]Fact, 0, 100)

				if err == nil {
					defer func() { _ = rows.Close() }()

					for rows.Next() {
						id := uint64(0)
						codekladr := ""
						city_id := ""
						region_id := ""
						x := uint64(0)
						err := rows.Scan(&id, &codekladr, &city_id, &region_id, &x)
						if err == nil {
							switch id {
							case 1:
								facts = append(facts, Fact{"city_id", city_id})
							case 2:
								facts = append(facts, Fact{"region_id", region_id})
							//case 3:
							//	facts = append(facts, Fact( "country_id", reg.country_id)

							case 4:
								facts = append(facts, Fact{"code_kladr", codekladr})
							}
						}

					}
				}
				if len(facts) > 0 {
					answer.RegisterFacts(facts)
				}
			} //for _, word := range arr_words {

		} else { // есть маркеры

			lStruct := make(map[int]AssArray, 100)
			lStructWords := make(map[int][]string, 100)

			for i := 0; i < len(blocks); i++ {

				l := make(AssArray, 3)

				if len(blocks[i]) > 1 && blocks[i][0:1] == "#" {
					l["ltype"] = blocks[i][1:2]
					l["text"] = strings.Trim(blocks[i][2:], " \t\n\r")
				} else {
					l["text"] = strings.Trim(blocks[i], " \t\n\r")
				}
				lStructWords[i] = Str2ImportantWords(l["text"])
				lStruct[i] = l
			}

			// поставим соседнему символу маркер гео. Пустые строки принимают, также как разделители
			for i := 0; i < len(lStruct)-1; i++ {
				if v, ok := lStruct[i+1]["ltype"]; ok {
					lStruct[i]["rtype"] = v
				}
			}

			for i, v := range lStruct {
				if len(lStructWords[i]) == 0 { // нет слова между двумя маркерами
					//	unset(struct [i])
					delete(lStructWords, i)
					delete(lStruct, i)
					continue
				}

				vl, okl := v["ltype"]
				vr, okr := v["rtype"]

				if len(lStructWords[i]) > 3 || !(okl || okr) { // маркеры не верны
					if answer, err = findByAddress(AssArray{"f": v["text"]}, store, answer); err != nil {
						return nil, err
					}
					continue
				}
				facts := make(AssArray, 8)
				if okl {
					if okr && vr != vl {
						switch vr {
						case "c":
							facts["city"] = v["text"]
						case "r":
							facts["region"] = v["text"]
						default:
							facts["0"] = v["text"]
						}
					}

					switch vl {
					case "c":
						facts["city"] = v["text"]
					case "r":
						facts["region"] = v["text"]
					default:
						facts["0"] = v["text"]
					}

				} else {
					if okr {
						switch vr {
						case "c":
							facts["city"] = v["text"]
						case "r":
							facts["region"] = v["text"]
						default:
							facts["0"] = v["text"]
						}
					}
				}
				if answer, err = findByAddress(facts, store, answer); err != nil {
					return nil, err
				}
			} // есть маркеры
		}
	}
	return answer, nil
}

// поиск по имени. Возвращает id или ошибку
func GeoLocation(town string, store IStore) (model.Geo, error) {

	var answer *Trust
	var err error

	// набранная пользователем строка
	if strings.Trim(town, " \t\n\r") != "" {
		if answer, err = findByAddress(AssArray{"a": town}, store); err != nil { // добавляем уверенности
			//fmt.Println("Пустой проход FindByAddress")
			return model.Geo{
				0,
				0,
				fmt.Sprint(town+" Ошибка findByAddress ", err),
			}, err
		}
	}
	repRegion := NewRegion(store)

	k := answer.trust_one()
	if k.ves == 0 {
		return model.Geo{}, err
	}
	kladrModel := k.model

	if kladrModel.City_id.Valid && kladrModel.City_id.Int64 > 0 {
		//fmt.Println("Ok by City_id")

		inf, err := repRegion.GetInfo(uint64(kladrModel.Region_id))
		name := "не найдено " + fmt.Sprint(err)
		if err == nil {
			name = inf.Name
		}
		return model.Geo{
			uint64(kladrModel.City_id.Int64),
			uint64(kladrModel.Region_id),
			name,
		}, nil

	}

	if kladrModel.Region_id > 0 {

		if tmp, err := repRegion.GetInfo(kladrModel.Region_id); err == nil {
			//	fmt.Println("Last pass Fake city from Region_id")
			return model.Geo{
				uint64(0),
				uint64(kladrModel.Region_id),
				tmp.Name,
			}, nil
		} else {
			return model.Geo{
				0,
				0,
				fmt.Sprint(town+" Ошибка repRegion.GetInfo ", err),
			}, err
		}

	}
	return model.Geo{
		0,
		0,
		fmt.Sprint(town+" Ошибка kladrModel ", kladrModel),
	}, err

}
