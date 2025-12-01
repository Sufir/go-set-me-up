package pkg

import (
	"errors"
	"fmt"
)

type LoadMode int

const (
	ModeOverride LoadMode = iota + 1
	ModeFillMissing
)

type Source interface {
	Load(target any) error
}

type Loader struct {
	sources []Source
}

func NewLoader(sources ...Source) *Loader {
	return &Loader{sources: sources}
}

func (l *Loader) Load(cfg any) error {
	var collectedErrors []error
	for index, source := range l.sources {
		if loadError := source.Load(cfg); loadError != nil {
			wrappedError := NewLoaderSourceFailedError(index, fmt.Sprintf("%T", source), loadError)
			collectedErrors = append(collectedErrors, wrappedError)
		}
	}

	if len(collectedErrors) > 0 {
		aggregatedError := errors.Join(collectedErrors...)
		return NewAggregatedLoadFailedError(aggregatedError)
	}

	return nil
}
