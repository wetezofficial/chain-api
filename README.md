Chain API
============================

### Requirements

- Go >= 1.16
- Redis >= 6.2.0
- MySQL >= 5.7

### Clone code

```shell
mkdir starnet
cd starnet

git clone https://github.com/wetezofficial/starnet
git clone https://github.com/wetezofficial/chain-api
```

### Build chain-api

```shell
cd chain-api
make build
# output file: ./build/prod/starnet-chain-api
```

### Run as a service on remote server

```shell
# copy ./build/prod/starnet-chain-api to remote server /usr/local/bin/starnet-chain-api
# copy ./starnet-chain-api.service to remote server /etc/systemd/system/starnet-chain-api.service
# copy ./config-example.toml to remote server /etc/starnet-chain-api/config.toml

# Run following commands to start starnet-chain-api.service on remote server
remote server # vim config.toml
remote server # systemctl start starnet-chain-api.service
remote server # systemctl enable starnet-chain-api.service
```
