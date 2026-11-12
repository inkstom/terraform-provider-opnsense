package ipsec_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccECDSAKeyPEM(t *testing.T) (pub, priv string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})),
		string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}))
}

func TestAccIpsecPoolKeyPairManualSpdResources(t *testing.T) {
	pub1, priv1 := testAccECDSAKeyPEM(t)
	pub2, _ := testAccECDSAKeyPEM(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIpsecPoolKeyPairSpdConfig(`["192.0.2.53", "192.0.2.54"]`, pub1, priv1, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_ipsec_pool.test", "addresses", "10.200.0.0/24"),
					resource.TestCheckResourceAttr("opnsense_ipsec_pool.test", "dns.#", "2"),
					resource.TestCheckResourceAttr("data.opnsense_ipsec_pool.test", "dns.#", "2"),
					resource.TestCheckResourceAttr("opnsense_ipsec_key_pair.test", "key_type", "ecdsa"),
					resource.TestCheckResourceAttr("opnsense_ipsec_key_pair.test", "key_size", "256"),
					resource.TestCheckResourceAttrSet("opnsense_ipsec_key_pair.test", "key_fingerprint"),
					resource.TestCheckResourceAttrPair("data.opnsense_ipsec_key_pair.test", "key_fingerprint", "opnsense_ipsec_key_pair.test", "key_fingerprint"),
					resource.TestCheckResourceAttr("opnsense_ipsec_manual_spd.test", "request_id", "100"),
					resource.TestCheckResourceAttr("opnsense_ipsec_manual_spd.test", "protocol", "esp"),
					resource.TestCheckResourceAttr("data.opnsense_ipsec_manual_spd.test", "source", "10.0.0.0/24"),
				),
			},
			{ResourceName: "opnsense_ipsec_pool.test", ImportState: true, ImportStateVerify: true},
			{
				ResourceName:      "opnsense_ipsec_key_pair.test",
				ImportState:       true,
				ImportStateVerify: true,
				// OPNsense stores the PEM without the trailing newline.
				ImportStateVerifyIgnore: []string{"public_key", "private_key"},
			},
			{ResourceName: "opnsense_ipsec_manual_spd.test", ImportState: true, ImportStateVerify: true},
			{
				// Swap to a public-key-only pair (a peer's key) and clear the pool's DNS.
				Config: testAccIpsecPoolKeyPairSpdConfig(`[]`, pub2, "", 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_ipsec_pool.test", "dns.#", "0"),
					resource.TestCheckResourceAttr("opnsense_ipsec_key_pair.test", "private_key", ""),
					resource.TestCheckResourceAttr("opnsense_ipsec_manual_spd.test", "request_id", "200"),
				),
			},
		},
	})
}

func testAccIpsecPoolKeyPairSpdConfig(dns, pub, priv string, reqid int) string {
	return fmt.Sprintf(`
resource "opnsense_ipsec_pool" "test" {
  name      = "tf-acc-pool"
  addresses = "10.200.0.0/24"
  dns       = %[1]s
}

resource "opnsense_ipsec_key_pair" "test" {
  name        = "tf-acc-keypair"
  key_type    = "ecdsa"
  public_key  = %[2]q
  private_key = %[3]q
}

resource "opnsense_ipsec_manual_spd" "test" {
  request_id  = %[4]d
  source      = "10.0.0.0/24"
  destination = "10.1.0.0/24"
  description = "tf acc"
}

data "opnsense_ipsec_pool" "test" {
  id = opnsense_ipsec_pool.test.id
}

data "opnsense_ipsec_key_pair" "test" {
  id = opnsense_ipsec_key_pair.test.id
}

data "opnsense_ipsec_manual_spd" "test" {
  id = opnsense_ipsec_manual_spd.test.id
}
`, dns, pub, priv, reqid)
}
