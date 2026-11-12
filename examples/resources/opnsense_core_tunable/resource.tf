# Drop packets to closed ports instead of answering with RST.
resource "opnsense_core_tunable" "blackhole_tcp" {
  tunable     = "net.inet.tcp.blackhole"
  value       = "2"
  description = "Silently drop TCP to closed ports"
}
