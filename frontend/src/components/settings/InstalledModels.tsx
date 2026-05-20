import React from 'react';

interface InstalledModelsProps {
  availableModels: string[];
  onRemoveRequest: (modelName: string) => void;
}

const InstalledModels: React.FC<InstalledModelsProps> = ({ availableModels, onRemoveRequest }) => {
  return (
    <div>
      <h4 style={{ fontSize: '15px', color: '#706558', marginBottom: '12px', textAlign: 'left' }}>Installed Models</h4>
      <div style={{ border: '1px solid #e1dacb', borderRadius: '12px', background: '#ffffff', overflow: 'hidden' }}>
        {availableModels.length === 0 || availableModels[0] === 'No models available' ? (
          <div style={{ padding: '24px', color: '#706558', fontSize: '14px', textAlign: 'center' }}>
            No models installed. Pull a model above to get started.
          </div>
        ) : (
          availableModels.map((modelName, index) => (
            <div 
              key={modelName} 
              className="model-list-item" 
              style={{ 
                borderBottom: index === availableModels.length - 1 ? 'none' : '1px solid #f1ede4',
                margin: 0,
                borderRadius: 0
              }}
            >
              <div className="model-info">
                <span className="model-name">{modelName}</span>
                <span className="status-badge status-installed">Installed</span>
              </div>
              
              <div className="action-area">
                <button 
                  onClick={() => onRemoveRequest(modelName)}
                  className="remove-btn"
                >
                  Remove
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default InstalledModels;