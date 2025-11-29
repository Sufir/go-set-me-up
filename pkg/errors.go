package pkg

import (
	"errors"
	"fmt"
)

var (
	ErrLoaderSourceFailed   = errors.New("loader source failed")
	ErrLoadAggregatedFailed = errors.New("aggregated load failed")
	ErrInvalidTarget        = errors.New("invalid target")
	ErrSourceFieldFailed    = errors.New("source field failed")
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

type InvalidTargetError struct {
	Reason string
}

func NewInvalidTargetError(reason string) error {
	typedError := &InvalidTargetError{Reason: reason}
	return fmt.Errorf("%w: %w", ErrInvalidTarget, typedError)
}

func (invalidTargetError *InvalidTargetError) Error() string {
	return fmt.Sprintf("invalid target: %s", invalidTargetError.Reason)
}

type SourceFieldFailedError struct {
	OriginalError error
	SourceName    string
	Key           string
	Value         string
	Path          string
}

func NewEnvFieldFailedError(key string, value string, path string, originalError error) error {
	typedError := &SourceFieldFailedError{SourceName: "env", Key: key, Value: value, Path: path, OriginalError: originalError}
	return fmt.Errorf("%w: %w", ErrSourceFieldFailed, typedError)
}

func NewDictFieldFailedError(path string, originalError error) error {
	typedError := &SourceFieldFailedError{SourceName: "dict", Path: path, OriginalError: originalError}
	return fmt.Errorf("%w: %w", ErrSourceFieldFailed, typedError)
}

func NewFlagsFieldFailedError(name string, value string, path string, originalError error) error {
	typedError := &SourceFieldFailedError{SourceName: "flags", Key: name, Value: value, Path: path, OriginalError: originalError}
	return fmt.Errorf("%w: %w", ErrSourceFieldFailed, typedError)
}

func NewJSONFieldFailedError(path string, originalError error) error {
	typedError := &SourceFieldFailedError{SourceName: "json", Path: path, OriginalError: originalError}
	return fmt.Errorf("%w: %w", ErrSourceFieldFailed, typedError)
}

func (e *SourceFieldFailedError) Error() string {
	if e.SourceName == "env" {
		return fmt.Sprintf("env %s=%s field %s: %v", e.Key, e.Value, e.Path, e.OriginalError)
	}
	if e.SourceName == "flags" {
		if e.Path != "" {
			return fmt.Sprintf("flags %s=%s field %s: %v", e.Key, e.Value, e.Path, e.OriginalError)
		}
		return fmt.Sprintf("flags %s=%s: %v", e.Key, e.Value, e.OriginalError)
	}
	if e.SourceName == "dict" {
		return fmt.Sprintf("dict field %s: %v", e.Path, e.OriginalError)
	}
	return fmt.Sprintf("json field %s: %v", e.Path, e.OriginalError)
}

func (e *SourceFieldFailedError) Unwrap() error {
	return e.OriginalError
}
