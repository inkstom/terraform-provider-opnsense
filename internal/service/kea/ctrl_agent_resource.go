package kea

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ctrlAgentResource{}
var _ resource.ResourceWithConfigure = &ctrlAgentResource{}
var _ resource.ResourceWithImportState = &ctrlAgentResource{}

func newCtrlAgentResource() resource.Resource {
	return &ctrlAgentResource{}
}

type ctrlAgentResource struct {
	client opnsense.Client
}

func (r *ctrlAgentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kea_ctrl_agent"
}

func (r *ctrlAgentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ctrlAgentResourceSchema()
}

func (r *ctrlAgentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ctrlAgentResource) read(ctx context.Context) (*ctrlAgentResourceModel, error) {
	result, err := r.client.Kea().CtrlAgentGet(ctx)
	if err != nil {
		return nil, err
	}
	return convertCtrlAgentStructToSchema(&result.Ctrlagent.General), nil
}

func (r *ctrlAgentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Import Required",
		"opnsense_kea_ctrl_agent is a singleton that always exists on the firewall. "+
			"Import it before managing it:\n\n"+
			"  import {\n    to = opnsense_kea_ctrl_agent.<name>\n    id = \""+ctrlAgentID+"\"\n  }")
}

func (r *ctrlAgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	model, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Kea control agent settings, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *ctrlAgentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ctrlAgentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Kea().CtrlAgentSet(ctx, convertCtrlAgentSchemaToStruct(data))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kea control agent settings, got error: %s", err))
		return
	}
	if res != nil && res.Result != "saved" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Kea control agent settings: %+v", res))
		return
	}
	if err := r.client.Kea().Client().ReconfigureService(ctx, "/kea/service/reconfigure"); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Saved Kea control agent settings but reconfigure failed: %s", err))
		return
	}

	model, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated Kea control agent settings, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// Delete only forgets the settings; it never changes the firewall.
func (r *ctrlAgentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Removed From State Only",
		"opnsense_kea_ctrl_agent was removed from Terraform state. The firewall's Kea control agent settings were not changed.")
}

func (r *ctrlAgentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != ctrlAgentID {
		resp.Diagnostics.AddError("Invalid Import ID",
			fmt.Sprintf("opnsense_kea_ctrl_agent must be imported with ID %q, got %q.", ctrlAgentID, req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(ctrlAgentID))...)
}
