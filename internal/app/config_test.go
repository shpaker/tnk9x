package app

import (
	"bytes"
	"encoding/csv"
	"os"
	"slices"
	"strconv"
	"testing"
	"unicode/utf8"

	"github.com/shpaker/tnk9x/internal/types"
)

// yandexPurchasesFile — товары для загрузки в консоль Яндекс Игр;
// путь от корня репозитория (TestMain переходит туда)
const yandexPurchasesFile = "web/yandex/console/purchases.csv"

// yandexDescriptionLimit — длина описания товара в консоли
const yandexDescriptionLimit = 200

// Языки конфигурации: без списка — английский, язык по умолчанию
// и замены — только из списка
func TestParseLanguageConfig(t *testing.T) {
	config, err := parseLanguageConfig(appConfigSchema{})
	if err != nil || config.Default != "en" || !config.IsSupported("en") {
		t.Errorf("empty config: %+v, %v", config, err)
	}

	config, err = parseLanguageConfig(appConfigSchema{
		Languages:         []string{"en", "ru"},
		LanguageFallbacks: map[string]string{"kk": "ru"},
	})
	if err != nil || config.Default != "en" ||
		config.Fallbacks["kk"] != types.Language("ru") {
		t.Errorf("config: %+v, %v", config, err)
	}

	invalid := []appConfigSchema{
		{Languages: []string{"en"}, DefaultLanguage: "ru"},
		{
			Languages:         []string{"en"},
			LanguageFallbacks: map[string]string{"kk": "ru"},
		},
	}
	for _, schema := range invalid {
		if _, err := parseLanguageConfig(schema); err == nil {
			t.Errorf("%+v must be refused", schema)
		}
	}
}

// Файл локали есть у каждого языка конфигурации игры
func TestLoadConfig_Languages(t *testing.T) {
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	for _, language := range cfg.Languages.Languages {
		if _, err := assetsFS().Open("locales/" + string(language) + ".yml"); err != nil {
			t.Errorf("no locale for %q: %v", language, err)
		}
	}
}

// Товары магазина: вид известен, у пачки номер, у жетонов
// количество, ID уникальны
func TestParseProducts(t *testing.T) {
	products, err := parseProducts(shopConfigSchema{
		Products: []productConfigSchema{
			{ID: "no_ads", Kind: "no_ads"},
			{ID: "pack_2", Kind: "pack", Pack: 2},
			{ID: "tokens_5", Kind: "tokens", Amount: 5},
		},
	})
	if err != nil || len(products) != 3 ||
		products[1].Kind != types.ProductKindPack || products[2].Amount != 5 {
		t.Errorf("products %+v, %v", products, err)
	}

	invalid := [][]productConfigSchema{
		{{Kind: "no_ads"}},
		{{ID: "a", Kind: "skin"}},
		{{ID: "a", Kind: "pack"}},
		{{ID: "a", Kind: "tokens"}},
		{{ID: "a", Kind: "no_ads"}, {ID: "a", Kind: "all_levels"}},
	}
	for _, schema := range invalid {
		if _, err := parseProducts(shopConfigSchema{Products: schema}); err == nil {
			t.Errorf("%+v must be refused", schema)
		}
	}
}

// Товары config.yml разобраны, их пачки есть в кампании
func TestLoadConfig_Products(t *testing.T) {
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Products) == 0 {
		t.Fatal("config.yml must list shop products")
	}

	campaign := types.NewCampaignEntity("TEST", []types.PackSpec{{Name: "ONE"}})
	ok := []types.ProductSpec{{ID: "p", Kind: types.ProductKindPack, Pack: 1}}
	if err := validateProductPacks(ok, campaign); err != nil {
		t.Errorf("valid pack refused: %v", err)
	}
	missing := []types.ProductSpec{
		{ID: "p", Kind: types.ProductKindPack, Pack: 2},
	}
	if err := validateProductPacks(missing, campaign); err == nil {
		t.Error("missing pack must be refused")
	}
}

// Товары консоли Яндекса совпадают с товарами config.yml: ID те же
// и в том же порядке, у каждого названия, описания в пределах лимита
// и цена — целое положительное число
func TestYandexPurchasesFile(t *testing.T) {
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	data, err := os.ReadFile(yandexPurchasesFile)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil || len(records) == 0 {
		t.Fatalf("read: %v", err)
	}

	header := []string{
		"id", "price", "title_en", "title_ru",
		"description_en", "description_ru",
	}
	if !slices.Equal(records[0], header) {
		t.Fatalf("header %v, want %v", records[0], header)
	}

	var ids []string
	for _, record := range records[1:] {
		ids = append(ids, record[0])
		if price, err := strconv.Atoi(record[1]); err != nil || price <= 0 {
			t.Errorf("%s: price %q", record[0], record[1])
		}
		if record[2] == "" || record[3] == "" {
			t.Errorf("%s: title required", record[0])
		}
		for _, description := range record[4:] {
			if utf8.RuneCountInString(description) > yandexDescriptionLimit {
				t.Errorf("%s: description too long", record[0])
			}
		}
	}

	var want []string
	for _, product := range cfg.Products {
		want = append(want, product.ID)
	}
	if !slices.Equal(ids, want) {
		t.Errorf("purchases.csv ids %v, config.yml %v", ids, want)
	}
}
