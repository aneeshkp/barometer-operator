package controller

import (
	"github.com/aneeshkp/collectd-operator/pkg/controller/collectd"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, collectd.Add)
}
