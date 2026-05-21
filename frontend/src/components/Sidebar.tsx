import React, { useState, useEffect } from 'react';
import { GetActiveVersion } from '../../wailsjs/go/main/App';
import './Sidebar.css';

interface SidebarProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
}

const Sidebar: React.FC<SidebarProps> = ({ activeTab, setActiveTab }) => {
  const [version, setVersion] = useState<string>('v0.0.0');

  useEffect(() => {
    const fetchVersion = async () => {
      try {
        const v = await GetActiveVersion();
        setVersion(v);
      } catch (err) {
        console.error(err);
      }
    };
    fetchVersion();
  }, [activeTab]);

  return (
    <nav className="sidebar">
      <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
        <div className="sidebar-top" style={{ flex: 1 }}>
          
          <div className="sidebar-logo" style={{ 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'baseline', 
            width: '100%',
            padding: '20px 15px',
            boxSizing: 'border-box'
          }}>
            <span style={{ fontSize: '10px', padding: '2px 6px', visibility: 'hidden' }}>
              {version}
            </span>
            <h2 style={{ margin: 0, fontSize: '1.5rem', textAlign: 'center', flex: 1 }}>
              Cria AI
            </h2>
            <span style={{ 
              fontSize: '10px', 
              color: '#76695b', 
              background: '#f1ede4', 
              padding: '2px 6px', 
              borderRadius: '4px', 
              fontWeight: 600 
            }}>
              {version}
            </span>
          </div>

          <ul className="nav-menu">
            <li className={activeTab === 'chat' ? 'active' : ''} onClick={() => setActiveTab('chat')}>New Chat</li>
            <li className={activeTab === 'agent' ? 'active' : ''} onClick={() => setActiveTab('agent')}>AI Agent</li>
            <li className={activeTab === 'upgrade' ? 'active' : ''} onClick={() => setActiveTab('upgrade')}>Upgrade Cria</li>
            <li className={activeTab === 'history' ? 'active' : ''} onClick={() => setActiveTab('history')}>Upgrade History</li>
          </ul>
        </div>
        
        <div className="sidebar-bottom">
          <ul className="nav-menu">
            <li className={activeTab === 'settings' ? 'active' : ''} onClick={() => setActiveTab('settings')}>Settings</li>
          </ul>
        </div>
      </div>
    </nav>
  );
};

export default Sidebar;