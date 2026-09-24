package domain

import (
	"strings"
	"testing"
)

func TestCategoryTable(t *testing.T) {
	if !CategoryCleaningYard.PhotoRequired() || CategoryWaterHeat.PhotoRequired() {
		t.Fatal("photo policy")
	}
	if CategoryElevator.Severity() != SeverityCritical || CategoryCityTerritory.Severity() != SeverityWarning {
		t.Fatal("severity table")
	}
	if CategoryWaterHeat.Title() != "Водоснабжение и отопление" {
		t.Fatal(CategoryWaterHeat.Title())
	}
	if Category("water_heat").Valid() || Authority("uk").Valid() {
		t.Fatal("case-sensitive enums")
	}
}

func TestCategoryAuthority(t *testing.T) {
	if CategoryCityTerritory.Authority() != AuthorityMunicipality {
		t.Fatal("city territory belongs to municipality")
	}
	for _, c := range []Category{CategoryWaterHeat, CategoryElectricity, CategoryElevator, CategoryCleaningYard, CategoryBuildingStructure} {
		if c.Authority() != AuthorityUK {
			t.Fatalf("%s must route to UK", c)
		}
	}
}

func TestReasoning(t *testing.T) {
	got := AuthorityMunicipality.Reasoning()
	if !strings.HasPrefix(got, "За эту проблему отвечает муниципалитет: ") {
		t.Fatal(got)
	}
}

func TestValidDescription(t *testing.T) {
	if _, ok := ValidDescription("          x         "); ok {
		t.Fatal("trim before length")
	}
	if d, ok := ValidDescription("  Нет воды уже день  "); !ok || d != "Нет воды уже день" {
		t.Fatal(d)
	}
	if _, ok := ValidDescription(strings.Repeat("я", 2001)); ok {
		t.Fatal("max 2000 runes")
	}
	if _, ok := ValidDescription(strings.Repeat("я", 2000)); !ok {
		t.Fatal("2000 runes ok")
	}
}
