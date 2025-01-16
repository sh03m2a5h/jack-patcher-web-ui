import { AlsaDevice } from "./models";
import { connectAlsaToJack, disconnectAlsaFromJack, getJackConnection, getAlsaDevices, startJackServer, stopJackServer } from "./utils/wrapper";
import express from "express";

const app = express();
const PORT = 3000;
app.use(express.json());

const alsaDevices: AlsaDevice[] = [];

// サーバー起動時にJACKサーバーを開始
app.listen(PORT, async () => {
    console.log(`API server running on http://localhost:${PORT}`);
    await startJackServer();

    // ALSAデバイス一覧取得
    const devices = await getAlsaDevices();
    console.log(devices);
    alsaDevices.push(...devices);
});

// ALSAデバイス一覧取得
app.get("/api/alsa-devices", async (req, res) => {
    res.json(alsaDevices);
});

// JACK接続一覧取得
app.get("/api/jack-connection", async (req, res) => {
    try {
        const connections = await getJackConnection();
        if (!connections) {
            res.status(404).json({ message: "JACK接続が見つかりません" });
            return;
        }
        res.json(connections);
    }
    catch (error) {
        res.status(500).json({ message: (error as Error).message });
    }
});

// ALSAデバイスをJACKに接続
app.post("/api/connect-alsa-to-jack", async (req, res) => {
    try {
        const { cardName } = req.body;
        const device = alsaDevices.find((d) => d.cardName === cardName);
        if (!device) {
            throw new Error(`ALSAデバイス ${cardName} が見つかりません`);
        }
        await connectAlsaToJack(device);
        res.json({ message: `ALSAデバイス ${req.body} をJACKに接続しました` });
    } catch (error) {
        res.status(500).json({ message: (error as Error).message });
    }
});

// ALSAデバイスをJACKから切断
app.post("/api/disconnect-alsa-from-jack", async (req, res) => {
    const { deviceName } = req.body;
    try {
        await disconnectAlsaFromJack(deviceName);
        res.json({ message: `ALSAデバイス ${deviceName} をJACKから切断しました` });
    } catch (error) {
        res.status(500).json({ message: (error as Error).message });
    }
});

// プロセス終了時にJACKサーバーを停止
process.on("SIGINT", async () => {
    console.log("Shutting down...");
    await stopJackServer();
    process.exit(0);
});

process.on("SIGTERM", async () => {
    console.log("Shutting down...");
    await stopJackServer();
    process.exit(0);
});
