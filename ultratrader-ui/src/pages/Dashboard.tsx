import { useEffect, useState } from 'react';
import { fetchWithAuth } from '../utils/api';

const Dashboard = () => {
  const [wsHealth, setWsHealth] = useState<{ connected: boolean; last_message_time: string; staleness_ms: number } | null>(null);

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        const response = await fetchWithAuth('/api/ws-health');
        if (response.ok) {
          const data = await response.json();
          setWsHealth(data);
        }
      } catch (error) {
        console.error('Failed to fetch WS health:', error);
      }
    };

    fetchHealth();
    const interval = setInterval(fetchHealth, 5000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold text-gray-800">BobTrader Dashboard</h1>
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-600">WebSocket Status:</span>
          {wsHealth ? (
            <div className={`flex items-center gap-1 ${wsHealth.connected ? 'text-green-600' : 'text-red-600'}`}>
              <div className={`w-3 h-3 rounded-full ${wsHealth.connected ? 'bg-green-500' : 'bg-red-500'}`}></div>
              <span className="text-sm font-medium">{wsHealth.connected ? 'Connected' : 'Disconnected'}</span>
              {wsHealth.connected && (
                  <span className="text-xs text-gray-500 ml-2">(Delay: {wsHealth.staleness_ms}ms)</span>
              )}
            </div>
          ) : (
            <span className="text-sm text-gray-400">Loading...</span>
          )}
        </div>
      </div>
      <p className="text-gray-600">Welcome to the autonomous trading platform.</p>
    </div>
  );
};

export default Dashboard;
