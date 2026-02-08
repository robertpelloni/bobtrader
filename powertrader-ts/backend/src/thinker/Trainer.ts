import { IExchangeConnector } from "../engine/connector/IExchangeConnector";
import { ConfigManager } from "../config/ConfigManager";
import fs from 'fs';
import path from 'path';

export class Trainer {
    private connector: IExchangeConnector;
    private config: ConfigManager;
    private tfChoices = ["1hour", "2hour", "4hour", "8hour", "12hour", "1day", "1week"];

    constructor(connector: IExchangeConnector) {
        this.connector = connector;
        this.config = ConfigManager.getInstance();
    }

    public async trainAll(): Promise<void> {
        const coins = this.config.get("trading.coins") as string[];
        for (const coin of coins) {
            await this.trainCoin(coin);
        }
    }

    public async trainCoin(coin: string): Promise<void> {
        console.log(`[Trainer] Starting training for ${coin}...`);
        const neuralDir = this.config.get("trading.main_neural_dir") || "hub_data";
        const coinDir = path.join(neuralDir, coin);

        if (!fs.existsSync(coinDir)) {
            fs.mkdirSync(coinDir, { recursive: true });
        }

        for (const tf of this.tfChoices) {
            try {
                // 1. Fetch History
                const history = await this.connector.fetchOHLCV(`${coin}-USD`, tf, 1000);
                if (history.length < 100) {
                    console.log(`[Trainer] Not enough data for ${coin} ${tf}`);
                    continue;
                }

                // 2. Extract Patterns (Mock logic matching Thinker expectations)
                // Real kNN logic requires normalizing sequences
                const patterns: string[] = [];
                const weights: number[] = [];

                for (let i = 2; i < history.length - 1; i++) {
                    const p1 = (history[i-1].close - history[i-1].open) / history[i-1].open;
                    const p2 = (history[i].close - history[i].open) / history[i].open;

                    const nextHigh = (history[i+1].high - history[i].close) / history[i].close;
                    const nextLow = (history[i+1].low - history[i].close) / history[i].close;

                    // Format: "p1 p2{}high{}low"
                    patterns.push(`${p1} ${p2}{}${nextHigh}{}${nextLow}`);
                    weights.push(1.0); // Default weight
                }

                // 3. Save to File (Thinker compatibility)
                const memPath = path.join(coinDir, `memories_${tf}.txt`);
                const wgtPath = path.join(coinDir, `memory_weights_${tf}.txt`);

                fs.writeFileSync(memPath, patterns.join("~"));
                fs.writeFileSync(wgtPath, weights.join(" "));

                console.log(`[Trainer] Saved ${patterns.length} patterns for ${coin} ${tf}`);

            } catch (e) {
                console.error(`[Trainer] Error training ${coin} ${tf}:`, e);
            }
        }

        // Write timestamp tag
        fs.writeFileSync(path.join(coinDir, "trainer_last_training_time.txt"), (Date.now() / 1000).toString());
    }
}
