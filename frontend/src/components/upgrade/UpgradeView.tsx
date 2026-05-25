import React, { useState, useEffect, useRef } from 'react';
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
import { StartUpgradePipeline, ApproveHITL, RejectHITL, GetOllamaModels } from '../../../wailsjs/go/main/App';
import AgentModelConfig from './AgentModelConfig';
import AlertDialog from '../common/AlertDialog';

interface PipelineEvent {
    type: 'status' | 'system_msg' | 'toast' | 'message' | 'hitl' | 'sound';
    icon?: string;
    role?: string;
    content?: string;
    action?: string;
    params?: any;
    data?: Record<string, string>;
}

const STAGES = ['DESIGN', 'IMPLEMENTATION', 'INTEGRATION', 'COMPLETE'];

const UpgradeView: React.FC = () => {
    const [task, setTask] = useState('');
    const [isRunning, setIsRunning] = useState(false);
    const [logs, setLogs] = useState<PipelineEvent[]>([]);
    const [currentStage, setCurrentStage] = useState(0);
    const [awaitingApproval, setAwaitingApproval] = useState(false);

    const [showConfig, setShowConfig] = useState(false);
    const [availableModels, setAvailableModels] = useState<string[]>([]);
    const [agentModels, setAgentModels] = useState<Record<string, string>>({});

    const [feedbackText, setFeedbackText] = useState('');
    const [showRejectForm, setShowRejectForm] = useState(false);
    const [expandedSpecs, setExpandedSpecs] = useState<Record<number, boolean>>({});

    const [alertOpen, setAlertOpen] = useState(false);
    const [alertTitle, setAlertTitle] = useState('');
    const [alertMessage, setAlertMessage] = useState('');

    const logEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const fetchModels = async () => {
            try {
                const models = await GetOllamaModels();
                setAvailableModels(models || []);
            } catch (err) {
                setAvailableModels([]);
            }
        };
        fetchModels();
    }, []);

    useEffect(() => {
        logEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [logs]);

    useEffect(() => {
        EventsOn('pipeline-event', (event: PipelineEvent) => {
            if (event.type === 'status') {
                const content = event.content?.toUpperCase() || '';
                if (content.includes('DESIGN')) setCurrentStage(0);
                else if (content.includes('IMPLEMENTATION')) setCurrentStage(1);
                else if (content.includes('INTEGRATION')) setCurrentStage(2);
            } else if (event.type === 'toast') {
                if (event.content?.includes('Complete')) {
                    setCurrentStage(3);
                    setIsRunning(false);
                } else if (event.icon === '❌') {
                    setIsRunning(false);
                    setAlertTitle('Pipeline Error');
                    setAlertMessage(event.content || 'An error occurred during the pipeline execution.');
                    setAlertOpen(true);
                }
            } else if (event.type === 'hitl') {
                setAwaitingApproval(true);
                setShowRejectForm(false);
                setFeedbackText('');
                setLogs(prev => [...prev, event]);
            } else if (event.type === 'system_msg' || event.type === 'message') {
                setLogs(prev => [...prev, event]);
            }
        });

        return () => EventsOff('pipeline-event');
    }, []);

    const handleModelChange = (agentId: string, model: string) => {
        setAgentModels(prev => ({
            ...prev,
            [agentId]: model === 'default' ? '' : model
        }));
    };

    const handleStart = async () => {
        if (!task.trim()) return;

        if (availableModels.length === 0) {
            setAlertTitle('No Models Found');
            setAlertMessage('Please download at least one model in the Settings menu before running the pipeline.');
            setAlertOpen(true);
            return;
        }

        setIsRunning(true);
        setLogs([]);
        setCurrentStage(0);

        setLogs(prev => [...prev, { type: 'message', role: 'User', content: task, icon: '👤' }]);

        await StartUpgradePipeline(task);
    };

    const handleApprove = async () => {
        setAwaitingApproval(false);
        await ApproveHITL();
    };

    const handleReject = async () => {
        if (!feedbackText.trim()) return;
        setAwaitingApproval(false);

        setLogs(prev => [...prev, {
            type: 'message',
            role: 'User',
            content: `[Feedback for Architect]\n${feedbackText}`,
            icon: '👤'
        }]);

        await RejectHITL(feedbackText);
    };

    return (
        <div className="placeholder-view" style={{ width: '100%', maxWidth: '900px', margin: '0 auto', height: '100vh', display: 'flex', flexDirection: 'column' }}>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', width: '100%' }}>
                <div>
                    <h2 style={{ marginBottom: '4px' }}>Upgrade Cria</h2>
                    <p style={{ color: '#706558', margin: 0 }}>Run the self-evolution pipeline to add new tools, agents, or pipeline stages.</p>
                </div>
                <button
                    onClick={() => setShowConfig(!showConfig)}
                    className="primary-action-btn"
                    style={{ background: showConfig ? '#e6dfd3' : '#f1ede4', color: '#2b2722', border: '1px solid #dcd3c1', boxShadow: 'none' }}
                >
                    {showConfig ? 'Close Roster' : '⚙️ Crew Setup'}
                </button>
            </div>

            <div className="settings-separator" style={{ marginTop: 0 }}></div>

            {showConfig && (
                <AgentModelConfig
                    availableModels={availableModels}
                    agentModels={agentModels}
                    onModelChange={handleModelChange}
                />
            )}

            <div style={{ position: 'relative', display: 'flex', width: '100%', margin: '50px 0 60px 0' }}>
                <div style={{ position: 'absolute', top: '16px', left: '12.5%', right: '12.5%', height: '4px', background: '#e1dacb', zIndex: 1 }} />
                <div style={{ position: 'absolute', top: '16px', left: '12.5%', width: `${(currentStage / (STAGES.length - 1)) * 75}%`, height: '4px', background: '#76695b', zIndex: 1, transition: 'width 0.5s ease' }} />
                {STAGES.map((stage, idx) => (
                    <div key={stage} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', position: 'relative', zIndex: 2 }}>
                        <div style={{ width: '36px', height: '36px', borderRadius: '50%', background: currentStage >= idx ? '#76695b' : '#f1ede4', border: '4px solid #f9f6f0', display: 'flex', justifyContent: 'center', alignItems: 'center', transition: 'background 0.5s ease', position: 'relative' }}>
                            {currentStage === idx && (
                                <div style={{ position: 'absolute', top: '-36px', fontSize: '28px', animation: isRunning ? 'bounce 0.4s infinite alternate' : 'none' }}>
                                    🏃
                                </div>
                            )}
                        </div>
                        <span style={{ fontSize: '13px', marginTop: '12px', fontWeight: currentStage >= idx ? 700 : 500, color: currentStage >= idx ? '#2b2722' : '#a0978c', letterSpacing: '0.5px', textAlign: 'center' }}>
                            {stage}
                        </span>
                    </div>
                ))}
            </div>

            <style>
                {`
          @keyframes bounce {
            from { transform: translateY(0); }
            to { transform: translateY(-6px); }
          }
        `}
            </style>

            <div style={{ flex: 1, width: '100%', overflowY: 'auto', padding: '10px 20px 20px 0', display: 'flex', flexDirection: 'column', gap: '16px' }}>
                {logs.map((log, index) => {
                    const isUser = log.role === 'User';
                    return (
                        <div key={index} style={{ display: 'flex', flexDirection: 'column', alignItems: isUser ? 'flex-end' : 'flex-start', width: '100%' }}>
                            <div style={{
                                background: isUser ? '#76695b' : '#ffffff',
                                color: isUser ? '#ffffff' : '#2b2722',
                                padding: '14px 18px',
                                borderRadius: isUser ? '16px 16px 0 16px' : '16px 16px 16px 0',
                                maxWidth: '80%',
                                border: isUser ? 'none' : '1px solid #e1dacb',
                                boxShadow: '0 2px 8px rgba(43, 39, 34, 0.04)'
                            }}>
                                {!isUser && (
                                    <div style={{ fontWeight: 700, fontSize: '13px', marginBottom: '6px', color: '#76695b', display: 'flex', alignItems: 'center', gap: '6px' }}>
                                        <span>{log.icon || '🤖'}</span>
                                        <span>{log.role || 'System'}</span>
                                        {log.action && <span style={{ opacity: 0.7, fontSize: '11px', background: '#f1ede4', padding: '2px 6px', borderRadius: '4px' }}>{log.action}</span>}
                                    </div>
                                )}

                                <div style={{ fontSize: '14px', lineHeight: '1.6', whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                                    {log.content}
                                </div>

                                {log.type === 'hitl' && log.data?.spec_content && (
                                    <div style={{ marginTop: '12px' }}>
                                        <button
                                            onClick={() => setExpandedSpecs(prev => ({ ...prev, [index]: !prev[index] }))}
                                            style={{
                                                background: 'transparent',
                                                border: '1px solid #dcd3c1',
                                                color: '#76695b',
                                                padding: '6px 12px',
                                                borderRadius: '6px',
                                                cursor: 'pointer',
                                                fontSize: '13px',
                                                fontWeight: 600,
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: '6px'
                                            }}
                                        >
                                            <span style={{ transform: expandedSpecs[index] ? 'rotate(90deg)' : 'none', transition: 'transform 0.2s' }}>▶</span>
                                            {expandedSpecs[index] ? 'Hide Design Spec' : 'View Design Spec'}
                                            <span style={{ opacity: 0.6, fontSize: '11px', fontFamily: 'monospace' }}>{log.data.spec_path}</span>
                                        </button>
                                        {expandedSpecs[index] && (
                                            <pre style={{
                                                marginTop: '10px',
                                                padding: '14px',
                                                background: '#f9f6f0',
                                                border: '1px solid #e1dacb',
                                                borderRadius: '8px',
                                                fontSize: '12px',
                                                lineHeight: '1.6',
                                                color: '#2b2722',
                                                whiteSpace: 'pre-wrap',
                                                wordBreak: 'break-word',
                                                maxHeight: '420px',
                                                overflowY: 'auto',
                                                fontFamily: 'Menlo, Consolas, monospace'
                                            }}>
                                                {log.data.spec_content}
                                            </pre>
                                        )}
                                    </div>
                                )}

                                {log.type === 'hitl' && awaitingApproval && (
                                    <div style={{ marginTop: '16px', borderTop: '1px solid #e1dacb', paddingTop: '16px' }}>
                                        {!showRejectForm ? (
                                            <div style={{ display: 'flex', gap: '8px' }}>
                                                <button onClick={handleApprove} className="primary-action-btn" style={{ flex: 1, background: '#4CAF50' }}>Approve</button>
                                                <button onClick={() => setShowRejectForm(true)} className="primary-action-btn" style={{ flex: 1, background: '#c53030' }}>Reject & Feedback</button>
                                            </div>
                                        ) : (
                                            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                                                <textarea
                                                    value={feedbackText}
                                                    onChange={(e) => setFeedbackText(e.target.value)}
                                                    placeholder="Tell the Architect what needs to be changed..."
                                                    style={{ width: '100%', padding: '10px', borderRadius: '8px', border: '1px solid #e1dacb', background: '#f9f6f0', color: '#2b2722', resize: 'vertical', minHeight: '60px', fontFamily: 'inherit', fontSize: '13px', outline: 'none' }}
                                                />
                                                <div style={{ display: 'flex', gap: '8px' }}>
                                                    <button onClick={handleReject} className="primary-action-btn" style={{ flex: 1, background: '#d97706' }} disabled={!feedbackText.trim()}>Send Feedback</button>
                                                    <button onClick={() => setShowRejectForm(false)} className="primary-action-btn" style={{ background: '#76695b' }}>Cancel</button>
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                )}
                            </div>
                        </div>
                    );
                })}
                <div ref={logEndRef} />
            </div>

            <div className="chat-input-wrapper" style={{ marginTop: '20px', width: '100%' }}>
                <input
                    type="text"
                    value={task}
                    onChange={(e) => setTask(e.target.value)}
                    placeholder="Enter an upgrade task (e.g., 'Add a new code complexity analysis tool')"
                    className="chat-input"
                    disabled={isRunning && !awaitingApproval}
                    onKeyDown={(e) => {
                        if (e.key === 'Enter' && task.trim() && !isRunning) {
                            handleStart();
                        }
                    }}
                />
                <div className="chat-input-bottom">
                    <div className="chat-input-actions-left"></div>
                    <button
                        onClick={handleStart}
                        className="primary-action-btn"
                        disabled={!task.trim() || (isRunning && !awaitingApproval)}
                    >
                        {isRunning ? 'Running...' : 'Execute Upgrade'}
                    </button>
                </div>
            </div>

            <AlertDialog
                isOpen={alertOpen}
                title={alertTitle}
                message={alertMessage}
                onClose={() => setAlertOpen(false)}
            />
        </div>
    );
};

export default UpgradeView;