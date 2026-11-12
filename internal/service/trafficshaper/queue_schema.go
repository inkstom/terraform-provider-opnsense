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
			"enabled": optBool("Enable the queue.", true),
			"number":  computedString("Queue number assigned by OPNsense (from 10000)."),
			"pipe": schema.StringAttribute{
				MarkdownDescription: "UUID of the `opnsense_trafficshaper_pipe` this queue shares.",
				Required:            true,
			},
			"weight": schema.Int64Attribute{
				MarkdownDescription: "Share of the pipe relative to its other queues (1-100). Defaults to `100`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(100),
				Validators:          []validator.Int64{int64validator.Between(1, 100)},
			},
			"mask": optString("Dynamic per-host queues: `none`, `src-ip`, `dst-ip`, `src-ip6` or `dst-ip6`.", "none",
				stringvalidator.OneOf(maskValues...)),
			"buckets":          optInt("Hash table size for dynamic queues; unset uses the system default."),
			"codel_enable":     optBool("Enable CoDel active queue management.", false),
			"codel_target":     optInt("CoDel target delay in ms; unset uses 5."),
			"codel_interval":   optInt("CoDel interval in ms; unset uses 100."),
			"codel_ecn_enable": optBool("Mark packets with ECN instead of dropping them.", false),
			"pie_enable":       optBool("Enable PIE active queue management.", false),
			"description":      optString("Description.", ""),
			"id":               computedString("UUID of the queue."),
		},
	}
}

func queueDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Traffic shaper queues (Firewall > Shaper > Queues).",

		Attributes: map[string]dschema.Attribute{
			"id":               dsID("UUID of the queue."),
			"enabled":          dsBool("Whether the queue is enabled."),
			"number":           dsString("Queue number."),
			"pipe":             dsString("UUID of the pipe."),
			"weight":           dsInt("Weight."),
			"mask":             dsString("Dynamic queue mask."),
			"buckets":          dsInt("Hash table size."),
			"codel_enable":     dsBool("Whether CoDel is enabled."),
			"codel_target":     dsInt("CoDel target delay."),
			"codel_interval":   dsInt("CoDel interval."),
			"codel_ecn_enable": dsBool("Whether ECN is enabled."),
			"pie_enable":       dsBool("Whether PIE is enabled."),
			"description":      dsString("Description."),
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
