import React, { useState } from 'react';
import { RemoveModel } from '../../../wailsjs/go/main/App';
import OllamaPathSection from './OllamaPathSection';
import CustomModelPull from './CustomModelPull';
import InstalledModels from './InstalledModels';
import ConfirmModal from './ConfirmModal';

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
  const [isRemoving, setIsRemoving] = useState(false); // ★ 삭제 진행 상태 추가

  const handleRemoveModelRequest = (modelName: string) => {
    const exactName = availableModels.find(m => m === modelName || m.startsWith(modelName + ':')) || modelName;
    setModelToRemove(exactName);
    setIsModalOpen(true);
  };

  const confirmRemoveModel = async () => {
    setIsRemoving(true); // ★ 모달창 버튼을 로딩 상태로 변경
    
    try {
      const result = await RemoveModel(modelToRemove);
      if (result === "Success") {
        await fetchModels(); // 모델 리스트 새로고침 완료 대기
      } else {
        alert(`Failed to remove model: ${result}`);
      }
    } catch (err) {
      console.error(err);
      alert('Error removing model.');
    } finally {
      setIsRemoving(false); // ★ 로딩 상태 해제
      setIsModalOpen(false); // ★ 모달창 닫기
    }
  };

  return (
    <div className="placeholder-view" style={{ width: '100%', maxWidth: '800px', margin: '0 auto' }}>
      <h2>Settings</h2>
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

      <ConfirmModal 
        isOpen={isModalOpen}
        modelName={modelToRemove}
        onClose={() => !isRemoving && setIsModalOpen(false)} // 삭제 중일 땐 바깥 클릭 등으로 닫히는 것 방지
        onConfirm={confirmRemoveModel}
        isRemoving={isRemoving} // ★ 모달에 상태 전달
      />
    </div>
  );
};

export default SettingsView;