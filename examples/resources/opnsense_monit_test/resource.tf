resource "opnsense_monit_test" "high_cpu" {
  name      = "HighProcessCPU"
  condition = "cpu is greater than 80% for 5 cycles"
  action    = "alert"
}
