# Terraform Provider for OPNsense

Manage OPNsense® firewall configuration with Terraform, through the OPNsense API.

This is a fork of [browningluke/terraform-provider-opnsense](https://github.com/browningluke/terraform-provider-opnsense), maintained by Thomas Inks. It covers almost all of the OPNsense core API: 89 resources and 91 data sources, and 85 of the 89 rows in the [coverage table](#current-api-coverage) below. Every resource has acceptance tests that run against real OPNsense 26.1 and 26.7 VMs in CI. [FORK.md](./FORK.md) lists what was added and fixed relative to upstream.

> [!WARNING]
> Versions before 1.0 may change resource schemas. Try changes against a non-production firewall first, and run `terraform plan` after importing existing objects to check for unexpected diffs.

- [Installation](#installation)
- [Example Usage](#example-usage)
- [OPNsense Versions](#opnsense-versions)
- [Moving from the Upstream Provider](#moving-from-the-upstream-provider)
- [Testing](#testing)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [Current API Coverage](#current-api-coverage)
  - [Core API](#core-api)
  - [Plugin API](#plugin-api)
- [License](#license)

## Installation

The provider isn't on the Terraform Registry. Build it from source (Go 1.24 or later) and Terraform picks it up from your local plugin directory:

```sh
git clone https://github.com/inkstom/terraform-provider-opnsense
cd terraform-provider-opnsense
make build-local   # installs dev.io/inkstom/opnsense v1.0.0 into ~/.terraform.d/plugins
```

```hcl
terraform {
  required_providers {
    opnsense = {
      source  = "dev.io/inkstom/opnsense"
      version = "1.0.0"
    }
  }
}
```

Run `make build-local` again after pulling to update it, then `terraform init -upgrade`.

## Example Usage

```hcl
provider "opnsense" {
  uri = "https://opnsense.example.com"

  # Either reference the API credentials literally
  api_key    = "<api key>"
  api_secret = "<api secret>"

  # Or set them in the environment:
  # export OPNSENSE_API_KEY="<api key>"
  # export OPNSENSE_API_SECRET="<api secret>"
}

resource "opnsense_firewall_filter" "allow_https" {
  enabled     = true
  description = "Allow inbound HTTPS traffic"

  interface = {
    interface = ["wan"]
  }

  filter = {
    action    = "pass"
    direction = "in"
    protocol  = "TCP"

    source = {
      net = "any"
    }

    destination = {
      net  = "192.168.1.100"
      port = "https"
    }

    log = true
  }
}
```

[`examples/complete/wireguard-site`](./examples/complete/wireguard-site/main.tf) builds a site-to-site WireGuard tunnel end to end: keys, instance, interface assignment, gateway and policy route.

## OPNsense Versions

CI tests against OPNsense 26.1 and 26.7. Most resources also work on 24.7 and 25.x, but only those two are tested. A few resources need a newer firewall than that:

- `opnsense_interfaces_assignment` needs 26.7.7, and so does anything that assigns an interface for you (bridges, captive portal zones on new VLANs).
- `opnsense_routing_gateway_group` needs 26.7.
- `opnsense_dnsmasq_tag`, `_domain` and `_boot` need 25.7.

Each resource's documentation states its minimum version when it has one, and [FORK.md](./FORK.md) has the full list. Some attributes only exist on newer releases (for example `suppress` on syslog destinations, or `roaming` on captive portal zones). On older firewalls, leave those unset and the provider won't send them.

## Moving from the Upstream Provider

If you already manage OPNsense with `browningluke/opnsense`, install this provider as above, change `source` in `required_providers`, and move existing state over:

```sh
terraform state replace-provider registry.terraform.io/browningluke/opnsense dev.io/inkstom/opnsense
terraform init
terraform plan
```

The plan should be empty. One known exception: `auth_id` and `eap_id` on `opnsense_ipsec_auth_local` and `opnsense_ipsec_auth_remote` now default to `""`. If you left them unset, the first plan shows a harmless in-place update and later plans are clean.

## Testing

Every pull request runs the acceptance tests for the packages it changes (all of them when `go.mod` changes) against a fresh OPNsense 26.1 VM in GitHub Actions. Every merge to `main` runs the full suite on both 26.1 and 26.7. The tests create, read, import, update and delete real objects through the API. Data sources are compared attribute by attribute with their resources. Tests for APIs that older releases lack are skipped there and run on 26.7.

To run them yourself, point them at a disposable OPNsense VM. They create and delete objects, and some change global settings before restoring them:

```sh
export OPNSENSE_URI=https://<test-vm> OPNSENSE_API_KEY=... OPNSENSE_API_SECRET=... OPNSENSE_ALLOW_INSECURE=true
make testacc PKG=routing
make testacc PKG=interfaces TEST=TestAccInterfacesAssignmentResource
```

A few end-to-end flows also run against a mock API, with no firewall needed (Terraform must be on your `PATH`):

```sh
go test ./internal/acctest/mockopnsense/
```

## Documentation

- **[Resource and data source reference](./docs/)**, generated from the schemas with `make docs`
- **[Examples](./examples/)**, one per resource
- **[FORK.md](./FORK.md)**: differences from upstream, OPNsense quirks the provider works around, and known limitations

## Contributing

Bug reports and pull requests go to [inkstom/terraform-provider-opnsense](https://github.com/inkstom/terraform-provider-opnsense/issues). Most features need a change in the API client too, [inkstom/opnsense-go](https://github.com/inkstom/opnsense-go). See the [Contributing Guide](./CONTRIBUTING.md) for the development setup.

## Current API Coverage

Coverage of the OPNsense API, by controller. The core table follows the API reference; settings that OPNsense still keeps on its older, non-API pages can't be managed by any provider.

- ✅ = Fully implemented
- 🚧 = Missing acceptance tests
- ❌ = Not implemented
- ➖ = Not applicable (no settings of its own, or read-only)

### Core API

| Module/Controller/Resource       | Resource | Data Source |
|----------------------------------|----------|-------------|
| `Auth/Group`                     | ✅        | ✅           |
| `Auth/Priv`                      | ✅        | ✅           |
| `Auth/User`                      | ✅        | ✅           |
| `Captiveportal/Settings`         | ➖        | ➖           |
| `Captiveportal/Service/Template` | ✅        | ✅           |
| `Captiveportal/Settings/Zone`    | ✅        | ✅           |
| `Core/Hasync`                    | ✅        | ✅           |
| `Core/Snapshots`                 | ❌        | ❌           |
| `Core/Tunables`                  | ✅        | ✅           |
| `Cron/Job`                       | ✅        | ✅           |
| `Dhcrelay/Settings`              | ➖        | ➖           |
| `Dhcrelay/Settings/Dest`         | ✅        | ✅           |
| `Dhcrelay/Settings/Relay`        | ✅        | ✅           |
| `Diagnostics/Interface`          | ➖        | ✅          |
| `Diagnostics/Lvtemplate`         | ❌        | ❌           |
| `Diagnostics/Lvtemplate/Item`    | ❌        | ❌           |
| `Dnsmasq/Settings`               | ✅        | ✅           |
| `Dnsmasq/Settings/Boot`          | ✅        | ✅           |
| `Dnsmasq/Settings/Domain`        | ✅        | ✅           |
| `Dnsmasq/Settings/Host`          | ✅        | ✅           |
| `Dnsmasq/Settings/Option`        | ✅        | ✅           |
| `Dnsmasq/Settings/Range`         | ✅        | ✅           |
| `Dnsmasq/Settings/Tag`           | ✅        | ✅           |
| `Firewall/Alias`                 | ✅        | ✅           |
| `Firewall/Category`              | ✅       | ✅          |
| `Firewall/Filter`                | ✅        | ✅           |
| `Firewall/Group`                 | ✅        | ✅           |
| `Firewall/NPTv6`                 | ✅        | ✅           |
| `Firewall/Source NAT`            | ✅        | ✅           |
| `Firewall/One-to-One NAT`        | ✅        | ✅           |
| `Interfaces/Assignment`          | ✅       | ✅          |
| `Interfaces/Bridge`              | ✅        | ✅           |
| `Interfaces/Gif`                 | ✅        | ✅           |
| `Interfaces/Gre`                 | ✅        | ✅           |
| `Interfaces/Lagg`                | ❌        | ❌           |
| `Interfaces/Loopback`            | ✅        | ✅           |
| `Interfaces/Neighbor`            | ✅        | ✅           |
| `Interfaces/Overview`            |          | ✅           |
| `Interfaces/Vip`                 | ✅        | ✅           |
| `Interfaces/Vlan`                | ✅        | ✅           |
| `Interfaces/Vxlan`               | ✅        | ✅           |
| `Ipsec/Settings`                 | ✅        | ✅           |
| `Ipsec/Connections/Local`        | ✅        | ✅           |
| `Ipsec/Connections/Remote`       | ✅        | ✅           |
| `Ipsec/Connections/Child`        | ✅        | ✅           |
| `Ipsec/Connections/Connection`   | ✅        | ✅           |
| `Ipsec/KeyPairs`                 | ✅        | ✅           |
| `Ipsec/ManualSpd`                | ✅        | ✅           |
| `Ipsec/Pools`                    | ✅        | ✅           |
| `Ipsec/Psk`                      | ✅        | ✅           |
| `Ipsec/Vti`                      | ✅        | ✅           |
| `Kea/CtrlAgent`                  | ✅        | ✅           |
| `Kea/Dhcpv4/Peer`                | ✅        | ✅           |
| `Kea/Dhcpv4/Reservation`         | ✅        | ✅           |
| `Kea/Dhcpv4/Subnet`              | ✅        | ✅           |
| `Kea/Dhcpv6/PD Pool`             | ✅        | ✅           |
| `Kea/Dhcpv6/Peer`                | ✅        | ✅           |
| `Kea/Dhcpv6/Reservation`         | ✅        | ✅           |
| `Kea/Dhcpv6/Subnet`              | ✅        | ✅           |
| `Monit/Settings`                 | ✅        | ✅           |
| `Monit/Settings/Alert`           | ✅        | ✅           |
| `Monit/Settings/Service`         | ✅        | ✅           |
| `Monit/Settings/Test`            | ✅        | ✅           |
| `Openvpn/Client Overwrites`      | ✅        | ✅           |
| `Openvpn/Instances`              | ✅        | ✅           |
| `Openvpn/Instances/Static Key`   | ✅        | ✅           |
| `Openvpn/Instances/Generate Key` | ✅ (ephemeral) |        |
| `Routes/Route`                   | ✅        | ✅           |
| `Routing/Gateway`                | ✅       | ✅          |
| `Routing/Gateway Group`          | ✅       | ✅          |
| `Syslog/Settings`                | ✅        | ✅           |
| `Syslog/Settings/Destination`    | ✅        | ✅           |
| `Trafficshaper/Pipe`             | ✅        | ✅           |
| `Trafficshaper/Queue`            | ✅        | ✅           |
| `Trafficshaper/Rule`             | ✅        | ✅           |
| `Trust/Settings`                 | ✅        | ✅           |
| `Trust/CA`                       | ✅        | ✅           |
| `Trust/Cert`                     | ✅        | ✅           |
| `Unbound/Settings`               | ✅        | ✅           |
| `Unbound/Settings/Domain Override` | ✅      | ✅           |
| `Unbound/Settings/Forward`       | ✅        | ✅           |
| `Unbound/Settings/Host Alias`    | ✅        | ✅           |
| `Unbound/Settings/Host Override` | ✅        | ✅           |
| `Unbound/Settings/ACL`           | ✅        | ✅           |
| `Wireguard/Settings`             | ✅        | ✅           |
| `Wireguard/Client`               | ✅       | ✅          |
| `Wireguard/Server`               | ✅       | ✅          |
| `Wireguard/Generate Key Pair`    | ✅       |             |
| `Wireguard/Generate PSK`         | ✅       |             |

### Plugin API

The following is a non-exhaustive list of the plugin APIs OPNsense supports. The table shows those which are 'highest priority'. [Open an issue](https://github.com/inkstom/terraform-provider-opnsense/issues) if you need a plugin that isn't listed.

| Plugin/Controller/Resource     | Resource | Data Source |
|--------------------------------|----------|-------------|
| `Acmeclient/Settings`          | ❌        | ❌           |
| `Acmeclient/Account`           | ❌        | ❌           |
| `Acmeclient/Validation`        | ❌        | ❌           |
| `Acmeclient/Certificates`      | ❌        | ❌           |
| `Acmeclient/Action`            | ❌        | ❌           |
| `Haproxy/Maintenance`          | ❌        | ❌           |
| `Haproxy/Settings`             | ❌        | ❌           |
| `Haproxy/Settings/Acl`         | ❌        | ❌           |
| `Haproxy/Settings/Action`      | ❌        | ❌           |
| `Haproxy/Settings/Backend`     | ❌        | ❌           |
| `Haproxy/Settings/Cpu`         | ❌        | ❌           |
| `Haproxy/Settings/Errorfile`   | ❌        | ❌           |
| `Haproxy/Settings/Fcgi`        | ❌        | ❌           |
| `Haproxy/Settings/Frontend`    | ❌        | ❌           |
| `Haproxy/Settings/Group`       | ❌        | ❌           |
| `Haproxy/Settings/Healthcheck` | ❌        | ❌           |
| `Haproxy/Settings/Lua`         | ❌        | ❌           |
| `Haproxy/Settings/Mapfile`     | ❌        | ❌           |
| `Haproxy/Settings/Server`      | ❌        | ❌           |
| `Haproxy/Settings/User`        | ❌        | ❌           |
| `Quagga/General`               | ❌        | ❌           |
| `Quagga/Bfd`                   | ❌        | ❌           |
| `Quagga/Bfd/Neighbor`          | ❌        | ❌           |
| `Quagga/Bgp`                   | ❌        | ❌           |
| `Quagga/Bgp/AS Path`           | 🚧       | 🚧          |
| `Quagga/Bgp/Community List`    | 🚧       | 🚧          |
| `Quagga/Bgp/Neighbor`          | 🚧       | 🚧          |
| `Quagga/Bgp/Peer Group`        | ❌        | ❌           |
| `Quagga/Bgp/Prefix List`       | 🚧       | 🚧          |
| `Quagga/Bgp/Route Map`         | 🚧       | 🚧          |
| `Quagga/Ospf`                  | ❌        | ❌           |
| `Quagga/Ospf/Interface`        | ❌        | ❌           |
| `Quagga/Ospf/Neighbor`         | ❌        | ❌           |
| `Quagga/Ospf/Network`          | ❌        | ❌           |
| `Quagga/Ospf/Prefix List`      | ❌        | ❌           |
| `Quagga/Ospf/Redistribution`   | ❌        | ❌           |
| `Quagga/Ospf/Route Map`        | ❌        | ❌           |
| `Quagga/Ospf6`                 | ❌        | ❌           |
| `Quagga/Ospf6/Interface`       | ❌        | ❌           |
| `Quagga/Ospf6/Neighbor`        | ❌        | ❌           |
| `Quagga/Ospf6/Network`         | ❌        | ❌           |
| `Quagga/Ospf6/Prefix List`     | ❌        | ❌           |
| `Quagga/Ospf6/Redistribution`  | ❌        | ❌           |
| `Quagga/Rip`                   | ❌        | ❌           |
| `Quagga/Static`                | ❌        | ❌           |
| `Quagga/Static/Route`          | ❌        | ❌           |

The complete OPNsense API documentation can be found at: [docs.opnsense.org](https://docs.opnsense.org/development/api.html)

## License

Mozilla Public License 2.0, the same as upstream. See [LICENSE](./LICENSE). The original provider is by Luke Browning and its contributors.
