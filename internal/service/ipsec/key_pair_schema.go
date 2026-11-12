package ipsec

import (
	"strings"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/ipsec"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type keyPairResourceModel struct {
	Name           types.String `tfsdk:"name"`
	KeyType        types.String `tfsdk:"key_type"`
	PublicKey      types.String `tfsdk:"public_key"`
	PrivateKey     types.String `tfsdk:"private_key"`
	KeySize        types.String `tfsdk:"key_size"`
	KeyFingerprint types.String `tfsdk:"key_fingerprint"`

	Id types.String `tfsdk:"id"`
}

func keyPairResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "IPsec raw public key pairs (VPN > IPsec > Key Pairs), for public-key authentication without certificates. " +
			"Generate the keys with the `tls_private_key` resource, for example.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Key pair name.",
				Required:            true,
			},
			"key_type": schema.StringAttribute{
				MarkdownDescription: "`rsa` or `ecdsa`; must match the keys.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf("rsa", "ecdsa")},
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "PEM public key (`-----BEGIN PUBLIC KEY-----`).",
				Required:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "PEM private key (PKCS#8, `-----BEGIN PRIVATE KEY-----`). Leave empty for a peer's key, where only the public key is known. Defaults to `\"\"`.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				Default:             stringdefault.StaticString(""),
			},
			"key_size": schema.StringAttribute{
				MarkdownDescription: "Key size in bits, computed by OPNsense.",
				Computed:            true,
			},
			"key_fingerprint": schema.StringAttribute{
				MarkdownDescription: "Key fingerprint, computed by OPNsense.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the key pair.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func keyPairDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: "IPsec raw public key pairs (VPN > IPsec > Key Pairs).",

		Attributes: map[string]dschema.Attribute{
			"id":              dschema.StringAttribute{MarkdownDescription: "UUID of the key pair.", Required: true},
			"name":            dschema.StringAttribute{MarkdownDescription: "Key pair name.", Computed: true},
			"key_type":        dschema.StringAttribute{MarkdownDescription: "Key type.", Computed: true},
			"public_key":      dschema.StringAttribute{MarkdownDescription: "PEM public key.", Computed: true},
			"private_key":     dschema.StringAttribute{MarkdownDescription: "PEM private key.", Computed: true, Sensitive: true},
			"key_size":        dschema.StringAttribute{MarkdownDescription: "Key size in bits.", Computed: true},
			"key_fingerprint": dschema.StringAttribute{MarkdownDescription: "Key fingerprint.", Computed: true},
		},
	}
}

func convertKeyPairSchemaToStruct(d *keyPairResourceModel) (*ipsec.IPsecKeyPair, error) {
	return &ipsec.IPsecKeyPair{
		Name:       d.Name.ValueString(),
		KeyType:    api.SelectedMap(d.KeyType.ValueString()),
		PublicKey:  d.PublicKey.ValueString(),
		PrivateKey: d.PrivateKey.ValueString(),
	}, nil
}

func convertKeyPairStructToSchema(d *ipsec.IPsecKeyPair) (*keyPairResourceModel, error) {
	return &keyPairResourceModel{
		Name:           types.StringValue(d.Name),
		KeyType:        types.StringValue(d.KeyType.String()),
		PublicKey:      types.StringValue(d.PublicKey),
		PrivateKey:     types.StringValue(d.PrivateKey),
		KeySize:        types.StringValue(d.KeySize),
		KeyFingerprint: types.StringValue(d.KeyFingerprint),
	}, nil
}

// mergeKeyPairWriteOnly keeps the configured PEM text when OPNsense returns
// the same key with different surrounding whitespace (e.g. without the
// trailing newline tls_private_key emits), so plans stay clean.
func mergeKeyPairWriteOnly(prior, read *keyPairResourceModel) {
	if prior == nil {
		return
	}
	samePEM := func(a, b types.String) bool {
		return !a.IsNull() && !a.IsUnknown() && strings.TrimSpace(a.ValueString()) == strings.TrimSpace(b.ValueString())
	}
	if samePEM(prior.PublicKey, read.PublicKey) {
		read.PublicKey = prior.PublicKey
	}
	if samePEM(prior.PrivateKey, read.PrivateKey) {
		read.PrivateKey = prior.PrivateKey
	}
}
