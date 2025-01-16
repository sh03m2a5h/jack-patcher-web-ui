export interface JackStatus {
    status: string;
}

export interface NumRange {
    min: number;
    max: number;
}

export interface AlsaDeviceParams {
    rate: number | NumRange;
    periods: number | NumRange;
    [key: string]: string | number | NumRange;
}

export interface AlsaDevice {
    card: string;
    cardName: string;
    device: string;
    description: string;
    playbackParams?: AlsaDeviceParams;
    captureParams?: AlsaDeviceParams;
}

export interface JackConnection {
    src: string;
    dest: string;
}

export interface JackConnectionState {
    alsaDevices: string[];
    connections: JackConnection[];
}
