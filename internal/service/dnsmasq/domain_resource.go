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

var _ resource.Resource = &domainResource{}
var _ resource.ResourceWithConfigure = &domainResource{}
var _ resource.ResourceWithImportState = &domainResource{}

func newDomainResource() resource.Resource {
	return &domainResource{}
}

type domainResource struct {
	client opnsense.Client
}

func (r *domainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnsmasq_domain"
}

func (r *domainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = domainResourceSchema()
}

func (r *domainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *domainResource) readBack(ctx context.Context, id string, prior *domainResourceModel) (*domainResourceModel, error) {
	obj, err := r.client.Dnsmasq().GetDomain(ctx, id)
	if err != nil {
		return nil, err
	}
	model, err := convertDomainStructToSchema(obj)
	if err != nil {
		return nil, err
	}
	model.Id = types.StringValue(id)
	mergeDomainWriteOnly(prior, model)
	return model, nil
}

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *domainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := convertDomainSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to parse dnsmasq domain override, got error: %s", err))
		return
	}

	id, err := r.client.Dnsmasq().AddDomain(ctx, obj)
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
			fmt.Sprintf("Unable to create dnsmasq domain override, got error: %s", err))
		return
	}

	model, err := r.readBack(ctx, id, data)
	if err != nil {
		data.Id = types.StringValue(id)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Created dnsmasq domain override %s but could not read it back, got error: %s", id, err))
		return
	}

	tflog.Trace(ctx, "created a resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *domainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := r.readBack(ctx, data.Id.ValueString(), data)
	if err != nil {
		var notFoundError *errs.NotFoundError
		if errors.As(err, &notFoundError) {
			tflog.Warn(ctx, "dnsmasq domain override not present in remote, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read dnsmasq domain override, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *domainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := convertDomainSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to parse dnsmasq domain override, got error: %s", err))
		return
	}

	if err := r.client.Dnsmasq().UpdateDomain(ctx, data.Id.ValueString(), obj); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update dnsmasq domain override, got error: %s", err))
		return
	}

	model, err := r.readBack(ctx, data.Id.ValueString(), data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Updated dnsmasq domain override but could not read it back, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *domainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Dnsmasq().DeleteDomain(ctx, data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete dnsmasq domain override, got error: %s", err))
		return
	}
}

func (r *domainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
