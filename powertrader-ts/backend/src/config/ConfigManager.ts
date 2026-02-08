import fs from 'fs';
import path from 'path';
import yaml from 'yaml';

export class ConfigManager {
    private static instance: ConfigManager;
    private config: any = {};
    private configPath: string;

    private constructor() {
        this.configPath = path.join(process.cwd(), 'config.yaml');
        this.load();
    }

    public static getInstance(): ConfigManager {
        if (!ConfigManager.instance) {
            ConfigManager.instance = new ConfigManager();
        }
        return ConfigManager.instance;
    }

    public load(): void {
        try {
            if (fs.existsSync(this.configPath)) {
                const fileContents = fs.readFileSync(this.configPath, 'utf8');
                this.config = yaml.parse(fileContents);
            } else {
                console.log("Config file not found, creating default.");
                this.createDefault();
            }
        } catch (e) {
            console.error("Error loading config:", e);
        }
    }

    public get(key: string): any {
        return key.split('.').reduce((o, i) => o?.[i], this.config);
    }

    private createDefault(): void {
        this.config = {
            trading: {
                coins: ["BTC", "ETH", "XRP", "BNB", "DOGE"],
                trade_start_level: 3,
                start_allocation_pct: 0.005,
                dca_multiplier: 2.0,
                dca_levels: [-2.5, -5.0, -10.0, -20.0, -30.0, -40.0, -50.0],
                max_dca_buys_per_24h: 2
            },
            system: {
                log_level: "INFO"
            }
        };
        fs.writeFileSync(this.configPath, yaml.stringify(this.config));
    }
}
