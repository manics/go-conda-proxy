# Go Conda Proxy

A Conda proxy written in Go.

## Build

Build binaries and run tests

```
make
```

## Usage

Create/update repodata cache and list of allowed files

```
./conda-parser -cfg config.yaml -force
```

Force an update of repodata cache and list of allowed files, ignore `max_age_minutes` in `config.yml`.

```
./conda-parser -cfg config.yaml -force
```

Run conda-proxy, this uses the `repodata-cache` directory/files created by `conda-parser`.

```
./conda-proxy -cfg config.yaml
```

## Development

```
go run cmd/conda-parser/main.go -cfg config.yaml
```

```
go run cmd/conda-proxy/main.go -cfg config.yaml
```
