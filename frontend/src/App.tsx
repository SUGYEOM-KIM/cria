import { useState, useEffect } from 'react';
import { GetOllamaModels } from '../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
import Sidebar from './components/Sidebar';
import ChatView from './components/ChatView';
import SettingsView from './components/settings/SettingsView';
import UpgradeView from './components/upgrade/UpgradeView';
import VersionHistoryView from './components/upgrade/VersionHistoryView';
import './App.css';

function App() {
  const [activeTab, setActiveTab] = useState('chat');
  const [sidebarRefresh, setSidebarRefresh] = useState<number>(0);
  const [selectedModel, setSelectedModel] = useState('');
  const [availableModels, setAvailableModels] = useState<string[]>([]);
  const [downloadProgress, setDownloadProgress] = useState<Record<string, string>>({});
  const [isDownloading, setIsDownloading] = useState<Record<string, boolean>>({});
  const popularModels = ['llama3', 'mistral', 'gemma:2b', 'phi3'];

  const fetchModels = async () => {
    try {
      const models = await GetOllamaModels();
      if (models && models.length > 0) {
        setAvailableModels(models);
        if (!selectedModel || selectedModel === 'No models available' || !models.includes(selectedModel)) {
          setSelectedModel(models[0]);
        }
      } else {
        setAvailableModels(prev => (prev.length > 0 && prev[0] !== 'No models available') ? prev : ['No models available']);
        setSelectedModel(prev => prev || 'No models available');
      }
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    fetchModels();

    EventsOn('ollama-ready', () => {
      fetchModels();
    });

    popularModels.forEach(model => {
      EventsOn(`download-progress-${model}`, (data) => {
        setDownloadProgress(prev => ({ ...prev, [model]: data }));
      });
    });

    return () => {
      EventsOff('ollama-ready');
      popularModels.forEach(model => {
        EventsOff(`download-progress-${model}`);
      });
    };
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
            downloadProgress={downloadProgress}
            setDownloadProgress={setDownloadProgress}
            isDownloading={isDownloading}
            setIsDownloading={setIsDownloading}
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
        return <UpgradeView />;
      case 'history':
        return <VersionHistoryView onApplySuccess={() => setSidebarRefresh((prev: number) => prev + 1)} />;
      default:
        return null;
    }
  };

  return (
    <div className="app-container">
      <Sidebar activeTab={activeTab} setActiveTab={setActiveTab} refreshKey={sidebarRefresh} />
      <main className="main-content">
        {renderContent()}
      </main>
    </div>
  );
}

export default App;