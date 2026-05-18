import React, { useRef, useEffect, useState } from 'react';
import characterImg from '../assets/images/detective_cria.png';
import { ChatWithModel } from '../../wailsjs/go/main/App';
import { Message } from '../types';

interface ChatViewProps {
  selectedModel: string;
  availableModels: string[];
  setSelectedModel: (model: string) => void;
}

const ChatView: React.FC<ChatViewProps> = ({ selectedModel, availableModels, setSelectedModel }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputText, setInputText] = useState('');
  const [isAiTyping, setIsAiTyping] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, isAiTyping]);

  const handleSendMessage = async () => {
    if (!inputText.trim()) return;
    if (selectedModel === 'No models available' || !selectedModel) {
      alert("Please download a model first from the Settings tab.");
      return;
    }

    const currentPrompt = inputText;
    setMessages(prev => [...prev, { role: 'user', content: currentPrompt }]);
    setInputText('');
    setIsAiTyping(true);

    try {
      const aiResponse = await ChatWithModel(selectedModel, currentPrompt);
      setMessages(prev => [...prev, { role: 'ai', content: aiResponse }]);
    } catch (err) {
      console.error("Chat error:", err);
      setMessages(prev => [...prev, { role: 'ai', content: "Error: Failed to get response from AI." }]);
    } finally {
      setIsAiTyping(false);
    }
  };

  return (
    <div className="chat-container">
      {messages.length === 0 ? (
        <div className="home-view">
          <img src={characterImg} alt="My Character" className="character-image" />
          <h1>Hello! I am Cria.</h1>
          <p>Have a great day! How can I help you today?</p>
        </div>
      ) : (
        <div className="messages-view">
          {messages.map((msg, index) => (
            <div 
              key={index} 
              className={`message-bubble ${msg.role === 'user' ? 'user-message' : 'ai-message'}`}
              style={{
                alignSelf: msg.role === 'user' ? 'flex-end' : 'flex-start',
                background: msg.role === 'user' ? '#e6dfd3' : '#ffffff',
                border: msg.role === 'ai' ? '1px solid #e1dacb' : 'none',
                borderRadius: msg.role === 'user' ? '12px 12px 0 12px' : '12px 12px 12px 0',
                padding: '12px 16px',
                marginBottom: '12px',
                maxWidth: '75%',
                wordBreak: 'break-word',
                color: '#2b2722',
                boxShadow: msg.role === 'ai' ? '0 2px 4px rgba(0,0,0,0.02)' : 'none'
              }}
            >
              {msg.role === 'ai' && <div style={{ fontSize: '12px', fontWeight: 'bold', color: '#706558', marginBottom: '4px' }}>🤖 {selectedModel}</div>}
              {msg.content}
            </div>
          ))}
          {isAiTyping && (
            <div style={{ alignSelf: 'flex-start', padding: '12px', color: '#706558', fontStyle: 'italic' }}>
              Cria is thinking...
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>
      )}

      <div className="chat-input-wrapper">
        <input 
          type="text" 
          value={inputText}
          onChange={(e) => setInputText(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !isAiTyping) {
              handleSendMessage();
            }
          }}
          disabled={isAiTyping}
          placeholder={isAiTyping ? "Please wait..." : "Message Cria..."}
          className="chat-input"
        />
        <div className="chat-input-bottom">
          <div className="chat-input-actions-left"></div>
          <div className="chat-input-actions-right">
            <select 
              value={selectedModel} 
              onChange={(e) => setSelectedModel(e.target.value)}
              className="model-selector-inline"
              disabled={isAiTyping}
            >
              {availableModels.map((model) => (
                <option key={model} value={model}>
                  {model}
                </option>
              ))}
            </select>
            <button 
              onClick={handleSendMessage} 
              className="chat-submit-btn"
              disabled={isAiTyping || !inputText.trim()}
              style={{ opacity: (isAiTyping || !inputText.trim()) ? 0.5 : 1 }}
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <line x1="22" y1="2" x2="11" y2="13"></line>
                <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ChatView;