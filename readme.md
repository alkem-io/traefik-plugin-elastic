This repository includes a traefik middleware plugin that writes to a configurable elasticsearch endpoint and index.
The repository is based on [the official sample for developing traefik plugins.](https://github.com/traefik/plugindemo)

[![Build Status](https://github.com/alkem-io/traefik-plugin-elastic/workflows/Main/badge.svg?branch=master)](https://github.com/alkem-io/traefik-plugin-elastic/actions)

### Configuration

The following declaration (given here in YAML) defines traefik elastic plugin:

```yaml
# Static configuration

experimental:
  plugins:
    traefik-plugin-elastic:
      moduleName: github.com/alkem-io/traefik-plugin-elastic
      version: v0.1.1
```

Here is an example of a file provider dynamic configuration (given here in YAML), where the interesting part is the `http.middlewares` section:

```yaml
# Dynamic configuration

http:
  routers:
    my-router:
      rule: host(`demo.localhost`)
      service: service-foo
      entryPoints:
        - web
      middlewares:
        - traefik-plugin-middleware

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000
  
  middlewares:
    traefik-plugin-middleware:
      plugin:
        traefik-plugin-elastic:
          ElasticsearchURL: http://localhost:9200
          Message: Test Elasticsearch
          IndexName: test-index
          VerifyTLS: false
          Username: elastic
          Password: elastic_user_password
          APIKey: api_key

```
