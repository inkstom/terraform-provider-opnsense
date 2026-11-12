package trafficshaper

import (
	"strconv"

	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
	return schema.BoolAttribute{
		MarkdownDescription: desc + " Defaults to `" + strconv.FormatBool(def) + "`.",
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(def),
	}
}

// optInt is an optional number with no default: null leaves it to OPNsense.
func optInt(desc string) schema.Int64Attribute {
	return schema.Int64Attribute{MarkdownDescription: desc, Optional: true}
}

func optSet(desc string, def ...string) schema.SetAttribute {
	vals := make([]attr.Value, len(def))
	for i, d := range def {
		vals[i] = types.StringValue(d)
	}
	return schema.SetAttribute{
		MarkdownDescription: desc,
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Default:             setdefault.StaticValue(types.SetValueMust(types.StringType, vals)),
	}
}

func emptySetDefault() defaults.Set {
	return setdefault.StaticValue(tools.EmptySetValue(types.StringType))
}

func computedString(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func intOrEmpty(v types.Int64) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return tools.Int64ToString(v.ValueInt64())
}

func dsString(desc string) dschema.StringAttribute {
	return dschema.StringAttribute{MarkdownDescription: desc, Computed: true}
}

func dsBool(desc string) dschema.BoolAttribute {
	return dschema.BoolAttribute{MarkdownDescription: desc, Computed: true}
}

func dsInt(desc string) dschema.Int64Attribute {
	return dschema.Int64Attribute{MarkdownDescription: desc, Computed: true}
}

func dsSet(desc string) dschema.SetAttribute {
	return dschema.SetAttribute{MarkdownDescription: desc, ElementType: types.StringType, Computed: true}
}

func dsID(desc string) dschema.StringAttribute {
	return dschema.StringAttribute{MarkdownDescription: desc, Required: true}
}
