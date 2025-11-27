package dict

import (
	"testing"

	"github.com/Sufir/go-set-me-up/pkg"
	"github.com/Sufir/go-set-me-up/pkg/source/testcommon"
)

func buildDictMapFromInput(input []testcommon.DataEntry) map[string]any {
	root := map[string]any{}
	for _, entry := range input {
		current := root
		for i := 0; i < len(entry.Path)-1; i++ {
			name := entry.Path[i]
			next, ok := current[name]
			if !ok {
				next = map[string]any{}
				current[name] = next
			}
			current = next.(map[string]any)
		}
		leaf := entry.Path[len(entry.Path)-1]
		current[leaf] = entry.Value
	}
	return root
}

func executeDictScenario(t *testing.T, configuration any, mode pkg.LoadMode, input []testcommon.DataEntry) error {
	inputMap := buildDictMapFromInput(input)
	inputMap["UNUSED"] = "42"
	source := NewDictSource(inputMap)
	return source.Load(configuration, mode)
}

func TestDictSource_Common_Scenarios(t *testing.T) {
	scenarioGroups := [][]testcommon.Scenario{
		testcommon.BuildBasicPrimitivesScenarios(),
		testcommon.BuildPointerLeafScenarios(),
		testcommon.BuildBytesScenarios(),
		testcommon.BuildNestedValueScenarios(),
		testcommon.BuildNestedPointerScenarios(),
		testcommon.BuildModeScenarios(),
		testcommon.BuildAggregatedErrorScenarios(),
		testcommon.BuildUnknownKeysScenarios(),
		testcommon.BuildTextUnmarshalerScenarios(),
	}
	for _, group := range scenarioGroups {
		for _, scenario := range group {
			t.Run(scenario.Name, func(t *testing.T) {
				testcommon.RunScenario(t, scenario, executeDictScenario)
			})
		}
	}
}
