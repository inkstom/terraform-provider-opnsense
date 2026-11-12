package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/auth"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// privilegeResource manages who holds one privilege (System > Access >
// Privileges). It is authoritative for that privilege: users and groups not
// listed lose it, and destroying the resource removes it from everyone.

var _ resource.Resource = &privilegeResource{}
var _ resource.ResourceWithConfigure = &privilegeResource{}
var _ resource.ResourceWithImportState = &privilegeResource{}

func newPrivilegeResource() resource.Resource { return &privilegeResource{} }

type privilegeResource struct{ client opnsense.Client }

type privilegeResourceModel struct {
	Privilege types.String `tfsdk:"privilege"`
	Users     types.Set    `tfsdk:"users"`
	Groups    types.Set    `tfsdk:"groups"`
	Id        types.String `tfsdk:"id"`
}

func (r *privilegeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_privilege"
}

func (r *privilegeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Assigns one privilege to users and groups (System > Access > Privileges). " +
			"This resource is **authoritative** for the privilege: users and groups not listed lose it, and destroying the " +
			"resource removes it from everyone, including built-in grants such as `page-all` for `admins`. " +
			"It edits the same data as `privileges` on `opnsense_auth_user` and `opnsense_auth_group`; manage each privilege in one place only.",
		Attributes: map[string]schema.Attribute{
			"privilege": schema.StringAttribute{
				MarkdownDescription: "Privilege ID, e.g. `page-diagnostics-arptable`. Changing it forces a new resource.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"users": schema.SetAttribute{
				MarkdownDescription: "UUIDs of users holding the privilege (`opnsense_auth_user.<name>.id`). Defaults to `[]`.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
			},
			"groups": schema.SetAttribute{
				MarkdownDescription: "UUIDs of groups holding the privilege (`opnsense_auth_group.<name>.id`). Defaults to `[]`.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Same as `privilege`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func configureClient(providerData any, add func(string, string)) opnsense.Client {
	if providerData == nil {
		return nil
	}
	apiClient, ok := providerData.(*api.Client)
	if !ok {
		add("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *opnsense.Client, got: %T. Please report this issue to the provider developers.", providerData))
		return nil
	}
	return opnsense.NewClient(apiClient)
}

func (r *privilegeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := configureClient(req.ProviderData, resp.Diagnostics.AddError); c != nil {
		r.client = c
	}
}

func readPrivilege(ctx context.Context, c opnsense.Client, id string) (*privilegeResourceModel, error) {
	res, err := c.Auth().PrivilegeGetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	if res.Item.Id == "" {
		return nil, fmt.Errorf("privilege %q not found", id)
	}
	return &privilegeResourceModel{
		Privilege: types.StringValue(res.Item.Id),
		Users:     tools.StringSliceToSet(res.Item.Users),
		Groups:    tools.StringSliceToSet(res.Item.Groups),
		Id:        types.StringValue(res.Item.Id),
	}, nil
}

func writePrivilege(ctx context.Context, c opnsense.Client, id string, users, groups []string) error {
	res, err := c.Auth().PrivilegeSetItem(ctx, id, &auth.PrivilegeSetItem{
		Users:  strings.Join(users, ","),
		Groups: strings.Join(groups, ","),
	})
	if err != nil {
		return err
	}
	if res != nil && res.Result != "saved" {
		return fmt.Errorf("OPNsense did not save privilege %q: result %q", id, res.Result)
	}
	return nil
}

func (r *privilegeResource) apply(ctx context.Context, plan *privilegeResourceModel) (*privilegeResourceModel, error) {
	id := plan.Privilege.ValueString()
	if err := writePrivilege(ctx, r.client, id, tools.SetToStringSlice(plan.Users), tools.SetToStringSlice(plan.Groups)); err != nil {
		return nil, err
	}
	return readPrivilege(ctx, r.client, id)
}

func (r *privilegeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan privilegeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model, err := r.apply(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set privilege, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *privilegeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state privilegeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Id.ValueString()
	if id == "" {
		id = state.Privilege.ValueString()
	}
	model, err := readPrivilege(ctx, r.client, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read privilege, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *privilegeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan privilegeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model, err := r.apply(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set privilege, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *privilegeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state privilegeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := writePrivilege(ctx, r.client, state.Privilege.ValueString(), nil, nil); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to clear privilege, got error: %s", err))
	}
}

func (r *privilegeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("privilege"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// ---------- data source ----------

var _ datasource.DataSource = &privilegeDataSource{}
var _ datasource.DataSourceWithConfigure = &privilegeDataSource{}

func newPrivilegeDataSource() datasource.DataSource { return &privilegeDataSource{} }

type privilegeDataSource struct{ client opnsense.Client }

func (d *privilegeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_privilege"
}

func (d *privilegeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Who holds one privilege (System > Access > Privileges).",
		Attributes: map[string]dschema.Attribute{
			"privilege": dschema.StringAttribute{MarkdownDescription: "Privilege ID, e.g. `page-all`.", Required: true},
			"users":     dschema.SetAttribute{MarkdownDescription: "UUIDs of users holding it.", ElementType: types.StringType, Computed: true},
			"groups":    dschema.SetAttribute{MarkdownDescription: "UUIDs of groups holding it.", ElementType: types.StringType, Computed: true},
			"id":        dschema.StringAttribute{MarkdownDescription: "Same as `privilege`.", Computed: true},
		},
	}
}

func (d *privilegeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if c := configureClient(req.ProviderData, resp.Diagnostics.AddError); c != nil {
		d.client = c
	}
}

func (d *privilegeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg privilegeResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model, err := readPrivilege(ctx, d.client, cfg.Privilege.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read privilege, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
