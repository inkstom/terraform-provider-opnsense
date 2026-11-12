package wireguard

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Key material is generated once on the firewall (wg genkey / wg genpsk) and then lives only in
// Terraform state. Read is a no-op; changing `keepers` forces a new key.

var _ resource.Resource = &keyPairResource{}
var _ resource.ResourceWithConfigure = &keyPairResource{}
var _ resource.Resource = &pskResource{}
var _ resource.ResourceWithConfigure = &pskResource{}

func configureClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) opnsense.Client {
	if req.ProviderData == nil {
		return nil
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *opnsense.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return nil
	}
	return opnsense.NewClient(apiClient)
}

func keepersAttribute() schema.MapAttribute {
	return schema.MapAttribute{
		MarkdownDescription: "Arbitrary map of values that, when changed, will trigger generation of a new key.",
		Optional:            true,
		ElementType:         types.StringType,
		PlanModifiers:       []planmodifier.Map{mapplanmodifier.RequiresReplace()},
	}
}

func computedKey(desc string, sensitive bool) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Computed:            true,
		Sensitive:           sensitive,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

// --- opnsense_wireguard_key_pair ---------------------------------------------------------------

type keyPairResourceModel struct {
	Keepers    types.Map    `tfsdk:"keepers"`
	PublicKey  types.String `tfsdk:"public_key"`
	PrivateKey types.String `tfsdk:"private_key"`
	Id         types.String `tfsdk:"id"`
}

func newKeyPairResource() resource.Resource { return &keyPairResource{} }

type keyPairResource struct{ client opnsense.Client }

func (r *keyPairResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wireguard_key_pair"
}

func (r *keyPairResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates a WireGuard key pair on the firewall (`/api/wireguard/server/keyPair`). " +
			"Nothing is stored on OPNsense; the keys live in Terraform state, so protect your state. " +
			"Feed the result into `opnsense_wireguard_server`.",
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"keepers":     keepersAttribute(),
			"public_key":  computedKey("Base64 public key.", false),
			"private_key": computedKey("Base64 private key.", true),
			"id":          computedKey("Same as `public_key`.", false),
		},
	}
}

func (r *keyPairResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := configureClient(req, resp); c != nil {
		r.client = c
	}
}

func (r *keyPairResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data keyPairResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kp, err := r.client.Wireguard().ServiceGenKeyPair(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to generate key pair, got error: %s", err))
		return
	}
	if kp.PubKey == "" || kp.PrivKey == "" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Key pair generation returned no key (status: %q)", kp.Status))
		return
	}

	data.PublicKey = types.StringValue(kp.PubKey)
	data.PrivateKey = types.StringValue(kp.PrivKey)
	data.Id = data.PublicKey
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *keyPairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *keyPairResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Only `keepers` is configurable and it forces replacement; nothing to do.
	var data keyPairResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *keyPairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// --- opnsense_wireguard_psk ----------------------------------------------------------------------

type pskResourceModel struct {
	Keepers types.Map    `tfsdk:"keepers"`
	PSK     types.String `tfsdk:"psk"`
	Id      types.String `tfsdk:"id"`
}

func newPskResource() resource.Resource { return &pskResource{} }

type pskResource struct{ client opnsense.Client }

func (r *pskResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wireguard_psk"
}

func (r *pskResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates a WireGuard pre-shared key on the firewall (`/api/wireguard/client/psk`). " +
			"Nothing is stored on OPNsense; the key lives in Terraform state. Feed it into `opnsense_wireguard_client.psk`.",
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"keepers": keepersAttribute(),
			"psk":     computedKey("Base64 pre-shared key.", true),
			"id":      computedKey("Random identifier for this key.", false),
		},
	}
}

func (r *pskResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := configureClient(req, resp); c != nil {
		r.client = c
	}
}

func (r *pskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data pskResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Wireguard().ServiceGenPsk(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to generate PSK, got error: %s", err))
		return
	}
	if res.Psk == "" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("PSK generation returned no key (status: %q)", res.Status))
		return
	}

	data.PSK = types.StringValue(res.Psk)
	// Don't expose the secret as the ID; derive a stable, non-secret identifier.
	data.Id = types.StringValue(fmt.Sprintf("psk-%x", fnv32(res.Psk)))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *pskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *pskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data pskResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *pskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// fnv32 is a tiny non-cryptographic hash; it only needs to be stable, not secret-preserving
// beyond not echoing the key itself. 32 bits leaks nothing usable about a 256-bit key.
func fnv32(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}
