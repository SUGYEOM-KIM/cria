import React, { useState, useEffect } from 'react';
import { GetOllamaPath, UpdateOllamaPath, SelectFolder, DownloadModel } from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

interface SettingsViewProps {
  availableModels: string[];
  fetchModels: () => Promise<void>;
  popularModels: string[];
}

const SettingsView: React.FC<SettingsViewProps> = ({ availableModels, fetchModels, popularModels }) => {
  const [ollamaPath, setOllamaPath] = useState('');
  const [isModelsExpanded, setIsModelsExpanded] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState<Record<string, string>>({});
  const [isDownloading, setIsDownloading] = useState<Record<string, boolean>>({});

  useEffect(() => {
    const fetchPath = async () => {
      try {
        const path = await GetOllamaPath();
        setOllamaPath(path);
      } catch (err) {
        console.error(err);
      }
    };
    fetchPath();

    popularModels.forEach(model => {
      EventsOn(`download-progress-${model}`, (data) => {
        setDownloadProgress(prev => ({ ...prev, [model]: data }));
      });
    });

    return () => {
      popularModels.forEach(model => {
        EventsOff(`download-progress-${model}`);
      });
    };
  }, [popularModels]);

  const handleBrowseFolder = async () => {
    try {
      const selected = await SelectFolder();
      if (selected) {
        setOllamaPath(selected);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const handleSavePath = async () => {
    try {
      const success = await UpdateOllamaPath(ollamaPath);
      if (success) {
        alert('Settings saved. Restarting Ollama engine...');
        setTimeout(() => {
          fetchModels();
        }, 3000);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const startDownload = async (modelName: string) => {
    setIsDownloading(prev => ({ ...prev, [modelName]: true }));
    setDownloadProgress(prev => ({ ...prev, [modelName]: 'Starting...' }));
    
    try {
      const result = await DownloadModel(modelName);
      if (result === "Success") {
        fetchModels();
      } else {
        alert(`Failed to download ${modelName}.`);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setIsDownloading(prev => ({ ...prev, [modelName]: false }));
    }
  };

  return (
    <div className="placeholder-view" style={{ width: '100%', maxWidth: '800px', margin: '0 auto' }}>
      <h2>Settings</h2>
      <div className="settings-separator"></div>
      
      <div style={{ width: '100%', marginBottom: '40px' }}>
        <label style={{ display: 'block', marginBottom: '8px', fontWeight: '600', textAlign: 'left' }}>Ollama Models Path</label>
        <div className="settings-path-container">
          <input 
            type="text" 
            value={ollamaPath} 
            onChange={(e) => setOllamaPath(e.target.value)}
            className="settings-input-field" 
          />
          <button onClick={handleBrowseFolder} className="settings-browse-btn">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
            </svg>
          </button>
        </div>
        <button onClick={handleSavePath} className="chat-submit-btn" style={{ borderRadius: '8px', width: 'auto', height: 'auto', padding: '12px 24px' }}>
          Save Settings
        </button>
      </div>

      <div style={{ width: '100%' }}>
        <h3 style={{ fontSize: '18px', color: '#2b2722', marginBottom: '16px', textAlign: 'left' }}>Model Management</h3>
        
        <div className="models-accordion">
          <div 
            className="accordion-header"
            onClick={() => setIsModelsExpanded(!isModelsExpanded)}
          >
            <span>Popular Models</span>
            <span>{isModelsExpanded ? '▲' : '▼'}</span>
          </div>
          
          {isModelsExpanded && (
            <div className="accordion-content">
              {popularModels.map((modelName) => {
                const isInstalled = availableModels.includes(modelName) || availableModels.some(m => m.startsWith(modelName + ':'));
                const downloading = isDownloading[modelName];
                const progress = downloadProgress[modelName];

                return (
                  <div key={modelName} className="model-list-item">
                    <div className="model-info">
                      <span className="model-name">{modelName}</span>
                      {isInstalled ? (
                        <span className="status-badge status-installed">Installed</span>
                      ) : (
                        <span className="status-badge status-not-installed">Not Installed</span>
                      )}
                    </div>
                    
                    <div className="action-area">
                      {downloading ? (
                        <span className="progress-text">{progress || 'Downloading...'}</span>
                      ) : !isInstalled ? (
                        <button 
                          onClick={() => startDownload(modelName)}
                          className="download-btn"
                        >
                          Download
                        </button>
                      ) : null}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SettingsView;