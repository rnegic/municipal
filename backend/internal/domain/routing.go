package domain

import (
	"strings"
	"unicode/utf8"
)

type Category string

const (
	CategoryWaterHeat         Category = "WATER_HEAT"
	CategoryElectricity       Category = "ELECTRICITY"
	CategoryElevator          Category = "ELEVATOR"
	CategoryCleaningYard      Category = "CLEANING_YARD"
	CategoryBuildingStructure Category = "BUILDING_STRUCTURE"
	CategoryCityTerritory     Category = "CITY_TERRITORY"
)

var categories = map[Category]struct {
	title     string
	severity  Severity
	photo     bool
	authority Authority
}{
	CategoryWaterHeat:         {"Водоснабжение и отопление", SeverityCritical, false, AuthorityUK},
	CategoryElectricity:       {"Электричество", SeverityCritical, false, AuthorityUK},
	CategoryElevator:          {"Лифт", SeverityCritical, false, AuthorityUK},
	CategoryCleaningYard:      {"Уборка, двор, подъезд", SeverityWarning, true, AuthorityUK},
	CategoryBuildingStructure: {"Конструктив здания", SeverityWarning, true, AuthorityUK},
	CategoryCityTerritory:     {"Городская территория", SeverityWarning, true, AuthorityMunicipality},
}

func (c Category) Valid() bool          { _, ok := categories[c]; return ok }
func (c Category) Title() string        { return categories[c].title }
func (c Category) Severity() Severity   { return categories[c].severity }
func (c Category) PhotoRequired() bool  { return categories[c].photo }
func (c Category) Authority() Authority { return categories[c].authority }

type Authority string

const (
	AuthorityUK           Authority = "UK"
	AuthorityFKR          Authority = "FKR"
	AuthorityRSO          Authority = "RSO"
	AuthorityMunicipality Authority = "MUNICIPALITY"
	AuthorityOwner        Authority = "OWNER"
)

var authorities = map[Authority]struct{ name, scope string }{
	AuthorityUK:           {"управляющая компания", "подъезд, двор, лифт, стояки до вентиля"},
	AuthorityFKR:          {"Фонд капитального ремонта", "протекающая крыша, трещины в несущих стенах, фасад старого дома"},
	AuthorityRSO:          {"ресурсоснабжающая организация", "отключения воды, тепла и света на уровне района"},
	AuthorityMunicipality: {"муниципалитет", "дороги за пределами двора, открытые люки на проезжей части"},
	AuthorityOwner:        {"собственник квартиры", "трубы внутри квартиры после счётчика или вентиля"},
}

func (a Authority) Valid() bool  { _, ok := authorities[a]; return ok }
func (a Authority) Name() string { return authorities[a].name }
func (a Authority) Reasoning() string {
	return "За эту проблему отвечает " + authorities[a].name + ": " + authorities[a].scope + "."
}

type Prediction struct {
	Category Category
	P        float64
}

func ValidDescription(s string) (string, bool) {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	return s, n >= 10 && n <= 2000
}

// ValidTitle нормализует краткое название заявки. Оно используется только для
// отображения и не участвует в автоопределении категории.
func ValidTitle(s string) (string, bool) {
	s = strings.TrimSpace(s)
	n := utf8.RuneCountInString(s)
	return s, n >= 3 && n <= 120
}
