package dnsmasq

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &settingsResource{}
var _ resource.ResourceWithConfigure = &settingsResource{}
var _ resource.ResourceWithImportState = &settingsResource{}

func NewSettingsResource() resource.Resource {
	return &settingsResource{}
}

type settingsResource struct {
	client opnsense.Client
}

func (r *settingsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnsmasq_settings"
}

func (r *settingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = settingsResourceSchema()
}

func (r *settingsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create is blocked: the settings always exist, so they must be imported first.
// This prevents a new config from silently overwriting every dnsmasq setting.
func (r *settingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Import Required",
		"opnsense_dnsmasq_settings is a singleton that always exists on the firewall. "+
			"Import it before managing it:\n\n"+
			"  import {\n    to = opnsense_dnsmasq_settings.<name>\n    id = \""+settingsID+"\"\n  }")
}

func (r *settingsResource) read(ctx context.Context) (*settingsResourceModel, error) {
	result, err := r.client.Dnsmasq().GeneralSettingsGet(ctx)
	if err != nil {
		return nil, err
	}
	return convertSettingsStructToSchema(&result.Dnsmasq), nil
}

func (r *settingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	model, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read dnsmasq settings, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Dnsmasq().GeneralSettingsSet(ctx, convertSettingsSchemaToStruct(data))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update dnsmasq settings, got error: %s", err))
		return
	}
	if res != nil && res.Result != "saved" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update dnsmasq settings: %+v", res))
		return
	}
	if _, err := r.client.Dnsmasq().ServiceReconfigure(ctx); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Saved dnsmasq settings but reconfigure failed: %s", err))
		return
	}

	model, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated dnsmasq settings, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Delete only forgets the settings; it never changes the firewall.
func (r *settingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Removed From State Only",
		"opnsense_dnsmasq_settings was removed from Terraform state. The firewall's dnsmasq settings were not changed.")
}

func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != settingsID {
		resp.Diagnostics.AddError("Invalid Import ID",
			fmt.Sprintf("opnsense_dnsmasq_settings must be imported with ID %q, got %q.", settingsID, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(settingsID))...)
}
