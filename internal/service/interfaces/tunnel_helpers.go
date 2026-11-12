package interfaces

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Small schema helpers for the loopback, gif, gre, vxlan and neighbor resources.

func optStr(desc, def string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc + " Defaults to `\"" + def + "\"`.",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(def),
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

func reqStr(desc string) schema.StringAttribute {
	return schema.StringAttribute{MarkdownDescription: desc, Required: true}
}

// computedStr is assigned by OPNsense (device names) and stable afterwards.
func computedStr(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func dsStr(desc string) dschema.StringAttribute {
	return dschema.StringAttribute{MarkdownDescription: desc, Computed: true}
}

func dsBool(desc string) dschema.BoolAttribute {
	return dschema.BoolAttribute{MarkdownDescription: desc, Computed: true}
}

func dsID(desc string) dschema.StringAttribute {
	return dschema.StringAttribute{MarkdownDescription: desc, Required: true}
}
