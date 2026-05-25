import React, { useState, useEffect } from 'react';
import { GetTranslationLanguage, SaveTranslationLanguage } from '../../../wailsjs/go/main/App';
import './Settings.css';

const LANGUAGES = [
    'English',
    'Korean',
    'Japanese',
    'Chinese (Simplified)',
    'Chinese (Traditional)',
    'Spanish',
    'French',
    'German',
    'Italian',
    'Portuguese',
    'Russian',
    'Arabic',
    'Hindi',
    'Vietnamese',
    'Thai',
    'Indonesian'
];

const TranslationLanguageSection: React.FC = () => {
    const [language, setLanguage] = useState('');
    const [loaded, setLoaded] = useState(false);

    useEffect(() => {
        const fetchLang = async () => {
            try {
                const saved = await GetTranslationLanguage();
                setLanguage(saved || 'English');
            } catch (err) {
                console.error(err);
                setLanguage('English');
            } finally {
                setLoaded(true);
            }
        };
        fetchLang();
    }, []);

    useEffect(() => {
        if (!loaded) return;
        const handle = setTimeout(() => {
            SaveTranslationLanguage(language).catch(err => console.error('SaveTranslationLanguage failed:', err));
        }, 300);
        return () => clearTimeout(handle);
    }, [language, loaded]);

    return (
        <div style={{ width: '100%', marginBottom: '40px' }}>
            <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600, textAlign: 'left' }}>Translation Language</label>
            <p style={{ margin: '0 0 12px 0', color: '#a0978c', fontSize: '13px', textAlign: 'left' }}>
                Target language used by the Translator agent and the "Translate" button on design specs.
            </p>
            <select
                value={language}
                onChange={(e) => setLanguage(e.target.value)}
                className="settings-input-field"
                style={{ width: '100%', cursor: 'pointer' }}
                disabled={!loaded}
            >
                {LANGUAGES.map(lang => (
                    <option key={lang} value={lang}>{lang}</option>
                ))}
            </select>
        </div>
    );
};

export default TranslationLanguageSection;
