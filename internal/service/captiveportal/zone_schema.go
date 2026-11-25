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
			"enabled": optBool("Enable this zone.", true),
			"zone_id": schema.StringAttribute{
				MarkdownDescription: "Internal number used for this zone. Assigned by OPNsense, from `0` to `19`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"interfaces": schema.SetAttribute{
				MarkdownDescription: "Select interface(s) to enable for captive portal. Use interface identifiers (e.g. `opt2`). Interfaces that get their address via DHCP cannot be selected.",
				ElementType:         types.StringType,
				Required:            true,
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			},
			"disable_rules":          keepBool("If this option is set, no automatic firewall rules for portal redirection and traffic blocking will be generated. This option allows you to override the default portal behavior for advanced use cases, such as redirections for DNS. See the documentation to see which rules you should implement in this scenario. Requires OPNsense 26.1 or later."),
			"auth_servers":           optSet("Select authentication methods to use, leave empty for no authentication needed. Use authentication server names, e.g. `Local Database` or an LDAP/RADIUS server."),
			"roaming":                keepBool("Allow connecting IPv6 clients to use multiple IPs bound to the same MAC over the course of its session. This option is needed for maximum IPv6 compatibility. Requires OPNsense 26.7 or later; OPNsense defaults to `true`."),
			"always_send_accounting": optBool("This will make the captive portal always send accounting requests, rather than just when there is a need for accounting (e.g. when there is a daily session limit).", false),
			"enforce_group":          optString("Restrict access to users in the selected (local) group.\n\n**NOTE:** please be aware that users (or vouchers) which aren't administered locally will be denied when using this option. Set the group by GID.", ""),
			"idle_timeout": schema.Int64Attribute{
				MarkdownDescription: "Clients will be disconnected after this amount of inactivity, in minutes. They may log in again immediately, though. Enter `0` to disable idle timeout. Must be between `0` and `10080` (a week). Defaults to `0`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"hard_timeout": schema.Int64Attribute{
				MarkdownDescription: "Clients will be disconnected after this amount of time, in minutes, regardless of activity. They may log in again immediately, though. Enter `0` to disable hard timeout (not recommended unless an idle timeout is set). Must be between `0` and `10080` (a week). Defaults to `0`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"concurrent_logins": optBool("If this option is set, users can login on multiple machines at once. If disabled subsequent logins will cause machines previously logged in with the same username to be disconnected.", true),
			"certificate":       optString("If provided, all traffic will be transmitted over an HTTPS connection to protect against eavesdroppers. Set the certificate reference (`refid`); empty serves HTTP.", ""),
			"server_name":       optString("Hostname (of this machine) to redirect login page to, leave blank to use this interface IP address, otherwise make sure the client can access DNS to resolve this location. When using a SSL certificate, make sure both this name and the cert name are equal.", ""),
			"cors_policy": schema.StringAttribute{
				MarkdownDescription: "Control which browser origins may read responses from the Captive Portal Access API. Same-origin access and clients that do not send an Origin header are unaffected. One of `same_origin` (Same origin only), `listed` (Specific origins) or `any` (Any origin); OPNsense defaults to `any`. Requires OPNsense 26.7 or later. Unset keeps the firewall's value.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators:          []validator.String{stringvalidator.OneOf("same_origin", "listed", "any")},
			},
			"cors_origins": schema.SetAttribute{
				MarkdownDescription: "Origins allowed when Cross-origin access is set to Specific origins (`cors_policy` = `listed`). Specify only the scheme, hostname and optional port, for example `https://login.example.com`. Requires OPNsense 26.7 or later. Unset keeps the firewall's value.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"allowed_addresses":      optSet("Avoid authentication for addresses and subnets in this list."),
			"allowed_mac_addresses":  optSet("Avoid authentication for physical addresses in this list."),
			"extended_pre_auth_data": optBool("Offer extended data to the login template before authentication (mac addresses for upstream use).", false),
			"template":               optString("Use custom template package for user login. Set the template UUID (`opnsense_captiveportal_template.id`); empty uses the default page.", ""),
			"description":            optString("Description to identify this zone.", ""),
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
			"enabled":                b("Enable this zone."),
			"zone_id":                s("Internal number used for this zone."),
			"interfaces":             set("Interface(s) enabled for captive portal."),
			"disable_rules":          b("If set, no automatic firewall rules for portal redirection and traffic blocking are generated."),
			"auth_servers":           set("Authentication methods to use; empty means no authentication needed."),
			"roaming":                b("Allow connecting IPv6 clients to use multiple IPs bound to the same MAC over the course of its session."),
			"always_send_accounting": b("Always send accounting requests, rather than just when there is a need for accounting."),
			"enforce_group":          s("Restrict access to users in this (local) group, by GID."),
			"idle_timeout":           i("Clients are disconnected after this amount of inactivity, in minutes. `0` disables idle timeout."),
			"hard_timeout":           i("Clients are disconnected after this amount of time, in minutes, regardless of activity. `0` disables hard timeout."),
			"concurrent_logins":      b("If set, users can login on multiple machines at once."),
			"certificate":            s("SSL certificate reference; if provided, all traffic is transmitted over HTTPS."),
			"server_name":            s("Hostname (of this machine) to redirect login page to; blank uses the interface IP address."),
			"cors_policy":            s("Control which browser origins may read responses from the Captive Portal Access API: `same_origin`, `listed` or `any`."),
			"cors_origins":           set("Origins allowed when Cross-origin access is set to Specific origins."),
			"allowed_addresses":      set("Avoid authentication for addresses and subnets in this list."),
			"allowed_mac_addresses":  set("Avoid authentication for physical addresses in this list."),
			"extended_pre_auth_data": b("Offer extended data to the login template before authentication (mac addresses for upstream use)."),
			"template":               s("Custom template package for user login, by UUID."),
			"description":            s("Description to identify this zone."),
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
