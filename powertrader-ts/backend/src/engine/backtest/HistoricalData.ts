import axios from 'axios';

export class HistoricalData {
    private static baseUrl = "https://api.kucoin.com";

    public static async fetch(pair: string, interval: string, start: number, end: number): Promise<any[]> {
        const symbol = pair.replace("USD", "USDT");
        let type = interval.replace('h', 'hour').replace('m', 'min').replace('d', 'day');
        if (type === '1h') type = '1hour';

        const allCandles: any[] = [];
        let currentEnd = end;

        // KuCoin API expects seconds
        const startSec = Math.floor(start / 1000);

        console.log(`[HistoricalData] Fetching ${pair} ${interval} from ${new Date(start).toISOString()} to ${new Date(end).toISOString()}`);

        while (currentEnd > start) {
            try {
                const endSec = Math.floor(currentEnd / 1000);

                // Fetch chunk
                const response = await axios.get(`${this.baseUrl}/api/v1/market/candles`, {
                    params: {
                        symbol: symbol,
                        type: type,
                        startAt: startSec,
                        endAt: endSec
                    }
                });

                if (response.data && response.data.data && response.data.data.length > 0) {
                    const candles = response.data.data; // [time, open, close, ...] (newest first)

                    // Convert to our format
                    const parsed = candles.map((c: string[]) => ({
                        timestamp: parseInt(c[0]) * 1000,
                        open: parseFloat(c[1]),
                        close: parseFloat(c[2]),
                        high: parseFloat(c[3]),
                        low: parseFloat(c[4]),
                        volume: parseFloat(c[5])
                    }));

                    // Add to list
                    allCandles.push(...parsed);

                    // The last candle in response is the oldest in this batch
                    const oldestTime = parsed[parsed.length - 1].timestamp;

                    // Move currentEnd back
                    currentEnd = oldestTime - 1000;

                    // If we got fewer than 100 candles, probably end of data
                    if (candles.length < 100) {
                        break;
                    }

                    // Rate limit protection
                    await new Promise(r => setTimeout(r, 200));
                } else {
                    break;
                }
            } catch (e) {
                console.error("[HistoricalData] Error fetching history:", e);
                break;
            }
        }

        // Sort by time ascending
        const sorted = allCandles.sort((a, b) => a.timestamp - b.timestamp);

        // Remove duplicates (based on timestamp)
        const unique = sorted.filter((v, i, a) => i === 0 || v.timestamp !== a[i-1].timestamp);

        console.log(`[HistoricalData] Fetched ${unique.length} candles.`);
        return unique;
    }
}
