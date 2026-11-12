# About this fork

This is a fork of [browningluke/terraform-provider-opnsense](https://github.com/browningluke/terraform-provider-opnsense). The work lives on the `feat/gateways-assignments-wireguard-keys` branch. It depends on a matching client fork, [inkstom/opnsense-go](https://github.com/inkstom/opnsense-go) (branch `feat/routing-gateways-and-assignments`), wired in through a `replace` in `go.mod`. The replace can be dropped once upstream merges the client changes.

## What's new

| Resource / data source | API | Min OPNsense |
|---|---|---|
| `opnsense_routing_gateway` | `routing/settings/*Gateway` | 24.x |
| `opnsense_routing_gateway_group` | `routing/group_settings/*` | 26.7 |
| `opnsense_interfaces_assignment` | `interfaces/assignment/*` | 26.7.7 |
| `opnsense_wireguard_key_pair` (resource) | `wireguard/server/keyPair` | any |
| `opnsense_wireguard_psk` (resource) | `wireguard/client/psk` | any |
| `opnsense_wireguard_server`: `+carp_depend_on`, `+debug`, `+device` | | any |
| `opnsense_dnsmasq_range` | `dnsmasq/settings/*_range` | 25.x |
| `opnsense_dnsmasq_option` | `dnsmasq/settings/*_option` | 25.x |
| `opnsense_dnsmasq_settings` (singleton, import with `dnsmasq_settings`) | `dnsmasq/settings/get,set` | 25.x |
| `opnsense_auth_group` | `auth/group/*` | 25.x |
| `opnsense_auth_user` | `auth/user/*` | 25.x |
| `opnsense_auth_privilege` (authoritative per privilege) | `auth/priv/get_item,set_item` | 25.x |
| `opnsense_core_tunable` | `core/tunables/*` | 25.x |
| `opnsense_dhcrelay_destination` | `dhcrelay/settings/*_dest` | 25.x |
| `opnsense_dhcrelay_relay` | `dhcrelay/settings/*_relay` | 25.x |
| `opnsense_syslog_destination` (`levels` is multi-select; `+suppress`) | `syslog/settings/*Destination` | 25.x |
| `opnsense_monit_service` | `monit/settings/*Service` | 25.x |
| `opnsense_monit_test` | `monit/settings/*Test` | 25.x |
| `opnsense_monit_alert` | `monit/settings/*Alert` | 25.x |
| `opnsense_interfaces_loopback` | `interfaces/loopback_settings/*` | 25.x |
| `opnsense_interfaces_gif` | `interfaces/gif_settings/*` | 25.x |
| `opnsense_interfaces_gre` | `interfaces/gre_settings/*` | 25.x |
| `opnsense_interfaces_vxlan` | `interfaces/vxlan_settings/*` | 25.x |
| `opnsense_interfaces_neighbor` | `interfaces/neighbor_settings/*` | 25.x |
| `opnsense_firewall_group` | `firewall/group/*` | 24.7 |

End-to-end example: [`examples/complete/wireguard-site/main.tf`](examples/complete/wireguard-site/main.tf).

## Verified

On 2026-11-11 against a disposable OPNsense **26.7.9** VM (QEMU/KVM, single NIC, same layout as upstream CI):

- Provider acceptance tests, `TF_ACC=1`: all 16 tests in `routing`, `wireguard` and `interfaces` pass, including gateway groups, interface assignment and WireGuard keys.
- Client integration tests in `opnsense-go` for `routing`, `interfaces` and `wireguard` pass.
- `examples/complete/wireguard-site` applies cleanly (7 resources), shows no drift on a second plan, and destroys cleanly.

Two fixes came out of that run: a retry for OPNsense's 20-second gateway-list cache in `opnsense_routing_gateway_group`, and a missing `tunnel_address` in the WireGuard keys test.

## Build and use locally

```sh
make build-local   # installs as dev.io/browningluke/opnsense v1.0.0
```

## Testing

```sh
# Mock-API Terraform tests (no firewall needed; needs terraform on PATH)
go test ./internal/acctest/mockopnsense/ -v

# Live acceptance tests: use a disposable OPNsense 26.7.7+ VM, not production
export OPNSENSE_URI=https://fw-test OPNSENSE_API_KEY=... OPNSENSE_API_SECRET=... OPNSENSE_ALLOW_INSECURE=true
make testacc PKG=routing
make testacc PKG=interfaces TEST=TestAccInterfacesAssignmentResource
make testacc PKG=wireguard TEST=TestAccWireguardKeysResource
```
