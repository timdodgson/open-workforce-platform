'use client';
import { useState, useRef, useEffect } from 'react';

interface Message {
  role: 'user' | 'assistant';
  content: string;
}

export default function ChatWidget() {
  const [messages, setMessages] = useState<Message[]>([
    { role: 'assistant', content: 'I\'m the PFRS Optimisation Assistant. I can help you design experiments, generate CLI commands, and interpret results.\n\nTry: "I want to test whether adding more SA workers improves results" or "Generate a sweep over beam widths 3, 5, 8, 12"' },
  ]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!input.trim() || loading) return;

    const userMessage: Message = { role: 'user', content: input.trim() };
    const updatedMessages = [...messages, userMessage];
    setMessages(updatedMessages);
    setInput('');
    setLoading(true);
    setError(null);

    try {
      // Send only user/assistant exchanges, excluding the initial greeting.
      const apiMessages = updatedMessages.filter((m, i) => !(i === 0 && m.role === 'assistant'));

      const res = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ messages: apiMessages }),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({ error: `HTTP ${res.status}` }));
        setError(body.error || `Request failed (${res.status})`);
        setLoading(false);
        return;
      }

      const data = await res.json();
      setMessages([...updatedMessages, { role: 'assistant', content: data.response }]);
    } catch (err) {
      setError('Network error — check your connection.');
    }

    setLoading(false);
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
  }

  function renderContent(content: string) {
    // Split on code blocks and render them with copy buttons.
    const parts = content.split(/(```[\s\S]*?```)/g);
    return parts.map((part, i) => {
      if (part.startsWith('```')) {
        const lines = part.split('\n');
        const lang = lines[0].replace('```', '').trim();
        const code = lines.slice(1, -1).join('\n');
        return (
          <div key={i} className="relative my-2">
            <div className="flex items-center justify-between bg-gray-900 rounded-t px-3 py-1 border border-gray-700 border-b-0">
              <span className="text-[10px] text-gray-500">{lang || 'shell'}</span>
              <button
                onClick={() => copyToClipboard(code)}
                className="text-[10px] text-gray-400 hover:text-white transition-colors"
              >
                Copy
              </button>
            </div>
            <pre className="bg-gray-900 p-3 rounded-b border border-gray-700 overflow-x-auto text-xs text-gray-300 font-mono">
              {code}
            </pre>
          </div>
        );
      }
      // Render paragraphs with basic markdown.
      return (
        <div key={i} className="whitespace-pre-wrap text-sm text-gray-300 leading-relaxed">
          {part.split('\n').map((line, j) => {
            if (line.startsWith('# ')) return <h3 key={j} className="text-white font-bold mt-3 mb-1">{line.slice(2)}</h3>;
            if (line.startsWith('## ')) return <h4 key={j} className="text-blue-400 font-semibold mt-2 mb-1">{line.slice(3)}</h4>;
            if (line.startsWith('**') && line.endsWith('**')) return <p key={j} className="font-semibold text-white">{line.slice(2, -2)}</p>;
            if (line.startsWith('- ')) return <p key={j} className="pl-3">• {line.slice(2)}</p>;
            if (line.trim() === '') return <br key={j} />;
            return <span key={j}>{line}{'\n'}</span>;
          })}
        </div>
      );
    });
  }

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)]">
      {/* Header */}
      <div className="border-b border-gray-700 p-4">
        <h2 className="text-sm font-bold text-blue-400">🧪 Optimisation Assistant</h2>
        <p className="text-[10px] text-gray-500 mt-1">
          Experiment planner • CLI generator • Results interpreter
        </p>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((msg, i) => (
          <div key={i} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            <div className={`max-w-[85%] rounded-lg p-3 ${
              msg.role === 'user'
                ? 'bg-blue-600/20 border border-blue-600/30'
                : 'bg-gray-800 border border-gray-700'
            }`}>
              {msg.role === 'assistant' && (
                <span className="text-[9px] text-gray-500 block mb-1">PFRS Assistant</span>
              )}
              {renderContent(msg.content)}
            </div>
          </div>
        ))}
        {loading && (
          <div className="flex justify-start">
            <div className="bg-gray-800 border border-gray-700 rounded-lg p-3">
              <span className="text-[9px] text-gray-500 block mb-1">PFRS Assistant</span>
              <div className="flex gap-1">
                <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
              </div>
            </div>
          </div>
        )}
        {error && (
          <div className="flex justify-center">
            <div className="bg-red-900/20 border border-red-600/30 rounded-lg p-3 text-xs text-red-400">
              {error}
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <form onSubmit={handleSubmit} className="border-t border-gray-700 p-4">
        <div className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={e => setInput(e.target.value)}
            placeholder="Describe your experiment idea..."
            className="flex-1 bg-gray-800 border border-gray-700 rounded-lg px-4 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
            disabled={loading}
          />
          <button
            type="submit"
            disabled={loading || !input.trim()}
            className="bg-blue-600 hover:bg-blue-500 disabled:bg-gray-700 disabled:text-gray-500 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            Send
          </button>
        </div>
      </form>
    </div>
  );
}
