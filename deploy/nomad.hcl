job "python-executor" {
  datacenters = ["dc1"]
  type        = "service"

  group "api" {
    count = 1

    network {
      port "http" {
        to = 8080
      }
    }

    service {
      name = "python-executor"
      port = "http"

      tags = [
        "api",
        "python-executor",
      ]

      check {
        type     = "http"
        path     = "/health"
        interval = "10s"
        timeout  = "2s"
      }
    }

    task "server" {
      driver = "docker"

      config {
        image = "python-executor:latest"

        ports = ["http"]

        # Required for Docker-in-Docker
        privileged = true

        # Mount Docker socket
        volumes = [
          "/var/run/docker.sock:/var/run/docker.sock"
        ]
      }

      env {
        PYEXEC_PORT            = "${NOMAD_PORT_http}"
        PYEXEC_HOST            = "0.0.0.0"
        PYEXEC_LOG_LEVEL       = "info"
        PYEXEC_DOCKER_SOCKET   = "/var/run/docker.sock"
        PYEXEC_DEFAULT_TIMEOUT = "300"
        PYEXEC_DEFAULT_MEMORY_MB  = "1024"
        PYEXEC_DEFAULT_DISK_MB    = "2048"
        PYEXEC_DEFAULT_CPU_SHARES = "1024"
        PYEXEC_DEFAULT_IMAGE   = "python:3.12-slim"

        # Optional: Consul integration
        PYEXEC_CONSUL_ADDR     = "${attr.unique.network.ip-address}:8500"
        PYEXEC_CONSUL_PREFIX   = "python-executor"
      }

      resources {
        cpu    = 500  # MHz
        memory = 512  # MB
      }

      # Graceful shutdown
      kill_timeout = "30s"
    }
  }
}
