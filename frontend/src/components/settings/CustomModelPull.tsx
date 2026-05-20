import React, { useState } from 'react';
import { DownloadModel } from '../../../wailsjs/go/main/App';
import { BrowserOpenURL, EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
import './Settings.css';

interface CustomModelPullProps {
    fetchModels: () => Promise<void>;
    downloadProgress: Record<string, string>;
    setDownloadProgress: React.Dispatch<React.SetStateAction<Record<string, string>>>;
    isDownloading: Record<string, boolean>;
    setIsDownloading: React.Dispatch<React.SetStateAction<Record<string, boolean>>>;
}

const CustomModelPull: React.FC<CustomModelPullProps> = ({
    fetchModels,
    downloadProgress,
    setDownloadProgress,
    isDownloading,
    setIsDownloading
}) => {
    const [customModel, setCustomModel] = useState('');

    const startDownload = async (modelName: string) => {
        setIsDownloading(prev => ({ ...prev, [modelName]: true }));
        setDownloadProgress(prev => ({ ...prev, [modelName]: 'Starting...' }));

        EventsOn(`download-progress-${modelName}`, (data: string) => {
            setDownloadProgress(prev => ({ ...prev, [modelName]: data }));
        });

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
            EventsOff(`download-progress-${modelName}`);
        }
    };

    return (
        <div className="custom-model-pull" style={{ marginBottom: '32px' }}>

            <div style={{ marginBottom: '12px', textAlign: 'left' }}>
                <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', fontWeight: '600', color: '#706558' }}>
                    Pull Custom Model from Ollama
                </label>
                <span style={{ fontSize: '13px', color: '#a0978c' }}>
                    Find models to download at{' '}
                    <span
                        onClick={() => BrowserOpenURL('https://ollama.com/search')}
                        style={{
                            color: '#76695b',
                            fontWeight: '600',
                            textDecoration: 'underline',
                            cursor: 'pointer',
                            transition: 'color 0.2s ease'
                        }}
                        onMouseOver={(e) => e.currentTarget.style.color = '#2b2722'}
                        onMouseOut={(e) => e.currentTarget.style.color = '#76695b'}
                    >
                        ollama.com/search
                    </span>
                </span>
            </div>

            <div className="settings-path-container" style={{ display: 'flex', gap: '8px', width: '100%' }}>
                <div className="input-with-icon" style={{ position: 'relative', flex: 1, display: 'flex', minWidth: 0 }}>
                    <svg style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: '#a0978c' }} width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                        <polyline points="7 10 12 15 17 10"></polyline>
                        <line x1="12" y1="15" x2="12" y2="3"></line>
                    </svg>
                    <input
                        type="text"
                        value={customModel}
                        onChange={(e) => setCustomModel(e.target.value)}
                        placeholder="e.g., llava:7b, qwen2.5:14b..."
                        className="settings-input-field"
                        style={{ flex: 1, paddingLeft: '38px', minWidth: 0 }}
                        onKeyDown={(e) => {
                            if (e.key === 'Enter' && customModel.trim()) {
                                startDownload(customModel.trim());
                                setCustomModel('');
                            }
                        }}
                    />
                </div>
                <button
                    onClick={() => {
                        if (customModel.trim()) {
                            startDownload(customModel.trim());
                            setCustomModel('');
                        }
                    }}
                    className="primary-action-btn"
                    disabled={!customModel.trim()}
                >
                    Pull
                </button>
            </div>

            <div className="download-cards-container">
                {Object.keys(isDownloading).map(model => {
                    if (isDownloading[model]) {
                        const progress = downloadProgress[model];
                        const isStarting = !progress || !progress.includes('%');

                        return (
                            <div key={model} className="enhanced-download-card">
                                <div className="download-card-header">
                                    <div className="download-card-title">
                                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                            <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
                                            <polyline points="2 17 12 22 22 17"></polyline>
                                            <polyline points="2 12 12 17 22 12"></polyline>
                                        </svg>
                                        <span>{model}</span>
                                    </div>
                                    <span className="download-card-status">
                                        {isStarting ? 'Preparing...' : 'Downloading...'}
                                    </span>
                                </div>

                                {isStarting ? (
                                    <div className="download-starting-state">
                                        <div className="loading-spinner"></div>
                                        <span>Connecting to Ollama registry...</span>
                                    </div>
                                ) : (
                                    <div className="enhanced-progress-container">
                                        <div className="enhanced-progress-bar">
                                            <div className="enhanced-progress-fill" style={{ width: progress }}></div>
                                        </div>
                                        <span className="enhanced-progress-percentage">{progress}</span>
                                    </div>
                                )}
                            </div>
                        );
                    }
                    return null;
                })}
            </div>
        </div>
    );
};

export default CustomModelPull;