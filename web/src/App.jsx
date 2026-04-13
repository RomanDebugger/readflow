import React, { useState, useRef, useEffect } from 'react';
import axios from 'axios';
import { Cpu, MessageSquare, ChevronRight, Upload, Zap, FileText, X, Send, Sliders, Hash, BookOpen, AlertCircle, LayoutPanelLeft } from 'lucide-react';
import './App.css';
import ReactMarkdown from 'react-markdown'
export default function App() {
  const [chunks, setChunks] = useState([]);
  const [chat, setChat] = useState([]);
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [chatLoading, setChatLoading] = useState(false);
  const [threshold, setThreshold] = useState(0.4);
  const [fileName, setFileName] = useState('');
  const [pdfUrl, setPdfUrl] = useState(null);
  const [pdfPage, setPdfPage] = useState(1);
  const [view, setView] = useState('pdf');
  const [error, setError] = useState('');

  const chatEndRef = useRef(null);
  const inputRef = useRef(null);
  const fileRef = useRef(null);

  useEffect(() => { chatEndRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [chat, chatLoading]);

  const processFile = async (file) => {
    if (!file) return;
    setFileName(file.name);
    setChunks([]);
    setChat([]);
    setError('');

    const localUrl = URL.createObjectURL(file);
    setPdfUrl(localUrl);
    setPdfPage(1);
    setView('pdf');

    const fd = new FormData();
    fd.append('file', file);
    setLoading(true);
    try {
      const res = await axios.post('/api/process', fd);
      setChunks(res.data);
    } catch (e) {
      setError('Backend unreachable — ensure localhost:8000 is running.');
    }
    setLoading(false);
  };

  const sendChat = async () => {
    const q = query.trim();
    if (!q || chatLoading) return;
    const newMsgs = [...chat, { role: 'user', content: q }];
    setChat(newMsgs);
    setQuery('');
    setChatLoading(true);
    try {
      const context = chunks.filter(c => c.quality >= threshold);
      const res = await axios.post('/api/chat', { query: q, context });
      setChat([...newMsgs, { role: 'assistant', content: res.data.response }]);
    } catch (e) {
      setChat([...newMsgs, { role: 'assistant', content: 'Could not reach inference engine.' }]);
    }
    setChatLoading(false);
    inputRef.current?.focus();
  };

  const handleNavClick = (page) => {
    setView('pdf');
    setPdfPage(page);
  };

  const visibleChunks = chunks.filter(c => c.quality >= threshold);
  const titles = chunks.filter(c => c.type === 'title');

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="logo-row">
            <div className="logo-icon"><Zap size={14} color="#000" strokeWidth={2.5} /></div>
            <span className="logo-text">READFLOW</span>
          </div>
          <span className="version-text">Go-Refinery v1.2</span>
        </div>

        <div className="nav-list">
          <div className="section-label">{titles.length ? 'Document Anchors' : 'No document loaded'}</div>
          {titles.map((t, i) => (
            <button key={i} className="nav-btn" onClick={() => handleNavClick(t.page)}>
              <ChevronRight size={12} className="nav-icon" />
              <span className="truncate">{t.text}</span>
            </button>
          ))}
          {!titles.length && !loading && (
            <div className="skeleton-list">
              {[BookOpen, FileText, LayoutPanelLeft].map((Icon, i) => (
                <div key={i} className="skeleton-item"><Icon size={11} /><div className="skeleton-line" style={{ width: `${60 + i * 15}%` }} /></div>
              ))}
            </div>
          )}
        </div>
      </aside>

      <main className="inspector">
        <header className="inspector-toolbar">
          <div className="toolbar-left">
            <button className={`tab-btn ${view === 'pdf' ? 'active' : ''}`} onClick={() => setView('pdf')}>
              <FileText size={14} /> Source PDF
            </button>
            <button className={`tab-btn ${view === 'data' ? 'active' : ''}`} onClick={() => setView('data')} disabled={!chunks.length}>
              <LayoutPanelLeft size={14} /> Extracted Data
            </button>
          </div>

          <div className="toolbar-right">
            <button className="upload-btn" onClick={() => fileRef.current?.click()}><Upload size={13} /> Upload File</button>
            <input ref={fileRef} type="file" accept=".pdf" onChange={(e) => processFile(e.target.files[0])} hidden />

            <div className="filter-ctrl">
              <Sliders size={13} className="text-muted" />
              <div className="filter-slider-group">
                <span className="filter-label">Signal Filter ≥ <b>{threshold.toFixed(2)}</b></span>
                <input type="range" min="0" max="1" step="0.05" value={threshold} onChange={e => setThreshold(parseFloat(e.target.value))} />
              </div>
            </div>
          </div>
        </header>

        <div className="inspector-content">
          {!loading && !pdfUrl && (
            <div className="drop-zone" onClick={() => fileRef.current?.click()}>
              <div className="drop-icon"><Upload size={22} /></div>
              <h3>Drop your document here</h3>
              <p>Process via Go-Refinery & Chat with the Document</p>
              {error && <div className="error-banner"><AlertCircle size={13} /> {error}</div>}
            </div>
          )}

          {loading && (
            <div className="loading-state">
              <div className="spinner" />
              <span className="pulse-text">Refining spatial document data...</span>
            </div>
          )}

          {!loading && pdfUrl && view === 'pdf' && (
            <div className="pdf-container">
              <iframe
                key={pdfPage}
                src={`${pdfUrl}#page=${pdfPage}&toolbar=0&navpanes=0`}
                title="PDF Viewer"
                className="pdf-frame"
              />
            </div>
          )}

          {!loading && chunks.length > 0 && view === 'data' && (
            <div className="chunk-grid">
              {visibleChunks.map((c) => (
                <div key={c.chunk_id} className={`chunk-card ${c.type}`}>
                  <div className="chunk-header">
                    <span className="chunk-badge">{c.type.toUpperCase()} · PG {c.page}</span>
                    <span className="score-text">QUAL: {c.quality.toFixed(2)}</span>
                  </div>
                  <p className="chunk-body">{c.text}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </main>

      <aside className="console">
        <header className="console-header">
          <div className="console-title"><div className={`status-dot ${chunks.length ? 'active' : ''}`} /> GEMINI CONSOLE</div>
          {chat.length > 0 && <button className="clear-btn" onClick={() => setChat([])}><X size={11} /> CLEAR</button>}
        </header>

        <div className="chat-area">
          {chat.length === 0 && (
            <div className="chat-empty">
              <MessageSquare size={28} className="text-muted" />
              <p>{chunks.length ? 'Query the extracted structural data...' : 'Upload a document to begin'}</p>
            </div>
          )}

          {chat.map((m, i) => (
            <div key={i} className={`bubble ${m.role}`}>
              {m.role === 'assistant' && <div className="bubble-label">GEMINI</div>}
              <ReactMarkdown>{m.content}</ReactMarkdown>
            </div>
          ))}

          {chatLoading && <div className="bubble assistant loading"><div className="dot-pulse" /><div className="dot-pulse" style={{ animationDelay: '0.2s' }} /><div className="dot-pulse" style={{ animationDelay: '0.4s' }} /></div>}
          <div ref={chatEndRef} />
        </div>

        <footer className="chat-input-area">
          <div className={`input-wrapper ${query.trim() && chunks.length ? 'ready' : ''}`}>
            <textarea ref={inputRef} rows={1} value={query} onChange={e => { setQuery(e.target.value); e.target.style.height = 'auto'; e.target.style.height = Math.min(e.target.scrollHeight, 120) + 'px'; }} onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendChat(); } }} placeholder="Ask about the document…" disabled={!chunks.length || chatLoading} />
            <button className="send-btn" onClick={sendChat} disabled={!query.trim() || chatLoading || !chunks.length}>
              <Send size={13} />
            </button>
          </div>
        </footer>
      </aside>
    </div>
  );
}