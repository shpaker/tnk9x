package models

type SpriteData struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
	W int `yaml:"w"`
	H int `yaml:"h"`
}

type SpritesConfig map[string]map[string]SpriteData
