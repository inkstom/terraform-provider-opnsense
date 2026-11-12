package dnsmasq

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

var _ resource.Resource = &tagResource{}
var _ resource.ResourceWithConfigure = &tagResource{}
var _ resource.ResourceWithImportState = &tagResource{}

func newTagResource() resource.Resource {
	return &tagResource{}
}

type tagResource struct {
	client opnsense.Client
}

func (r *tagResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnsmasq_tag"
}

func (r *tagResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tagResourceSchema()
}

func (r *tagResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *tagResource) readBack(ctx context.Context, id string, prior *tagResourceModel) (*tagResourceModel, error) {
	obj, err := r.client.Dnsmasq().GetTag(ctx, id)
	if err != nil {
		return nil, err
	}
	model, err := convertTagStructToSchema(obj)
	if err != nil {
		return nil, err
	}
	model.Id = types.StringValue(id)
	mergeTagWriteOnly(prior, model)
	return model, nil
}

func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *tagResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := convertTagSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to parse dnsmasq tag, got error: %s", err))
		return
	}

	id, err := r.client.Dnsmasq().AddTag(ctx, obj)
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
			fmt.Sprintf("Unable to create dnsmasq tag, got error: %s", err))
		return
	}

	model, err := r.readBack(ctx, id, data)
	if err != nil {
		data.Id = types.StringValue(id)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Created dnsmasq tag %s but could not read it back, got error: %s", id, err))
		return
	}

	tflog.Trace(ctx, "created a resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *tagResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := r.readBack(ctx, data.Id.ValueString(), data)
	if err != nil {
		var notFoundError *errs.NotFoundError
		if errors.As(err, &notFoundError) {
			tflog.Warn(ctx, "dnsmasq tag not present in remote, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read dnsmasq tag, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *tagResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := convertTagSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to parse dnsmasq tag, got error: %s", err))
		return
	}

	if err := r.client.Dnsmasq().UpdateTag(ctx, data.Id.ValueString(), obj); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update dnsmasq tag, got error: %s", err))
		return
	}

	model, err := r.readBack(ctx, data.Id.ValueString(), data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Updated dnsmasq tag but could not read it back, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *tagResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Dnsmasq().DeleteTag(ctx, data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete dnsmasq tag, got error: %s", err))
		return
	}
}

func (r *tagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
