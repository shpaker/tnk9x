package adapters

// IInputAdapter интерфейс для адаптера ввода
type IInputAdapter interface {
	Update()
}

// IRendererAdapter интерфейс для адаптера рендеринга
type IRendererAdapter interface {
	DrawAll(screen interface{})
}
