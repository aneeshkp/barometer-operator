package selectors

import (
	"k8s.io/apimachinery/pkg/labels"
)

//LabelApplicationKey ..
const (
	LabelApplicationKey = "application"
	LabelAppKey         = "app"

	LabelResourceKey = "barometer_cr"
	LabelAppValue    = "barometer"
)

//LabelsForCollectd ... Set labels in a map
func LabelsForCollectd(name string) map[string]string {
	return map[string]string{
		LabelAppKey:         name,
		LabelResourceKey:    name,
		LabelApplicationKey: LabelAppValue,
	}
}

//ResourcesByCollectdName return a selector that matches resources for a Collectd resource
func ResourcesByCollectdName(name string) labels.Selector {
	set := map[string]string{
		LabelAppKey:      name,
		LabelResourceKey: name,
	}
	return labels.SelectorFromSet(set)
}

//ResourcesByApplicationKey ...
func ResourcesByApplicationKey() labels.Selector {
	set := map[string]string{
		LabelAppKey: LabelAppValue,
	}
	return labels.SelectorFromSet(set)
}
