import React from 'react';

interface AlertDialogProps {
  isOpen: boolean;
  title: string;
  message: React.ReactNode;
  buttonText?: string;
  onClose: () => void;
}

const AlertDialog: React.FC<AlertDialogProps> = ({
  isOpen,
  title,
  message,
  buttonText = 'OK',
  onClose
}) => {
  if (!isOpen) return null;

  return (
    <div style={{
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
      backgroundColor: 'rgba(240, 237, 230, 0.7)',
      backdropFilter: 'blur(3px)',
      display: 'flex', justifyContent: 'center', alignItems: 'center',
      zIndex: 9999
    }}>
      <div style={{
        background: '#fff', padding: '32px', borderRadius: '16px',
        boxShadow: '0 10px 30px rgba(0,0,0,0.1)', width: '340px', textAlign: 'center'
      }}>
        <div style={{ fontSize: '32px', marginBottom: '12px' }}>✅</div>
        <h3 style={{ margin: '0 0 16px 0', fontSize: '20px', color: '#2b2722' }}>{title}</h3>
        <div style={{ marginBottom: '28px', color: '#555', fontSize: '15px', lineHeight: '1.5' }}>
          {message}
        </div>
        <button
          onClick={onClose}
          style={{
            padding: '10px 24px', borderRadius: '8px', border: 'none',
            background: '#76695b', color: '#fff', cursor: 'pointer',
            fontWeight: 600, fontSize: '14px', width: '100%'
          }}
        >
          {buttonText}
        </button>
      </div>
    </div>
  );
};

export default AlertDialog;