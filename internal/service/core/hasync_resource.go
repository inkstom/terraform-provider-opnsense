package core

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &hasyncResource{}
var _ resource.ResourceWithConfigure = &hasyncResource{}
var _ resource.ResourceWithImportState = &hasyncResource{}

func newHasyncResource() resource.Resource {
	return &hasyncResource{}
}

type hasyncResource struct {
	client opnsense.Client
}

func (r *hasyncResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_core_hasync"
}

func (r *hasyncResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = hasyncResourceSchema()
}

func (r *hasyncResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	r.client = opnsense.NewClient(apiClient)
}

func (r *hasyncResource) read(ctx context.Context) (*hasyncResourceModel, error) {
	result, err := r.client.Core().HasyncGet(ctx)
	if err != nil {
		return nil, err
	}
	return convertHasyncStructToSchema(&result.Hasync), nil
}

func (r *hasyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Import Required",
		"opnsense_core_hasync is a singleton that always exists on the firewall. "+
			"Import it before managing it:\n\n"+
			"  import {\n    to = opnsense_core_hasync.<name>\n    id = \""+hasyncID+"\"\n  }")
}

func (r *hasyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	model, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HA sync settings, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *hasyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *hasyncResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Core().HasyncSet(ctx, convertHasyncSchemaToStruct(data))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update HA sync settings, got error: %s", err))
		return
	}
	if res != nil && res.Result != "saved" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update HA sync settings: %+v", res))
		return
	}
	if _, err := r.client.Core().HasyncReconfigure(ctx); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Saved HA sync settings but reconfigure failed: %s", err))
		return
	}

	model, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated HA sync settings, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Delete only forgets the settings; it never changes the firewall.
func (r *hasyncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Removed From State Only",
		"opnsense_core_hasync was removed from Terraform state. The firewall's HA settings were not changed.")
}

func (r *hasyncResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != hasyncID {
		resp.Diagnostics.AddError("Invalid Import ID",
			fmt.Sprintf("opnsense_core_hasync must be imported with ID %q, got %q.", hasyncID, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(hasyncID))...)
}
