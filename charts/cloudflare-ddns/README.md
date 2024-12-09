# Cloudflare DDNS Helm Chart

This Helm chart deploys a Cloudflare Dynamic DNS (DDNS) updater to your Kubernetes cluster. It uses a Docker image to periodically update DNS records in your Cloudflare account based on the external IPs of your Kubernetes nodes.

## Prerequisites

- Kubernetes cluster
- Helm 3.x or later
- Cloudflare API token with necessary permissions

## Installation

1. **Add the Helm repository**:
   ```bash
   helm repo add cddns https://cloudflare-ddns.aureum.cloud
   helm repo update
   ```

2. **Install the chart**:
   ```bash
   helm install my-cloudflare-ddns cddns/cloudflare-ddns \
     --values values.yaml
   ```

   Replace `values.yaml` with your custom configuration file.

## Configuration

The following table describes the configurable parameters of the chart and their default values:

| Parameter                        | Description                                   | Default                                             |
|----------------------------------|-----------------------------------------------|-----------------------------------------------------|
| `image.repository`               | Docker image repository                       | `ghcr.io/aureum-cloud/cloudflare-ddns`             |
| `image.tag`                      | Docker image tag                              | `latest`                                            |
| `image.pullPolicy`               | Image pull policy                             | `IfNotPresent`                                      |
| `externalConfigMap`              | Name of an external ConfigMap (optional)      | `""`                                                |
| `externalSecret`                 | Name of an external Secret (optional)         | `""`                                                |
| `config.configJson`              | Configuration JSON for Cloudflare DDNS        | See default in [values.yaml](#default-values)       |
| `secrets.cloudflareApiToken`     | Cloudflare API token                          | `""`                                                |

### Example `values.yaml`

```yaml
config:
  configJson: |
    {
      "ipv4_enabled": true,
      "ipv6_enabled": false,
      "ttl": 300,
      "zones": [
        {
          "id": "your_zone_id_here",
          "name": "your_root_domain_here",
          "update_root_domain": false,
          "subdomains": [
            "your_subdomain_here"
          ]
        }
      ]
    }

secrets:
  cloudflareApiToken: "your_cloudflare_api_token_here"
```

### Using `externalConfigMap`

An `externalConfigMap` allows you to provide configuration data from an existing Kubernetes ConfigMap. For example, if you have a `config.json` file:

1. **Create the ConfigMap**:
   ```bash
   kubectl create configmap cloudflare-ddns-config --from-file=config.json
   ```

2. **Reference the ConfigMap in `values.yaml`**:
   ```yaml
   externalConfigMap: cloudflare-ddns-config
   ```

   The `config.json` file should follow the expected format, such as:
   ```json
   {
     "ipv4_enabled": true,
     "ipv6_enabled": false,
     "ttl": 300,
     "zones": [
       {
         "id": "your_zone_id_here",
         "name": "your_root_domain_here",
         "update_root_domain": false,
         "subdomains": [
           "your_subdomain_here"
         ]
       }
     ]
   }
   ```

## Secrets Management

The `externalSecret` value is an alternative to authenticate with the Cloudflare API. For security, consider using Kubernetes secrets:

1. **Create a Kubernetes Secret**:
   ```bash
   kubectl create secret generic cloudflare-secrets --from-literal=CLOUDFLARE_API_TOKEN=<your-cloudflare-api-token>
   ```

2. **Reference the Secret in `values.yaml`**:
   ```yaml
   externalSecret: cloudflare-secrets
   ```