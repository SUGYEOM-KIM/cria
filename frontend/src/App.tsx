import { useState } from 'react';
import characterImg from './assets/images/logo-universal.png';
import './App.css';

function App() {
  const [activeTab, setActiveTab] = useState('chat');
  const [messages, setMessages] = useState<string[]>([]);
  const [inputText, setInputText] = useState('');

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

            <div className="chat-input-area">
              <input 
                type="text" 
                value={inputText}
                onChange={(e) => setInputText(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleSendMessage()}
                placeholder="Message Cria..."
                className="chat-input"
              />
              <button onClick={handleSendMessage} className="chat-submit-btn">
                Send
              </button>
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