# API実装にあたってのJACKのコマンド

## /control/server

jackdのサーバー制御を行う。

### GET

サーバーの状態を取得する。

```shell
jack_control status
```

### POST

サーバーを起動する。

```shell
jack_control start
jack_control ds dummy
jack_control dps rate 96000
jack_control dps period 128
jack_control dps nperiods 4
```

### DELETE

サーバーを停止する。

```shell
jack_control stop
jack_control exit
```

## /alsa/devices

ALSAのデバイスを取得する。

### GET

```shell
aplay -l
cat /proc/asound/card{N}/stream{M}
```

## /device/alsa{device_name}

ALSAデバイスをJACKに接続する。

### POST

```shell
jack_load {device_name} zalsa_in -i "-d {device_name} -r 44100 -p 128 -n 4"
jack_load {device_name} zalsa_out -i "-d {device_name} -r 44100 -p 128 -n 4"
```

### DELETE

```shell
jack_unload {device_name}
```

## /ports

JACKのポートを取得する。

### GET

```shell
jack_lsp
```

## /patches

### GET

```shell
jack_lsp -c
```

### POST

```shell
jack_connect {source} {destination}
```

### DELETE

```shell
jack_disconnect {source} {destination}
```