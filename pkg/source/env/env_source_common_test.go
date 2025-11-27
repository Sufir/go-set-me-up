package env

import (
	"reflect"
	"testing"

	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/Sufir/go-set-me-up/pkg/source/testcommon"
)

func buildEnvKeyFromPath(prefix string, configuration any, path []string) string {
	segments := []string{}
	if prefix != "" {
		segments = append(segments, convertToEnvVar(prefix))
	}
	configurationValue := reflect.ValueOf(configuration)
	if configurationValue.Kind() == reflect.Ptr {
		configurationValue = configurationValue.Elem()
	}
	configurationType := configurationValue.Type()
	currentType := configurationType
	for i := 0; i < len(path)-1; i++ {
		fieldName := path[i]
		field, ok := currentType.FieldByName(fieldName)
		if !ok {
			segmentName := convertToEnvVar(fieldName)
			segments = append(segments, segmentName)
			continue
		}
		segmentCandidate := field.Tag.Get("envSegment")
		if segmentCandidate == "" {
			segmentCandidate = field.Name
		}
		segmentName := convertToEnvVar(segmentCandidate)
		segments = append(segments, segmentName)
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
		tagValue := leafField.Tag.Get("env")
		if tagValue != "" {
			leafCandidate = tagValue
		}
	}
	leaf := convertToEnvVar(leafCandidate)
	return buildKey(segments, leaf)
}

func executeEnvScenario(t *testing.T, configuration any, mode pkg.LoadMode, input []testcommon.DataEntry) error {
	prefix := "app"
	source := NewEnvSource(prefix, ",")
	for _, entry := range input {
		key := buildEnvKeyFromPath(prefix, configuration, entry.Path)
		t.Setenv(key, entry.Value)
	}
	t.Setenv(buildKey([]string{convertToEnvVar(prefix)}, "UNUSED"), "42")
	return source.Load(configuration, mode)
}

func TestEnvSource_Common_Scenarios(t *testing.T) {
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
		testcommon.BuildTextUnmarshalerScenarios(),
	}
	for _, group := range scenarioGroups {
		for _, scenario := range group {
			t.Run(scenario.Name, func(t *testing.T) {
				testcommon.RunScenario(t, scenario, executeEnvScenario)
			})
		}
	}
}
