// The code in this file was automatically generated by running httprequest-generate-client.
// DO NOT EDIT

package jemclient

import (
	"context"

	"github.com/CanonicalLtd/jimm/params"
	"gopkg.in/httprequest.v1"
)

type client struct {
	Client httprequest.Client
}

// AddController adds a new controller.
func (c *client) AddController(ctx context.Context, p *params.AddController) error {
	return c.Client.Call(ctx, p, nil)
}

// DeleteController removes an existing controller.
func (c *client) DeleteController(ctx context.Context, p *params.DeleteController) error {
	return c.Client.Call(ctx, p, nil)
}

// DeleteModel deletes an model from JEM.
func (c *client) DeleteModel(ctx context.Context, p *params.DeleteModel) error {
	return c.Client.Call(ctx, p, nil)
}

// GetAuditEntries return the list of audit log entries based on the requested query.
func (c *client) GetAuditEntries(ctx context.Context, p *params.AuditLogRequest) (params.AuditLogEntries, error) {
	var r params.AuditLogEntries
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// GetController returns information on a controller.
func (c *client) GetController(ctx context.Context, p *params.GetController) (*params.ControllerResponse, error) {
	var r *params.ControllerResponse
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

func (c *client) GetControllerDeprecated(ctx context.Context, p *params.GetControllerDeprecated) (*params.DeprecatedBody, error) {
	var r *params.DeprecatedBody
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// GetControllerPerm returns the ACL for a given controller.
// Only the owner (arg.EntityPath.User) can read the ACL.
func (c *client) GetControllerPerm(ctx context.Context, p *params.GetControllerPerm) (params.ACL, error) {
	var r params.ACL
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// GetModel returns information on a given model.
func (c *client) GetModel(ctx context.Context, p *params.GetModel) (*params.ModelResponse, error) {
	var r *params.ModelResponse
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// GetModelName returns the name of the model identified by the provided uuid.
func (c *client) GetModelName(ctx context.Context, p *params.ModelNameRequest) (params.ModelNameResponse, error) {
	var r params.ModelNameResponse
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// GetModelPerm returns the ACL for a given model.
// Only the owner (arg.EntityPath.User) can read the ACL.
func (c *client) GetModelPerm(ctx context.Context, p *params.GetModelPerm) (params.ACL, error) {
	var r params.ACL
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// GetModelStatuses return the list of all models created between 2 dates (or all).
func (c *client) GetModelStatuses(ctx context.Context, p *params.ModelStatusesRequest) (params.ModelStatuses, error) {
	var r params.ModelStatuses
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// JujuStatus retrieves and returns the status of the specifed model.
func (c *client) JujuStatus(ctx context.Context, p *params.JujuStatus) (*params.JujuStatusResponse, error) {
	var r *params.JujuStatusResponse
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// ListController returns all the controllers stored in JEM.
// Currently the ProviderType field in each ControllerResponse is not
// populated.
func (c *client) ListController(ctx context.Context, p *params.ListController) (*params.ListControllerResponse, error) {
	var r *params.ListControllerResponse
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// ListModels returns all the models stored in JEM.
// Note that the models returned don't include the username or password.
// To gain access to a specific model, that model should be retrieved
// explicitly.
func (c *client) ListModels(ctx context.Context, p *params.ListModels) (*params.ListModelsResponse, error) {
	var r *params.ListModelsResponse
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// LogLevel returns the current logging level of the running service.
func (c *client) LogLevel(ctx context.Context, p *params.LogLevel) (params.Level, error) {
	var r params.Level
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// Migrate starts a migration of a model from its current
// controller to a different one. The migration will not have
// completed by the time the Migrate call returns.
func (c *client) Migrate(ctx context.Context, p *params.Migrate) error {
	return c.Client.Call(ctx, p, nil)
}

// MissingModels returns a list of models present on the given controller
// that are not in the local database.
func (c *client) MissingModels(ctx context.Context, p *params.MissingModelsRequest) (params.MissingModels, error) {
	var r params.MissingModels
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

// NewModel creates a new model inside an existing Controller.
func (c *client) NewModel(ctx context.Context, p *params.NewModel) (*params.ModelResponse, error) {
	var r *params.ModelResponse
	err := c.Client.Call(ctx, p, &r)
	return r, err
}

func (c *client) SetControllerDeprecated(ctx context.Context, p *params.SetControllerDeprecated) error {
	return c.Client.Call(ctx, p, nil)
}

// SetControllerPerm sets the permissions on a controller entity.
// Only the owner (arg.EntityPath.User) can change the permissions
// on an an entity. The owner can always read an entity, even
// if it has empty ACL.
func (c *client) SetControllerPerm(ctx context.Context, p *params.SetControllerPerm) error {
	return c.Client.Call(ctx, p, nil)
}

// SetLogLevel configures the logging level of the running service.
func (c *client) SetLogLevel(ctx context.Context, p *params.SetLogLevel) error {
	return c.Client.Call(ctx, p, nil)
}

// SetModelPerm sets the permissions on a controller entity.
// Only the owner (arg.EntityPath.User) can change the permissions
// on an an entity. The owner can always read an entity, even
// if it has empty ACL.
// TODO remove this.
func (c *client) SetModelPerm(ctx context.Context, p *params.SetModelPerm) error {
	return c.Client.Call(ctx, p, nil)
}

// UpdateCredential stores the provided credential under the provided,
// user, cloud and name. If there is already a credential with that name
// it is overwritten.
func (c *client) UpdateCredential(ctx context.Context, p *params.UpdateCredential) error {
	return c.Client.Call(ctx, p, nil)
}

// WhoAmI returns authentication information on the client that is
// making the WhoAmI call.
func (c *client) WhoAmI(ctx context.Context, p *params.WhoAmI) (params.WhoAmIResponse, error) {
	var r params.WhoAmIResponse
	err := c.Client.Call(ctx, p, &r)
	return r, err
}
