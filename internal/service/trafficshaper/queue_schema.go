package trafficshaper

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/trafficshaper"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type queueResourceModel struct {
	Enabled        types.Bool   `tfsdk:"enabled"`
	Number         types.String `tfsdk:"number"`
	Pipe           types.String `tfsdk:"pipe"`
	Weight         types.Int64  `tfsdk:"weight"`
	Mask           types.String `tfsdk:"mask"`
	Buckets        types.Int64  `tfsdk:"buckets"`
	CodelEnable    types.Bool   `tfsdk:"codel_enable"`
	CodelTarget    types.Int64  `tfsdk:"codel_target"`
	CodelInterval  types.Int64  `tfsdk:"codel_interval"`
	CodelEcnEnable types.Bool   `tfsdk:"codel_ecn_enable"`
	PieEnable      types.Bool   `tfsdk:"pie_enable"`
	Description    types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

func queueResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Traffic shaper queues (Firewall > Shaper > Queues): weighted shares of a pipe. Deleting one removes it from the configuration and the ruleset, but OPNsense leaves the now-unused queue in dummynet until the shaper is reset (Firewall > Shaper > Status > Reset) or the firewall reboots; the provider does not reset it because that briefly flushes every ipfw rule, including captive portal ones.",

		Attributes: map[string]schema.Attribute{
			"enabled": optBool("Enable this queue and its related rules.", true),
			"number":  computedString("Queue number assigned by OPNsense (from 10000)."),
			"pipe": schema.StringAttribute{
				MarkdownDescription: "Connected pipe for this queue. Set to the UUID (`id`) of an `opnsense_trafficshaper_pipe`.",
				Required:            true,
			},
			"weight": schema.Int64Attribute{
				MarkdownDescription: "Weight of this queue (`1`..`100`), used to prioritize within a pipe (1 is low, 100 is high). Defaults to `100`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(100),
				Validators:          []validator.Int64{int64validator.Between(1, 100)},
			},
			"mask": optString("Dynamic queue creation by source or destination address. Choose destination to evenly share every IP in destination field of rules the specified bandwidth. A pipe with 1Mbit e.g. would let 4 clients lend 250Kbit each. Normally this is used for download queues. Choose source to evenly share every IP in the source field of rules the specified bandwidth. Normally this is used for upload queues. Leave this value empty if you want to specify multiple queues with different weights. One of `none` (empty), `src-ip` (source), `dst-ip` (destination), `src-ip6` (source, IPv6) or `dst-ip6` (destination, IPv6).", "none",
				stringvalidator.OneOf(maskValues...)),
			"buckets":          optInt("Specifies the size of the hash table used for storing the various dynamic queues configured with the mask setting. Accepted range is 1-65535; unset uses the system default."),
			"codel_enable":     optBool("Enable CoDel active queue management. Cannot be combined with `pie_enable`.", false),
			"codel_target":     optInt("Minimum acceptable persistent queue delay (in ms), leave empty for default. Accepted range is 1-10000; unset uses 5."),
			"codel_interval":   optInt("Interval before dropping packets (in ms), leave empty for default. Accepted range is 1-10000; unset uses 100."),
			"codel_ecn_enable": optBool("Explicit Congestion Notification. Marks packets instead of dropping them.", false),
			"pie_enable":       optBool("Enable PIE active queue management. Cannot be combined with `codel_enable`.", false),
			"description":      optString("Description to identify this queue. OPNsense requires a value, and rules can only target queues that have one.", ""),
			"id":               computedString("UUID of the queue."),
		},
	}
}

func queueDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Traffic shaper queues (Firewall > Shaper > Queues).",

		Attributes: map[string]dschema.Attribute{
			"id":               dsID("UUID of the queue."),
			"enabled":          dsBool("Enable this queue and its related rules."),
			"number":           dsString("Queue number."),
			"pipe":             dsString("Connected pipe for this queue (UUID of the pipe)."),
			"weight":           dsInt("Weight of this queue (`1`..`100`), used to prioritize within a pipe (1 is low, 100 is high)."),
			"mask":             dsString("Dynamic queue creation by source or destination address. One of `none` (empty), `src-ip` (source), `dst-ip` (destination), `src-ip6` (source, IPv6) or `dst-ip6` (destination, IPv6)."),
			"buckets":          dsInt("Specifies the size of the hash table used for storing the various dynamic queues configured with the mask setting."),
			"codel_enable":     dsBool("Enable CoDel active queue management."),
			"codel_target":     dsInt("Minimum acceptable persistent queue delay (in ms)."),
			"codel_interval":   dsInt("Interval before dropping packets (in ms)."),
			"codel_ecn_enable": dsBool("Explicit Congestion Notification."),
			"pie_enable":       dsBool("Enable PIE active queue management."),
			"description":      dsString("Description to identify this queue."),
		},
	}
}

func convertQueueSchemaToStruct(d *queueResourceModel) (*trafficshaper.Queue, error) {
	return &trafficshaper.Queue{
		Enabled:        tools.BoolToString(d.Enabled.ValueBool()),
		Pipe:           api.SelectedMap(d.Pipe.ValueString()),
		Weight:         tools.Int64ToString(d.Weight.ValueInt64()),
		Mask:           api.SelectedMap(d.Mask.ValueString()),
		Buckets:        intOrEmpty(d.Buckets),
		CodelEnable:    tools.BoolToString(d.CodelEnable.ValueBool()),
		CodelTarget:    intOrEmpty(d.CodelTarget),
		CodelInterval:  intOrEmpty(d.CodelInterval),
		CodelEcnEnable: tools.BoolToString(d.CodelEcnEnable.ValueBool()),
		PieEnable:      tools.BoolToString(d.PieEnable.ValueBool()),
		Description:    d.Description.ValueString(),
	}, nil
}

func convertQueueStructToSchema(d *trafficshaper.Queue) (*queueResourceModel, error) {
	return &queueResourceModel{
		Enabled:        types.BoolValue(tools.StringToBool(d.Enabled)),
		Number:         types.StringValue(d.Number),
		Pipe:           types.StringValue(d.Pipe.String()),
		Weight:         types.Int64Value(tools.StringToInt64(d.Weight)),
		Mask:           types.StringValue(d.Mask.String()),
		Buckets:        tools.StringToInt64Null(d.Buckets),
		CodelEnable:    types.BoolValue(tools.StringToBool(d.CodelEnable)),
		CodelTarget:    tools.StringToInt64Null(d.CodelTarget),
		CodelInterval:  tools.StringToInt64Null(d.CodelInterval),
		CodelEcnEnable: types.BoolValue(tools.StringToBool(d.CodelEcnEnable)),
		PieEnable:      types.BoolValue(tools.StringToBool(d.PieEnable)),
		Description:    types.StringValue(d.Description),
	}, nil
}

func mergeQueueWriteOnly(prior, read *queueResourceModel) {}
