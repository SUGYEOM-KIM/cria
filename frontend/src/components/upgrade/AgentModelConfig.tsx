import React from 'react';
import './AgentModelConfig.css';

interface AgentModelConfigProps {
  availableModels: string[];
  agentModels: Record<string, string>;
  onModelChange: (agentId: string, model: string) => void;
}

const AGENTS = [
  { id: 'architect', name: 'Global Architect', role: 'Architecture Design', color: '#76695b' },
  { id: 'planner', name: 'Unit Planner', role: 'Task Planning', color: '#6e7d88' },
  { id: 'developer', name: 'Developer', role: 'Code Implementation', color: '#5c6b73' },
  { id: 'integrator', name: 'Integrator', role: 'Module Merge', color: '#83786c' },
  { id: 'critic', name: 'Critic & Verifier', role: 'All Feedback & Verification', color: '#717c66' },
  { id: 'watchdog', name: 'Watchdog', role: 'Failure Recovery', color: '#a67c6d' },
  { id: 'translator', name: 'Translator', role: 'Message Localization', color: '#8a6f9c' }
];

const AgentModelConfig: React.FC<AgentModelConfigProps> = ({ availableModels, agentModels, onModelChange }) => {
  return (
    <div className="agent-config-container">
      <div className="agent-config-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '16px', marginBottom: '24px' }}>
        <div>
          <h3 style={{ margin: 0, color: '#2b2722', fontSize: '18px' }}>Agent Roster</h3>
          <p style={{ margin: '4px 0 0 0', color: '#706558', fontSize: '13px' }}>Assign specific models to your AI crew.</p>
        </div>
        
        <div style={{ background: '#f9f6f0', padding: '12px 16px', borderRadius: '8px', border: '1px solid #e1dacb', minWidth: '240px' }}>
          <div style={{ color: '#706558', fontSize: '12px', fontWeight: 700, marginBottom: '6px', letterSpacing: '0.5px' }}>GLOBAL DEFAULT MODEL</div>
          <select 
            className="agent-model-select"
            value={agentModels['global'] || (availableModels.length > 0 ? availableModels[0] : 'default')}
            onChange={(e) => onModelChange('global', e.target.value)}
            style={{ background: '#ffffff', borderColor: '#dcd3c1', color: '#2b2722' }}
          >
            {availableModels.length === 0 && <option value="default">No models available</option>}
            {availableModels.map(m => (
              <option key={m} value={m}>{m}</option>
            ))}
          </select>
        </div>
      </div>
      
      <div className="agent-character-grid">
        {AGENTS.map((agent) => (
          <div key={agent.id} className="agent-character-card">
            <div className="character-avatar-box" style={{ backgroundColor: agent.color }}>
              <div className="character-avatar-placeholder">
                {agent.name.charAt(0)}
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