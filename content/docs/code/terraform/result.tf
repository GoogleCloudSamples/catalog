output "domain_name_servers" {
  value = google_dns_managed_zone.default.name_servers
}

output "certificate_map_id" {
  value = google_certificate_manager_certificate_map.default.id
}
