import React, { useState, useEffect } from 'react';
import { GetUpgradeHistory, RollbackUpgrade, GetActiveCommit, ApplyUpgrade, GetInitialCommit } from '../../../wailsjs/go/main/App';
import ConfirmDialog from '../common/ConfirmDialog';
import AlertDialog from '../common/AlertDialog';

interface UpgradeHistory {
  version: string;
  hash: string;
  message: string;
  date: string;
  time: string;
  isAutoUpgrade: boolean;
}

interface Props {
  onApplySuccess: () => void;
}

const VersionHistoryView: React.FC<Props> = ({ onApplySuccess }) => {
  const [history, setHistory] = useState<UpgradeHistory[]>([]);
  const [activeCommit, setActiveCommit] = useState<string>('');
  const [initialCommit, setInitialCommit] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);
  const [isAlertOpen, setIsAlertOpen] = useState(false);
  const [alertContent, setAlertContent] = useState({ title: '', message: '' });
  const [selectedItem, setSelectedItem] = useState<UpgradeHistory | null>(null);

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const currentActive = await GetActiveCommit();
      setActiveCommit(currentActive || '');

      const firstCommit = await GetInitialCommit();
      setInitialCommit(firstCommit || '');

      const data = await GetUpgradeHistory();
      setHistory(data || []);
    } catch (err) {
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleUndoClick = (item: UpgradeHistory) => {
    setSelectedItem(item);
    setIsConfirmOpen(true);
  };

  const executeUndo = async () => {
    if (!selectedItem) return;
    setIsConfirmOpen(false);
    try {
      await RollbackUpgrade(selectedItem.hash);
      setHistory([]);
      setAlertContent({
        title: 'Undo Successful',
        message: `Modification [${selectedItem.message}] has been removed.`
      });
      setIsAlertOpen(true);
      fetchData();
    } catch (err) {
      setAlertContent({ title: 'Undo Failed', message: String(err) });
      setIsAlertOpen(true);
    }
  };

  const handleApplyClick = async (item: UpgradeHistory) => {
    try {
      const applyVersion = item.version || item.hash;
      await ApplyUpgrade(item.hash, applyVersion);
      onApplySuccess();
      setAlertContent({
        title: 'Application Rebuilt',
        message: `Version ${applyVersion} build triggered successfully. Restarting application...`
      });
      setIsAlertOpen(true);
      fetchData();
    } catch (err) {
      setAlertContent({
        title: 'Apply Failed',
        message: String(err)
      });
      setIsAlertOpen(true);
    }
  };

  const getFilteredHistory = () => {
    if (!history || history.length === 0) return [];

    const cleanActiveCommit = activeCommit.replace('-dirty', '').toLowerCase();
    const cleanInitialCommit = initialCommit.replace('-dirty', '').toLowerCase();

    let validRange = history;
    if (cleanInitialCommit && cleanInitialCommit !== 'dev-mode-hash') {
      const initialIdx = history.findIndex(item =>
        item.hash.toLowerCase().startsWith(cleanInitialCommit) ||
        cleanInitialCommit.startsWith(item.hash.toLowerCase())
      );
      if (initialIdx !== -1) {
        validRange = history.slice(0, initialIdx + 1);
      }
    }

    return validRange.filter((item) => {
      const isCurrent = Boolean(
        cleanActiveCommit && cleanActiveCommit !== 'dev-mode-hash' &&
        (item.hash.toLowerCase().startsWith(cleanActiveCommit) || cleanActiveCommit.startsWith(item.hash.toLowerCase()))
      );

      const isInitial = Boolean(
        cleanInitialCommit && cleanInitialCommit !== 'dev-mode-hash' &&
        (item.hash.toLowerCase().startsWith(cleanInitialCommit) || cleanInitialCommit.startsWith(item.hash.toLowerCase()))
      );

      const isAiCommit = item.isAutoUpgrade || 
                         (item.message && item.message.toLowerCase().includes('auto-upgrade'));

      return isCurrent || isInitial || isAiCommit;
    });
  };

  const visibleHistory = getFilteredHistory();

  return (
    <div style={{ padding: '30px', color: '#2b2722', width: '100%', boxSizing: 'border-box' }}>
      <ConfirmDialog
        isOpen={isConfirmOpen}
        title="Undo Modification"
        message={<>Undo <strong>{selectedItem?.message}</strong>? History will be deleted.</>}
        confirmText="Undo"
        onConfirm={executeUndo}
        onCancel={() => setIsConfirmOpen(false)}
      />
      <AlertDialog
        isOpen={isAlertOpen}
        title={alertContent.title}
        message={alertContent.message}
        onClose={() => setIsAlertOpen(false)}
      />

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '32px' }}>
        <div style={{ textAlign: 'left' }}>
          <h2 style={{ margin: '0 0 16px 0', fontSize: '24px', fontWeight: 'bold' }}>Upgrade History</h2>
          <p style={{ margin: 0, color: '#706558', fontSize: '14px' }}>Manage automated codebase modifications.</p>
        </div>
        <button onClick={fetchData} style={{ background: '#fff', border: '1px solid #e1dacb', padding: '8px 16px', borderRadius: '6px', cursor: 'pointer', fontSize: '14px', display: 'flex', alignItems: 'center', gap: '6px' }}>
          <span>🔄</span> Refresh
        </button>
      </div>

      <div style={{ background: '#fff', borderRadius: '8px', border: '1px solid #e1dacb', width: '100%', overflow: 'hidden' }}>
        {isLoading ? (
          <div style={{ padding: '30px', textAlign: 'center' }}>Loading...</div>
        ) : visibleHistory.length === 0 ? (
          <div style={{ padding: '30px', textAlign: 'center' }}>No history found.</div>
        ) : (
          visibleHistory.map((item, idx) => {
            const cleanActiveCommit = activeCommit.replace('-dirty', '').toLowerCase();
            const cleanInitialCommit = initialCommit.replace('-dirty', '').toLowerCase();
            
            const isCurrent = Boolean(
              cleanActiveCommit && cleanActiveCommit !== 'dev-mode-hash' &&
              (item.hash.toLowerCase().startsWith(cleanActiveCommit) || cleanActiveCommit.startsWith(item.hash.toLowerCase()))
            );

            const isInitial = Boolean(
              item.hash && cleanInitialCommit && cleanInitialCommit !== 'dev-mode-hash' && (
                cleanInitialCommit.startsWith(item.hash.toLowerCase()) ||
                item.hash.toLowerCase().startsWith(cleanInitialCommit)
              )
            );
            
            const canUndo = visibleHistory.length > 0 && visibleHistory[0].hash === item.hash && !isCurrent && !isInitial;

            return (
              <div key={item.hash} style={{ 
                display: 'flex', 
                justifyContent: 'space-between', 
                alignItems: 'center', 
                padding: '16px 20px', 
                borderBottom: idx === visibleHistory.length - 1 ? 'none' : '1px solid #e1dacb', 
                background: isCurrent ? '#fdfbf7' : 'transparent'
              }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '16px', flex: 1 }}>
                  <span style={{ 
                    fontSize: '12px', 
                    color: isCurrent ? '#bd9313' : '#1e8e3e', 
                    background: isCurrent ? '#fef7e0' : '#e6f4ea', 
                    padding: '4px 8px', 
                    borderRadius: '4px', 
                    fontWeight: 'bold', 
                    minWidth: '60px', 
                    textAlign: 'center' 
                  }}>
                    {item.version || 'No Tag'}
                  </span>
                  <span style={{ fontSize: '12px', color: '#4a90e2', background: '#f0f7ff', padding: '4px 8px', borderRadius: '4px', fontFamily: 'monospace' }}>
                    {item.hash ? item.hash.substring(0, 7) : ''}
                  </span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flex: 1, textAlign: 'left' }}>
                    <span style={{ fontSize: '15px', color: '#2b2722', fontWeight: 500 }}>
                      {item.message}
                    </span>
                    {isCurrent && (
                      <span style={{ fontSize: '11px', color: '#bd9313', background: '#fffbeb', border: '1px solid #fde8bb', padding: '1px 6px', borderRadius: '3px', fontWeight: 'bold' }}>
                        Running
                      </span>
                    )}
                  </div>
                  <span style={{ fontSize: '13px', color: '#a0968a', minWidth: '150px', textAlign: 'right' }}>
                    {item.date} {item.time}
                  </span>
                </div>
                
                <div style={{ display: 'flex', gap: '8px', marginLeft: '20px' }}>
                  {isCurrent ? (
                    <button disabled style={{ background: '#f1ede4', border: '1px solid #e1dacb', color: '#a0968a', padding: '6px 16px', borderRadius: '4px', fontSize: '13px', cursor: 'not-allowed' }}>
                      Active
                    </button>
                  ) : (
                    <button onClick={() => handleApplyClick(item)} style={{ background: '#1e8e3e', border: '1px solid #1e8e3e', color: '#fff', padding: '6px 16px', borderRadius: '4px', cursor: 'pointer', fontSize: '13px' }}>
                      Apply
                    </button>
                  )}
                  <button onClick={() => handleUndoClick(item)} disabled={!canUndo} style={{ background: 'transparent', border: canUndo ? '1px solid #ef4444' : '1px solid #e1dacb', color: canUndo ? '#ef4444' : '#a0968a', padding: '6px 16px', borderRadius: '4px', cursor: canUndo ? 'pointer' : 'not-allowed', fontSize: '13px' }}>
                    Undo
                  </button>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};

export default VersionHistoryView;