package controller

import (
	"github.com/aneeshkp/barometer-operator/pkg/controller/collectd"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, collectd.Add)
}
