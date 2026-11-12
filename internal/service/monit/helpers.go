package monit

import (
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func strDefault(desc, def string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc + " Defaults to `\"" + def + "\"`.",
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(def),
	}
}

func boolDefault(desc string, def bool) schema.BoolAttribute {
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

func setDefault(desc string) schema.SetAttribute {
	return schema.SetAttribute{
		MarkdownDescription: desc + " Defaults to `[]`.",
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		Default:             setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
	}
}

func idAttr(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: desc,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func dsStr(desc string) dschema.StringAttribute {
	return dschema.StringAttribute{MarkdownDescription: desc, Computed: true}
}

func dsBool(desc string) dschema.BoolAttribute {
	return dschema.BoolAttribute{MarkdownDescription: desc, Computed: true}
}

func dsSet(desc string) dschema.SetAttribute {
	return dschema.SetAttribute{MarkdownDescription: desc, ElementType: types.StringType, Computed: true}
}

func dsID(desc string) dschema.StringAttribute {
	return dschema.StringAttribute{MarkdownDescription: desc, Required: true}
}
