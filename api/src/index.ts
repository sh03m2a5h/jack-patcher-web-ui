import { AlsaDevice, connectAlsaToJack, disconnectAlsaFromJack, getAlsaDevices, startJackServer, stopJackServer } from "./utils/wrapper";
import express from "express";

const app = express();
const PORT = 3000;

const alsaDevices: AlsaDevice[] = [];

// サーバー起動時にJACKサーバーを開始
app.listen(PORT, async () => {
    console.log(`API server running on http://localhost:${PORT}`);
    await startJackServer();

    // ALSAデバイス一覧取得
    const devices = await getAlsaDevices();
    alsaDevices.push(...devices);
});

// ALSAデバイス一覧取得
app.get("/api/alsa-devices", async (req, res) => {
    res.json(alsaDevices);
});

// ALSAデバイスをJACKに接続
app.post("/api/connect-alsa-to-jack", async (req, res) => {
    const { deviceName } = req.body;
    try {
        await connectAlsaToJack(deviceName);
        res.json({ message: `ALSAデバイス ${deviceName} をJACKに接続しました` });
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
