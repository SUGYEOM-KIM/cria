import { useState } from 'react';
import characterImg from './assets/images/logo-universal.png';
import './App.css';

function App() {
  const [activeTab, setActiveTab] = useState('chat');
  const [messages, setMessages] = useState<string[]>([]);
  const [inputText, setInputText] = useState('');
  const [selectedModel, setSelectedModel] = useState('llama3');
  const availableModels = ['llama3', 'mistral', 'gemma', 'phi3'];

  const handleSendMessage = () => {
    if (!inputText.trim()) return;
    setMessages([...messages, inputText]);
    setInputText('');
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
            <h2>🤖 AI Agent</h2>
            <p>A screen to select or configure various AI agents.</p>
          </div>
        );
      case 'upgrade':
        return (
          <div className="placeholder-view">
            <h2>🚀 Upgrade Cria</h2>
            <p>A screen to train the agent or upgrade its features.</p>
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <div className="app-container">
      <nav className="sidebar">
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
      </nav>

      <main className="main-content">
        {renderContent()}
      </main>
    </div>
  );
}

export default App;