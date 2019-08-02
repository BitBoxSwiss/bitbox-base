---
layout: default
title: NGINX
parent: Supporting Applications
nav_order: 730
---
## NGINX: Reverse Proxy

To avoid exposing applications directly to the network, all network communication is proxied through [NGINX](https://www.nginx.com/), a powerful open-source web server and proxy.
The BitBox Base only uses the reverse-proxy functionality, routing both HTTP and TCP traffic.
By only exposing a single, battle-tested server to the network, the attack surface is minimized significantly.
Together with the strict `iptables` firewall rules, all unknown communication patterns are ignored.

### Installation

NGINX is installed using the standard Armbian package and the configuration for the default web-server landing page is deleted.

```bash
apt install -y nginx
rm -f /etc/nginx/sites-enabled/default
```

### Configuration

The reverse-proxy rules are stored in the main configuration file `/etc/nginx/nginx.conf`.
Additional services (called "sites") are configured using individual `.conf` files in the directory `/etc/nginx/sites-available/`.
To enable individual sites, a symbolic link is created in the directory `/etc/nginx/sites-enabled/`, pointing to the corresponding `.conf` file.
The main configuration file in turn includes all `conf` files from this directory.

#### Main configuration

The file `/etc/nginx/nginx.conf` contains the main configuration.

* **General configuration**: information is available in the [NGINX documentation](https://nginx.org/en/docs/ngx_core_module.html)
* **TCP reverse-proxy** is used for the Electrum server: as `electrs` does not provide TLS encryption, NGINX is used to route TCP communication from the insecure internal port `50001` over the public TLS port `50002` which uses TLS with a self-signed SSL certificate.
  For Bitcoin testnet, ports `60001`/`51002` are used.
* **HTTP reverse-proxy** is used for specific web content like the Grafana dashboard.
  The top block specifies the general configuration like MIME types and logfile locations.
  Specific configurations are included from site-specific `*.conf` files.

```nginx
user www-data;
worker_processes 1;
pid /run/nginx.pid;
include /etc/nginx/modules-enabled/*.conf;

events {
  worker_connections 768;
}

stream {
  ssl_certificate /data/ssl/nginx-selfsigned.crt;
  ssl_certificate_key /data/ssl/nginx-selfsigned.key;
  ssl_session_cache shared:SSL:1m;
  ssl_session_timeout 4h;
  ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
  ssl_prefer_server_ciphers on;

  upstream electrs {
    server 127.0.0.1:50001;
  }
  server {
    listen 50002 ssl;
    proxy_pass electrs;
  }

  upstream electrs_testnet {
    server 127.0.0.1:60001;
  }
  server {
    listen 51002 ssl;
    proxy_pass electrs_testnet;
  }
}

http {
  include /etc/nginx/mime.types;
  default_type application/octet-stream;
  access_log /var/log/nginx/access.log;
  error_log /var/log/nginx/error.log;
  include /etc/nginx/sites-enabled/*.conf;
}
```

#### Grafana dashboard

Grafana serves a monitoring dashboard over its own built-in web-server, which can be exposed publicly with the `/etc/nginx/sites-available/grafana.conf` configuration.
To enable this rule, a symbolic link is created in the directory `/etc/nginx/sites-enabled/`, and deleted to disable it.

This reverse-proxy rule causes NGINX to

* listen for HTTP requests on port 80
* redirect the root folder to the Grafana dashboard URI (by returning HTTP status 301: Moved Permanently)
* route all traffic from public port 80 to Grafana's internal port 3000

```nginx
server {
  listen 80;
  location = / {
    return 301 http://$host/info/d/BitBoxBase/;
  }
  location /info/ {
    proxy_pass http://127.0.0.1:3000/;
  }
}
```

### Service management

After installation, NGINX is already configured to be managed by systemd, with its own service configuration located at `/lib/systemd/system/nginx.service`.
Service files provided by package installation should not be altered manually.
Systemd provides a method to extend/overwrite configuration values by using a drop-in file.
The standard configuration is extended in `cat /etc/systemd/system/nginx.service.d/override.conf` to start it after the Grafana service and reliably restart the application.

```bash
[Unit]
After=grafana-server.service startup-checks.service

[Service]
Restart=always
RestartSec=10
PrivateTmp=true
```
