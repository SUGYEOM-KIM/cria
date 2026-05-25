import React, { useState, useEffect } from 'react';
import { RemoveModel } from '../../../wailsjs/go/main/App';
import OllamaPathSection from './OllamaPathSection';
import CustomModelPull from './CustomModelPull';
import InstalledModels from './InstalledModels';
import TranslationLanguageSection from './TranslationLanguageSection';
import ConfirmDialog from '../common/ConfirmDialog'; 

interface SettingsViewProps {
  availableModels: string[];
  fetchModels: () => Promise<void>;
  downloadProgress: Record<string, string>;
  setDownloadProgress: React.Dispatch<React.SetStateAction<Record<string, string>>>;
  isDownloading: Record<string, boolean>;
  setIsDownloading: React.Dispatch<React.SetStateAction<Record<string, boolean>>>;
}

const SettingsView: React.FC<SettingsViewProps> = ({ 
  availableModels, 
  fetchModels, 
  downloadProgress,
  setDownloadProgress,
  isDownloading,
  setIsDownloading
}) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modelToRemove, setModelToRemove] = useState('');
  const [isRemoving, setIsRemoving] = useState(false);

  useEffect(() => {
    fetchModels();
  }, []);

  const handleRemoveModelRequest = (modelName: string) => {
    const exactName = availableModels.find(m => m === modelName || m.startsWith(modelName + ':')) || modelName;
    setModelToRemove(exactName);
    setIsModalOpen(true);
  };

  const confirmRemoveModel = async () => {
    setIsRemoving(true);
    
    try {
      const result = await RemoveModel(modelToRemove);
      if (result === "Success") {
        await fetchModels();
      } else {
        alert(`Failed to remove model: ${result}`);
      }
    } catch (err) {
      console.error(err);
      alert('Error removing model.');
    } finally {
      setIsRemoving(false);
      setIsModalOpen(false);
    }
  };

  return (
    <div className="placeholder-view" style={{ width: '100%', maxWidth: '800px', margin: '0 auto' }}>
      <h2>Settings</h2>
      <div className="settings-separator"></div>

      <h3 style={{ fontSize: '18px', color: '#2b2722', marginBottom: '16px', textAlign: 'left' }}>Localization</h3>
      <TranslationLanguageSection />

      <div className="settings-separator"></div>

      <OllamaPathSection fetchModels={fetchModels} />

      <div style={{ width: '100%' }}>
        <h3 style={{ fontSize: '18px', color: '#2b2722', marginBottom: '16px', textAlign: 'left' }}>Model Management</h3>
        
        <CustomModelPull 
          fetchModels={fetchModels}
          downloadProgress={downloadProgress}
          setDownloadProgress={setDownloadProgress}
          isDownloading={isDownloading}
          setIsDownloading={setIsDownloading}
        />
        
        <InstalledModels 
          availableModels={availableModels} 
          onRemoveRequest={handleRemoveModelRequest} 
        />
      </div>

      <ConfirmDialog
        isOpen={isModalOpen}
        title="Remove Model"
        message={
          <>
            Are you sure you want to remove<br/>
            <strong>{modelToRemove}</strong>?
          </>
        }
        confirmText={isRemoving ? 'Removing...' : 'Remove'}
        onConfirm={confirmRemoveModel}
        onCancel={() => !isRemoving && setIsModalOpen(false)}
      />
    </div>
  );
};

export default SettingsView;