package flags

import (
	"os"
	"reflect"
	"testing"

	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/Sufir/go-set-me-up/pkg/source/testcommon"
)

func executeFlagsScenario(t *testing.T, configuration any, mode pkg.LoadMode, input []testcommon.DataEntry) error {
	t.Helper()
	args := []string{"app"}
	for _, entry := range input {
		key := buildFlagNameFromPath(configuration, entry.Path)
		short := flagShortFromPath(configuration, entry.Path)
		prefix := "--"
		if short != "" {
			key = short
			prefix = "-"
		}
		args = append(args, prefix+key)
		if entry.Value != "" {
			args = append(args, entry.Value)
		}
	}
	old := osArgsSwap(args)
	defer osArgsSwap(old)
	source := NewSource(mode)
	return source.Load(configuration)
}

func buildFlagNameFromPath(configuration any, path []string) string {
	configurationValue := reflect.ValueOf(configuration)
	if configurationValue.Kind() == reflect.Ptr {
		configurationValue = configurationValue.Elem()
	}
	currentType := configurationValue.Type()
	for i := 0; i < len(path)-1; i++ {
		fieldName := path[i]
		field, ok := currentType.FieldByName(fieldName)
		if !ok {
			continue
		}
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		currentType = fieldType
	}
	leafFieldName := path[len(path)-1]
	leafField, ok := currentType.FieldByName(leafFieldName)
	leafCandidate := leafFieldName
	if ok {
		tagValue := leafField.Tag.Get("flag")
		if tagValue != "" {
			leafCandidate = tagValue
		}
	}
	return leafCandidate
}

func flagShortFromPath(configuration any, path []string) string {
	configurationValue := reflect.ValueOf(configuration)
	if configurationValue.Kind() == reflect.Ptr {
		configurationValue = configurationValue.Elem()
	}
	currentType := configurationValue.Type()
	for i := 0; i < len(path)-1; i++ {
		fieldName := path[i]
		field, ok := currentType.FieldByName(fieldName)
		if !ok {
			continue
		}
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		currentType = fieldType
	}
	leafFieldName := path[len(path)-1]
	leafField, ok := currentType.FieldByName(leafFieldName)
	if ok {
		tagShort := leafField.Tag.Get("flagShort")
		return tagShort
	}
	return ""
}

func TestFlagsSource_Common_Scenarios(t *testing.T) {
	scenarioGroups := [][]testcommon.Scenario{
		testcommon.BuildBasicPrimitivesScenarios(),
		testcommon.BuildPointerLeafScenarios(),
		testcommon.BuildBytesScenarios(),
		testcommon.BuildNestedValueScenarios(),
		testcommon.BuildNestedPointerScenarios(),
		testcommon.BuildModeScenarios(),
		testcommon.BuildAggregatedErrorScenarios(),
		testcommon.BuildUnknownKeysScenarios(),
		testcommon.BuildEmptyValuesScenarios(),
		testcommon.BuildInvalidPrimitiveCastScenarios(),
		testcommon.BuildTextUnmarshalerScenarios(),
	}
	for _, group := range scenarioGroups {
		for _, scenario := range group {
			t.Run(scenario.Name, func(t *testing.T) {
				testcommon.RunScenario(t, scenario, executeFlagsScenario)
			})
		}
	}
}

func osArgsSwap(newArgs []string) []string {
	old := os.Args
	os.Args = newArgs
	return old
}
