# Cloudflare DDNS

**Cloudflare DDNS** is a Go application designed to update DNS records on Cloudflare according to your current public IPv4 and/or IPv6 addresses. It reads configuration from a JSON file, retrieves the current IP addresses, and updates DNS records as needed.

## Features

- **Fetches Public IPs**: Retrieves your public IPv4 and IPv6 addresses.
- **Configurable TTL**: Allows setting a custom Time-To-Live (TTL) for DNS records.
- **Automatic Updates**: Supports periodic updates of DNS records with the latest IP address.

## Requirements

- **Cloudflare API Token**: Set an environment variable `CLOUDFLARE_API_TOKEN` with your Cloudflare API token.

## Configuration

The script expects a JSON configuration file located at `/etc/config/config.json`. Below is an example configuration:

```json
{
  "ipv4_enabled": true,
  "ipv6_enabled": false,
  "ttl": 300,
  "zones": [
    {
      "id": "d9e353b268c23a9737f5b40b31f92a6f",
      "name": "aureum.cloud",
      "update_root_domain": true,
      "subdomains": [
        "www"
      ]
    }
  ]
}
```

### Configuration Fields

- **ipv4_enabled**: Set to `true` to enable IPv4 address updates.
- **ipv6_enabled**: Set to `true` to enable IPv6 address updates.
- **ttl**: Time-To-Live for the DNS records (in seconds). If set to less than 30, defaults to 300 seconds.
- **zones**: An array of DNS zones to update.
  - **id**: Cloudflare zone ID.
  - **name**: Domain name of the zone.
  - **update_root_domain**: Set to `true` to update the root domain.
  - **subdomains**: List of subdomains to update.

## Usage

### Single Update

To perform a single update of DNS records:

```bash
go run main.go
```

### Repeated Updates

To repeatedly update DNS records at the interval specified by the TTL:

```bash
go run main.go --repeat
```

### Running with Docker

To run this script in a Docker container, use the following command. Be sure to replace `<path-to-config>` with the path to your local configuration file and `<your_cloudflare_api_token>` with your actual Cloudflare API token:

```bash
docker run -e CLOUDFLARE_API_TOKEN=<your_cloudflare_api_token> -v <path-to-config>:/etc/config/config.json ghcr.io/aureum-cloud/cloudflare-ddns:latest
```
