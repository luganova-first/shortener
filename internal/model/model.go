package model

// Хранилище сокращений:
// Shorted -- ключ -- хеш сокращения, значение -- сокращаемый URL
// Full -- обратный к Shorted map, ключ -- сокращаемый URL, значение -- хеш сокращения
// Full требуется для быстрого поиска по значениям сокращаемых URL
type Storage struct {
	Shorted map[string]string
	Full    map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		Shorted: make(map[string]string),
		Full:    make(map[string]string),
	}
}

func (s *Storage) GetShort(key string) string {
	return s.Shorted[key]
}

func (s *Storage) GetFull(key string) string {
	return s.Full[key]
}
