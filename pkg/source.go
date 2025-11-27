package pkg

type Source interface {
	Load(target any, mode LoadMode) error
}
