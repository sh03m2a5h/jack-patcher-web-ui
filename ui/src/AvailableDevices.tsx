import { useRecoilValueLoadable } from "recoil";
import './AvailableDevices.css'
import { AlsaDevice } from "../../api/src/models";
import { devicesState, connectionState } from "./stores/devices";


export default function AvailableDevices() {
  const { state: devicesLoadingState, contents: devices } = useRecoilValueLoadable(devicesState);
  const connectionsStateLoadable = useRecoilValueLoadable(connectionState);

  if (devicesLoadingState === "loading") {
    return <div className="available-devices">Loading...</div>;
  }
  if (devicesLoadingState === "hasError" || devicesLoadingState !== "hasValue") {
    return <div className="available-devices">An error has occurred</div>;
  }

  console.log(devicesLoadingState, devices);
  
  const connectionData = connectionsStateLoadable.state === "hasValue" ? connectionsStateLoadable.getValue().alsaDevices ?? [] : [] as string[];

  const deviceConnect = (device: AlsaDevice) => {
    console.log(device);
    fetch("/api/connect-alsa-to-jack", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        card: device.card,
        cardName: device.cardName,
      }),
    });
  }

  return (
    <div className="available-devices">
      {
        <ul>{devices.map(
          (device) => (
            <li key={device.card + device.device} className={`device ${connectionData.includes(device.cardName) ? 'device-connected': ''}`} onDoubleClick={() => deviceConnect(device)}>
              {device.cardName} ({device.card}): {device.description}
            </li>
          )
        )}</ul>
      }
    </div>
  );
}