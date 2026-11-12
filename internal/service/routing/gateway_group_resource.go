package routing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/errs"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gatewayGroupResource{}
var _ resource.ResourceWithConfigure = &gatewayGroupResource{}
var _ resource.ResourceWithImportState = &gatewayGroupResource{}

func newGatewayGroupResource() resource.Resource {
	return &gatewayGroupResource{}
}

// gatewayGroupResource defines the resource implementation.
type gatewayGroupResource struct {
	client opnsense.Client
}

func (r *gatewayGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_gateway_group"
}

func (r *gatewayGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = gatewayGroupResourceSchema()
}

func (r *gatewayGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *opnsense.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = opnsense.NewClient(apiClient)
}

// readInto fetches the gateway group and converts it, preserving the given ID.
func (r *gatewayGroupResource) readInto(ctx context.Context, id string) (*gatewayGroupResourceModel, error) {
	obj, err := r.client.Routing().GetGatewayGroup(ctx, id)
	if err != nil {
		return nil, err
	}
	model, err := convertGatewayGroupStructToSchema(obj)
	if err != nil {
		return nil, err
	}
	model.Id = types.StringValue(id)
	return model, nil
}

func (r *gatewayGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *gatewayGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := convertGatewayGroupSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to parse gateway group, got error: %s", err))
		return
	}

	var id string
	err = retryStaleGatewayList(ctx, func() error {
		var addErr error
		id, addErr = r.client.Routing().AddGatewayGroup(ctx, obj)
		if addErr != nil && id != "" {
			// Saved but reconfigure failed: never retry, the group exists.
			return permanent{addErr}
		}
		return addErr
	})
	var p permanent
	if errors.As(err, &p) {
		err = p.err
	}
	if err != nil {
		if id != "" {
			// Created but reconfigure failed: keep it in state so it isn't orphaned.
			data.Id = types.StringValue(id)
			if readModel, readErr := r.readInto(ctx, id); readErr == nil {
				data = readModel
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create gateway group, got error: %s", err))
		return
	}

	data.Id = types.StringValue(id)
	tflog.Trace(ctx, "created gateway group")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gatewayGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *gatewayGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := r.readInto(ctx, data.Id.ValueString())
	if err != nil {
		var notFoundError *errs.NotFoundError
		if errors.As(err, &notFoundError) {
			tflog.Warn(ctx, "gateway group not present in remote, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read gateway group, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *gatewayGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *gatewayGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := convertGatewayGroupSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to parse gateway group, got error: %s", err))
		return
	}

	if err := retryStaleGatewayList(ctx, func() error {
		return r.client.Routing().UpdateGatewayGroup(ctx, data.Id.ValueString(), obj)
	}); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update gateway group, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gatewayGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *gatewayGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Routing().DeleteGatewayGroup(ctx, data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete gateway group, got error: %s", err))
		return
	}
}

func (r *gatewayGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// OPNsense builds the gateway-group member options from the configd action
// "interface gateways list", which is cached for 20 seconds (cache_ttl in
// actions_interface.conf). A gateway created moments earlier in the same
// apply is rejected with "Option [NAME] not in list." until the cache expires,
// so retry that specific validation error for a bounded time.
const (
	staleGatewayListTimeout  = 60 * time.Second
	staleGatewayListInterval = 5 * time.Second
)

type permanent struct{ err error }

func (p permanent) Error() string { return p.err.Error() }

func isStaleGatewayListError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not in list")
}

func retryStaleGatewayList(ctx context.Context, fn func() error) error {
	deadline := time.Now().Add(staleGatewayListTimeout)
	for {
		err := fn()
		var p permanent
		if err == nil || errors.As(err, &p) || !isStaleGatewayListError(err) || time.Now().After(deadline) {
			return err
		}
		tflog.Info(ctx, "gateway not yet in OPNsense's cached gateway list, retrying", map[string]any{"error": err.Error()})
		select {
		case <-ctx.Done():
			return err
		case <-time.After(staleGatewayListInterval):
		}
	}
}
