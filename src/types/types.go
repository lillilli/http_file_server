package types

// Lifecycle - интерфейс структуры с жизененным циклом
type Lifecycle interface {
	Start() error
	Stop() error
}

// Error - структура сервисной ошибки
type Error struct {
	Message string `json:"errors"`
}
