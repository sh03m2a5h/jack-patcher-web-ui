// APIで利用する各種コマンドをラップする関数

import { exec } from 'child_process';
import { promisify } from 'util';
import { AlsaDevice, AlsaDeviceParams, JackConnection, JackConnectionState, JackStatus, NumRange } from '../models';

const execAsync = promisify(exec);
const snake2camel = (str: string) => str.toLowerCase().replace(/_./g, (s) => s.charAt(1).toUpperCase());

function parseParamData(paramData: string): string | number | NumRange {
  if (/^\d+$/.test(paramData)) {
    return parseInt(paramData, 10);
  } else if (/^\d+-\d+$/.test(paramData)) {
    const [min, max] = paramData.split('-').map((v) => parseInt(v, 10));
    return { min, max };
  } else if (/\d+\s+\d+/.test(paramData)) {
    const match = paramData.match(/(\d+)\s+(\d+)/);
    if (match) {
      const [, min, max] = match;
      return { min: parseInt(min, 10), max: parseInt(max, 10) };
    } else {
      return paramData;
    }
  } else {
    return paramData;
  }
}

/**
 * JACKサーバーの状態を取得する
 * @returns {Promise<JackStatus>} サーバー状態オブジェクト
 */
async function getJackStatus(): Promise<JackStatus> {
  try {
    const { stdout } = await execAsync('jack_control status');
    return { status: stdout.trim() };
  } catch (error) {
    throw new Error(`JACKサーバーの状態取得に失敗しました: ${(error as Error).message}`);
  }
}

/**
 * JACKサーバーを起動する
 * @param options JACKサーバーの設定オプション
 */
async function startJackServer(options: { rate?: number; period?: number } = {}): Promise<void> {
  const { rate = 96000, period = 128 } = options;
  try {
    await execAsync('jack_control start');
    await execAsync('jack_control ds dummy');
    await execAsync(`jack_control dps rate ${rate}`);
    await execAsync(`jack_control dps period ${period}`);
  } catch (error) {
    throw new Error(`JACKサーバーの起動に失敗しました: ${(error as Error).message}`);
  }
}

/**
 * JACKサーバーを停止する
 */
async function stopJackServer(): Promise<void> {
  try {
    await execAsync('jack_control stop');
    await execAsync('jack_control exit');
  } catch (error) {
    throw new Error(`JACKサーバーの停止に失敗しました: ${(error as Error).message}`);
  }
}

/**
 * ALSAのデバイス情報を取得する
 * @param mode 'playback'または'capture'
 * @returns {Promise<AlsaDevice[]>} ALSAデバイス情報の配列
 */
async function getAlsaDevices(/*mode: 'playback' | 'capture'*/): Promise<AlsaDevice[]> {
  try {
    // const command = mode === 'playback' ? 'aplay' : 'arecord';
    // const { stdout } = await execAsync(command + ' -l');
    const { stdout } = await execAsync("aplay -l");
    const devices: AlsaDevice[] = [];
    const lines = stdout.split('\n');
    for (const line of lines) {
      const match = line.match(/card (\d+): (\S+) \[(.+?)\], device (\d+): (.+) \[(.+?)\]/);
      if (match) {
        const cardNum = parseInt(match[1], 10);
        const deviceNum = parseInt(match[4], 10);

        const { stderr: pOut } = await execAsync(`timeout 1 aplay --dump-hw-params -D hw:${cardNum},${deviceNum} /dev/zero`).catch((e) => e);
        const { stderr: cOut } = await execAsync(`timeout 1 arecord --dump-hw-params -D hw:${cardNum},${deviceNum} /dev/null`).catch((e) => e);
        const pParams = String(pOut).match(/-{20}(.*)-{20}/s)?.[1];
        const cParams = String(cOut).match(/-{20}(.*)-{20}/s)?.[1];

        let playbackParams: AlsaDeviceParams | null = null;
        let captureParams: AlsaDeviceParams | null = null;

        if (pParams) {
          const playbackMatch = pParams.match(/(.*):(\s+?)(.*)/gm);
          if (playbackMatch) {
            playbackParams = { rate: 48000, periods: 1 };
            for (const param of playbackMatch) {
              const [key, value] = param.split(':');
              playbackParams[snake2camel(key)] = parseParamData(value.trim());
            }
          }
        }
        if (cParams) {
          const captureMatch = cParams.match(/(.*):(\s+?)(.*)/gm);
          if (captureMatch) {
            captureParams = { rate: 48000, periods: 1 };
            for (const param of captureMatch) {
              const [key, value] = param.split(':');
              captureParams[snake2camel(key)] = parseParamData(value.trim());
            }
          }
        }


        devices.push({
          card: match[1],
          cardName: match[2],
          device: match[4],
          description: match[6],
          playbackParams: playbackParams ? playbackParams : undefined,
          captureParams: captureParams ? captureParams : undefined,
        });
      }
    }
    return devices;
  } catch (error) {
    throw new Error(`ALSAデバイス情報の取得に失敗しました: ${(error as Error).message}`);
  }
}

/**
 * ALSAデバイスの接続状況を取得する
 * @returns {Promise<JackConnectionState | null>} 接続状況のオブジェクト
 */
async function getJackConnection(): Promise<JackConnectionState | null> {
  try {
    const { stdout } = await execAsync(`jack_lsp -c | grep alsa_`).catch((e) => e);
    if (!stdout) {
      return null;
    }
    const lines = stdout.trim().split('\n');
    const alsaDevices: string[] = [];
    const connections: JackConnection[] = [];
    let currentDevice = '';
    for (const line of lines) {
      const match = line.match(/^alsa_(\S+)_(src|sink):(.+)$/);
      if (match) {
        currentDevice = match[1];
        if (!alsaDevices.includes(currentDevice))
          alsaDevices.push(currentDevice);
      } else {
        const dest = line.trim();
        connections.push({ src: currentDevice, dest });
      }
    }
    return { alsaDevices, connections };
  } catch (error) {
    throw new Error(`ALSAデバイスの接続状況の取得に失敗しました: ${(error as Error).message}`);
  }
}

/**
 * ALSAデバイスをJACKに接続する
 * @param options 接続設定
 */
async function connectAlsaToJack(device: AlsaDevice, periods: number = 128, nperiods: number = 2): Promise<void> {
  const deviceId = `hw:${device.card}`;
  const cardName = `alsa_${device.cardName}`;
  const samplingRate = device.playbackParams?.rate || 48000;
  try {
    if (device.captureParams) {
      execAsync(`alsa_in -j ${cardName}_src -d ${deviceId} -r ${samplingRate} -p ${periods} -n ${nperiods} > /dev/null`);
    }
    if (device.playbackParams) {
      execAsync(`alsa_out -j ${cardName}_sink -d ${deviceId} -r ${samplingRate} -p ${periods} -n ${nperiods} > /dev/null`);
    }
    // await execAsync(`jack_load ${cardName} audioadapter -i "-d ${deviceId} -r ${samplingRate} -p ${periods} -n ${nperiods}"`);
  } catch (error) {
    throw new Error(`ALSAデバイスのJACK接続に失敗しました: ${(error as Error).message}`);
  }
}

/**
 * ALSAデバイスのJACK接続を解除する
 * @param deviceName デバイス名
 */
async function disconnectAlsaFromJack(deviceName: string): Promise<void> {
  try {
    await execAsync(`jack_unload ${deviceName}_src`);
    await execAsync(`jack_unload ${deviceName}_sink`);
  } catch (error) {
    throw new Error(`ALSAデバイスのJACK接続解除に失敗しました: ${(error as Error).message}`);
  }
}

/**
 * JACKのポートを接続する
 * @param source 接続元ポート名
 * @param destination 接続先ポート名
 */
async function connectJackPorts(source: string, destination: string): Promise<void> {
  try {
    await execAsync(`jack_connect ${source} ${destination}`);
  } catch (error) {
    throw new Error(`JACKポートの接続に失敗しました: ${(error as Error).message}`);
  }
}

/**
 * JACKのポート接続を解除する
 * @param source 接続元ポート名
 * @param destination 接続先ポート名
 */
async function disconnectJackPorts(source: string, destination: string): Promise<void> {
  try {
    await execAsync(`jack_disconnect ${source} ${destination}`);
  } catch (error) {
    throw new Error(`JACKポート接続の解除に失敗しました: ${(error as Error).message}`);
  }
}

export {
  getJackStatus,
  startJackServer,
  stopJackServer,
  getAlsaDevices,
  getJackConnection,
  connectAlsaToJack,
  disconnectAlsaFromJack,
  connectJackPorts,
  disconnectJackPorts
};
