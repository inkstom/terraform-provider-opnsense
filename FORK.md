# Differences from upstream

This provider is a fork of [browningluke/terraform-provider-opnsense](https://github.com/browningluke/terraform-provider-opnsense), branched from upstream `main` at `33c2a03` (2026-07-30). It is 64 commits ahead and nothing behind. The API client is a fork too: [inkstom/opnsense-go](https://github.com/inkstom/opnsense-go), pinned in `go.mod` with a `replace` to an exact commit.

Go module paths are still `github.com/browningluke/...`. That keeps upstream changes easy to merge. If the fork stops tracking upstream, rename the modules then.

## Added resources

Minimum OPNsense versions come from the controller and model history in opnsense/core. CI tests 26.1 and 26.7.

| Resource (each has a data source unless noted) | API | Min OPNsense |
|---|---|---|
| `opnsense_auth_group` | `auth/group/*` | 25.x |
| `opnsense_auth_user` | `auth/user/*` | 25.x |
| `opnsense_auth_privilege` (authoritative per privilege) | `auth/priv/get_item,set_item` | 25.x |
| `opnsense_captiveportal_zone` (`roaming`, `cors_*` need 26.7) | `captiveportal/settings/*Zone` | 24.7 |
| `opnsense_captiveportal_template` | `captiveportal/template/*` (26.7), `captiveportal/service/*Template` (earlier) | 24.7 |
| `opnsense_core_hasync` (singleton) | `core/hasync/*` | 24.7 |
| `opnsense_core_tunable` | `core/tunables/*` | 25.x |
| `opnsense_dhcrelay_destination` | `dhcrelay/settings/*_dest` | 25.x |
| `opnsense_dhcrelay_relay` | `dhcrelay/settings/*_relay` | 25.x |
| `opnsense_dnsmasq_range` | `dnsmasq/settings/*_range` | 25.x |
| `opnsense_dnsmasq_option` | `dnsmasq/settings/*_option` | 25.x |
| `opnsense_dnsmasq_settings` (singleton) | `dnsmasq/settings/get,set` | 25.x |
| `opnsense_dnsmasq_tag` | `dnsmasq/settings/*_tag` | 25.7 |
| `opnsense_dnsmasq_domain` | `dnsmasq/settings/*_domain` | 25.7 |
| `opnsense_dnsmasq_boot` | `dnsmasq/settings/*_boot` | 25.7 |
| `opnsense_firewall_group` | `firewall/group/*` | 24.7 |
| `opnsense_firewall_npt` | `firewall/npt/*` | 24.7 |
| `opnsense_interfaces_assignment` | `interfaces/assignment/*` | 26.7.7 |
| `opnsense_interfaces_bridge` | `interfaces/bridge_settings/*` | 24.7 |
| `opnsense_interfaces_loopback` | `interfaces/loopback_settings/*` | 25.x |
| `opnsense_interfaces_gif` | `interfaces/gif_settings/*` | 25.x |
| `opnsense_interfaces_gre` | `interfaces/gre_settings/*` | 25.x |
| `opnsense_interfaces_vxlan` | `interfaces/vxlan_settings/*` | 25.x |
| `opnsense_interfaces_neighbor` | `interfaces/neighbor_settings/*` | 25.x |
| `opnsense_ipsec_pool` | `ipsec/pools/*` | 24.7 |
| `opnsense_ipsec_key_pair` | `ipsec/key_pairs/*` | 24.7 |
| `opnsense_ipsec_manual_spd` | `ipsec/manual_spd/*` | 24.7 |
| `opnsense_ipsec_settings` (singleton, general section) | `ipsec/settings/get,set` | 24.7 |
| `opnsense_kea_ctrl_agent` (singleton) | `kea/ctrl_agent/get,set` | 24.7 |
| `opnsense_monit_service` | `monit/settings/*Service` | 25.x |
| `opnsense_monit_test` | `monit/settings/*Test` | 25.x |
| `opnsense_monit_alert` | `monit/settings/*Alert` | 25.x |
| `opnsense_monit_settings` (singleton) | `monit/settings/get,set` | 24.7 |
| `opnsense_routing_gateway` | `routing/settings/*Gateway` | 24.x |
| `opnsense_routing_gateway_group` | `routing/group_settings/*` | 26.7 |
| `opnsense_syslog_destination` | `syslog/settings/*Destination` | 25.x |
| `opnsense_syslog_settings` (singleton) | `syslog/settings/get,set` | 24.7 |
| `opnsense_trafficshaper_pipe` | `trafficshaper/settings/*Pipe` | 24.7 |
| `opnsense_trafficshaper_queue` | `trafficshaper/settings/*Queue` | 24.7 |
| `opnsense_trafficshaper_rule` | `trafficshaper/settings/*Rule` | 24.7 |
| `opnsense_wireguard_key_pair` (resource only) | `wireguard/server/keyPair` | any |
| `opnsense_wireguard_psk` (resource only) | `wireguard/client/psk` | any |

New data sources for resources upstream already had: `opnsense_ipsec_connection`, `_auth_local`, `_auth_remote`, `_child`, `_psk` and `_vti`.

New attributes on existing resources: `carp_depend_on`, `debug` and `device` on `opnsense_wireguard_server`.

Singletons always exist on the firewall, so Terraform can't create them. Import them first; the import ID is the resource name without `opnsense_` (for example `core_hasync` or `dnsmasq_settings`). Destroying one only removes it from state. Attributes left unset keep whatever the firewall has.

[`examples/complete/wireguard-site`](examples/complete/wireguard-site/main.tf) uses several of these together.

## Bugs fixed

In the provider:

- `opnsense_ipsec_auth_local` and `_auth_remote` showed an update on every plan when `auth_id` or `eap_id` was unset, which is normal for PSK authentication. Both now default to `""`.
- `opnsense_ipsec_psk.pre_shared_key` wasn't marked sensitive, so the key appeared in plan output.
- `opnsense_routing_gateway_group` failed when created in the same apply as its gateways, because OPNsense caches the gateway list for 20 seconds. It now retries for up to a minute.
- Descriptions on firewall NAT rules were limited to letters, digits, spaces and dots, which rejected rules OPNsense itself accepts. WireGuard servers no longer require a gateway when `disable_routes` is set.

In the client (inkstom/opnsense-go):

- On OPNsense 26.7.6 and later, `reconfigure` reports a failure when it can't stop a service that isn't running. With IPsec disabled, every IPsec write failed and left a half-saved object. The client treats that one message as success.
- Multi-select fields: auth group members and source networks, syslog levels and dnsmasq `ra_mode` were handled as single values, and the blank placeholder option in some lists was sent as a real value.
- Bridge updates sent an empty device name, which OPNsense rejects. It's now omitted.
- Captive portal template endpoints moved in 26.7. The client tries the new path and falls back to the old one.

In CI:

- A 26.7 acceptance lane runs the full suite after every merge to `main`, next to the existing 26.1 one.
- Tests for APIs that 26.1 lacks are skipped there with `acctest.MinVersionPreCheck` and run in the 26.7 lane.
- Change detection on pull requests fetches `main` with full history, so PRs based on an older `main` are tested correctly.
- New service packages are assigned to a test shard, so their tests actually run.

## OPNsense behaviour to know about

- **Traffic shaper:** a deleted pipe or queue stays active in dummynet until the shaper is reset (Firewall > Shaper > Reset, or `trafficshaper/service/flushreload`). The reset briefly clears every ipfw rule, captive portal included, so the provider doesn't do it for you.
- **HA sync:** `disable_preempt` is saved immediately but only takes effect at the next boot.
- **Kea control agent:** OPNsense only runs it while Kea DHCPv4 or DHCPv6 is enabled.
- **Captive portal templates:** OPNsense strips files identical to its default template before storing the zip, so the stored copy never matches the upload. The provider tracks the zip you configured and re-uploads it when that changes. Edits made on the firewall aren't detected.
- **Interface assignments:** assign a VLAN by its explicit device name. If you let OPNsense generate the name, the VLAN resource stores an empty `device`.

## Not covered

- Core snapshots: these are ZFS boot environments, and the CI image uses UFS, so they can't be tested.
- LAGG: needs spare physical ports, which the single-port test VMs don't have.
- Log viewer templates (`Diagnostics/Lvtemplate`).
- Plugins (ACME, HAProxy, FRR and others).

## Releases

There are no releases or Registry listing for the fork yet. Install from source with `make build-local`, as the README describes. The `v0.x` tags in this repository come from upstream.
