package auth

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/auth"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type groupResourceModel struct {
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Privileges     types.Set    `tfsdk:"privileges"`
	Members        types.Set    `tfsdk:"members"`
	SourceNetworks types.Set    `tfsdk:"source_networks"`
	GroupId        types.String `tfsdk:"gid"`

	Id types.String `tfsdk:"id"`
}

func groupResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Local groups (System > Access > Groups). Members are referenced by user ID (`uid`), so use `opnsense_auth_user.<name>.uid`.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Group name.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Group description. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"privileges": schema.SetAttribute{
				MarkdownDescription: "Privilege IDs granted to the group, e.g. `page-all` or `page-diagnostics-arptable`. Defaults to `[]`.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
			},
			"members": schema.SetAttribute{
				MarkdownDescription: "User IDs (`uid`) of the members, e.g. `[\"0\"]` for root. Leave unset to leave membership " +
					"unmanaged here, for example when users join through `opnsense_auth_user.group_memberships`.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"source_networks": schema.SetAttribute{
				MarkdownDescription: "Networks (CIDR) members may log in from. Empty allows any. Defaults to `[]`.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
			},
			"gid": schema.StringAttribute{
				MarkdownDescription: "Numeric group ID assigned by OPNsense.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the group.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func groupDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Local groups (System > Access > Groups).",

		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				MarkdownDescription: "UUID of the group.",
				Required:            true,
			},
			"name":            dschema.StringAttribute{MarkdownDescription: "Group name.", Computed: true},
			"description":     dschema.StringAttribute{MarkdownDescription: "Group description.", Computed: true},
			"privileges":      dschema.SetAttribute{MarkdownDescription: "Privilege IDs granted to the group.", ElementType: types.StringType, Computed: true},
			"members":         dschema.SetAttribute{MarkdownDescription: "User IDs of the members.", ElementType: types.StringType, Computed: true},
			"source_networks": dschema.SetAttribute{MarkdownDescription: "Networks members may log in from.", ElementType: types.StringType, Computed: true},
			"gid":             dschema.StringAttribute{MarkdownDescription: "Numeric group ID.", Computed: true},
		},
	}
}

func convertGroupSchemaToStruct(d *groupResourceModel) (*auth.Group, error) {
	return &auth.Group{
		Name:           d.Name.ValueString(),
		Description:    d.Description.ValueString(),
		Privilege:      api.SelectedMapList(tools.SetToStringSlice(d.Privileges)),
		Member:         api.SelectedMapList(tools.SetToStringSlice(d.Members)),
		SourceNetworks: api.SelectedMapList(tools.SetToStringSlice(d.SourceNetworks)),
	}, nil
}

func convertGroupStructToSchema(d *auth.Group) (*groupResourceModel, error) {
	return &groupResourceModel{
		Name:           types.StringValue(d.Name),
		Description:    types.StringValue(d.Description),
		Privileges:     tools.StringSliceToSet(d.Privilege),
		Members:        tools.StringSliceToSet(d.Member),
		SourceNetworks: tools.StringSliceToSet(d.SourceNetworks),
		GroupId:        types.StringValue(d.GroupId),
	}, nil
}

// mergeGroupWriteOnly leaves members null when they are not managed here
// (prior has a name but no members). On import, prior has no name yet and the
// live members are kept.
func mergeGroupWriteOnly(prior, read *groupResourceModel) {
	if prior != nil && !prior.Name.IsNull() && prior.Members.IsNull() {
		read.Members = types.SetNull(types.StringType)
	}
}
