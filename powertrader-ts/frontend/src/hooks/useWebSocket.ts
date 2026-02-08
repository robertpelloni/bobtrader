import { useEffect, useState, useRef, useCallback } from 'react';

export const useWebSocket = (url: string) => {
    const [isConnected, setIsConnected] = useState(false);
    const [lastMessage, setLastMessage] = useState<any>(null);
    const ws = useRef<WebSocket | null>(null);

    useEffect(() => {
        ws.current = new WebSocket(url);

        ws.current.onopen = () => {
            console.log('WebSocket Connected');
            setIsConnected(true);
        };

        ws.current.onclose = () => {
            console.log('WebSocket Disconnected');
            setIsConnected(false);
        };

        ws.current.onmessage = (event) => {
            const data = JSON.parse(event.data);
            setLastMessage(data);
        };

        return () => {
            ws.current?.close();
        };
    }, [url]);

    const sendMessage = useCallback((msg: any) => {
        if (ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify(msg));
        }
    }, []);

    return { isConnected, lastMessage, sendMessage };
};
