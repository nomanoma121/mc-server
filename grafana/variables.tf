variable "grafana_auth_key" {
  description = "The API key for Grafana authentication"
  type        = string
}

variable "grafana_endpoint" {
  description = "The endpoint for Grafana"
  type        = string
  default     = "http://monitor.nomanoma-dev.com"
}

variable "discord_webhook_url" {
  description = "The Discord webhook URL for alerts"
  type        = string
}
