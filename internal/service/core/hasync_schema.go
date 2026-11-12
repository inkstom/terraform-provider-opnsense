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

const hasyncID = "core_hasync"

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

func hasyncResourceSchema() schema.Schema {
	password := hasyncString("Password of `username` on the backup firewall.")
	password.Sensitive = true
	return schema.Schema{
		MarkdownDescription: "High availability settings (System > High Availability > Settings): pfsync state synchronisation and configuration sync to a backup firewall. Saving reconfigures pfsync. " +
			"This is a singleton that always exists, so it must be imported before Terraform manages it (`id = \"" + hasyncID + "\"`); " +
			"destroying it only removes it from state.",

		Attributes: map[string]schema.Attribute{
			"disable_preempt":   hasyncBool("Disable CARP preemption, so a recovered master doesn't take its addresses back. OPNsense applies this at boot (`net.inet.carp.preempt`), not on save."),
			"disconnect_dialup": hasyncBool("Disconnect dial-up (PPP) interfaces when this firewall becomes a CARP backup."),
			"pfsync_interface":  hasyncString("Interface to synchronise firewall states over, e.g. a dedicated `opt1`; `\"\"` disables pfsync."),
			"pfsync_peer_ip":    hasyncString("Peer address for pfsync, instead of multicast."),
			"pfsync_version": hasyncString("pfsync protocol version: `1301`, `1400` or `1500`. Both firewalls must match.",
				stringvalidator.OneOf("1301", "1400", "1500")),
			"pfsync_defer":      hasyncBool("Defer packets until the peer acknowledges the state (for active/active setups)."),
			"synchronize_to_ip": hasyncString("Address of the backup firewall to push configuration to; `\"\"` disables configuration sync."),
			"verify_peer":       hasyncBool("Verify the backup firewall's TLS certificate."),
			"username":          hasyncString("User on the backup firewall used for configuration sync."),
			"password":          password,
			"sync_items": schema.SetAttribute{
				MarkdownDescription: "Sections to synchronise, e.g. `aliases`, `rules`, `nat`, `certs`, `dhcpd`, `cron`, `ipsec`, `wireguard`. Unset keeps the current value.",
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
			"disable_preempt":   b("Whether CARP preemption is disabled."),
			"disconnect_dialup": b("Whether dial-up interfaces disconnect on backup."),
			"pfsync_interface":  s("pfsync interface."),
			"pfsync_peer_ip":    s("pfsync peer address."),
			"pfsync_version":    s("pfsync protocol version."),
			"pfsync_defer":      b("Whether pfsync defers packets."),
			"synchronize_to_ip": s("Backup firewall address for configuration sync."),
			"verify_peer":       b("Whether the backup's certificate is verified."),
			"username":          s("Configuration sync user."),
			"sync_items":        dschema.SetAttribute{MarkdownDescription: "Synchronised sections.", ElementType: types.StringType, Computed: true},
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
