import WebSocket, { WebSocketServer } from 'ws';
import { Server } from 'http';

export class WebSocketManager {
    private static instance: WebSocketManager;
    private wss: WebSocketServer | null = null;

    private constructor() {}

    public static getInstance(): WebSocketManager {
        if (!WebSocketManager.instance) {
            WebSocketManager.instance = new WebSocketManager();
        }
        return WebSocketManager.instance;
    }

    public initialize(server: Server): void {
        this.wss = new WebSocketServer({ server });

        this.wss.on('connection', (ws) => {
            console.log('[WS] Client connected');
            ws.send(JSON.stringify({ type: 'WELCOME', message: 'Connected to PowerTrader AI Stream' }));

            ws.on('close', () => {
                // console.log('[WS] Client disconnected');
            });
        });

        console.log('[WS] WebSocket Server initialized');
    }

    public broadcast(type: string, payload: any): void {
        if (!this.wss) return;

        const message = JSON.stringify({ type, payload, timestamp: Date.now() });

        this.wss.clients.forEach((client) => {
            if (client.readyState === WebSocket.OPEN) {
                client.send(message);
            }
        });
    }
}
