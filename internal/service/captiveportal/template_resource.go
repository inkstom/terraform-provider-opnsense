package captiveportal

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Templates are zip files. OPNsense strips files identical to its default
// template before storing one, so the stored content never equals what was
// uploaded; the resource keeps the configured content in state and does not
// diff it against the firewall.

var _ resource.Resource = &templateResource{}
var _ resource.ResourceWithConfigure = &templateResource{}
var _ resource.ResourceWithImportState = &templateResource{}

func newTemplateResource() resource.Resource {
	return &templateResource{}
}

type templateResource struct {
	client opnsense.Client
}

type templateResourceModel struct {
	Name          types.String `tfsdk:"name"`
	ContentBase64 types.String `tfsdk:"content_base64"`
	ContentSha256 types.String `tfsdk:"content_sha256"`
	FileId        types.String `tfsdk:"file_id"`
	Id            types.String `tfsdk:"id"`
}

func (r *templateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_captiveportal_template"
}

func (r *templateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Captive portal templates (Services > Captive Portal > Templates): a zip with a custom login page (`index.html` plus assets). " +
			"Reference one by `id` in `opnsense_captiveportal_zone.template`. OPNsense removes files that match its default template before storing " +
			"the zip, so changes made on the firewall are not detected; Terraform re-uploads when `content_base64` changes.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Template name.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"content_base64": schema.StringAttribute{
				MarkdownDescription: "The template zip, base64 encoded, e.g. `filebase64(\"portal.zip\")`.",
				Required:            true,
			},
			"content_sha256": schema.StringAttribute{
				MarkdownDescription: "SHA-256 of the decoded zip as configured.",
				Computed:            true,
			},
			"file_id": schema.StringAttribute{
				MarkdownDescription: "File ID OPNsense assigned to the template.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "UUID of the template.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *templateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	r.client = opnsense.NewClient(apiClient)
}

func contentSha256(b64 string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("content_base64 is not valid base64: %w", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func (r *templateResource) save(ctx context.Context, data *templateResourceModel, uuid string) error {
	sum, err := contentSha256(data.ContentBase64.ValueString())
	if err != nil {
		return err
	}
	id, err := r.client.Captiveportal().SaveTemplate(ctx, uuid, data.Name.ValueString(), data.ContentBase64.ValueString())
	if err != nil {
		return err
	}
	tpl, err := r.client.Captiveportal().GetTemplate(ctx, id)
	if err != nil {
		return err
	}
	if tpl == nil {
		return fmt.Errorf("template %s not found after saving", id)
	}
	data.Id = types.StringValue(id)
	data.FileId = types.StringValue(tpl.FileId)
	data.ContentSha256 = types.StringValue(sum)
	return nil
}

func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data templateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.save(ctx, &data, ""); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create captive portal template, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tpl, err := r.client.Captiveportal().GetTemplate(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read captive portal template, got error: %s", err))
		return
	}
	if tpl == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Name = types.StringValue(tpl.Name)
	data.FileId = types.StringValue(tpl.FileId)
	// After import there is no configured content; adopt what the firewall stores.
	if data.ContentBase64.IsNull() || data.ContentBase64.ValueString() == "" {
		data.ContentBase64 = types.StringValue(tpl.Content)
		if sum, err := contentSha256(tpl.Content); err == nil {
			data.ContentSha256 = types.StringValue(sum)
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state templateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.save(ctx, &data, state.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update captive portal template, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Captiveportal().DeleteTemplate(ctx, data.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete captive portal template, got error: %s", err))
	}
}

func (r *templateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Data source

var _ datasource.DataSource = &templateDataSource{}
var _ datasource.DataSourceWithConfigure = &templateDataSource{}

func newTemplateDataSource() datasource.DataSource {
	return &templateDataSource{}
}

type templateDataSource struct {
	client opnsense.Client
}

type templateDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	FileId        types.String `tfsdk:"file_id"`
	ContentBase64 types.String `tfsdk:"content_base64"`
}

func (d *templateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_captiveportal_template"
}

func (d *templateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Captive portal templates (Services > Captive Portal > Templates).",
		Attributes: map[string]dschema.Attribute{
			"id":             dschema.StringAttribute{MarkdownDescription: "UUID of the template.", Required: true},
			"name":           dschema.StringAttribute{MarkdownDescription: "Template name.", Computed: true},
			"file_id":        dschema.StringAttribute{MarkdownDescription: "File ID.", Computed: true},
			"content_base64": dschema.StringAttribute{MarkdownDescription: "Stored zip (without files identical to the default template), base64 encoded.", Computed: true},
		},
	}
}

func (d *templateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	d.client = opnsense.NewClient(apiClient)
}

func (d *templateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data templateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tpl, err := d.client.Captiveportal().GetTemplate(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read captive portal template, got error: %s", err))
		return
	}
	if tpl == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Captive portal template %s does not exist.", data.Id.ValueString()))
		return
	}
	data.Name = types.StringValue(tpl.Name)
	data.FileId = types.StringValue(tpl.FileId)
	data.ContentBase64 = types.StringValue(tpl.Content)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
