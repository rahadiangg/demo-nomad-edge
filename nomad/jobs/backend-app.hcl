job "backend-app" {

  region      = "global"
  datacenters = ["oc-sg"]

  group "backend-app" {

    network {
      mode = "bridge"
      port "http" {
        to = 8080
      }
    }

    task "backend-app" {

      driver = "exec"

      env {
        DB_USER = "demo-nomad-edge"
        DB_PASS = "demo-nomad-edge"
        DB_NAME = "demo-nomad-edge"
      }

      template {
        env         = true
        destination = "secrets/.env"
        change_mode = "restart"
        data        = <<EOH
DB_HOST={{ range nomadService "postgresql" }}{{ .Address }}{{- end }}
DB_PORT= {{ range nomadService "postgresql" }}{{ .Port }}{{- end }}
EOH
      }

      config {
        command = "local/backend-app_v0.0.1_${attr.kernel.name}_${attr.cpu.arch}"
      }

      artifact {
        source = "https://github.com/rahadiangg/demo-nomad-edge/releases/download/v0.0.1/backend-app_v0.0.1_${attr.kernel.name}_${attr.cpu.arch}"
      }

      service {
        provider = "nomad"
        port     = "http"

        tags = [
          "traefik.enable=true",
          "traefik.http.routers.backend-app.rule=Host(`demo-nomad-edge.madebydian.com`)",
        ]
      }

    }
  }
}