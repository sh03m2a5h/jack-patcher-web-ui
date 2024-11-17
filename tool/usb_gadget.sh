#!/bin/bash

# USB ガジェットの設定を行うスクリプト
# このスクリプトは、Raspberry Pi 4B で動作確認済みです。
# 他の環境での動作は保証しません。

# USB ガジェットの設定開始
cd /sys/kernel/config/usb_gadget/
mkdir -p g_audio
cd g_audio

# ベンダーIDとプロダクトIDの設定
echo 0x1d6b > idVendor # Linux Foundation
echo 0x0104 > idProduct # Multifunction Composite Gadget
echo 0x0100 > bcdDevice
echo 0x0200 > bcdUSB

# デバイス情報の設定
mkdir -p strings/0x409
echo "0123456789" > strings/0x409/serialnumber
echo "Linux" > strings/0x409/manufacturer
echo "Dual Audio Gadget" > strings/0x409/product

# 構成の設定
mkdir -p configs/c.1
mkdir -p configs/c.1/strings/0x409
echo "USB_Gadget_Audio" > configs/c.1/strings/0x409/configuration
echo 500 > configs/c.1/MaxPower  # 電力を 500mA に設定

# ========== ↓ UAC2 の設定 ↓ ==========

# ----- 1つ目のオーディオ機能の追加 -----
mkdir -p functions/uac2.usb0

# サポートするサンプルレートの設定
echo 44100 > functions/uac2.usb0/c_srate
echo 44100 > functions/uac2.usb0/p_srate

# サンプルサイズ（ビット深度）は単一の値のみ指定可能
# ここでは24ビット（3バイト）を指定
echo 3 > functions/uac2.usb0/c_ssize   # 24ビット
echo 3 > functions/uac2.usb0/p_ssize

# チャンネルマスクの設定（ステレオ）
echo 0x00000003 > functions/uac2.usb0/c_chmask
echo 0x00000003 > functions/uac2.usb0/p_chmask

# # サンプルレートの有効化（0に設定してホスト側で指定可能にする）
# echo 0 > functions/uac2.usb0/c_srate_active
# echo 0 > functions/uac2.usb0/p_srate_active

# ----- 2つ目のオーディオ機能の追加 -----
mkdir -p functions/uac2.usb1

echo 96000 > functions/uac2.usb1/c_srate
echo 96000 > functions/uac2.usb1/p_srate

echo 3 > functions/uac2.usb1/c_ssize
echo 3 > functions/uac2.usb1/p_ssize

echo 0x00000003 > functions/uac2.usb1/c_chmask
echo 0x00000003 > functions/uac2.usb1/p_chmask

# echo 0 > functions/uac2.usb1/c_srate_active
# echo 0 > functions/uac2.usb1/p_srate_active

# ========== ↑ UAC2 の設定 ↑ ==========

# USB ガジェットの構成に各オーディオ機能を追加
ln -s functions/uac2.usb0 configs/c.1/
ln -s functions/uac2.usb1 configs/c.1/

# USB デバイスの有効化
ls /sys/class/udc > UDC