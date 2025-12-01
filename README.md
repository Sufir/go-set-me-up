# go-set-me-up
[![Lint](https://github.com/Sufir/go-set-me-up/actions/workflows/lint.yml/badge.svg)](https://github.com/Sufir/go-set-me-up/actions/workflows/lint.yml)  [![Coverage main](https://img.shields.io/endpoint?url=https://sufir.github.io/go-set-me-up/main/coverage-badge.json)](https://sufir.github.io/go-set-me-up/main/)

## Overview
GoSetUpMe is a small Go library that populates configuration structs from multiple sources in a consistent way. It supports environment variables, CLI flags, in-memory dictionaries, and JSON files. The library includes a type-casting engine that converts string inputs into native Go types and provides two loading modes: override and fill-missing.

## Core Features
- Unified loader that chains multiple sources
- Sources: environment, flags, dictionary, JSON file
- Type casting for primitives, complex numbers, byte slices/arrays, and `encoding.TextUnmarshaler`
- Modes: `Override` (always set) and `FillMissing` (only zero values)
- Clear aggregated error reporting

## Supported Tags

| Tag | Sources | Purpose | Types | Default | Example |
| --- | --- | --- | --- | --- | --- |
| `env` | `env` | Environment variable name for a field; `"-"` disables the field | Any supported types | None | `Port int \`env:"PORT"\`` |
| `envSegment` | `env` | Segment name for nested structs, used in key construction | Structs and pointers to structs | Field name | `Outer.Inner.Value` with `env:"VALUE"` and `envSegment:"outer"` → key `APP_OUTER_VALUE` |
| `envDefault` | `env` | Fallback string used when the variable is missing | Leaf fields | None | `A int \`env:"A" envDefault:"10"\`` |
| `envDelim` | `env` | Delimiter for `[]string`, `[]int`, `[N]int` inputs | Slices and int arrays | Source delimiter (`,` by default) | `B []int \`env:"B" envDelim:":"\`` |
| `flag` | `flags` | Long flag name; `"-"` disables the field | Leaf fields | None | `Port int \`flag:"port"\`` |
| `flagShort` | `flags` | Short flag alias | Leaf fields | None | `Port int \`flag:"port" flagShort:"p"\`` |
| `flagDefault` | `flags` | Fallback string used when the flag is absent | Leaf fields | None | `A int \`flag:"a" flagDefault:"10"\`` |
| `flagDelim` | `flags` | Delimiter for `[]string`, `[]int`, `[N]int` inputs | Slices and int arrays | `,` | `B []int \`flag:"b" flagDelim:":"\`` |
| `json` | `json-file` | JSON tag name; `"-"` disables the field; only the part before the comma is used | Any leaf fields | None | `Port int \`json:"Port,omitempty"\`` |

- `env` specifics: an empty environment value is treated as present and wins over `envDefault`. For numeric and boolean types this yields a parse error; for strings it sets an empty string (`pkg/source/env/env_source.go`:102, 155–193).
- `flags` specifics: a flag without a value for a boolean field is treated as `true`; for non-boolean types it is an empty-value error. Supported syntaxes include `--name=value`, `--name value`, `-n=value`, `-n value`, and `--no-name` to set `false` (`pkg/source/flags/flags_source.go`:45–108, 136–201).
- `json` specifics: standard `json` tags are used; only the name before the comma is matched, additional options like `omitempty` are ignored for name matching (`pkg/source/json-file/json_file_source.go`:54, 115–123).

## Sources

- `env` — environment variables. Construct via `env.NewSource(prefix, delimiter, mode)`. Prefix and segments are converted to upper snake-case; keys are built as `PREFIX_SEGMENT_LEAF`. Supports `env`, `envSegment`, `envDefault`, `envDelim`. Empty env values take precedence over defaults and may cause parse errors for non-string types (`pkg/source/env/env_source.go`:57–76, 124–131, 154–193).
- `flags` — command-line arguments. Construct via `flags.NewSource(mode)`. Supports long/short forms, auto-boolean flags, negation via `--no-name`, and values via `=` or the next argument. Tags: `flag`, `flagShort`, `flagDefault`, `flagDelim` (`pkg/source/flags/flags_source.go`:45–108, 136–201).
  - Delimiter configuration: use `flags.NewSourceWithDelimiter(mode, delimiter)` or `flags.NewSourceWithCasterAndDelimiter(mode, delimiter, caster)` to set a default delimiter at the source level; a per-field `flagDelim` tag overrides the source default.
- `dict` — `map[string]any` dictionary. Construct via `dict.NewSource(dict, mode)`. Keys may be the field name (`FieldName`), upper snake-case (`UPPER_SNAKE`), or lower snake-case (`lower_snake`). Nested structs are provided via nested maps. No tags used (`pkg/source/dict/dict_source.go`:83–95, 48–81).
- `json-file` — JSON file. Construct via `jsonfile.NewSource(path, mode)`. Values are matched by `json` tags. For `[]byte` a base64 string is expected; for `[N]byte` — an array of numbers. Pointers to structs are allocated automatically when needed (`pkg/source/json-file/json_file_source.go`:22–45, 54–90, 98–110; `pkg/source/json-file/json_file_source_test.go`:143–173).

## Quick Start

Environment source

```go
package main

import (
    "github.com/Sufir/go-set-me-up/pkg"
    "github.com/Sufir/go-set-me-up/pkg/source/env"
)

type ApplicationConfiguration struct {
    Name  string `env:"NAME"`
    Port  int    `env:"PORT"`
    Debug bool   `env:"DEBUG"`
}

func main() {
    environmentSource := env.NewSource("app", ",", pkg.ModeOverride)
    loader := pkg.NewLoader(environmentSource)
    configuration := &ApplicationConfiguration{}
    if err := loader.Load(configuration); err != nil {
        panic(err)
    }
}
```

## Multiple Sources
Load configuration from several sources in a single pass. Sources are applied left-to-right.

```go
package main

import (
    "github.com/Sufir/go-set-me-up/pkg"
    jsonfile "github.com/Sufir/go-set-me-up/pkg/source/json-file"
    "github.com/Sufir/go-set-me-up/pkg/source/dict"
    "github.com/Sufir/go-set-me-up/pkg/source/env"
    "github.com/Sufir/go-set-me-up/pkg/source/flags"
)

type ApplicationConfiguration struct {
    Name  string `json:"Name" env:"NAME" flag:"name"`
    Port  int    `json:"Port" env:"PORT" flag:"port"`
    Debug bool   `json:"Debug" env:"DEBUG" flag:"debug"`
}

func main() {
    loader := pkg.NewLoader(
        jsonfile.NewSource("config.json", pkg.ModeFillMissing),
        dict.NewSource(map[string]any{"Port": 8080}, pkg.ModeFillMissing),
        env.NewSource("app", ",", pkg.ModeOverride),
        flags.NewSource(pkg.ModeOverride),
    )

    configuration := &ApplicationConfiguration{}
    if err := loader.Load(configuration); err != nil {
        panic(err)
    }
}
```

Behavior and order
- Sources apply in the order passed to `NewLoader`; later sources see the results of earlier ones.
- `ModeOverride` sets values regardless of existing non-zero values.
- `ModeFillMissing` only sets zero-valued fields.
- When both modes are mixed, early sources can provide defaults (`FillMissing`), while later sources (`Override`) refine or replace.
- Errors from all sources are aggregated; messages include the source and the field path.

## Custom Type Option

Add your own string-to-type converter by implementing `pkg.TypeCasterOption`. The option declares which target types it supports and performs the conversion from string to a `reflect.Value`.

Example: add support for `time.Duration` values.

```go
package mycaster

import (
    "reflect"
    "strings"
    "time"

    "github.com/Sufir/go-set-me-up/pkg"
)

type DurationOption struct{}

func (DurationOption) Supports(targetType reflect.Type) bool {
    return targetType == reflect.TypeOf(time.Duration(0))
}

func (DurationOption) Cast(value string, targetType reflect.Type) (reflect.Value, error) {
    parsedDuration, parseError := time.ParseDuration(strings.TrimSpace(value))
    if parseError != nil {
        return reflect.Value{}, pkg.ErrParseFailed{Type: targetType, Value: value, Cause: parseError}
    }

    return reflect.ValueOf(parsedDuration), nil
}
```

Use the custom option with any source by constructing a caster via `pkg.NewTypeCaster` and passing it into source constructors.

```go
package main

import (
    "time"

    "github.com/Sufir/go-set-me-up/pkg"
    "github.com/Sufir/go-set-me-up/pkg/source/dict"
    "github.com/Sufir/go-set-me-up/pkg/source/env"
    "github.com/Sufir/go-set-me-up/pkg/source/flags"
    "github.com/your/module/mycaster"
)

type ApplicationConfiguration struct {
    Timeout time.Duration `env:"TIMEOUT" flag:"timeout"`
}

func main() {
    typeCaster := pkg.NewTypeCaster(mycaster.DurationOption{})

    environmentSource := env.NewSourceWithCaster("app", ",", pkg.ModeOverride, typeCaster)
    flagsSource := flags.NewSourceWithCaster(pkg.ModeOverride, typeCaster)
    dictionarySource := dict.NewSourceWithCaster(map[string]any{"Timeout": "1s"}, pkg.ModeOverride, typeCaster)

    loader := pkg.NewLoader(environmentSource, flagsSource, dictionarySource)

    configuration := &ApplicationConfiguration{}
    loadError := loader.Load(configuration)
    if loadError != nil {
        panic(loadError)
    }
}
```

## Custom Source (YAML)

This example shows a minimal custom source that reads a YAML file and applies values to a configuration struct. It treats YAML-provided values as present and respects the library load modes.

```go
package yamlfile

import (
    "os"
    "reflect"
    "gopkg.in/yaml.v3"
    "github.com/Sufir/go-set-me-up/pkg"
    "github.com/Sufir/go-set-me-up/pkg/source/sourceutil"
)

type Source struct {
    path string
    mode pkg.LoadMode
}

func NewSource(path string, mode pkg.LoadMode) *Source {
    return &Source{path: path, mode: sourceutil.DefaultMode(mode)}
}

func (source Source) Load(cfg any) error {
    elem, err := sourceutil.EnsureTargetStruct(cfg)
    if err != nil {
        return err
    }
    
    data, readErr := os.ReadFile(source.path)
    if readErr != nil {
        return pkg.NewAggregatedLoadFailedError(readErr)
    }

    holder := reflect.New(elem.Type())
    if err := yaml.Unmarshal(data, holder.Interface()); err != nil {
        return pkg.NewAggregatedLoadFailedError(err)
    }

    return nil
}
```
