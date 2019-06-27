package test

//https://github.com/operator-framework/operator-sdk/blob/master/doc/user/unit-testing.md
import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/RHsyseng/operator-utils/pkg/validation"
	collectdv1alpha1 "github.com/aneeshkp/barometer-operator/pkg/apis/collectd/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

var crdTypeMap = map[string]interface{}{
	"collectd_v1alpha1_collectd_crd.yaml": &collectdv1alpha1.Collectd{},
}

func TestCRDSchemas(t *testing.T) {
	for crdFileName, collectdType := range crdTypeMap {
		schema := getSchema(t, crdFileName)
		missingEntries := schema.GetMissingEntries(collectdType)
		for _, missing := range missingEntries {
			if strings.HasPrefix(missing.Path, "/status/conditions/transitionTime/") {
				//skill detailed properties of transition Time.
			} else {
				assert.Fail(t, "Discrepancy between CRD and Struct",
					"Missing or incorrect schema validation at %v, expected type %v  in CRD file %v", missing.Path, missing.Type, crdFileName)
			}
		}
	}
}

func TestSampleCustomResources(t *testing.T) {

	var crFileName, crdFileName string = "collectd_v1alpha1_collectd_cr.yaml", "collectd_v1alpha1_collectd_crd.yaml"
	assert.NotEmpty(t, crdFileName, "No matching CRD file found for CR suffixed: %s", crFileName)

	schema := getSchema(t, crdFileName)
	yamlString, err := ioutil.ReadFile("../deploy/crds/" + crFileName)
	assert.NoError(t, err, "Error reading %v CR yaml", crFileName)
	var input map[string]interface{}
	assert.NoError(t, yaml.Unmarshal([]byte(yamlString), &input))
	assert.NoError(t, schema.Validate(input), "File %v does not validate against the CRD schema", crFileName)
}

func getSchema(t *testing.T, crdFile string) validation.Schema {

	yamlString, err := ioutil.ReadFile("../deploy/crds/" + crdFile)
	assert.NoError(t, err, "Error reading CRD yaml %v", yamlString)

	schema, err := validation.New([]byte(yamlString))
	assert.NoError(t, err)

	return schema
}
