# API実装にあたってのJACK・ALSAのコマンド

## JACKのサーバー制御
APIサーバーの起動と同時にJACKサーバーも起動し、APIサーバーの終了と同時にJACKサーバーも終了するようにする。

### サーバーの状態を取得

```shell
$ jack_control status
```

### サーバーを起動

```shell
$ jack_control start
$ jack_control ds dummy
$ jack_control dps rate 96000
$ jack_control dps period 128
```

### サーバーを停止

```shell
$ jack_control stop
$ jack_control exit
```

## ALSAのデバイス周りの操作
入出力の対応パラメータについては、APIサーバー起動時に取得しJACKに接続する前にDBに保存しておく。
カード名をハッシュ化してIDとして扱う。

### 入力側対応パラメータを取得

```shell
$ aplay -l
**** List of PLAYBACK Hardware Devices ****
card 0: Loopback [Loopback], device 0: Loopback PCM [Loopback PCM]
  Subdevices: 8/8
  Subdevice #0: subdevice #0
  Subdevice #1: subdevice #1
  Subdevice #2: subdevice #2
  Subdevice #3: subdevice #3
  Subdevice #4: subdevice #4
  Subdevice #5: subdevice #5
  Subdevice #6: subdevice #6
  Subdevice #7: subdevice #7
card 0: Loopback [Loopback], device 1: Loopback PCM [Loopback PCM]
  Subdevices: 8/8
  Subdevice #0: subdevice #0
  Subdevice #1: subdevice #1
  Subdevice #2: subdevice #2
  Subdevice #3: subdevice #3
  Subdevice #4: subdevice #4
  Subdevice #5: subdevice #5
  Subdevice #6: subdevice #6
  Subdevice #7: subdevice #7
card 1: Headphones [bcm2835 Headphones], device 0: bcm2835 Headphones [bcm2835 Headphones]
  Subdevices: 8/8
  Subdevice #0: subdevice #0
  Subdevice #1: subdevice #1
  Subdevice #2: subdevice #2
  Subdevice #3: subdevice #3
  Subdevice #4: subdevice #4
  Subdevice #5: subdevice #5
  Subdevice #6: subdevice #6
  Subdevice #7: subdevice #7
card 2: vc4hdmi0 [vc4-hdmi-0], device 0: MAI PCM i2s-hifi-0 [MAI PCM i2s-hifi-0]
  Subdevices: 1/1
  Subdevice #0: subdevice #0
card 3: vc4hdmi1 [vc4-hdmi-1], device 0: MAI PCM i2s-hifi-0 [MAI PCM i2s-hifi-0]
  Subdevices: 1/1
  Subdevice #0: subdevice #0
card 4: Rubix24 [Rubix24], device 0: USB Audio [USB Audio]
  Subdevices: 1/1
  Subdevice #0: subdevice #0
card 5: UAC2Gadget [UAC2_Gadget], device 0: UAC2 PCM [UAC2 PCM]
  Subdevices: 1/1
  Subdevice #0: subdevice #0
card 6: UAC2Gadget_1 [UAC2_Gadget], device 0: UAC2 PCM [UAC2 PCM]
  Subdevices: 1/1
  Subdevice #0: subdevice #0

$ aplay --dump-hw-params -D {device_name} /dev/zero
Playing raw data '/dev/zero' : Unsigned 8 bit, Rate 8000 Hz, Mono
HW Params of device "hw:4,0":
--------------------
ACCESS:  MMAP_INTERLEAVED RW_INTERLEAVED
FORMAT:  S32_LE
SUBFORMAT:  STD
SAMPLE_BITS: 32
FRAME_BITS: 128
CHANNELS: 4
RATE: [44100 192000]
PERIOD_TIME: [125 1000000]
PERIOD_SIZE: [6 192000]
PERIOD_BYTES: [96 3072000]
PERIODS: [2 1024]
BUFFER_TIME: (62 2000000]
BUFFER_SIZE: [12 384000]
BUFFER_BYTES: [192 6144000]
TICK_TIME: ALL
--------------------
aplay: set_params:1352: Sample format non available
Available formats:
- S32_LE
```

### 出力側対応パラメータを取得

```shell
$ arecord -l
**** List of CAPTURE Hardware Devices ****
card 0: Loopback [Loopback], device 0: Loopback PCM [Loopback PCM]
  Subdevices: 8/8
  Subdevice #0: subdevice #0
  Subdevice #1: subdevice #1
  Subdevice #2: subdevice #2
  Subdevice #3: subdevice #3
  Subdevice #4: subdevice #4
  Subdevice #5: subdevice #5
  Subdevice #6: subdevice #6
  Subdevice #7: subdevice #7
card 0: Loopback [Loopback], device 1: Loopback PCM [Loopback PCM]
  Subdevices: 8/8
  Subdevice #0: subdevice #0
  Subdevice #1: subdevice #1
  Subdevice #2: subdevice #2
  Subdevice #3: subdevice #3
  Subdevice #4: subdevice #4
  Subdevice #5: subdevice #5
  Subdevice #6: subdevice #6
  Subdevice #7: subdevice #7
card 4: Rubix24 [Rubix24], device 0: USB Audio [USB Audio]
  Subdevices: 1/1
  Subdevice #0: subdevice #0
card 5: UAC2Gadget [UAC2_Gadget], device 0: UAC2 PCM [UAC2 PCM]
  Subdevices: 1/1
  Subdevice #0: subdevice #0
card 6: UAC2Gadget_1 [UAC2_Gadget], device 0: UAC2 PCM [UAC2 PCM]
  Subdevices: 1/1
  Subdevice #0: subdevice #0

$ arecord --dump-hw-params -D {device_name} /dev/null
Warning: Some sources (like microphones) may produce inaudible results
         with 8-bit sampling. Use '-f' argument to increase resolution
         e.g. '-f S16_LE'.
Recording WAVE '/dev/null' : Unsigned 8 bit, Rate 8000 Hz, Mono
HW Params of device "hw:4,0":
--------------------
ACCESS:  MMAP_INTERLEAVED RW_INTERLEAVED
FORMAT:  S32_LE
SUBFORMAT:  STD
SAMPLE_BITS: 32
FRAME_BITS: 64
CHANNELS: 2
RATE: [44100 192000]
PERIOD_TIME: [125 1000000]
PERIOD_SIZE: [8 192000]
PERIOD_BYTES: [64 1536000]
PERIODS: [2 1024]
BUFFER_TIME: (83 2000000]
BUFFER_SIZE: [16 384000]
BUFFER_BYTES: [128 3072000]
TICK_TIME: ALL
--------------------
arecord: set_params:1352: Sample format non available
Available formats:
- S32_LE
```

## ALSAデバイスをJACKに接続

### 入出力を両方とも接続

```shell
$ jack_load {device_name}_src alsa_in -i "-d {device_name} -r {sampling_rate} -p {periods} -n {nperiods}"
$ jack_load {device_name}_sink alsa_out -i "-d {device_name} -r {sampling_rate} -p {periods} -n {nperiods}"
```

### 接続済みのデバイスを解放

```shell
$ jack_unload {device_name}_src
$ jack_unload {device_name}_sink
```

## JACKのポート周りの操作

## 接続可能・接続済みのポートを取得

```shell
$ jack_lsp -c
```

### 入出力のポート同士を接続

```shell
$ jack_connect {source} {destination}
```

### 接続を解除

```shell
$ jack_disconnect {source} {destination}
```