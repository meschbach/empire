directory "system_units" {
  path = "/etc/systemd/system"
}

static_local_file {
  target = "/etc/systemd/system/consul.service"
  source = "consul.service"

  depends_on {
    type = "directory"
    name = "system_units"
  }
}
