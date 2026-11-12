package interfaces

import (
	"context"
	"errors"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/errs"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/browningluke/terraform-provider-opnsense/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &assignmentResource{}
var _ resource.ResourceWithConfigure = &assignmentResource{}
var _ resource.ResourceWithImportState = &assignmentResource{}
var _ resource.ResourceWithConfigValidators = &assignmentResource{}

// protectedInterfaces cannot be removed through this provider; OPNsense needs them.
var protectedInterfaces = map[string]bool{"wan": true, "lan": true}

func newAssignmentResource() resource.Resource {
	return &assignmentResource{}
}

// assignmentResource defines the resource implementation.
type assignmentResource struct {
	client opnsense.Client
}

func (r *assignmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces_assignment"
}

func (r *assignmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = assignmentResourceSchema()
}

func (r *assignmentResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredWhenStringEqualsOneOf(path.MatchRoot("ipv4_address"), path.MatchRoot("ipv4_type"), []string{"staticv4"}),
		validators.RequiredWhenStringEqualsOneOf(path.MatchRoot("ipv6_address"), path.MatchRoot("ipv6_type"), []string{"staticv6"}),
		validators.RequiredWhenStringEqualsOneOf(path.MatchRoot("track6_interface"), path.MatchRoot("ipv6_type"), []string{"track6", "idassoc6"}),
	}
}

func (r *assignmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// readInto fetches the interface assignment and converts it, preserving the given ID.
func (r *assignmentResource) readInto(ctx context.Context, id string) (*assignmentResourceModel, error) {
	obj, err := r.client.Interfaces().GetAssignment(ctx, id)
	if err != nil {
		return nil, err
	}
	model, err := convertAssignmentStructToSchema(obj)
	if err != nil {
		return nil, err
	}
	model.Id = types.StringValue(id)
	return model, nil
}

func (r *assignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *assignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := convertAssignmentSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to parse interface assignment, got error: %s", err))
		return
	}

	id, err := r.client.Interfaces().AddAssignment(ctx, obj)
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
			fmt.Sprintf("Unable to create interface assignment, got error: %s", err))
		return
	}

	data.Id = types.StringValue(id)
	tflog.Trace(ctx, "created interface assignment")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *assignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *assignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := r.readInto(ctx, data.Id.ValueString())
	if err != nil {
		var notFoundError *errs.NotFoundError
		if errors.As(err, &notFoundError) {
			tflog.Warn(ctx, "interface assignment not present in remote, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read interface assignment, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *assignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *assignmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := convertAssignmentSchemaToStruct(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to parse interface assignment, got error: %s", err))
		return
	}

	if err := r.client.Interfaces().UpdateAssignment(ctx, data.Id.ValueString(), obj); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update interface assignment, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *assignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *assignmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ifname := data.Id.ValueString()

	if protectedInterfaces[ifname] {
		resp.Diagnostics.AddError("Refusing to delete interface",
			fmt.Sprintf("The %q assignment cannot be removed by this provider. Use `terraform state rm` to stop managing it.", ifname))
		return
	}

	// OPNsense rejects removal of locked interfaces; clear the lock first.
	current, err := r.client.Interfaces().GetAssignment(ctx, ifname)
	if err != nil {
		var notFoundError *errs.NotFoundError
		if errors.As(err, &notFoundError) {
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read interface assignment before delete, got error: %s", err))
		return
	}
	if tools.StringToBool(current.Lock) {
		current.Lock = "0"
		current.Identifier = ""
		if err := r.client.Interfaces().UpdateAssignment(ctx, ifname, current); err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to unlock interface assignment, got error: %s", err))
			return
		}
	}

	if err := r.client.Interfaces().DeleteAssignment(ctx, ifname); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete interface assignment, got error: %s", err))
		return
	}
}

func (r *assignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
