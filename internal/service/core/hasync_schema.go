package core

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/core"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const hasyncID = "high_availability"

type hasyncResourceModel struct {
	DisablePreempt   types.Bool   `tfsdk:"disable_preempt"`
	DisconnectDialup types.Bool   `tfsdk:"disconnect_dialup"`
	PfsyncInterface  types.String `tfsdk:"pfsync_interface"`
	PfsyncPeerIp     types.String `tfsdk:"pfsync_peer_ip"`
	PfsyncVersion    types.String `tfsdk:"pfsync_version"`
	PfsyncDefer      types.Bool   `tfsdk:"pfsync_defer"`
	SynchronizeToIp  types.String `tfsdk:"synchronize_to_ip"`
	VerifyPeer       types.Bool   `tfsdk:"verify_peer"`
	Username         types.String `tfsdk:"username"`
	Password         types.String `tfsdk:"password"`
	SyncItems        types.Set    `tfsdk:"sync_items"`

	Id types.String `tfsdk:"id"`
}

// Every setting is optional: leave one unset and its current value on the
// firewall is kept, not managed.
func hasyncBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: desc + " Unset keeps the current value.",
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
	}
}

func hasyncString(desc string, v ...validator.String) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc + " Unset keeps the current value.",
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		Validators:          v,
	}
}

// OPNsense help text (core hasyncSettings form), shared by resource and data source.
var hasyncHelp = map[string]string{
	"disable_preempt":   "When this device is configured as CARP master it will try to switch to master when powering up, this option will keep this one slave if there already is a master on the network. A reboot is required to take effect.",
	"disconnect_dialup": "When this device is configured as CARP backup it will disconnect all PPP type interfaces and try to reconnect them when becoming master again.",
	"pfsync_interface":  "This enables state insertion, update, and deletion messages between firewalls by utilizing the selected interface for communication. Best choose a dedicated interface for this type of communication to prevent manipulation of states causing security issues.",
	"pfsync_peer_ip":    "Setting this option will force pfsync to synchronize its state table to this IP address. The default is directed multicast.",
	"pfsync_version":    "Newer versions of OPNsense offer additional attributes in the state synchronization, for compatibility reasons you can optionally choose an older version here. Always make sure both nodes use the same version to avoid inconsistent state tables.",
	"pfsync_defer":      "Defer transmission of the first packet in a state until a peer has acknowledged that the associated state has been inserted. This adds small delays but allows traffic to flow when more than one firewall might actively handle packets (\"active/active\").",
	"synchronize_to_ip": "Enter the IP address of the firewall to which the selected configuration sections should be synchronized. This should be empty on the backup machine. When an IP address is offered, both web GUI configurations should be equal (port and protocol).",
	"verify_peer":       "In most cases the target host will be a directly attached neighbor in which case TLS verification can be ignored.",
	"username":          "Enter the web GUI username of the system entered above for synchronizing your configuration.",
	"password":          "Enter the web GUI password of the system entered above for synchronizing your configuration.",
	"sync_items":        "Select the services that should be sent to the other host.",
}

func hasyncResourceSchema() schema.Schema {
	password := hasyncString(hasyncHelp["password"])
	password.Sensitive = true
	return schema.Schema{
		MarkdownDescription: "High availability settings (System > High Availability > Settings): pfsync state synchronisation and configuration sync to a backup firewall. Saving reconfigures pfsync. " +
			"This is a singleton that always exists, so it must be imported before Terraform manages it (`id = \"" + hasyncID + "\"`); " +
			"destroying it only removes it from state.",

		Attributes: map[string]schema.Attribute{
			"disable_preempt":   hasyncBool(hasyncHelp["disable_preempt"] + " Sets `net.inet.carp.preempt`."),
			"disconnect_dialup": hasyncBool(hasyncHelp["disconnect_dialup"]),
			"pfsync_interface":  hasyncString(hasyncHelp["pfsync_interface"] + " Interface name, e.g. `opt1`; `\"\"` disables pfsync."),
			"pfsync_peer_ip":    hasyncString(hasyncHelp["pfsync_peer_ip"]),
			"pfsync_version": hasyncString(hasyncHelp["pfsync_version"]+" One of `1301` (OPNsense 24.1 or below), `1400` (24.7 to 26.1) or `1500` (26.7 or above).",
				stringvalidator.OneOf("1301", "1400", "1500")),
			"pfsync_defer":      hasyncBool(hasyncHelp["pfsync_defer"]),
			"synchronize_to_ip": hasyncString(hasyncHelp["synchronize_to_ip"] + " `\"\"` disables configuration sync."),
			"verify_peer":       hasyncBool(hasyncHelp["verify_peer"] + " Set to `true` to verify the backup firewall's TLS certificate."),
			"username":          hasyncString(hasyncHelp["username"]),
			"password":          password,
			"sync_items": schema.SetAttribute{
				MarkdownDescription: hasyncHelp["sync_items"] + " Section keys, e.g. `aliases`, `rules`, `nat`, `certs`, `dhcpd`, `cron`, `ipsec`, `wireguard`. Unset keeps the current value.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `" + hasyncID + "`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func hasyncDataSourceSchema() dschema.Schema {
	b := func(d string) dschema.BoolAttribute {
		return dschema.BoolAttribute{MarkdownDescription: d, Computed: true}
	}
	s := func(d string) dschema.StringAttribute {
		return dschema.StringAttribute{MarkdownDescription: d, Computed: true}
	}
	return dschema.Schema{
		MarkdownDescription: "High availability settings (System > High Availability > Settings). The password is not exposed.",

		Attributes: map[string]dschema.Attribute{
			"disable_preempt":   b(hasyncHelp["disable_preempt"]),
			"disconnect_dialup": b(hasyncHelp["disconnect_dialup"]),
			"pfsync_interface":  s(hasyncHelp["pfsync_interface"]),
			"pfsync_peer_ip":    s(hasyncHelp["pfsync_peer_ip"]),
			"pfsync_version":    s(hasyncHelp["pfsync_version"]),
			"pfsync_defer":      b(hasyncHelp["pfsync_defer"]),
			"synchronize_to_ip": s(hasyncHelp["synchronize_to_ip"]),
			"verify_peer":       b(hasyncHelp["verify_peer"] + " `true` when the backup firewall's TLS certificate is verified."),
			"username":          s(hasyncHelp["username"]),
			"sync_items":        dschema.SetAttribute{MarkdownDescription: hasyncHelp["sync_items"], ElementType: types.StringType, Computed: true},
			"id":                s("Always `" + hasyncID + "`."),
		},
	}
}

func convertHasyncSchemaToStruct(d *hasyncResourceModel) *core.HasyncSettings {
	return &core.HasyncSettings{
		DisablePreempt:  tools.BoolToString(d.DisablePreempt.ValueBool()),
		DisconnectPPPs:  tools.BoolToString(d.DisconnectDialup.ValueBool()),
		PfsyncInterface: api.SelectedMap(d.PfsyncInterface.ValueString()),
		PfsyncPeerIp:    d.PfsyncPeerIp.ValueString(),
		PfsyncVersion:   api.SelectedMap(d.PfsyncVersion.ValueString()),
		PfsyncDefer:     tools.BoolToString(d.PfsyncDefer.ValueBool()),
		SynchronizeToIp: d.SynchronizeToIp.ValueString(),
		VerifyPeer:      tools.BoolToString(d.VerifyPeer.ValueBool()),
		Username:        d.Username.ValueString(),
		Password:        d.Password.ValueString(),
		SyncItems:       api.SelectedMapList(tools.SetToStringSlice(d.SyncItems)),
	}
}

func convertHasyncStructToSchema(d *core.HasyncSettings) *hasyncResourceModel {
	return &hasyncResourceModel{
		DisablePreempt:   types.BoolValue(tools.StringToBool(d.DisablePreempt)),
		DisconnectDialup: types.BoolValue(tools.StringToBool(d.DisconnectPPPs)),
		PfsyncInterface:  types.StringValue(d.PfsyncInterface.String()),
		PfsyncPeerIp:     types.StringValue(d.PfsyncPeerIp),
		PfsyncVersion:    types.StringValue(d.PfsyncVersion.String()),
		PfsyncDefer:      types.BoolValue(tools.StringToBool(d.PfsyncDefer)),
		SynchronizeToIp:  types.StringValue(d.SynchronizeToIp),
		VerifyPeer:       types.BoolValue(tools.StringToBool(d.VerifyPeer)),
		Username:         types.StringValue(d.Username),
		Password:         types.StringValue(d.Password),
		SyncItems:        tools.StringSliceToSet(d.SyncItems),
		Id:               types.StringValue(hasyncID),
	}
}
