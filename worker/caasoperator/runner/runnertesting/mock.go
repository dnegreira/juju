// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package runnertesting

import (
	"github.com/juju/errors"
	"github.com/juju/utils/proxy"
	"gopkg.in/juju/charm.v6"

	"github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/core/relation"
	"github.com/juju/juju/status"
	"github.com/juju/juju/worker/caasoperator/commands"
)

func NewMockContextAPI(apiAddresses []string, settings proxy.Settings) *MockContextAPI {
	return &MockContextAPI{
		apiAddresses: apiAddresses, settings: settings,
		appStatus: make(map[string]status.StatusInfo),
	}
}

type MockContextAPI struct {
	apiAddresses   []string
	settings       proxy.Settings
	appStatus      map[string]status.StatusInfo
	configSettings charm.Settings
	SpecYaml       string
	SpecUnitName   string
}

func (m *MockContextAPI) APIAddresses() ([]string, error) {
	return m.apiAddresses, nil
}

func (m *MockContextAPI) ProxySettings() (proxy.Settings, error) {
	return m.settings, nil
}

func (m *MockContextAPI) ConfigSettings() (charm.Settings, error) {
	return m.configSettings, nil
}

func (m *MockContextAPI) UpdateConfigSettings(settings charm.Settings) {
	m.configSettings = settings
}

func (m *MockContextAPI) NetworkInfo([]string, *int) (map[string]params.NetworkInfoResult, error) {
	return map[string]params.NetworkInfoResult{
		"db": {IngressAddresses: []string{"10.0.0.1"}},
	}, nil
}

func (m *MockContextAPI) ApplicationStatus(applicationName string) (params.ApplicationStatusResult, error) {
	statusInfo, ok := m.appStatus[applicationName]
	if !ok {
		return params.ApplicationStatusResult{}, errors.NotFoundf("application %v", applicationName)
	}
	return params.ApplicationStatusResult{Application: params.StatusResult{
		Status: string(statusInfo.Status),
		Info:   statusInfo.Message,
		Data:   statusInfo.Data,
	}}, nil
}

func (m *MockContextAPI) SetApplicationStatus(applicationName string, s status.Status, info string, data map[string]interface{}) error {
	if data == nil {
		data = map[string]interface{}{}
	}
	m.appStatus[applicationName] = status.StatusInfo{
		Status:  s,
		Message: info,
		Data:    data,
	}
	return nil
}

func (m *MockContextAPI) SetContainerSpec(specYaml, unitName string) error {
	m.SpecYaml = specYaml
	m.SpecUnitName = unitName
	return nil
}

func NewMockRelationUnitAPI(id int, endpoint string, suspended bool) *MockRelationUnitAPI {
	return &MockRelationUnitAPI{
		id:            id,
		endpoint:      endpoint,
		suspended:     suspended,
		localSettings: make(Settings),
	}
}

type MockRelationUnitAPI struct {
	id             int
	endpoint       string
	localSettings  Settings
	remoteSettings Settings
	status         relation.Status
	suspended      bool
}

func (m *MockRelationUnitAPI) Id() int {
	return m.id
}

func (m *MockRelationUnitAPI) Endpoint() string {
	return m.endpoint
}

func (m *MockRelationUnitAPI) LocalSettings() (commands.Settings, error) {
	return m.localSettings, nil
}

func (m *MockRelationUnitAPI) Suspended() bool {
	return m.suspended
}

func (m *MockRelationUnitAPI) SetStatus(status relation.Status) error {
	m.status = status
	return nil
}

func (m *MockRelationUnitAPI) Status() relation.Status {
	return m.status
}

func (m *MockRelationUnitAPI) RemoteSettings(unitName string) (commands.Settings, error) {
	result := make(Settings)
	for k, v := range m.remoteSettings {
		result[k] = v
	}
	return result, nil
}

func (m *MockRelationUnitAPI) WriteSettings(s commands.Settings) error {
	m.remoteSettings = Settings(s.Map())
	return nil
}

type Settings map[string]string

func (s Settings) Get(k string) (interface{}, bool) {
	v, f := s[k]
	return v, f
}

func (s Settings) Set(k, v string) {
	s[k] = v
}

func (s Settings) Delete(k string) {
	delete(s, k)
}

func (s Settings) Map() map[string]string {
	r := map[string]string{}
	for k, v := range s {
		r[k] = v
	}
	return r
}
