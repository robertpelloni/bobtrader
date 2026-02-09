export class RateLimiter {
    private calls: number[] = [];
    private maxCalls: number;
    private period: number; // in milliseconds

    constructor(maxCalls: number, period: number = 60000) {
        this.maxCalls = maxCalls;
        this.period = period;
    }

    public acquire(): boolean {
        const now = Date.now();
        this.calls = this.calls.filter(t => t > now - this.period);

        if (this.calls.length >= this.maxCalls) {
            return false;
        }

        this.calls.push(now);
        return true;
    }

    public reset(): void {
        this.calls = [];
    }
}
