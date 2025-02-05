// Copyright 2021, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smartagentreceiver

import (
	"path"
	"testing"
	"time"

	"github.com/signalfx/signalfx-agent/pkg/core/common/httpclient"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/consul"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/hadoop"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/redis"
	"github.com/signalfx/signalfx-agent/pkg/monitors/filesystems"
	"github.com/signalfx/signalfx-agent/pkg/monitors/haproxy"
	"github.com/signalfx/signalfx-agent/pkg/monitors/prometheusexporter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/ntpq"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/config/configtest"
)

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Receivers), 5)

	haproxyCfg := cfg.Receivers["smartagent/haproxy"].(*Config)
	expectedDimensionClients := []string{"nop/one", "nop/two"}
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/haproxy",
		},
		DimensionClients: expectedDimensionClients,
		monitorConfig: &haproxy.Config{
			MonitorConfig: config.MonitorConfig{
				Type:                "haproxy",
				IntervalSeconds:     123,
				DatapointsToExclude: []config.MetricFilter{},
			},
			Username:  "SomeUser",
			Password:  "secret",
			Path:      "stats?stats;csv",
			SSLVerify: true,
			Timeout:   timeutil.Duration(5 * time.Second),
		},
	}, haproxyCfg)
	require.NoError(t, haproxyCfg.validate())

	redisCfg := cfg.Receivers["smartagent/redis"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/redis",
		},
		DimensionClients: []string{},
		monitorConfig: &redis.Config{
			MonitorConfig: config.MonitorConfig{
				Type:                "collectd/redis",
				IntervalSeconds:     234,
				DatapointsToExclude: []config.MetricFilter{},
			},
			Host: "localhost",
			Port: 6379,
		},
	}, redisCfg)
	require.NoError(t, redisCfg.validate())

	hadoopCfg := cfg.Receivers["smartagent/hadoop"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/hadoop",
		},
		monitorConfig: &hadoop.Config{
			MonitorConfig: config.MonitorConfig{
				Type:                "collectd/hadoop",
				IntervalSeconds:     345,
				DatapointsToExclude: []config.MetricFilter{},
			},
			CommonConfig: python.CommonConfig{},
			Host:         "localhost",
			Port:         8088,
		},
	}, hadoopCfg)
	require.NoError(t, hadoopCfg.validate())

	etcdCfg := cfg.Receivers["smartagent/etcd"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/etcd",
		},
		monitorConfig: &prometheusexporter.Config{
			MonitorConfig: config.MonitorConfig{
				Type:                "etcd",
				IntervalSeconds:     456,
				DatapointsToExclude: []config.MetricFilter{},
			},
			HTTPConfig: httpclient.HTTPConfig{
				HTTPTimeout: timeutil.Duration(10 * time.Second),
			},
			Host:       "localhost",
			Port:       5309,
			MetricPath: "/metrics",
		},
	}, etcdCfg)
	require.NoError(t, etcdCfg.validate())

	tr := true
	ntpqCfg := cfg.Receivers["smartagent/ntpq"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/ntpq",
		},
		monitorConfig: &ntpq.Config{
			MonitorConfig: config.MonitorConfig{
				Type:                "telegraf/ntpq",
				IntervalSeconds:     567,
				DatapointsToExclude: []config.MetricFilter{},
			},
			DNSLookup: &tr,
		},
	}, ntpqCfg)
	require.NoError(t, ntpqCfg.validate())
}

func TestLoadInvalidConfigWithoutType(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "without_type.yaml"), factories,
	)
	require.Error(t, err)
	require.EqualError(t, err,
		"error reading receivers configuration for smartagent/withouttype: you must specify a \"type\" for a smartagent receiver")
	require.Nil(t, cfg)
}

func TestLoadInvalidConfigWithUnknownType(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "unknown_type.yaml"), factories,
	)
	require.Error(t, err)
	require.EqualError(t, err,
		"error reading receivers configuration for smartagent/unknowntype: no known monitor type \"notamonitor\"")
	require.Nil(t, cfg)
}

func TestLoadInvalidConfigWithUnexpectedTag(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "unexpected_tag.yaml"), factories,
	)
	require.Error(t, err)
	require.EqualError(t, err,
		"error reading receivers configuration for smartagent/unexpectedtag: failed creating Smart Agent Monitor custom config: yaml: unmarshal errors:\n  line 2: field notasupportedtag not found in type redis.Config")
	require.Nil(t, cfg)
}

func TestLoadInvalidConfigs(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "invalid_config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, len(cfg.Receivers), 2)

	negativeIntervalCfg := cfg.Receivers["smartagent/negativeintervalseconds"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/negativeintervalseconds",
		},
		monitorConfig: &redis.Config{
			MonitorConfig: config.MonitorConfig{
				Type:                "collectd/redis",
				IntervalSeconds:     -234,
				DatapointsToExclude: []config.MetricFilter{},
			},
		},
	}, negativeIntervalCfg)
	err = negativeIntervalCfg.validate()
	require.Error(t, err)
	require.EqualError(t, err, "intervalSeconds must be greater than 0s (-234 provided)")

	missingRequiredCfg := cfg.Receivers["smartagent/missingrequired"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/missingrequired",
		},
		monitorConfig: &consul.Config{
			MonitorConfig: config.MonitorConfig{
				Type:                "collectd/consul",
				IntervalSeconds:     0,
				DatapointsToExclude: []config.MetricFilter{},
			},
			Port:          5309,
			TelemetryHost: "0.0.0.0",
			TelemetryPort: 8125,
		},
	}, missingRequiredCfg)
	err = missingRequiredCfg.validate()
	require.Error(t, err)
	require.EqualError(t, err, "Validation error in field 'Config.host': host is a required field (got '')")
}

func TestLoadConfigWithEndpoints(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "endpoints_config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Receivers), 4)

	haproxyCfg := cfg.Receivers["smartagent/haproxy"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/haproxy",
		},
		Endpoint: "haproxyhost:2345",
		monitorConfig: &haproxy.Config{
			MonitorConfig: config.MonitorConfig{
				Type:                "haproxy",
				IntervalSeconds:     123,
				DatapointsToExclude: []config.MetricFilter{},
			},
			Host:      "haproxyhost",
			Port:      2345,
			Username:  "SomeUser",
			Password:  "secret",
			Path:      "stats?stats;csv",
			SSLVerify: true,
			Timeout:   timeutil.Duration(5 * time.Second),
		},
	}, haproxyCfg)
	require.NoError(t, haproxyCfg.validate())

	redisCfg := cfg.Receivers["smartagent/redis"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/redis",
		},
		Endpoint: "redishost",
		monitorConfig: &redis.Config{
			MonitorConfig: config.MonitorConfig{
				Type:                "collectd/redis",
				IntervalSeconds:     234,
				DatapointsToExclude: []config.MetricFilter{},
			},
			Host: "redishost",
			Port: 6379,
		},
	}, redisCfg)
	require.NoError(t, redisCfg.validate())

	hadoopCfg := cfg.Receivers["smartagent/hadoop"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/hadoop",
		},
		Endpoint: "hadoophost:12345",
		monitorConfig: &hadoop.Config{
			MonitorConfig: config.MonitorConfig{
				Type:                "collectd/hadoop",
				IntervalSeconds:     345,
				DatapointsToExclude: []config.MetricFilter{},
			},
			CommonConfig: python.CommonConfig{},
			Host:         "localhost",
			Port:         8088,
		},
	}, hadoopCfg)
	require.NoError(t, hadoopCfg.validate())

	etcdCfg := cfg.Receivers["smartagent/etcd"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/etcd",
		},
		Endpoint: "etcdhost:5555",
		monitorConfig: &prometheusexporter.Config{
			MonitorConfig: config.MonitorConfig{
				Type:                "etcd",
				IntervalSeconds:     456,
				DatapointsToExclude: []config.MetricFilter{},
			},
			HTTPConfig: httpclient.HTTPConfig{
				HTTPTimeout: timeutil.Duration(10 * time.Second),
			},
			Host:       "localhost",
			Port:       5555,
			MetricPath: "/metrics",
		},
	}, etcdCfg)
	require.NoError(t, etcdCfg.validate())
}

func TestLoadInvalidConfigWithInvalidEndpoint(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "invalid_endpoint.yaml"), factories,
	)
	require.Error(t, err)
	require.EqualError(t, err,
		"error reading receivers configuration for smartagent/haproxy: cannot determine port via Endpoint: strconv.ParseUint: parsing \"notaport\": invalid syntax")
	require.Nil(t, cfg)
}

func TestLoadInvalidConfigWithUnsupportedEndpoint(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "unsupported_endpoint.yaml"), factories,
	)
	require.Error(t, err)
	require.EqualError(t, err,
		"error reading receivers configuration for smartagent/nagios: unable to set monitor Host field using Endpoint-derived value of localhost: no field Host of type string detected")
	require.Nil(t, cfg)
}

func TestLoadInvalidConfigWithNonArrayDimensionClients(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "invalid_nonarray_dimension_clients.yaml"), factories,
	)
	require.Error(t, err)
	require.EqualError(t, err,
		"error reading receivers configuration for smartagent/haproxy: dimensionClients must be an array of compatible exporter names")
	require.Nil(t, cfg)
}

func TestLoadInvalidConfigWithNonStringArrayDimensionClients(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "invalid_float_dimension_clients.yaml"), factories,
	)
	require.Error(t, err)
	require.EqualError(t, err,
		"error reading receivers configuration for smartagent/haproxy: dimensionClients must be an array of compatible exporter names")
	require.Nil(t, cfg)
}

func TestFilteringConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "filtering_config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	fsCfg := cfg.Receivers["smartagent/filesystems"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/filesystems",
		},
		monitorConfig: &filesystems.Config{
			MonitorConfig: config.MonitorConfig{
				Type: "filesystems",
				DatapointsToExclude: []config.MetricFilter{
					{
						MetricName: "df_inodes.*",
						Dimensions: map[string]interface{}{
							"mountpoint": []interface{}{"*", "!/hostfs/var/lib/cni"},
						},
					},
				},
				ExtraGroups:  []string{"inodes"},
				ExtraMetrics: []string{"percent_bytes.reserved"},
			},
		},
	}, fsCfg)
	require.NoError(t, fsCfg.validate())
}

func TestInvalidFilteringConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Receivers[configmodels.Type(typeStr)] = factory
	cfg, err := configtest.LoadConfigFile(
		t, path.Join(".", "testdata", "invalid_filtering_config.yaml"), factories,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	fsCfg := cfg.Receivers["smartagent/filesystems"].(*Config)
	require.Equal(t, &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr + "/filesystems",
		},
		monitorConfig: &filesystems.Config{
			MonitorConfig: config.MonitorConfig{
				Type: "filesystems",
				DatapointsToExclude: []config.MetricFilter{
					{
						MetricNames: []string{"./[0-"},
					},
				},
			},
		},
	}, fsCfg)

	err = fsCfg.validate()
	require.Error(t, err)
	require.EqualError(t, err, "unexpected end of input")
}
