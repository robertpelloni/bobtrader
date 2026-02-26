import * as tf from '@tensorflow/tfjs-node';
import path from 'path';
import fs from 'fs';
import { ConfigManager } from '../config/ConfigManager';

export class DeepThinker {
    private model: tf.LayersModel | null = null;
    private config: ConfigManager;
    private modelPath: string;
    private lookbackWindow: number = 60; // Use last 60 candles

    constructor() {
        this.config = ConfigManager.getInstance();
        const hubDir = this.config.get("trading.hub_data_dir") || "hub_data";
        this.modelPath = path.resolve(process.cwd(), '..', hubDir, 'models/deep_thinker');
    }

    public async initialize(): Promise<void> {
        try {
            if (fs.existsSync(this.modelPath)) {
                this.model = await tf.loadLayersModel(`file://${this.modelPath}/model.json`);
                console.log("[DeepThinker] Loaded existing model.");
            } else {
                console.log("[DeepThinker] No existing model found. Creating new.");
                this.model = this.createModel();
            }
        } catch (e) {
            console.error("[DeepThinker] Error loading model:", e);
            this.model = this.createModel();
        }
    }

    private createModel(): tf.LayersModel {
        const model = tf.sequential();

        // LSTM Layer
        model.add(tf.layers.lstm({
            units: 50,
            returnSequences: true,
            inputShape: [this.lookbackWindow, 5] // [TimeSteps, Features(Open, High, Low, Close, Vol)]
        }));

        model.add(tf.layers.dropout({ rate: 0.2 }));

        model.add(tf.layers.lstm({
            units: 50,
            returnSequences: false
        }));

        model.add(tf.layers.dropout({ rate: 0.2 }));

        // Output Layer (Predicting Next Close Price)
        model.add(tf.layers.dense({ units: 1 }));

        model.compile({
            optimizer: tf.train.adam(0.001),
            loss: 'meanSquaredError'
        });

        return model;
    }

    public async train(data: any[]): Promise<any> {
        if (!this.model) this.model = this.createModel();
        if (data.length < this.lookbackWindow + 10) throw new Error("Not enough data to train");

        console.log(`[DeepThinker] Training on ${data.length} candles...`);

        // 1. Preprocess
        const { inputs, labels, max, min } = this.preprocess(data);

        // 2. Train
        const history = await this.model.fit(inputs, labels, {
            epochs: 20,
            batchSize: 32,
            validationSplit: 0.1,
            callbacks: {
                onEpochEnd: (epoch, logs) => console.log(`Epoch ${epoch}: loss=${logs?.loss.toFixed(5)}`)
            }
        });

        // 3. Save
        const dir = path.dirname(this.modelPath);
        if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });

        // Save main "latest" model
        await this.model.save(`file://${this.modelPath}`);
        fs.writeFileSync(`${this.modelPath}/meta.json`, JSON.stringify({ max, min }));

        // Save versioned backup
        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        const versionDir = path.join(dir, `deep_thinker_${timestamp}`);
        if (!fs.existsSync(versionDir)) fs.mkdirSync(versionDir, { recursive: true });
        await this.model.save(`file://${versionDir}`);
        fs.writeFileSync(`${versionDir}/meta.json`, JSON.stringify({ max, min }));

        console.log(`[DeepThinker] Model saved to ${this.modelPath} and backup at ${versionDir}`);

        inputs.dispose();
        labels.dispose();

        return history.history;
    }

    public async predict(recentCandles: any[]): Promise<number> {
        if (!this.model) await this.initialize();
        if (recentCandles.length < this.lookbackWindow) return 0; // Not enough data

        // Slice last window
        const window = recentCandles.slice(-this.lookbackWindow);

        // Load meta for normalization
        let meta = { max: 100000, min: 0 };
        try {
            if (fs.existsSync(`${this.modelPath}/meta.json`)) {
                meta = JSON.parse(fs.readFileSync(`${this.modelPath}/meta.json`, 'utf-8'));
            }
        } catch (e) {}

        // Normalize
        const normalized = window.map(c => [
            (c.open - meta.min) / (meta.max - meta.min),
            (c.high - meta.min) / (meta.max - meta.min),
            (c.low - meta.min) / (meta.max - meta.min),
            (c.close - meta.min) / (meta.max - meta.min),
            (c.volume - 0) / (10000000 - 0) // Simplified vol norm
        ]);

        const tensor = tf.tensor3d([normalized], [1, this.lookbackWindow, 5]);
        const prediction = this.model!.predict(tensor) as tf.Tensor;
        const val = prediction.dataSync()[0];

        tensor.dispose();
        prediction.dispose();

        // Denormalize
        return val * (meta.max - meta.min) + meta.min;
    }

    private preprocess(data: any[]): { inputs: tf.Tensor3D, labels: tf.Tensor2D, max: number, min: number } {
        // Find Global Min/Max for normalization
        let min = Infinity;
        let max = -Infinity;

        data.forEach(c => {
            min = Math.min(min, c.low);
            max = Math.max(max, c.high);
        });

        const inputs: number[][][] = [];
        const labels: number[] = [];

        for (let i = this.lookbackWindow; i < data.length; i++) {
            const window = data.slice(i - this.lookbackWindow, i);
            const target = data[i].close;

            const normWindow = window.map(c => [
                (c.open - min) / (max - min),
                (c.high - min) / (max - min),
                (c.low - min) / (max - min),
                (c.close - min) / (max - min),
                (c.volume - 0) / (10000000 - 0) // Rough volume norm
            ]);

            inputs.push(normWindow);
            labels.push((target - min) / (max - min));
        }

        return {
            inputs: tf.tensor3d(inputs),
            labels: tf.tensor2d(labels, [labels.length, 1]),
            max,
            min
        };
    }
}
