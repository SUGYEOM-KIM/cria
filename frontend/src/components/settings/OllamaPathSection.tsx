import React, { useState, useEffect } from 'react';
import { GetOllamaPath, UpdateOllamaPath, SelectFolder } from '../../../wailsjs/go/main/App';
import { EventsOnce } from '../../../wailsjs/runtime/runtime';
import './Settings.css';

interface OllamaPathSectionProps {
    fetchModels: () => Promise<void>;
}

const OllamaPathSection: React.FC<OllamaPathSectionProps> = ({ fetchModels }) => {
    const [ollamaPath, setOllamaPath] = useState('');
    const [toastMessage, setToastMessage] = useState('');
    const [isToastVisible, setIsToastVisible] = useState(false);

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
    }, []);

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

    const showToast = (message: string, durationMs: number = 3000) => {
        setToastMessage(message);
        setIsToastVisible(true);
        if (durationMs > 0) {
            setTimeout(() => {
                setIsToastVisible(false);
            }, durationMs);
        }
    };

    const handleSavePath = async () => {
        try {
            const success = await UpdateOllamaPath(ollamaPath);
            if (!success) return;

            showToast('Restarting Ollama engine...', 0);

            let resolved = false;
            const finish = (toastMsg: string) => {
                if (resolved) return;
                resolved = true;
                showToast(toastMsg);
            };

            EventsOnce('ollama-ready', () => finish('Settings saved.'));
            setTimeout(() => finish('Settings saved (timed out waiting for engine).'), 30000);
        } catch (err) {
            console.error(err);
        }
    };

    return (
        <div style={{ width: '100%', marginBottom: '40px' }}>
            <label style={{ display: 'block', marginBottom: '8px', fontWeight: '600', textAlign: 'left' }}>Ollama Models Path</label>

            <div className="settings-path-container" style={{ display: 'flex', gap: '8px', width: '100%', marginBottom: '12px' }}>
                <input
                    type="text"
                    value={ollamaPath}
                    onChange={(e) => setOllamaPath(e.target.value)}
                    className="settings-input-field"
                    style={{ flex: 1, minWidth: 0 }}
                />
                <button onClick={handleBrowseFolder} className="settings-browse-btn" style={{ flexShrink: 0 }}>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
                    </svg>
                </button>
            </div>

            <button onClick={handleSavePath} className="primary-action-btn" style={{ padding: '12px 24px' }}>
                Save Settings
            </button>

            {isToastVisible && (
                <div className="custom-toast">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                        <polyline points="22 4 12 14.01 9 11.01"></polyline>
                    </svg>
                    <span>{toastMessage}</span>
                </div>
            )}
        </div>
    );
};

export default OllamaPathSection;