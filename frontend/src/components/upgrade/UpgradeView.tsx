import React, { useState, useEffect, useRef } from 'react';
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
import { StartUpgradePipeline, ApproveHITL, RejectHITL, GetOllamaModels, TranslateText, LogClientEvent, GetAgentModels, SaveAgentModels } from '../../../wailsjs/go/main/App';
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
    const [agentModelsLoaded, setAgentModelsLoaded] = useState(false);

    const [feedbackText, setFeedbackText] = useState('');
    const [showRejectForm, setShowRejectForm] = useState(false);
    const [expandedSpecs, setExpandedSpecs] = useState<Record<number, boolean>>({});
    const [specTranslations, setSpecTranslations] = useState<Record<number, string>>({});
    const [translatingSpecs, setTranslatingSpecs] = useState<Record<number, boolean>>({});
    const [showTranslation, setShowTranslation] = useState<Record<number, boolean>>({});

    const [alertOpen, setAlertOpen] = useState(false);
    const [alertTitle, setAlertTitle] = useState('');
    const [alertMessage, setAlertMessage] = useState('');

    const logEndRef = useRef<HTMLDivElement>(null);

    const logEvent = (msg: string, level: string = 'user') => {
        LogClientEvent(level, `[UpgradeView] ${msg}`).catch(() => { });
    };

    useEffect(() => {
        const fetchModels = async () => {
            try {
                const models = await GetOllamaModels();
                setAvailableModels(models || []);
                logEvent(`models loaded count=${(models || []).length}`, 'debug');
            } catch (err) {
                setAvailableModels([]);
                logEvent(`models load failed: ${String(err)}`, 'error');
            }
        };
        const fetchAgentModels = async () => {
            try {
                const saved = await GetAgentModels();
                setAgentModels(saved || {});
                logEvent(`agent models loaded entries=${Object.keys(saved || {}).length}`, 'debug');
            } catch (err) {
                setAgentModels({});
                logEvent(`agent models load failed: ${String(err)}`, 'error');
            } finally {
                setAgentModelsLoaded(true);
            }
        };
        fetchModels();
        fetchAgentModels();
    }, []);

    useEffect(() => {
        if (!agentModelsLoaded) return;
        const handle = setTimeout(() => {
            logEvent(`SaveAgentModels triggered entries=${Object.keys(agentModels).length}`, 'state');
            SaveAgentModels(agentModels).catch(err => {
                logEvent(`SaveAgentModels failed: ${String(err)}`, 'error');
                console.error('SaveAgentModels failed:', err);
            });
        }, 400);
        return () => clearTimeout(handle);
    }, [agentModels, agentModelsLoaded]);

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
        logEvent(`model change agentId=${agentId} value=${model}`);
        setAgentModels(prev => ({
            ...prev,
            [agentId]: model === 'default' ? '' : model
        }));
    };

    const handleStart = async () => {
        if (!task.trim()) return;

        if (availableModels.length === 0) {
            logEvent(`start blocked: no models available task=${JSON.stringify(task)}`);
            setAlertTitle('No Models Found');
            setAlertMessage('Please download at least one model in the Settings menu before running the pipeline.');
            setAlertOpen(true);
            return;
        }

        logEvent(`start pipeline task=${JSON.stringify(task)}`);
        setIsRunning(true);
        setLogs([]);
        setCurrentStage(0);

        setLogs(prev => [...prev, { type: 'message', role: 'User', content: task, icon: '👤' }]);

        await StartUpgradePipeline(task);
    };

    const handleApprove = async () => {
        logEvent('HITL approve');
        setAwaitingApproval(false);
        await ApproveHITL();
    };

    const handleReject = async () => {
        if (!feedbackText.trim()) return;
        logEvent(`HITL reject feedbackLen=${feedbackText.length}`);
        setAwaitingApproval(false);

        setLogs(prev => [...prev, {
            type: 'message',
            role: 'User',
            content: `[Feedback for Architect]\n${feedbackText}`,
            icon: '👤'
        }]);

        await RejectHITL(feedbackText);
    };

    const handleTranslateSpec = async (index: number, specContent: string) => {
        if (specTranslations[index]) {
            const next = !showTranslation[index];
            logEvent(`spec translate toggle index=${index} showTranslation=${next}`);
            setShowTranslation(prev => ({ ...prev, [index]: next }));
            return;
        }
        const translatorModel = agentModels['translator'] || agentModels['global'] || '';
        logEvent(`spec translate request index=${index} model=${translatorModel || '(none)'} contentLen=${specContent.length}`);
        setTranslatingSpecs(prev => ({ ...prev, [index]: true }));
        try {
            const translated = await TranslateText(translatorModel, '', specContent);
            logEvent(`spec translate done index=${index} resultLen=${translated.length}`, 'state');
            setSpecTranslations(prev => ({ ...prev, [index]: translated }));
            setShowTranslation(prev => ({ ...prev, [index]: true }));
        } catch (err) {
            logEvent(`spec translate failed index=${index}: ${String(err)}`, 'error');
            console.error('TranslateText failed:', err);
            setAlertTitle('Translation Failed');
            setAlertMessage(String(err));
            setAlertOpen(true);
        } finally {
            setTranslatingSpecs(prev => ({ ...prev, [index]: false }));
        }
    };

    return (
        <div className="placeholder-view" style={{ width: '100%', maxWidth: '900px', margin: '0 auto', height: '100vh', display: 'flex', flexDirection: 'column' }}>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', width: '100%' }}>
                <div>
                    <h2 style={{ marginBottom: '4px' }}>Upgrade Cria</h2>
                    <p style={{ color: '#706558', margin: 0 }}>Run the self-evolution pipeline to add new tools, agents, or pipeline stages.</p>
                </div>
                <button
                    onClick={() => {
                        logEvent(`crew setup toggle to=${!showConfig}`);
                        setShowConfig(!showConfig);
                    }}
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
                                        <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap', alignItems: 'center' }}>
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
                                                <button
                                                    onClick={() => handleTranslateSpec(index, log.data!.spec_content)}
                                                    disabled={translatingSpecs[index]}
                                                    style={{
                                                        background: showTranslation[index] ? '#76695b' : 'transparent',
                                                        border: '1px solid #76695b',
                                                        color: showTranslation[index] ? '#ffffff' : '#76695b',
                                                        padding: '6px 12px',
                                                        borderRadius: '6px',
                                                        cursor: translatingSpecs[index] ? 'not-allowed' : 'pointer',
                                                        fontSize: '13px',
                                                        fontWeight: 600,
                                                        display: 'flex',
                                                        alignItems: 'center',
                                                        gap: '6px',
                                                        opacity: translatingSpecs[index] ? 0.7 : 1
                                                    }}
                                                >
                                                    {translatingSpecs[index]
                                                        ? '⏳ Translating...'
                                                        : specTranslations[index]
                                                            ? (showTranslation[index] ? 'Show Original' : 'Show Translation')
                                                            : '🌐 Translate'}
                                                </button>
                                            )}
                                        </div>
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
                                                {showTranslation[index] && specTranslations[index]
                                                    ? specTranslations[index]
                                                    : log.data.spec_content}
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