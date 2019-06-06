package configs

import (
	v1alpha1 "github.com/aneeshkp/barometer-operator/pkg/apis/collectd/v1alpha1"
)

const (
	collectdConfigPath = "/opt/collectd/etc/collectd.conf"
)

//ConfigForCollectd  ....
func ConfigForCollectd(m *v1alpha1.Collectd) string {
	config := `
	kind: ConfigMap
    apiVersion: v1
    metadata:
      name: collectd-config
      namespace: default
    data:
    node-collectd.conf: |-
        FQDNLookup false
        LoadPlugin syslog
        <Plugin syslog>
        LogLevel info
        </Plugin>

        LoadPlugin cpu

        LoadPlugin memory
    
        <Plugin "cpu">
        Interval 5
        ReportByState false
        ReportByCpu false
        </Plugin>

        <Plugin "memory">
        Interval 30
        ValuesAbsolute false
        ValuesPercentage true
        </Plugin>

        LoadPlugin processes
        <Plugin " processes">
        Process "docker"
        # Add any other processes you wish to monitor...
        </Plugin>
    
        #Last line (collectd requires ‘\n’ at the last line)`
	return config

	//var buff bytes.Buffer
	//collectdconfig := template.Must(template.New("collectdconfig").Parse(config))
	//collectdconfig.Execute(&buff, m.Spec)
	//return buff.String()
}
