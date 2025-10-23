package game

import (
	"testing"

	"github.com/shpaker/gonflict/internal/types"
)

func TestNewBlocksRepository(t *testing.T) {
	repo := NewBlocksRepository()

	if repo == nil {
		t.Fatal("NewBlocksRepository вернул nil")
	}

	blocks := repo.GetAllBlocks()
	if len(*blocks) != 0 {
		t.Errorf("Ожидалось 0 блоков, получено %d", len(*blocks))
	}
}

func TestAddAndGetBlocks(t *testing.T) {
	repo := NewBlocksRepository()

	// Создаем тестовый блок
	block := types.BlockEntity{
		ImageGetter: nil, // Для тестов не используем ImageGetter
		Data: &types.BlockData{
			Name:     types.Brick,
			Position: types.Position{X: 0, Y: 0},
		},
		Properties: &types.BlockProperties{
			Collidable: true,
		},
		WorldPosition: types.Position{X: 100, Y: 100},
	}

	repo.AddBlock(block)

	blocks := repo.GetAllBlocks()
	if len(*blocks) != 1 {
		t.Errorf("Ожидалось 1 блок в списке, получено %d", len(*blocks))
	}
}

func TestRemoveBlockByPointer(t *testing.T) {
	repo := NewBlocksRepository()

	// Создаем тестовый блок
	block := types.BlockEntity{
		ImageGetter: nil, // Для тестов не используем ImageGetter
		Data: &types.BlockData{
			Name:     types.Brick,
			Position: types.Position{X: 0, Y: 0},
		},
		Properties: &types.BlockProperties{
			Collidable: true,
		},
		WorldPosition: types.Position{X: 100, Y: 100},
	}

	repo.AddBlock(block)

	// Получаем указатель на блок в репозитории
	blocks := repo.GetAllBlocks()
	if len(*blocks) == 0 {
		t.Fatal("Блок не был добавлен в репозиторий")
	}

	blockPtr := &(*blocks)[0]

	// Удаляем блок по указателю
	err := repo.RemoveBlockByPointer(blockPtr)
	if err != nil {
		t.Errorf("Неожиданная ошибка при удалении по указателю: %v", err)
	}

	blocks = repo.GetAllBlocks()
	if len(*blocks) != 0 {
		t.Errorf("Ожидалось 0 блоков после удаления, получено %d", len(*blocks))
	}

	// Тест удаления несуществующего блока
	err = repo.RemoveBlockByPointer(blockPtr)
	if err == nil {
		t.Error("Ожидалась ошибка для несуществующего блока")
	}

	// Тест удаления nil указателя
	err = repo.RemoveBlockByPointer(nil)
	if err == nil {
		t.Error("Ожидалась ошибка для nil указателя")
	}
}
