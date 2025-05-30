resource "grafana_contact_point" "discord_alert" {
  name = "Discord Alert"

  discord {
    webhook_url = var.discord_webhook_url
    title = "Grafana Alert"
    message = "Alert: {{ .CommandLabels.alertname }}\n\n{{ .Alerts.Firing | len }} firing alerts"
  }
}
