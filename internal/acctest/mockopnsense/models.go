package mockopnsense

import "fmt"

func boolDefaults(names ...string) map[string]string {
	m := map[string]string{}
	for _, n := range names {
		m[n] = "0"
	}
	return m
}

func merged(ms ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range ms {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

// GatewayRoute mirrors /api/routing/settings (Gateways.xml).
func GatewayRoute() Route {
	f := map[string]Kind{"interface": Select, "ipprotocol": Select}
	for _, n := range []string{"disabled", "name", "descr", "gateway", "defaultgw", "fargw", "monitor_disable",
		"monitor_noroute", "monitor_killstates", "monitor_killstates_priority", "monitor", "force_down", "nosync",
		"priority", "weight", "latencylow", "latencyhigh", "losslow", "losshigh", "interval", "time_period",
		"loss_interval", "data_length"} {
		f[n] = Plain
	}
	return Route{
		Add: "/routing/settings/addGateway", Get: "/routing/settings/getGateway",
		Set: "/routing/settings/setGateway", Del: "/routing/settings/delGateway",
		Reconfigure: "/routing/settings/reconfigure",
		Model: Model{
			Monad:  "gateway_item",
			Fields: f,
			Defaults: merged(boolDefaults("disabled", "defaultgw", "fargw", "monitor_noroute", "monitor_killstates",
				"monitor_killstates_priority", "force_down", "nosync"),
				map[string]string{"interface": "wan", "ipprotocol": "inet", "monitor_disable": "1", "priority": "255", "weight": "1"}),
			Options: map[string][]string{"interface": {"wan", "lan"}, "ipprotocol": {"inet", "inet6"}},
		},
	}
}

// GatewayGroupRoute mirrors /api/routing/group_settings (GatewayGroups.xml).
func GatewayGroupRoute() Route {
	return Route{
		Add: "/routing/group_settings/add", Get: "/routing/group_settings/get",
		Set: "/routing/group_settings/set", Del: "/routing/group_settings/del",
		Reconfigure: "/routing/group_settings/reconfigure",
		Model: Model{
			Monad: "gateway_group",
			Fields: map[string]Kind{"name": Plain, "item": Multi, "item2": Multi, "item3": Multi, "item4": Multi,
				"item5": Multi, "trigger": Select, "poolopts": Select, "descr": Plain},
			Defaults: map[string]string{"trigger": "down"},
			Options:  map[string][]string{"trigger": {"down", "downloss", "downlatency", "downlosslatency"}},
		},
	}
}

// AssignmentRoute mirrors /api/interfaces/assignment (NetworkInterface.xml): items keyed by optN.
func AssignmentRoute() Route {
	f := map[string]Kind{"if": Select, "type4": Select, "type6": Select, "gateway": Select, "gatewayv6": Select,
		"track6-interface": Select}
	for _, n := range []string{"descr", "identifier", "enable", "lock", "blockpriv", "blockbogons", "spoofmac", "promisc",
		"mtu", "mss", "gateway_interface", "ipaddr", "ipaddrv6", "dhcphostname", "track6-prefix-id"} {
		f[n] = Plain
	}
	return Route{
		Add: "/interfaces/assignment/addItem", Get: "/interfaces/assignment/getItem",
		Set: "/interfaces/assignment/setItem", Del: "/interfaces/assignment/delItem",
		Reconfigure: "/interfaces/assignment/reconfigure",
		Model: Model{
			Monad:    "interface",
			Fields:   f,
			Defaults: merged(boolDefaults("blockpriv", "blockbogons", "promisc", "gateway_interface"), map[string]string{"enable": "1", "lock": "1"}),
			Volatile: []string{"identifier"},
			NewID: func(s *store) string {
				for n := 1; ; n++ {
					id := fmt.Sprintf("opt%d", n)
					if _, taken := s.items[id]; !taken {
						return id
					}
				}
			},
			OnAdd: func(s *store, id string, item map[string]string) { item["identifier"] = id },
		},
	}
}

// WireguardServerRoute mirrors /api/wireguard/server (Server.xml).
func WireguardServerRoute() Route {
	f := map[string]Kind{"dns": Multi, "tunneladdress": Multi, "peers": Multi, "carp_depend_on": Select}
	for _, n := range []string{"enabled", "name", "instance", "pubkey", "privkey", "port", "mtu", "disableroutes", "gateway", "debug"} {
		f[n] = Plain
	}
	return Route{
		Add: "/wireguard/server/addServer", Get: "/wireguard/server/getServer",
		Set: "/wireguard/server/setServer", Del: "/wireguard/server/delServer",
		Reconfigure: "/wireguard/service/reconfigure",
		Model: Model{
			Monad:    "server",
			Fields:   f,
			Defaults: merged(boolDefaults("disableroutes", "debug"), map[string]string{"enabled": "1"}),
			Volatile: []string{"instance"},
			OnAdd: func(s *store, id string, item map[string]string) {
				item["instance"] = fmt.Sprintf("%d", s.seq)
				s.seq++
			},
		},
	}
}
