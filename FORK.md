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

End-to-end example: [`examples/complete/wireguard-site/main.tf`](examples/complete/wireguard-site/main.tf).

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
