// APIで利用する各種コマンドをラップする関数

import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);
const snake2camel = (str: string) => str.toLowerCase().replace(/_./g, (s) => s.charAt(1).toUpperCase());

// 型定義
export interface JackStatus {
  status: string;
}

export interface Range {
  min: number;
  max: number;
}

export interface AlsaDeviceParams {
  rate: number | Range;
  periods: number | Range;
  [key: string]: string | number | Range;
}

export interface AlsaDevice {
  card: string;
  cardName: string;
  device: string;
  description: string;
  playbackParams?: AlsaDeviceParams[];
  captureParams?: AlsaDeviceParams[];
}

function parseParamData(paramData: string): string | number | Range {
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

        const { stderr: pOut } = await execAsync(`aplay --dump-hw-params -D hw:${cardNum},${deviceNum} /dev/zero`).catch((e) => e);
        const { stderr: cOut } = await execAsync(`arecord --dump-hw-params -D hw:${cardNum},${deviceNum} /dev/null`).catch((e) => e);
        const pParams = String(pOut).match(/-{20}(.*)-{20}/s)?.[1];
        const cParams = String(cOut).match(/-{20}(.*)-{20}/s)?.[1];

        const playbackParams: AlsaDeviceParams = { rate: 48000, periods: 1 };
        const captureParams: AlsaDeviceParams = { rate: 48000, periods: 1 };

        if (pParams) {
          const playbackMatch = pParams.match(/(.*):(\s+?)(.*)/gm);
          playbackMatch?.forEach((param) => {
            const [key, value] = param.split(':');
            playbackParams[snake2camel(key)] = parseParamData(value.trim());
          });
        }
        if (cParams) {
          const captureMatch = cParams.match(/(.*):(\s+?)(.*)/gm);
          captureMatch?.forEach((param) => {
            const [key, value] = param.split(':');
            captureParams[snake2camel(key)] = parseParamData(value.trim());
          });
        }


        devices.push({
          card: match[1],
          cardName: match[2],
          device: match[4],
          description: match[6],
          playbackParams: playbackParams ? [playbackParams] : undefined,
          captureParams: captureParams ? [captureParams] : undefined,
        });
      }
    }
    return devices;
  } catch (error) {
    throw new Error(`ALSAデバイス情報の取得に失敗しました: ${(error as Error).message}`);
  }
}

/**
 * ALSAデバイスをJACKに接続する
 * @param options 接続設定
 */
async function connectAlsaToJack(options: {
  deviceName: string;
  samplingRate: number;
  periods: number;
  nperiods: number;
}): Promise<void> {
  const { deviceName, samplingRate = 96000, periods = 128, nperiods = 4 } = options;
  const deviceId = `hw:${deviceName},${deviceName}`;
  try {
    await execAsync(`jack_load ${deviceName}_src alsa_in -i "-d ${deviceName} -r ${samplingRate} -p ${periods} -n ${nperiods}"`);
    await execAsync(`jack_load ${deviceName}_sink alsa_out -i "-d ${deviceName} -r ${samplingRate} -p ${periods} -n ${nperiods}"`);
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
  connectAlsaToJack,
  disconnectAlsaFromJack,
  connectJackPorts,
  disconnectJackPorts
};
