import { useState, useEffect } from 'react';
import characterImg from './assets/images/detective_cria.png';
import { GetOllamaModels, GetOllamaPath, UpdateOllamaPath, SelectFolder, DownloadModel } from '../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
import './App.css';

function App() {
  const [activeTab, setActiveTab] = useState('chat');
  const [messages, setMessages] = useState<string[]>([]);
  const [inputText, setInputText] = useState('');
  const [selectedModel, setSelectedModel] = useState('');
  const [availableModels, setAvailableModels] = useState<string[]>([]);
  const [ollamaPath, setOllamaPath] = useState('');

  const [isModelsExpanded, setIsModelsExpanded] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState<Record<string, string>>({});
  const [isDownloading, setIsDownloading] = useState<Record<string, boolean>>({});

  const popularModels = ['llama3', 'mistral', 'gemma:2b', 'phi3'];

  const fetchModels = async () => {
    try {
      const models = await GetOllamaModels();
      if (models && models.length > 0) {
        setAvailableModels(models);
        setSelectedModel(models[0]);
      } else {
        setAvailableModels(['No models available']);
        setSelectedModel('No models available');
      }
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    fetchModels();
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
  }, []);

  const handleSendMessage = () => {
    if (!inputText.trim()) return;
    setMessages([...messages, inputText]);
    setInputText('');
  };

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

  const renderContent = () => {
    switch (activeTab) {
      case 'chat':
        return (
          <div className="chat-container">
            {messages.length === 0 ? (
              <div className="home-view">
                <img src={characterImg} alt="My Character" className="character-image" />
                <h1>Hello! I am Cria.</h1>
                <p>Have a great day! How can I help you today?</p>
              </div>
            ) : (
              <div className="messages-view">
                {messages.map((msg, index) => (
                  <div key={index} className="user-message">
                    {msg}
                  </div>
                ))}
              </div>
            )}

            <div className="chat-input-wrapper">
              <input 
                type="text" 
                value={inputText}
                onChange={(e) => setInputText(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleSendMessage()}
                placeholder="Message Cria..."
                className="chat-input"
              />
              <div className="chat-input-bottom">
                <div className="chat-input-actions-left">
                </div>
                <div className="chat-input-actions-right">
                  <select 
                    value={selectedModel} 
                    onChange={(e) => setSelectedModel(e.target.value)}
                    className="model-selector-inline"
                  >
                    {availableModels.map((model) => (
                      <option key={model} value={model}>
                        {model}
                      </option>
                    ))}
                  </select>
                  <button onClick={handleSendMessage} className="chat-submit-btn">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <line x1="22" y1="2" x2="11" y2="13"></line>
                      <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
        );
      case 'agent':
        return (
          <div className="placeholder-view">
            <h2>AI Agent</h2>
            <p>A screen to select or configure various AI agents.</p>
          </div>
        );
      case 'upgrade':
        return (
          <div className="placeholder-view">
            <h2>Upgrade Cria</h2>
            <p>A screen to train the agent or upgrade its features.</p>
          </div>
        );
      case 'settings':
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
      default:
        return null;
    }
  };

  return (
    <div className="app-container">
      <nav className="sidebar">
        <div className="sidebar-top">
          <div className="sidebar-logo">
            <h2>Cria AI</h2>
          </div>
          <ul className="nav-menu">
            <li 
              className={activeTab === 'chat' ? 'active' : ''} 
              onClick={() => setActiveTab('chat')}
            >
              New Chat
            </li>
            <li 
              className={activeTab === 'agent' ? 'active' : ''} 
              onClick={() => setActiveTab('agent')}
            >
              AI Agent
            </li>
            <li 
              className={activeTab === 'upgrade' ? 'active' : ''} 
              onClick={() => setActiveTab('upgrade')}
            >
              Upgrade Cria
            </li>
          </ul>
        </div>
        <div className="sidebar-bottom">
          <ul className="nav-menu">
            <li 
              className={activeTab === 'settings' ? 'active' : ''} 
              onClick={() => setActiveTab('settings')}
            >
              Settings
            </li>
          </ul>
        </div>
      </nav>

      <main className="main-content">
        {renderContent()}
      </main>
    </div>
  );
}

export default App;