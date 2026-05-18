import { useState, useEffect } from 'react';
import { GetOllamaModels } from '../wailsjs/go/main/App';
import Sidebar from './components/Sidebar';
import ChatView from './components/ChatView';
import SettingsView from './components/SettingsView';
import './App.css';

function App() {
  const [activeTab, setActiveTab] = useState('chat');
  const [selectedModel, setSelectedModel] = useState('');
  const [availableModels, setAvailableModels] = useState<string[]>([]);
  const popularModels = ['llama3', 'mistral', 'gemma:2b', 'phi3'];

  const fetchModels = async () => {
    try {
      const models = await GetOllamaModels();
      if (models && models.length > 0) {
        setAvailableModels(models);
        if (!selectedModel || !models.includes(selectedModel)) {
           setSelectedModel(models[0]);
        }
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
  }, []);

  const renderContent = () => {
    switch (activeTab) {
      case 'chat':
        return (
          <ChatView 
            selectedModel={selectedModel} 
            availableModels={availableModels} 
            setSelectedModel={setSelectedModel} 
          />
        );
      case 'settings':
        return (
          <SettingsView 
            availableModels={availableModels} 
            fetchModels={fetchModels} 
            popularModels={popularModels} 
          />
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
      default:
        return null;
    }
  };

  return (
    <div className="app-container">
      <Sidebar activeTab={activeTab} setActiveTab={setActiveTab} />
      <main className="main-content">
        {renderContent()}
      </main>
    </div>
  );
}

export default App;