import React from 'react';

interface ConfirmModalProps {
  isOpen: boolean;
  modelName: string;
  onClose: () => void;
  onConfirm: () => void;
  isRemoving: boolean;
}

const ConfirmModal: React.FC<ConfirmModalProps> = ({ isOpen, modelName, onClose, onConfirm, isRemoving }) => {
  if (!isOpen) return null;
  return (
    <div className="modal-overlay">
      <div className="modal-box">
        <h3>Remove Model</h3>
        <p>Are you sure you want to remove <strong>{modelName}</strong>?</p>
        <div className="modal-actions">
          <button className="modal-btn-cancel" onClick={onClose} disabled={isRemoving}>
            Cancel
          </button>
          <button 
            className="modal-btn-confirm" 
            onClick={onConfirm}
            disabled={isRemoving}
            style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px', minWidth: '95px' }}
          >
            {isRemoving ? (
              <>
                <div className="loading-spinner" style={{ width: '14px', height: '14px', border: '2px solid rgba(255,255,255,0.3)', borderTopColor: '#ffffff' }}></div>
                Removing...
              </>
            ) : (
              'Remove'
            )}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmModal;