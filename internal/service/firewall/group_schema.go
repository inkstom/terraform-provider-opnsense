package firewall

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/firewall"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type groupResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Members     types.Set    `tfsdk:"members"`
	NoGroup     types.Bool   `tfsdk:"no_group"`
	Sequence    types.Int64  `tfsdk:"sequence"`
	Description types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func groupResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Interface groups (Firewall > Groups). A group gets its own rules tab, and rules on it apply to every member interface.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Only texts containing letters, digits and underscores with a maximum length of 15 characters are allowed and the name may not end with a digit. OPNsense also rejects names that start with a digit. Firewall rules and aliases refer to the group by this name, so renaming it can break those references.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 15)},
			},
			"members": schema.SetAttribute{
				MarkdownDescription: "Rules for WAN type interfaces in groups do not contain the reply-to mechanism upon which Multi-WAN typically relies. Member interfaces, by identifier (e.g. `wan`, `lan`, `opt1`). At least one is required.",
				ElementType:         types.StringType,
				Required:            true,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			},
			"no_group": schema.BoolAttribute{
				MarkdownDescription: "Prevent grouping these members in the interfaces menu section. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"sequence": schema.Int64Attribute{
				MarkdownDescription: "Priority sequence used in sorting the groups. Between `0` and `9999`. Defaults to `0`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "You may enter a description here for your reference (not parsed). Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
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
		MarkdownDescription: "Interface groups (Firewall > Groups).",

		Attributes: map[string]dschema.Attribute{
			"id":          dschema.StringAttribute{MarkdownDescription: "UUID of the group.", Required: true},
			"name":        dschema.StringAttribute{MarkdownDescription: "Only texts containing letters, digits and underscores with a maximum length of 15 characters are allowed and the name may not end with a digit.", Computed: true},
			"members":     dschema.SetAttribute{MarkdownDescription: "Rules for WAN type interfaces in groups do not contain the reply-to mechanism upon which Multi-WAN typically relies. Member interface identifiers.", ElementType: types.StringType, Computed: true},
			"no_group":    dschema.BoolAttribute{MarkdownDescription: "Prevent grouping these members in the interfaces menu section.", Computed: true},
			"sequence":    dschema.Int64Attribute{MarkdownDescription: "Priority sequence used in sorting the groups.", Computed: true},
			"description": dschema.StringAttribute{MarkdownDescription: "You may enter a description here for your reference (not parsed).", Computed: true},
		},
	}
}

func convertGroupSchemaToStruct(d *groupResourceModel) (*firewall.Group, error) {
	return &firewall.Group{
		Name:        d.Name.ValueString(),
		Members:     api.SelectedMapList(tools.SetToStringSlice(d.Members)),
		NoGroup:     tools.BoolToString(d.NoGroup.ValueBool()),
		Sequence:    tools.Int64ToString(d.Sequence.ValueInt64()),
		Description: d.Description.ValueString(),
	}, nil
}

func convertGroupStructToSchema(d *firewall.Group) (*groupResourceModel, error) {
	return &groupResourceModel{
		Name:        types.StringValue(d.Name),
		Members:     tools.StringSliceToSet(d.Members),
		NoGroup:     types.BoolValue(tools.StringToBool(d.NoGroup)),
		Sequence:    types.Int64Value(tools.StringToInt64(d.Sequence)),
		Description: types.StringValue(d.Description),
	}, nil
}

func mergeGroupWriteOnly(prior, read *groupResourceModel) {}
