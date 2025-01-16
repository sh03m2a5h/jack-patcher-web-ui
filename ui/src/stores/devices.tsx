import { atom, selector } from "recoil";
import * as models from "../../../api/src/models";
import { AppNode } from "../nodes/types";

export default models;

export const devicesState = atom({
  key: "devicesState",
  default: fetch("/api/alsa-devices").then((res) => res.json()) as Promise<models.AlsaDevice[]>,
});

export const connectionState = atom({
  key: "connectionState",
  default: fetch("/api/jack-connection").then((res) => res.json()) as Promise<models.JackConnectionState>,  
});

export const flowNodes = selector({
  key: "flowNodes",
  get: ({ get }) => {
    const devices = get(devicesState);
    const connections = get(connectionState);
    
    const nodes = connections.alsaDevices.flatMap((deviceName) => {
      const nodes: AppNode[] = [];
      const device = devices.find((d) => d.cardName === deviceName);
      if (!device) 
        return nodes;
      const getMaxChannels = (arg: string | number | models.NumRange): number => {
        if (typeof arg === "string") {
          return parseInt(arg);
        }
        if (typeof arg === "number") {
          return arg;
        }
        return arg.max;
      }
      if (device.playbackParams) {
        nodes.push({
          id: device.cardName + "-output",
          type: "audio-output",
          position: { x: 250, y: 300 },
          data: { deviceName: device.cardName, channelCount: getMaxChannels(device.playbackParams.channels) },
        });
      }
      if (device.captureParams) {
        nodes.push({
          id: device.cardName + "-input",
          type: "audio-input",
          position: { x: 250, y: 100 },
          data: { deviceName: device.cardName, channelCount: getMaxChannels(device.captureParams.channels) },
        });
      }
      return nodes;
    });
    return nodes;
  },
});
