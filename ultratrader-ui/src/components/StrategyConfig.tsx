import React, { useState } from 'react';

interface Parameter {
  name: string;
  type: 'number' | 'string' | 'boolean';
  value: any;
  description?: string;
}

interface StrategyConfigProps {
  strategyName: string;
  parameters: Parameter[];
  onSave: (strategyName: string, updatedParams: Record<string, any>) => void;
}

const StrategyConfig: React.FC<StrategyConfigProps> = ({ strategyName, parameters, onSave }) => {
  const [params, setParams] = useState<Record<string, any>>(
    parameters.reduce((acc, param) => ({ ...acc, [param.name]: param.value }), {})
  );

  const handleChange = (name: string, value: any) => {
    setParams(prev => ({ ...prev, [name]: value }));
  };

  const handleSave = () => {
    onSave(strategyName, params);
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow-md">
      <h3 className="text-xl font-bold mb-4">{strategyName} Configuration</h3>
      <div className="space-y-4">
        {parameters.map((param) => (
          <div key={param.name} className="flex flex-col">
            <label className="text-sm font-medium text-gray-700 mb-1" title={param.description}>
              {param.name}
            </label>
            {param.type === 'number' && (
              <input
                type="number"
                className="border rounded-md px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={params[param.name]}
                onChange={(e) => handleChange(param.name, Number(e.target.value))}
              />
            )}
            {param.type === 'string' && (
              <input
                type="text"
                className="border rounded-md px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={params[param.name]}
                onChange={(e) => handleChange(param.name, e.target.value)}
              />
            )}
            {param.type === 'boolean' && (
              <label className="flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  className="form-checkbox h-5 w-5 text-blue-600"
                  checked={params[param.name]}
                  onChange={(e) => handleChange(param.name, e.target.checked)}
                />
                <span className="ml-2 text-gray-700">{param.name}</span>
              </label>
            )}
          </div>
        ))}
      </div>
      <div className="mt-6">
        <button
          onClick={handleSave}
          className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
        >
          Save Configuration
        </button>
      </div>
    </div>
  );
};

export default StrategyConfig;
