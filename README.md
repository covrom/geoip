# geoip

Here is a ready to deploy ip geo location server, It works for both ip v4 and ip v6,
However the underlying database is not that huge,
It makes use of the [OpenGeoFeed database](https://github.com/sapics/ip-location-db/blob/master/geo-whois-asn-country/README.md#geofeed-database-update-daily) found [here](https://github.com/sapics/ip-location-db),

# How to use

Create `docker-compose.yml` like that:

```yaml
version: "3.3"
services:
  geoip:
    container_name: geoip
    image: your_regisrty_ip:registry_port/geoip:latest
    build:
      context: .
      dockerfile: ./Dockerfile
    restart: always
    ports:
      - 8000:8000
```
Build and run:

```bash
docker compose --progress=plain build --no-cache --force-rm --pull geoip
docker compose push geoip
docker compose up -d --no-build --no-deps --force-recreate geoip
```

## Web UI

Open in browser: http://localhost:8000

## API Request

Request:

```bash
curl -X POST http://localhost:8000/getIpInfo/140.82.114.3
```

Response:

```json
{"ip":"140.82.114.3","country":"US"}
```
