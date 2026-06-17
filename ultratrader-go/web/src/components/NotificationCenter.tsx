import { useState, useEffect } from 'react';
import { Bell, Shield, Zap } from 'lucide-react';

interface Event {
  id: string;
  type: 'signal' | 'order' | 'system';
  message: string;
  timestamp: string;
  level: 'info' | 'success' | 'warning' | 'error';
}

export const NotificationCenter = () => {
  const [events, setEvents] = useState<Event[]>([]);

  useEffect(() => {
    const fetchEvents = async () => {
      try {
        const [signalsRes, ordersRes] = await Promise.all([
          fetch('/api/signals'),
          fetch('/api/orders')
        ]);

        const signals = await signalsRes.json();
        const orders = await ordersRes.json();

        const combined: Event[] = [
          ...signals.map((s: any) => ({
            id: `sig-${s.timestamp}`,
            type: 'signal',
            message: `Signal: ${s.action.toUpperCase()} ${s.symbol} - ${s.reason}`,
            timestamp: s.timestamp,
            level: s.outcome === 'executed' ? 'success' : 'warning'
          })),
          ...orders.slice(-10).map((o: any) => ({
            id: `ord-${o.ID}`,
            type: 'order',
            message: `Order ${o.Status.toUpperCase()}: ${o.Side.toUpperCase()} ${o.Quantity} ${o.Symbol}`,
            timestamp: o.Timestamp,
            level: o.Status === 'filled' ? 'success' : 'info'
          }))
        ];

        setEvents(combined.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()).slice(0, 15));
      } catch (e) {
        console.error(e);
      }
    };

    fetchEvents();
    const interval = setInterval(fetchEvents, 5000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="bg-[#0e1729] rounded-xl border border-[#1e3050] overflow-hidden">
      <div className="p-4 bg-[#121e34] border-b border-[#1e3050] flex items-center gap-2">
        <Bell className="w-4 h-4 text-[#18ffff]" />
        <h2 className="text-sm font-bold uppercase tracking-widest">Live Activity Feed</h2>
      </div>
      <div className="max-h-[400px] overflow-y-auto">
        {events.length === 0 ? (
          <div className="p-10 text-center text-[#5d7490] text-sm italic">Waiting for market activity...</div>
        ) : (
          events.map(event => (
            <div key={event.id} className="p-4 border-b border-[#1e3050]/50 hover:bg-[#121e34] transition-colors flex gap-3">
              <div className="mt-1">
                {event.type === 'signal' ? <Zap className="w-3 h-3 text-[#ffab40]" /> : <Shield className="w-3 h-3 text-[#18ffff]" />}
              </div>
              <div className="flex-1">
                <p className="text-xs text-[#d0dced] leading-relaxed">{event.message}</p>
                <p className="text-[10px] text-[#5d7490] mt-1 font-mono">
                  {new Date(event.timestamp).toLocaleTimeString()}
                </p>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};
