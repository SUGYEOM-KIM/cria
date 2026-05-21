import React from 'react';

interface ConfirmDialogProps {
  isOpen: boolean;
  title: string;
  message: React.ReactNode;
  confirmText?: string;
  cancelText?: string;
  onConfirm: () => void;
  onCancel: () => void;
  confirmColor?: string;
}

const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  isOpen,
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  onConfirm,
  onCancel,
  confirmColor = '#d32f2f' // 기본값: 빨간색 (Settings의 Remove와 동일)
}) => {
  if (!isOpen) return null;

  return (
    <div style={{
      position: 'fixed',
      top: 0, left: 0, right: 0, bottom: 0,
      backgroundColor: 'rgba(240, 237, 230, 0.7)', // 앱 테마에 맞는 반투명 배경
      backdropFilter: 'blur(3px)',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      zIndex: 9999
    }}>
      <div style={{
        background: '#fff',
        padding: '32px',
        borderRadius: '16px',
        boxShadow: '0 10px 30px rgba(0,0,0,0.1)',
        width: '380px',
        textAlign: 'center'
      }}>
        <h3 style={{ margin: '0 0 16px 0', fontSize: '20px', color: '#2b2722' }}>{title}</h3>
        <div style={{ marginBottom: '28px', color: '#555', fontSize: '15px', lineHeight: '1.5' }}>
          {message}
        </div>
        <div style={{ display: 'flex', justifyContent: 'center', gap: '12px' }}>
          <button
            onClick={onCancel}
            style={{
              padding: '10px 24px',
              borderRadius: '8px',
              border: '1px solid #e1dacb',
              background: '#f9f6f0',
              color: '#706558',
              cursor: 'pointer',
              fontWeight: 600,
              fontSize: '14px'
            }}
          >
            {cancelText}
          </button>
          <button
            onClick={onConfirm}
            style={{
              padding: '10px 24px',
              borderRadius: '8px',
              border: 'none',
              background: confirmColor,
              color: '#fff',
              cursor: 'pointer',
              fontWeight: 600,
              fontSize: '14px'
            }}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmDialog;