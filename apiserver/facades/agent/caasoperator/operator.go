// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package caasoperator

import (
	"github.com/juju/errors"
	"gopkg.in/juju/names.v2"

	"github.com/juju/juju/apiserver/common"
	"github.com/juju/juju/apiserver/facade"
	"github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/state/watcher"
	"github.com/juju/juju/status"
)

type Facade struct {
	auth      facade.Authorizer
	resources facade.Resources
	state     CAASOperatorState
}

// NewStateFacade provides the signature required for facade registration.
func NewStateFacade(ctx facade.Context) (*Facade, error) {
	authorizer := ctx.Auth()
	resources := ctx.Resources()
	return NewFacade(resources, authorizer, stateShim{ctx.State()})
}

// NewFacade returns a new CAASOperator facade.
func NewFacade(
	resources facade.Resources,
	authorizer facade.Authorizer,
	st CAASOperatorState,
) (*Facade, error) {
	if !authorizer.AuthApplicationAgent() {
		return nil, common.ErrPerm
	}
	return &Facade{
		auth:      authorizer,
		resources: resources,
		state:     st,
	}, nil
}

// SetStatus sets the status of each given entity.
func (f *Facade) SetStatus(args params.SetStatus) (params.ErrorResults, error) {
	results := params.ErrorResults{
		Results: make([]params.ErrorResult, len(args.Entities)),
	}
	authTag := f.auth.GetAuthTag()
	for i, arg := range args.Entities {
		tag, err := names.ParseApplicationTag(arg.Tag)
		if err != nil {
			results.Results[i].Error = common.ServerError(err)
			continue
		}
		if tag != authTag {
			results.Results[i].Error = common.ServerError(common.ErrPerm)
			continue
		}
		info := status.StatusInfo{
			Status:  status.Status(arg.Status),
			Message: arg.Info,
			Data:    arg.Data,
		}
		results.Results[i].Error = common.ServerError(f.setStatus(tag, info))
	}
	return results, nil
}

func (f *Facade) setStatus(tag names.ApplicationTag, info status.StatusInfo) error {
	app, err := f.state.Application(tag.Id())
	if err != nil {
		return errors.Trace(err)
	}
	return app.SetStatus(info)
}

// Charm returns the charm info for all given applications.
func (f *Facade) Charm(args params.Entities) (params.ApplicationCharmResults, error) {
	results := params.ApplicationCharmResults{
		Results: make([]params.ApplicationCharmResult, len(args.Entities)),
	}
	authTag := f.auth.GetAuthTag()
	for i, entity := range args.Entities {
		tag, err := names.ParseApplicationTag(entity.Tag)
		if err != nil {
			results.Results[i].Error = common.ServerError(err)
			continue
		}
		if tag != authTag {
			results.Results[i].Error = common.ServerError(common.ErrPerm)
			continue
		}
		application, err := f.state.Application(tag.Id())
		if err != nil {
			results.Results[i].Error = common.ServerError(err)
			continue
		}
		charm, force, err := application.Charm()
		if err != nil {
			results.Results[i].Error = common.ServerError(err)
			continue
		}
		results.Results[i].Result = &params.ApplicationCharm{
			URL:          charm.URL().String(),
			ForceUpgrade: force,
			SHA256:       charm.BundleSha256(),
		}
	}
	return results, nil
}

// WatchApplicationConfig returns a NotifyWatcher that notifies when
// the application's config settings have changed.
func (f *Facade) WatchApplicationConfig(args params.Entities) (params.NotifyWatchResults, error) {
	results := params.NotifyWatchResults{
		Results: make([]params.NotifyWatchResult, len(args.Entities)),
	}
	authTag := f.auth.GetAuthTag()
	for i, arg := range args.Entities {
		watcherId, err := f.watchApplicationConfig(arg, authTag)
		if err != nil {
			results.Results[i].Error = common.ServerError(err)
			continue
		}
		results.Results[i].NotifyWatcherId = watcherId
	}
	return results, nil
}

func (f *Facade) watchApplicationConfig(arg params.Entity, authTag names.Tag) (string, error) {
	tag, err := names.ParseApplicationTag(arg.Tag)
	if err != nil {
		return "", err
	}
	if tag != authTag {
		return "", common.ErrPerm
	}
	application, err := f.state.Application(tag.Id())
	if err != nil {
		return "", err
	}
	w, err := application.WatchConfigSettings()
	if err != nil {
		return "", err
	}
	// Consume the initial event.
	if _, ok := <-w.Changes(); !ok {
		return "", watcher.EnsureErr(w)
	}
	return f.resources.Register(w), nil
}

// ApplicationConfig returns the application's config settings.
func (f *Facade) ApplicationConfig(args params.Entities) (params.ConfigSettingsResults, error) {
	results := params.ConfigSettingsResults{
		Results: make([]params.ConfigSettingsResult, len(args.Entities)),
	}
	authTag := f.auth.GetAuthTag()
	for i, arg := range args.Entities {
		tag, err := names.ParseApplicationTag(arg.Tag)
		if err != nil {
			results.Results[i].Error = common.ServerError(err)
			continue
		}
		if tag != authTag {
			results.Results[i].Error = common.ServerError(common.ErrPerm)
			continue
		}
		application, err := f.state.Application(tag.Id())
		if err != nil {
			results.Results[i].Error = common.ServerError(err)
			continue
		}
		settings, err := application.ConfigSettings()
		if err != nil {
			results.Results[i].Error = common.ServerError(err)
			continue
		}
		results.Results[i].Settings = params.ConfigSettings(settings)
	}
	return results, nil
}

// SetContainerSpec sets the container specs for a set of entities.
func (f *Facade) SetContainerSpec(args params.SetContainerSpecParams) (params.ErrorResults, error) {
	results := params.ErrorResults{
		Results: make([]params.ErrorResult, len(args.Entities)),
	}
	authTag := f.auth.GetAuthTag()
	canAccess := func(tag names.Tag) bool {
		if tag == authTag {
			return true
		}
		if tag, ok := tag.(names.UnitTag); ok {
			appName, err := names.UnitApplication(tag.Id())
			if err == nil && appName == authTag.Id() {
				return true
			}
		}
		return false
	}
	model, err := f.state.Model()
	if err != nil {
		return params.ErrorResults{}, errors.Trace(err)
	}
	for i, arg := range args.Entities {
		tag, err := names.ParseTag(arg.Tag)
		if err != nil {
			results.Results[i].Error = common.ServerError(err)
			continue
		}
		if !canAccess(tag) {
			results.Results[i].Error = common.ServerError(common.ErrPerm)
			continue
		}
		results.Results[i].Error = common.ServerError(
			model.SetContainerSpec(tag, arg.Value),
		)
	}
	return results, nil
}
