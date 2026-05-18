import React from 'react';

interface SidebarProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
}

const Sidebar: React.FC<SidebarProps> = ({ activeTab, setActiveTab }) => {
  return (
    <nav className="sidebar">
      <div className="sidebar-top">
        <div className="sidebar-logo">
          <h2>Cria AI</h2>
        </div>
        <ul className="nav-menu">
          <li 
            className={activeTab === 'chat' ? 'active' : ''} 
            onClick={() => setActiveTab('chat')}
          >
            New Chat
          </li>
          <li 
            className={activeTab === 'agent' ? 'active' : ''} 
            onClick={() => setActiveTab('agent')}
          >
            AI Agent
          </li>
          <li 
            className={activeTab === 'upgrade' ? 'active' : ''} 
            onClick={() => setActiveTab('upgrade')}
          >
            Upgrade Cria
          </li>
        </ul>
      </div>
      <div className="sidebar-bottom">
        <ul className="nav-menu">
          <li 
            className={activeTab === 'settings' ? 'active' : ''} 
            onClick={() => setActiveTab('settings')}
          >
            Settings
          </li>
        </ul>
      </div>
    </nav>
  );
};

export default Sidebar;