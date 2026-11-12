resource "opnsense_monit_test" "high_cpu" {
  name      = "HighProcessCPU"
  condition = "cpu is greater than 80% for 5 cycles"
  action    = "alert"
}

# Watch Unbound and restart it if it dies.
resource "opnsense_monit_service" "unbound" {
  name    = "unbound"
  type    = "process"
  pidfile = "/var/run/unbound.pid"
  start   = "/usr/local/sbin/configctl unbound start"
  stop    = "/usr/local/sbin/configctl unbound stop"
  tests   = [opnsense_monit_test.high_cpu.id]
}
