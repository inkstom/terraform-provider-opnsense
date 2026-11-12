package captiveportal

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/captiveportal"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type zoneResourceModel struct {
	Enabled              types.Bool   `tfsdk:"enabled"`
	ZoneId               types.String `tfsdk:"zone_id"`
	Interfaces           types.Set    `tfsdk:"interfaces"`
	DisableRules         types.Bool   `tfsdk:"disable_rules"`
	AuthServers          types.Set    `tfsdk:"auth_servers"`
	Roaming              types.Bool   `tfsdk:"roaming"`
	AlwaysSendAccounting types.Bool   `tfsdk:"always_send_accounting"`
	EnforceGroup         types.String `tfsdk:"enforce_group"`
	IdleTimeout          types.Int64  `tfsdk:"idle_timeout"`
	HardTimeout          types.Int64  `tfsdk:"hard_timeout"`
	ConcurrentLogins     types.Bool   `tfsdk:"concurrent_logins"`
	Certificate          types.String `tfsdk:"certificate"`
	ServerName           types.String `tfsdk:"server_name"`
	CorsPolicy           types.String `tfsdk:"cors_policy"`
	CorsOrigins          types.Set    `tfsdk:"cors_origins"`
	AllowedAddresses     types.Set    `tfsdk:"allowed_addresses"`
	AllowedMACAddresses  types.Set    `tfsdk:"allowed_mac_addresses"`
	ExtendedPreAuthData  types.Bool   `tfsdk:"extended_pre_auth_data"`
	Template             types.String `tfsdk:"template"`
	Description          types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func optString(desc, def string, v ...validator.String) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc + " Defaults to `\"" + def + "\"`.",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(def),
		Validators:          v,
	}
}

func optBool(desc string, def bool) schema.BoolAttribute {
	d := "false"
	if def {
		d = "true"
	}
	return schema.BoolAttribute{
		MarkdownDescription: desc + " Defaults to `" + d + "`.",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(def),
	}
}

func optSet(desc string) schema.SetAttribute {
	return schema.SetAttribute{
		MarkdownDescription: desc + " Defaults to `[]`.",
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
	}
}

// keepBool and friends are for settings newer OPNsense versions added: unset
// keeps the firewall's value (or its default on create).
func keepBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: desc + " Unset keeps the firewall's value.",
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
	}
}

func zoneResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Captive portal zones (Services > Captive Portal > Administration). Templates are uploaded in the web UI; reference one by UUID.",

		Attributes: map[string]schema.Attribute{
			"enabled": optBool("Enable the zone.", true),
			"zone_id": schema.StringAttribute{
				MarkdownDescription: "Zone number assigned by OPNsense.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"interfaces": schema.SetAttribute{
				MarkdownDescription: "Interfaces the portal runs on, by identifier (e.g. `opt2`). WAN cannot be used.",
				ElementType:         types.StringType,
				Required:            true,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			},
			"disable_rules":          keepBool("Don't generate the portal's firewall rules (OPNsense 26.1 or later)."),
			"auth_servers":           optSet("Authentication servers by name, e.g. `Local Database` or an LDAP/RADIUS server; empty allows anyone in after clicking through."),
			"roaming":                keepBool("Keep a session when the client's address changes (OPNsense 26.7 or later)."),
			"always_send_accounting": optBool("Always send RADIUS accounting requests.", false),
			"enforce_group":          optString("Only allow members of this group (by GID) to log in.", ""),
			"idle_timeout": schema.Int64Attribute{
				MarkdownDescription: "Disconnect idle clients after this many minutes; 0 disables. Defaults to `0`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"hard_timeout": schema.Int64Attribute{
				MarkdownDescription: "Disconnect clients after this many minutes regardless of activity; 0 disables. Defaults to `0`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"concurrent_logins": optBool("Allow a user to be logged in from several clients at once.", true),
			"certificate":       optString("Certificate reference for HTTPS on the portal; empty serves HTTP.", ""),
			"server_name":       optString("Host name for the portal (should match `certificate`).", ""),
			"cors_policy": schema.StringAttribute{
				MarkdownDescription: "CORS policy for the portal API: `same_origin`, `listed` or `any` (OPNsense 26.7 or later). Unset keeps the firewall's value.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators:          []validator.String{stringvalidator.OneOf("same_origin", "listed", "any")},
			},
			"cors_origins": schema.SetAttribute{
				MarkdownDescription: "Origins allowed when `cors_policy` is `listed` (OPNsense 26.7 or later). Unset keeps the firewall's value.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"allowed_addresses":      optSet("Addresses or networks allowed through without logging in."),
			"allowed_mac_addresses":  optSet("MAC addresses allowed through without logging in."),
			"extended_pre_auth_data": optBool("Send extended data (e.g. MAC address) to the authenticator.", false),
			"template":               optString("UUID of an uploaded portal template; empty uses the default page.", ""),
			"description":            optString("Description, shown as the zone name.", ""),
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the zone.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func zoneDataSourceSchema() dschema.Schema {
	s := func(d string) dschema.StringAttribute {
		return dschema.StringAttribute{MarkdownDescription: d, Computed: true}
	}
	b := func(d string) dschema.BoolAttribute {
		return dschema.BoolAttribute{MarkdownDescription: d, Computed: true}
	}
	i := func(d string) dschema.Int64Attribute {
		return dschema.Int64Attribute{MarkdownDescription: d, Computed: true}
	}
	set := func(d string) dschema.SetAttribute {
		return dschema.SetAttribute{MarkdownDescription: d, ElementType: types.StringType, Computed: true}
	}
	return dschema.Schema{
		MarkdownDescription: "Captive portal zones (Services > Captive Portal > Administration).",

		Attributes: map[string]dschema.Attribute{
			"id":                     dschema.StringAttribute{MarkdownDescription: "UUID of the zone.", Required: true},
			"enabled":                b("Whether the zone is enabled."),
			"zone_id":                s("Zone number."),
			"interfaces":             set("Interfaces."),
			"disable_rules":          b("Whether rule generation is disabled."),
			"auth_servers":           set("Authentication servers."),
			"roaming":                b("Whether roaming is enabled."),
			"always_send_accounting": b("Whether accounting is always sent."),
			"enforce_group":          s("Enforced group."),
			"idle_timeout":           i("Idle timeout in minutes."),
			"hard_timeout":           i("Hard timeout in minutes."),
			"concurrent_logins":      b("Whether concurrent logins are allowed."),
			"certificate":            s("Certificate reference."),
			"server_name":            s("Portal host name."),
			"cors_policy":            s("CORS policy."),
			"cors_origins":           set("Allowed CORS origins."),
			"allowed_addresses":      set("Addresses allowed without login."),
			"allowed_mac_addresses":  set("MAC addresses allowed without login."),
			"extended_pre_auth_data": b("Whether extended pre-auth data is sent."),
			"template":               s("Template UUID."),
			"description":            s("Description."),
		},
	}
}

func boolOrEmpty(v types.Bool) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return tools.BoolToString(v.ValueBool())
}

func stringOrEmpty(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

func convertZoneSchemaToStruct(d *zoneResourceModel) (*captiveportal.Zone, error) {
	var cors []string
	if !d.CorsOrigins.IsNull() && !d.CorsOrigins.IsUnknown() {
		cors = tools.SetToStringSlice(d.CorsOrigins)
	}
	return &captiveportal.Zone{
		Enabled:                  tools.BoolToString(d.Enabled.ValueBool()),
		Interfaces:               api.SelectedMapList(tools.SetToStringSlice(d.Interfaces)),
		DisableRules:             boolOrEmpty(d.DisableRules),
		AuthServers:              api.SelectedMapList(tools.SetToStringSlice(d.AuthServers)),
		Roaming:                  boolOrEmpty(d.Roaming),
		AlwaysSendAccountingReqs: tools.BoolToString(d.AlwaysSendAccounting.ValueBool()),
		AuthEnforceGroup:         api.SelectedMap(d.EnforceGroup.ValueString()),
		IdleTimeout:              tools.Int64ToString(d.IdleTimeout.ValueInt64()),
		HardTimeout:              tools.Int64ToString(d.HardTimeout.ValueInt64()),
		ConcurrentLogins:         tools.BoolToString(d.ConcurrentLogins.ValueBool()),
		Certificate:              api.SelectedMap(d.Certificate.ValueString()),
		ServerName:               d.ServerName.ValueString(),
		CorsPolicy:               api.SelectedMap(stringOrEmpty(d.CorsPolicy)),
		CorsOrigins:              api.SelectedMapList(cors),
		AllowedAddresses:         api.SelectedMapList(tools.SetToStringSlice(d.AllowedAddresses)),
		AllowedMACAddresses:      api.SelectedMapList(tools.SetToStringSlice(d.AllowedMACAddresses)),
		ExtendedPreAuthData:      tools.BoolToString(d.ExtendedPreAuthData.ValueBool()),
		Template:                 api.SelectedMap(d.Template.ValueString()),
		Description:              d.Description.ValueString(),
	}, nil
}

func convertZoneStructToSchema(d *captiveportal.Zone) (*zoneResourceModel, error) {
	return &zoneResourceModel{
		Enabled:              types.BoolValue(tools.StringToBool(d.Enabled)),
		ZoneId:               types.StringValue(d.ZoneId),
		Interfaces:           tools.StringSliceToSet(d.Interfaces),
		DisableRules:         types.BoolValue(tools.StringToBool(d.DisableRules)),
		AuthServers:          tools.StringSliceToSet(d.AuthServers),
		Roaming:              types.BoolValue(tools.StringToBool(d.Roaming)),
		AlwaysSendAccounting: types.BoolValue(tools.StringToBool(d.AlwaysSendAccountingReqs)),
		EnforceGroup:         types.StringValue(d.AuthEnforceGroup.String()),
		IdleTimeout:          types.Int64Value(tools.StringToInt64(d.IdleTimeout)),
		HardTimeout:          types.Int64Value(tools.StringToInt64(d.HardTimeout)),
		ConcurrentLogins:     types.BoolValue(tools.StringToBool(d.ConcurrentLogins)),
		Certificate:          types.StringValue(d.Certificate.String()),
		ServerName:           types.StringValue(d.ServerName),
		CorsPolicy:           types.StringValue(d.CorsPolicy.String()),
		CorsOrigins:          tools.StringSliceToSet(d.CorsOrigins),
		AllowedAddresses:     tools.StringSliceToSet(d.AllowedAddresses),
		AllowedMACAddresses:  tools.StringSliceToSet(d.AllowedMACAddresses),
		ExtendedPreAuthData:  types.BoolValue(tools.StringToBool(d.ExtendedPreAuthData)),
		Template:             types.StringValue(d.Template.String()),
		Description:          types.StringValue(d.Description),
	}, nil
}

func mergeZoneWriteOnly(prior, read *zoneResourceModel) {}
