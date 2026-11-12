package trafficshaper

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/trafficshaper"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type pipeResourceModel struct {
	Enabled         types.Bool   `tfsdk:"enabled"`
	Number          types.String `tfsdk:"number"`
	Bandwidth       types.Int64  `tfsdk:"bandwidth"`
	BandwidthMetric types.String `tfsdk:"bandwidth_metric"`
	QueueSize       types.Int64  `tfsdk:"queue_size"`
	Mask            types.String `tfsdk:"mask"`
	Buckets         types.Int64  `tfsdk:"buckets"`
	Scheduler       types.String `tfsdk:"scheduler"`
	CodelEnable     types.Bool   `tfsdk:"codel_enable"`
	CodelTarget     types.Int64  `tfsdk:"codel_target"`
	CodelInterval   types.Int64  `tfsdk:"codel_interval"`
	CodelEcnEnable  types.Bool   `tfsdk:"codel_ecn_enable"`
	PieEnable       types.Bool   `tfsdk:"pie_enable"`
	FqcodelQuantum  types.Int64  `tfsdk:"fqcodel_quantum"`
	FqcodelLimit    types.Int64  `tfsdk:"fqcodel_limit"`
	FqcodelFlows    types.Int64  `tfsdk:"fqcodel_flows"`
	Delay           types.Int64  `tfsdk:"delay"`
	Description     types.String `tfsdk:"description"`

	Id types.String `tfsdk:"id"`
}

var maskValues = []string{"none", "src-ip", "dst-ip", "src-ip6", "dst-ip6"}

func pipeResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Traffic shaper pipes (Firewall > Shaper > Pipes): a bandwidth limit that queues and rules send traffic through. Deleting one removes it from the configuration and the ruleset, but OPNsense leaves the now-unused pipe in dummynet until the shaper is reset (Firewall > Shaper > Status > Reset) or the firewall reboots; the provider does not reset it because that briefly flushes every ipfw rule, including captive portal ones.",

		Attributes: map[string]schema.Attribute{
			"enabled": optBool("Enable the pipe.", true),
			"number":  computedString("Pipe number assigned by OPNsense (from 10000)."),
			"bandwidth": schema.Int64Attribute{
				MarkdownDescription: "Bandwidth, in `bandwidth_metric` units.",
				Required:            true,
				Validators:          []validator.Int64{int64validator.AtLeast(1)},
			},
			"bandwidth_metric": optString("Unit of `bandwidth`: `bit`, `Kbit`, `Mbit` or `Gbit`.", "Mbit",
				stringvalidator.OneOf("bit", "Kbit", "Mbit", "Gbit")),
			"queue_size": optInt("Queue size in slots (2-100); unset uses the system default."),
			"mask": optString("Dynamic per-host pipes: `none`, `src-ip`, `dst-ip`, `src-ip6` or `dst-ip6`.", "none",
				stringvalidator.OneOf(maskValues...)),
			"buckets": optInt("Hash table size for dynamic pipes; unset uses the system default."),
			"scheduler": optString("Scheduler: `\"\"` (weighted fair queueing), `fifo`, `rr`, `qfq`, `fq_codel` or `fq_pie`.", "",
				stringvalidator.OneOf("", "fifo", "rr", "qfq", "fq_codel", "fq_pie")),
			"codel_enable":     optBool("Enable CoDel active queue management.", false),
			"codel_target":     optInt("CoDel target delay in ms; unset uses 5."),
			"codel_interval":   optInt("CoDel interval in ms; unset uses 100."),
			"codel_ecn_enable": optBool("Mark packets with ECN instead of dropping them (CoDel/FQ-CoDel/FQ-PIE).", false),
			"pie_enable":       optBool("Enable PIE active queue management.", false),
			"fqcodel_quantum":  optInt("FQ-CoDel quantum in bytes; unset uses 1514."),
			"fqcodel_limit":    optInt("FQ-CoDel queue limit in packets; unset uses 10240."),
			"fqcodel_flows":    optInt("FQ-CoDel flow count; unset uses 1024."),
			"delay":            optInt("Added delay in ms, for testing; unset adds none."),
			"description":      optString("Description.", ""),
			"id":               computedString("UUID of the pipe."),
		},
	}
}

func pipeDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Traffic shaper pipes (Firewall > Shaper > Pipes).",

		Attributes: map[string]dschema.Attribute{
			"id":               dsID("UUID of the pipe."),
			"enabled":          dsBool("Whether the pipe is enabled."),
			"number":           dsString("Pipe number."),
			"bandwidth":        dsInt("Bandwidth."),
			"bandwidth_metric": dsString("Bandwidth unit."),
			"queue_size":       dsInt("Queue size in slots."),
			"mask":             dsString("Dynamic pipe mask."),
			"buckets":          dsInt("Hash table size."),
			"scheduler":        dsString("Scheduler."),
			"codel_enable":     dsBool("Whether CoDel is enabled."),
			"codel_target":     dsInt("CoDel target delay."),
			"codel_interval":   dsInt("CoDel interval."),
			"codel_ecn_enable": dsBool("Whether ECN is enabled."),
			"pie_enable":       dsBool("Whether PIE is enabled."),
			"fqcodel_quantum":  dsInt("FQ-CoDel quantum."),
			"fqcodel_limit":    dsInt("FQ-CoDel limit."),
			"fqcodel_flows":    dsInt("FQ-CoDel flows."),
			"delay":            dsInt("Added delay."),
			"description":      dsString("Description."),
		},
	}
}

func convertPipeSchemaToStruct(d *pipeResourceModel) (*trafficshaper.Pipe, error) {
	return &trafficshaper.Pipe{
		Enabled:         tools.BoolToString(d.Enabled.ValueBool()),
		Bandwidth:       tools.Int64ToString(d.Bandwidth.ValueInt64()),
		BandwidthMetric: api.SelectedMap(d.BandwidthMetric.ValueString()),
		Queue:           intOrEmpty(d.QueueSize),
		Mask:            api.SelectedMap(d.Mask.ValueString()),
		Buckets:         intOrEmpty(d.Buckets),
		Scheduler:       api.SelectedMap(d.Scheduler.ValueString()),
		CodelEnable:     tools.BoolToString(d.CodelEnable.ValueBool()),
		CodelTarget:     intOrEmpty(d.CodelTarget),
		CodelInterval:   intOrEmpty(d.CodelInterval),
		CodelEcnEnable:  tools.BoolToString(d.CodelEcnEnable.ValueBool()),
		PieEnable:       tools.BoolToString(d.PieEnable.ValueBool()),
		FqcodelQuantum:  intOrEmpty(d.FqcodelQuantum),
		FqcodelLimit:    intOrEmpty(d.FqcodelLimit),
		FqcodelFlows:    intOrEmpty(d.FqcodelFlows),
		Delay:           intOrEmpty(d.Delay),
		Description:     d.Description.ValueString(),
	}, nil
}

func convertPipeStructToSchema(d *trafficshaper.Pipe) (*pipeResourceModel, error) {
	return &pipeResourceModel{
		Enabled:         types.BoolValue(tools.StringToBool(d.Enabled)),
		Number:          types.StringValue(d.Number),
		Bandwidth:       types.Int64Value(tools.StringToInt64(d.Bandwidth)),
		BandwidthMetric: types.StringValue(d.BandwidthMetric.String()),
		QueueSize:       tools.StringToInt64Null(d.Queue),
		Mask:            types.StringValue(d.Mask.String()),
		Buckets:         tools.StringToInt64Null(d.Buckets),
		Scheduler:       types.StringValue(d.Scheduler.String()),
		CodelEnable:     types.BoolValue(tools.StringToBool(d.CodelEnable)),
		CodelTarget:     tools.StringToInt64Null(d.CodelTarget),
		CodelInterval:   tools.StringToInt64Null(d.CodelInterval),
		CodelEcnEnable:  types.BoolValue(tools.StringToBool(d.CodelEcnEnable)),
		PieEnable:       types.BoolValue(tools.StringToBool(d.PieEnable)),
		FqcodelQuantum:  tools.StringToInt64Null(d.FqcodelQuantum),
		FqcodelLimit:    tools.StringToInt64Null(d.FqcodelLimit),
		FqcodelFlows:    tools.StringToInt64Null(d.FqcodelFlows),
		Delay:           tools.StringToInt64Null(d.Delay),
		Description:     types.StringValue(d.Description),
	}, nil
}

func mergePipeWriteOnly(prior, read *pipeResourceModel) {}
