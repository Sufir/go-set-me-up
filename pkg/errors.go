package pkg

import (
	"errors"
	"fmt"
)

var (
	ErrLoaderSourceFailed   = errors.New("loader source failed")
	ErrLoadAggregatedFailed = errors.New("aggregated load failed")
)

type LoaderSourceFailedError struct {
	OriginalError error
	SourceName    string
	SourceIndex   int
}

func NewLoaderSourceFailedError(sourceIndex int, sourceName string, originalError error) error {
	typedError := &LoaderSourceFailedError{SourceIndex: sourceIndex, SourceName: sourceName, OriginalError: originalError}
	return fmt.Errorf("%w: %w", ErrLoaderSourceFailed, typedError)
}

func (loaderSourceFailedError *LoaderSourceFailedError) Error() string {
	return fmt.Sprintf("source at index %d named %s failed: %v", loaderSourceFailedError.SourceIndex, loaderSourceFailedError.SourceName, loaderSourceFailedError.OriginalError)
}

func (loaderSourceFailedError *LoaderSourceFailedError) Unwrap() error {
	return loaderSourceFailedError.OriginalError
}

type AggregatedLoadFailedError struct {
	Aggregated error
}

func NewAggregatedLoadFailedError(aggregated error) error {
	typedError := &AggregatedLoadFailedError{Aggregated: aggregated}
	return fmt.Errorf("%w: %w", ErrLoadAggregatedFailed, typedError)
}

func (aggregatedLoadFailedError *AggregatedLoadFailedError) Error() string {
	return fmt.Sprintf("aggregated load failed: %v", aggregatedLoadFailedError.Aggregated)
}

func (aggregatedLoadFailedError *AggregatedLoadFailedError) Unwrap() error {
	return aggregatedLoadFailedError.Aggregated
}
