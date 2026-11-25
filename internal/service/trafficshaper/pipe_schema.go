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
			"enabled": optBool("Enable this pipe and its related queues and rules.", true),
			"number":  computedString("Pipe number assigned by OPNsense (from 10000)."),
			"bandwidth": schema.Int64Attribute{
				MarkdownDescription: "Total bandwidth for this pipe, in `bandwidth_metric` units.",
				Required:            true,
				Validators:          []validator.Int64{int64validator.AtLeast(1)},
			},
			"bandwidth_metric": optString("Unit of `bandwidth`: `bit`, `Kbit`, `Mbit` or `Gbit`. The OPNsense UI defaults to `Kbit`.", "Mbit",
				stringvalidator.OneOf("bit", "Kbit", "Mbit", "Gbit")),
			"queue_size": optInt("Number of dynamic queues, leave empty for default. Sets the dummynet queue size in slots (2-100); unset uses the system default."),
			"mask": optString("Dynamic pipe creation by source or destination address. Choose destination to give every IP in destination field of rules the specified bandwidth. A pipe with 1Mbit e.g. would let 3 clients lend 1Mbit each so 3Mbit max. Normally this is used for download pipes. Choose source to give every IP in the source field of rules the specified bandwidth. Normally this is used for upload pipes. Leave this value empty if you want to create a pipe with a fixed bandwidth. One of `none` (empty), `src-ip` (source), `dst-ip` (destination), `src-ip6` (source, IPv6) or `dst-ip6` (destination, IPv6).", "none",
				stringvalidator.OneOf(maskValues...)),
			"buckets": optInt("Specifies the size of the hash table used for storing the various dynamic pipes configured with the mask setting. Accepted range is 1-65535; unset uses the system default."),
			"scheduler": optString("Specify the scheduling algorithm to use. One of `\"\"` (Weighted Fair Queueing), `fifo` (FIFO), `rr` (Deficit Round Robin), `qfq` (QFQ), `fq_codel` (FlowQueue-CoDel) or `fq_pie` (FlowQueue-PIE).", "",
				stringvalidator.OneOf("", "fifo", "rr", "qfq", "fq_codel", "fq_pie")),
			"codel_enable":     optBool("Enable CoDel active queue management. Cannot be combined with `pie_enable`.", false),
			"codel_target":     optInt("Minimum acceptable persistent queue delay (in ms), leave empty for default. Accepted range is 1-10000; unset uses 5."),
			"codel_interval":   optInt("Interval before dropping packets (in ms), leave empty for default. Accepted range is 1-10000; unset uses 100."),
			"codel_ecn_enable": optBool("Explicit Congestion Notification. Marks packets instead of dropping them, with CoDel, FQ-CoDel or FQ-PIE.", false),
			"pie_enable":       optBool("Enable PIE active queue management. Cannot be combined with `codel_enable`.", false),
			"fqcodel_quantum":  optInt("The number of bytes a queue can serve before being moved to the tail of old queues list (bytes), leave empty for defaults. Applies to the `fq_codel` scheduler; accepted range is 1-65535; unset uses 1514."),
			"fqcodel_limit":    optInt("The hard size limit of all queues managed by this instance, leave empty for defaults. Counted in packets; applies to the `fq_codel` scheduler; accepted range is 1-65535; unset uses 10240."),
			"fqcodel_flows":    optInt("The number of flow queues that are created and managed, leave empty for defaults. Applies to the `fq_codel` scheduler; accepted range is 1-65535; unset uses 1024."),
			"delay":            optInt("Add delay in ms to this pipe. Accepted range is 1-3000; unset adds none."),
			"description":      optString("Description to identify this pipe. OPNsense requires a value, and rules can only target pipes that have one.", ""),
			"id":               computedString("UUID of the pipe."),
		},
	}
}

func pipeDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "Traffic shaper pipes (Firewall > Shaper > Pipes).",

		Attributes: map[string]dschema.Attribute{
			"id":               dsID("UUID of the pipe."),
			"enabled":          dsBool("Enable this pipe and its related queues and rules."),
			"number":           dsString("Pipe number."),
			"bandwidth":        dsInt("Total bandwidth for this pipe, in `bandwidth_metric` units."),
			"bandwidth_metric": dsString("Unit of `bandwidth`: `bit`, `Kbit`, `Mbit` or `Gbit`."),
			"queue_size":       dsInt("Number of dynamic queues (dummynet queue size in slots). Null when OPNsense uses the default."),
			"mask":             dsString("Dynamic pipe creation by source or destination address. One of `none` (empty), `src-ip` (source), `dst-ip` (destination), `src-ip6` (source, IPv6) or `dst-ip6` (destination, IPv6)."),
			"buckets":          dsInt("Specifies the size of the hash table used for storing the various dynamic pipes configured with the mask setting."),
			"scheduler":        dsString("Specify the scheduling algorithm to use. Empty means Weighted Fair Queueing."),
			"codel_enable":     dsBool("Enable CoDel active queue management."),
			"codel_target":     dsInt("Minimum acceptable persistent queue delay (in ms)."),
			"codel_interval":   dsInt("Interval before dropping packets (in ms)."),
			"codel_ecn_enable": dsBool("Explicit Congestion Notification."),
			"pie_enable":       dsBool("Enable PIE active queue management."),
			"fqcodel_quantum":  dsInt("The number of bytes a queue can serve before being moved to the tail of old queues list (bytes)."),
			"fqcodel_limit":    dsInt("The hard size limit of all queues managed by this instance."),
			"fqcodel_flows":    dsInt("The number of flow queues that are created and managed."),
			"delay":            dsInt("Add delay in ms to this pipe."),
			"description":      dsString("Description to identify this pipe."),
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
