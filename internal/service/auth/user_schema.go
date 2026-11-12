package auth

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/auth"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type userResourceModel struct {
	Name             types.String `tfsdk:"name"`
	Password         types.String `tfsdk:"password"`
	Disabled         types.Bool   `tfsdk:"disabled"`
	FullName         types.String `tfsdk:"full_name"`
	Email            types.String `tfsdk:"email"`
	Comment          types.String `tfsdk:"comment"`
	Expires          types.String `tfsdk:"expires"`
	AuthorizedKeys   types.String `tfsdk:"authorized_keys"`
	Shell            types.String `tfsdk:"shell"`
	LandingPage      types.String `tfsdk:"landing_page"`
	Language         types.String `tfsdk:"language"`
	Privileges       types.Set    `tfsdk:"privileges"`
	GroupMemberships types.Set    `tfsdk:"group_memberships"`
	UserId           types.String `tfsdk:"uid"`

	Id types.String `tfsdk:"id"`
}

func optString(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc + " Defaults to `\"\"`.",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(""),
	}
}

func userResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Local users (System > Access > Users). The password is write-only on the OPNsense side: it is " +
			"sent on create and when it changes, and kept in state, but changes made outside Terraform are not detected. " +
			"When `password` is omitted, OPNsense sets a random (scrambled) password, which suits API-only or SSH-key accounts.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Login name.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Login password. Omit to have OPNsense set a random password.",
				Optional:            true,
				Sensitive:           true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Prevent the user from logging in. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"full_name":       optString("Full name."),
			"email":           optString("E-mail address."),
			"comment":         optString("Free-form comment."),
			"expires":         optString("Expiration date as `MM/DD/YYYY`; empty never expires."),
			"authorized_keys": optString("SSH public keys, one per line. Used for console/SSH login when the user has a shell."),
			"shell":           optString("Login shell: `\"\"` (none), `/bin/csh`, `/bin/sh` or `/bin/tcsh`."),
			"landing_page":    optString("Page shown after login, e.g. `/ui/core/dashboard`."),
			"language":        optString("UI language code, e.g. `en_US`; empty uses the system default."),
			"privileges": schema.SetAttribute{
				MarkdownDescription: "Privilege IDs granted directly to the user. Prefer granting through groups. Defaults to `[]`.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
			},
			"group_memberships": schema.SetAttribute{
				MarkdownDescription: "Group IDs (`gid`) the user belongs to, e.g. `[\"1999\"]` for `admins`. " +
					"Membership is the same data as `members` on `opnsense_auth_group`: manage it on one side only. " +
					"Leave unset to leave it unmanaged here.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"uid": schema.StringAttribute{
				MarkdownDescription: "Numeric user ID assigned by OPNsense.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the user.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func userDataSourceSchema() dschema.Schema {
	str := func(desc string) dschema.StringAttribute {
		return dschema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	return dschema.Schema{
		MarkdownDescription: "Local users (System > Access > Users).",

		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				MarkdownDescription: "UUID of the user.",
				Required:            true,
			},
			"name":              str("Login name."),
			"password":          dschema.StringAttribute{MarkdownDescription: "Always null; OPNsense does not return passwords.", Computed: true, Sensitive: true},
			"disabled":          dschema.BoolAttribute{MarkdownDescription: "Whether login is disabled.", Computed: true},
			"full_name":         str("Full name."),
			"email":             str("E-mail address."),
			"comment":           str("Comment."),
			"expires":           str("Expiration date."),
			"authorized_keys":   str("SSH public keys."),
			"shell":             str("Login shell."),
			"landing_page":      str("Landing page."),
			"language":          str("UI language."),
			"privileges":        dschema.SetAttribute{MarkdownDescription: "Privilege IDs granted directly.", ElementType: types.StringType, Computed: true},
			"group_memberships": dschema.SetAttribute{MarkdownDescription: "Group IDs the user belongs to.", ElementType: types.StringType, Computed: true},
			"uid":               str("Numeric user ID."),
		},
	}
}

func convertUserSchemaToStruct(d *userResourceModel) (*auth.User, error) {
	u := &auth.User{
		Name:              d.Name.ValueString(),
		Disabled:          tools.BoolToString(d.Disabled.ValueBool()),
		Fullname:          d.FullName.ValueString(),
		Email:             d.Email.ValueString(),
		Comment:           d.Comment.ValueString(),
		Expires:           d.Expires.ValueString(),
		AuthorizedKeys:    d.AuthorizedKeys.ValueString(),
		Shell:             api.SelectedMap(d.Shell.ValueString()),
		LandingPage:       d.LandingPage.ValueString(),
		Language:          api.SelectedMap(d.Language.ValueString()),
		Privilege:         api.SelectedMapList(tools.SetToStringSlice(d.Privileges)),
		GroupMemberships:  api.SelectedMapList(tools.SetToStringSlice(d.GroupMemberships)),
		ScrambledPassword: "0",
	}
	if d.Password.IsNull() || d.Password.IsUnknown() {
		u.ScrambledPassword = "1"
	} else {
		u.Password = d.Password.ValueString()
	}
	return u, nil
}

func convertUserStructToSchema(d *auth.User) (*userResourceModel, error) {
	return &userResourceModel{
		Name:             types.StringValue(d.Name),
		Password:         types.StringNull(),
		Disabled:         types.BoolValue(tools.StringToBool(d.Disabled)),
		FullName:         types.StringValue(d.Fullname),
		Email:            types.StringValue(d.Email),
		Comment:          types.StringValue(d.Comment),
		Expires:          types.StringValue(d.Expires),
		AuthorizedKeys:   types.StringValue(d.AuthorizedKeys),
		Shell:            types.StringValue(d.Shell.String()),
		LandingPage:      types.StringValue(d.LandingPage),
		Language:         types.StringValue(d.Language.String()),
		Privileges:       tools.StringSliceToSet(d.Privilege),
		GroupMemberships: tools.StringSliceToSet(d.GroupMemberships),
		UserId:           types.StringValue(d.UserId),
	}, nil
}

// mergeUserWriteOnly keeps the configured password (the API never returns
// it) and leaves group_memberships null when not managed here. On import,
// prior has no name yet and the live memberships are kept.
func mergeUserWriteOnly(prior, read *userResourceModel) {
	if prior == nil {
		return
	}
	read.Password = prior.Password
	if !prior.Name.IsNull() && prior.GroupMemberships.IsNull() {
		read.GroupMemberships = types.SetNull(types.StringType)
	}
}
