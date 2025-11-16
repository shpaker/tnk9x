package types

type AnimationDataFrame struct {
	Image    string `yaml:"image"`
	Duration int    `yaml:"duration"`
}

type AnimationConfig struct {
	Offset   [2]float64 `yaml:"offset"`   // Смещение анимации относительно сущности [x, y]
	Duration int        `yaml:"duration"` // Длительность кадра в тиках
	Frames   []string   `yaml:"frames"`   // Список кадров анимации
	Repeats  *int       `yaml:"repeats"`  // Количество повторений (nil = бесконечно, число = проиграть N раз)
}

type AnimationData []AnimationDataFrame

type TilesetDataConfig struct {
	Size       int                        `yaml:"size"`
	Images     map[string][2]int          `yaml:"images"`
	Animations map[string]AnimationConfig `yaml:"animations"`
}
