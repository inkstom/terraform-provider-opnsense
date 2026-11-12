package ipsec

import (
	"context"
	"errors"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/errs"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &keyPairResource{}
var _ resource.ResourceWithConfigure = &keyPairResource{}
var _ resource.ResourceWithImportState = &keyPairResource{}

func newKeyPairResource() resource.Resource {
	return &keyPairResource{}
}

type keyPairResource struct {
	client opnsense.Client
}

func (r *keyPairResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_key_pair"
}

func (r *keyPairResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = keyPairResourceSchema()
}

func (r *keyPairResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// readBack fetches the object and converts it, keeping write-only values
// (which the API does not return) from prior.
func (r *keyPairResource) readBack(ctx context.Context, id string, prior *keyPairResourceModel) (*keyPairResourceModel, error) {
	obj, err := r.client.Ipsec().GetIPsecKeyPair(ctx, id)
	if err != nil {
		return nil, err
	}
	model, err := convertKeyPairStructToSchema(obj)
	if err != nil {
		return nil, err
	}
	model.Id = types.StringValue(id)
	mergeKeyPairWriteOnly(prior, model)
	return model, nil
}

func (r *keyPairResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *keyPairResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := convertKeyPairSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to parse IPsec key pair, got error: %s", err))
		return
	}

	id, err := r.client.Ipsec().AddIPsecKeyPair(ctx, obj)
	if err != nil {
		if id != "" {
			// Created, but a later step (such as reconfigure) failed: keep it
			// in state so it is not orphaned.
			data.Id = types.StringValue(id)
			if model, readErr := r.readBack(ctx, id, data); readErr == nil {
				data = model
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create IPsec key pair, got error: %s", err))
		return
	}

	model, err := r.readBack(ctx, id, data)
	if err != nil {
		data.Id = types.StringValue(id)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Created IPsec key pair %s but could not read it back, got error: %s", id, err))
		return
	}

	tflog.Trace(ctx, "created a resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *keyPairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *keyPairResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := r.readBack(ctx, data.Id.ValueString(), data)
	if err != nil {
		var notFoundError *errs.NotFoundError
		if errors.As(err, &notFoundError) {
			tflog.Warn(ctx, "IPsec key pair not present in remote, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read IPsec key pair, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *keyPairResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *keyPairResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := convertKeyPairSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to parse IPsec key pair, got error: %s", err))
		return
	}

	if err := r.client.Ipsec().UpdateIPsecKeyPair(ctx, data.Id.ValueString(), obj); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update IPsec key pair, got error: %s", err))
		return
	}

	model, err := r.readBack(ctx, data.Id.ValueString(), data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Updated IPsec key pair but could not read it back, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *keyPairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *keyPairResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Ipsec().DeleteIPsecKeyPair(ctx, data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete IPsec key pair, got error: %s", err))
		return
	}
}

func (r *keyPairResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
