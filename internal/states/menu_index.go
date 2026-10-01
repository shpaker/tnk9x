package states

// stepIndex сдвигает курсор меню на шаг вверх или вниз
// в пределах count пунктов
func stepIndex(index, count int, moveUp, moveDown bool) int {
	if moveUp && index > 0 {
		index--
	}
	if moveDown && index < count-1 {
		index++
	}
	return index
}
