resource "grafana_dashboard" "mc-monitor-metrics" {
  folder = grafana_folder.mc_monitor_folder.uid
  for_each = fileset("${path.module}/dashboards/", "*.json")
  config_json = file("${path.module}/dashboards/${each.value}")
}
