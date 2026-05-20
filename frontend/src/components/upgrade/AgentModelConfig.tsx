import React from 'react';
import './AgentModelConfig.css';

interface AgentModelConfigProps {
  availableModels: string[];
  agentModels: Record<string, string>;
  onModelChange: (agentId: string, model: string) => void;
}

const AGENTS = [
  { id: 'architect', name: 'Global Architect', role: 'System Design & Planning', color: '#4a5568' },
  { id: 'writer', name: 'Module Writer', role: 'Core Implementation', color: '#2b6cb0' },
  { id: 'tester', name: 'Reviewer & Tester', role: 'QA & Syntax Check', color: '#38a169' },
  { id: 'watchdog', name: 'Watchdog', role: 'Routing & Recovery', color: '#c53030' }
];

const AgentModelConfig: React.FC<AgentModelConfigProps> = ({ availableModels, agentModels, onModelChange }) => {
  return (
    <div className="agent-config-container">
      <div className="agent-config-header">
        <h3>Agent Roster</h3>
        <p>Assign specific models to your AI crew.</p>
      </div>
      
      <div className="agent-character-grid">
        {AGENTS.map((agent) => (
          <div key={agent.id} className="agent-character-card">
            <div className="character-avatar-box" style={{ backgroundColor: agent.color }}>
              <div className="character-avatar-placeholder">
                IMG
              </div>
            </div>
            
            <div className="agent-card-body">
              <div className="agent-card-title">{agent.name}</div>
              <div className="agent-card-role">{agent.role}</div>
              
              <div className="agent-model-select-wrapper">
                <select 
                  className="agent-model-select"
                  value={agentModels[agent.id] || 'default'}
                  onChange={(e) => onModelChange(agent.id, e.target.value)}
                >
                  <option value="default">Global Default</option>
                  {availableModels.map(m => (
                    <option key={m} value={m}>{m}</option>
                  ))}
                </select>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default AgentModelConfig;