job "edge-app-system" {

  type = "system"

  region      = "global"
  datacenters = ["edge"]

  constraint {
    attribute = "${meta.platform}"
    value     = "hostinger"
  }

  group "edge-app" {

    disconnect {
      lost_after = "6h"
      replace    = false
      reconcile  = "best_score"
    }

    task "edge-app" {

      driver = "raw_exec"

      env {
        DB_PATH                  = "/opt/edge-data.db"
        APP_INTERVAL_RANDOM_DATA = 5
        APP_INTERVAL_SEND_DATA   = 10
        PLATFORM                 = "${meta.platform}"
        BACKEND_URI              = "https://demo-nomad-edge.madebydian.com/transaction"
      }

      config {
        command = "edge-app_v0.0.1_${attr.kernel.name}_${attr.cpu.arch}"
      }

      artifact {
        source = "https://github.com/rahadiangg/demo-nomad-edge/releases/download/v0.0.1/edge-app_v0.0.1_${attr.kernel.name}_${attr.cpu.arch}"
      }
    }
  }
}