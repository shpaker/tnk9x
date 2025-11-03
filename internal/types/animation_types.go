package types

// AnimationDataFrame представляет кадр анимации
type AnimationDataFrame struct {
	Image    string `yaml:"image"`
	Duration int    `yaml:"duration"`
}

// AnimationConfig представляет конфигурацию анимации в новом формате
type AnimationConfig struct {
	Offset   [2]float64 `yaml:"offset"`   // Смещение анимации относительно сущности [x, y]
	Duration int        `yaml:"duration"` // Длительность кадра в тиках
	Frames   []string   `yaml:"frames"`   // Список кадров анимации
	Repeats  *int       `yaml:"repeats"`  // Количество повторений (nil = бесконечно, число = проиграть N раз)
}

// AnimationData представляет анимацию (старый формат - массив кадров)
type AnimationData []AnimationDataFrame

// TilesetDataConfig представляет конфигурацию тайлсета
type TilesetDataConfig struct {
	Size       int                        `yaml:"size"`
	Images     map[string][2]int          `yaml:"images"`
	Animations map[string]AnimationConfig `yaml:"animations"`
}
