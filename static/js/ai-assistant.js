/* Victor Akor Portfolio v3 — AI Sales Assistant */
'use strict';

const PORTFOLIO_CONTEXT = {
  name: 'Victor Akor',
  title: 'Senior Software Engineer & AI Specialist',
  skills: ['Go', 'Python', 'JavaScript', 'PostgreSQL', 'Docker', 'OpenCV', 'TensorFlow', 'REST APIs', 'Linux'],
  services: ['AI Engineering', 'Backend Engineering', 'Full Stack Development', 'Technical Consulting'],
  projects: [
    { name: 'Mall Surveillance System', tech: ['Python', 'OpenCV', 'TensorFlow'], type: 'AI/Computer Vision' },
    { name: 'AI Face Recognition System', tech: ['Python', 'Deep Learning'], type: 'AI' },
    { name: 'Eye Disease Detection', tech: ['Python', 'CNN', 'TensorFlow'], type: 'AI/Medical' },
    { name: 'Hackerthon Platform', tech: ['Go', 'PostgreSQL', 'JavaScript'], type: 'Full Stack' },
    { name: 'Text Analyzer', tech: ['Python', 'NLP'], type: 'AI/NLP' },
    { name: 'CLI Calculator', tech: ['Go'], type: 'CLI Tool' },
    { name: 'Gwinks Hub', tech: ['Go', 'PostgreSQL', 'JavaScript'], type: 'Full Stack' },
  ],
  experience: '5+ years',
  availability: 'Available for new projects',
};

const RESPONSES_FALLBACK = `Sorry, I'm having trouble connecting right now. You can still reach Victor directly through the contact form.`;

function createAssistant() {
  const widget = document.createElement('div');
  widget.id = 'aiAssistant';
  widget.innerHTML = `
    <style>
      #aiAssistant { position: fixed; bottom: 24px; right: 24px; z-index: 9999; font-family: var(--font-sans, 'Inter', sans-serif); }
      #aiToggle { width: 56px; height: 56px; border-radius: 50%; background: linear-gradient(135deg, #2563eb, #06b6d4); border: none; cursor: pointer; display: flex; align-items: center; justify-content: center; font-size: 1.4rem; box-shadow: 0 4px 24px rgba(37,99,235,0.4); transition: transform 0.2s, box-shadow 0.2s; position: relative; }
      #aiToggle:hover { transform: scale(1.08); box-shadow: 0 8px 32px rgba(37,99,235,0.5); }
      #aiToggle .pulse-ring { position: absolute; inset: -4px; border-radius: 50%; border: 2px solid rgba(37,99,235,0.4); animation: pulse 2s infinite; }
      #aiChat { position: absolute; bottom: 68px; right: 0; width: 360px; background: #0d1628; border: 1px solid rgba(255,255,255,0.08); border-radius: 20px; box-shadow: 0 20px 60px rgba(0,0,0,0.6); display: none; flex-direction: column; overflow: hidden; max-height: 520px; }
      #aiChat.open { display: flex; animation: fadeInUp 0.3s ease; }
      .ai-header { padding: 16px 20px; background: linear-gradient(135deg, rgba(37,99,235,0.15), rgba(6,182,212,0.08)); border-bottom: 1px solid rgba(255,255,255,0.06); display: flex; align-items: center; gap: 12px; }
      .ai-avatar { width: 36px; height: 36px; border-radius: 50%; background: linear-gradient(135deg, #2563eb, #06b6d4); display: flex; align-items: center; justify-content: center; font-size: 1rem; }
      .ai-info h4 { font-size: 0.875rem; font-weight: 700; color: #e2e8f0; margin: 0; }
      .ai-info p { font-size: 0.72rem; color: #10b981; margin: 0; }
      .ai-close { margin-left: auto; background: none; border: none; color: #64748b; cursor: pointer; font-size: 1.1rem; padding: 4px; }
      .ai-messages { flex: 1; overflow-y: auto; padding: 16px; display: flex; flex-direction: column; gap: 12px; }
      .ai-messages::-webkit-scrollbar { width: 4px; }
      .ai-messages::-webkit-scrollbar-track { background: transparent; }
      .ai-messages::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.1); border-radius: 2px; }
      .msg { max-width: 85%; padding: 10px 14px; border-radius: 14px; font-size: 0.85rem; line-height: 1.5; }
      .msg-bot { background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.06); color: #e2e8f0; align-self: flex-start; border-radius: 4px 14px 14px 14px; }
      .msg-user { background: #2563eb; color: #fff; align-self: flex-end; border-radius: 14px 4px 14px 14px; }
      .msg-typing { display: flex; gap: 4px; align-items: center; padding: 12px 14px; }
      .msg-typing span { width: 6px; height: 6px; background: #64748b; border-radius: 50%; animation: bounce 1.2s infinite; }
      .msg-typing span:nth-child(2) { animation-delay: 0.2s; }
      .msg-typing span:nth-child(3) { animation-delay: 0.4s; }
      .ai-suggestions { padding: 8px 16px; display: flex; flex-wrap: wrap; gap: 6px; border-top: 1px solid rgba(255,255,255,0.04); }
      .suggestion { padding: 5px 12px; background: rgba(37,99,235,0.1); border: 1px solid rgba(37,99,235,0.25); border-radius: 100px; font-size: 0.72rem; color: #93c5fd; cursor: pointer; transition: all 0.2s; white-space: nowrap; }
      .suggestion:hover { background: rgba(37,99,235,0.2); }
      .ai-input-row { padding: 12px 16px; border-top: 1px solid rgba(255,255,255,0.06); display: flex; gap: 8px; }
      #aiInput { flex: 1; padding: 10px 14px; background: rgba(255,255,255,0.04); border: 1px solid rgba(255,255,255,0.08); border-radius: 10px; color: #e2e8f0; font-size: 0.85rem; outline: none; font-family: inherit; }
      #aiInput:focus { border-color: #2563eb; }
      #aiSend { padding: 10px 14px; background: #2563eb; border: none; border-radius: 10px; color: #fff; cursor: pointer; font-size: 0.85rem; font-weight: 600; transition: background 0.2s; }
      #aiSend:hover { background: #1d4ed8; }
      .ai-lead-prompt { margin: 8px 0; padding: 12px; background: rgba(37,99,235,0.08); border: 1px solid rgba(37,99,235,0.2); border-radius: 10px; font-size: 0.8rem; color: #93c5fd; }
      .ai-lead-prompt a { color: #60a5fa; font-weight: 600; }
      @keyframes fadeInUp { from{opacity:0;transform:translateY(12px)} to{opacity:1;transform:translateY(0)} }
      @keyframes bounce { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-4px)} }
      @keyframes pulse { 0%,100%{opacity:1;transform:scale(1)} 50%{opacity:0.5;transform:scale(1.1)} }
    </style>
    <button id="aiToggle" aria-label="Open AI Assistant">
      <div class="pulse-ring"></div>
      🤖
    </button>
    <div id="aiChat" role="dialog" aria-label="AI Sales Assistant">
      <div class="ai-header">
        <div class="ai-avatar">🤖</div>
        <div class="ai-info">
          <h4>Victor's AI Assistant</h4>
          <p>● Online — Ask me anything</p>
        </div>
        <button class="ai-close" id="aiClose" aria-label="Close">✕</button>
      </div>
      <div class="ai-messages" id="aiMessages"></div>
      <div class="ai-suggestions" id="aiSuggestions">
        <button class="suggestion">Can Victor build AI systems?</button>
        <button class="suggestion">What's his Go experience?</button>
        <button class="suggestion">How much does it cost?</button>
        <button class="suggestion">Is he available?</button>
      </div>
      <div class="ai-input-row">
        <input id="aiInput" type="text" placeholder="Ask about Victor's work..." autocomplete="off" />
        <button id="aiSend">Send</button>
      </div>
    </div>
  `;
  document.body.appendChild(widget);

  const toggle = document.getElementById('aiToggle');
  const chat = document.getElementById('aiChat');
  const closeBtn = document.getElementById('aiClose');
  const input = document.getElementById('aiInput');
  const send = document.getElementById('aiSend');
  const messages = document.getElementById('aiMessages');
  const suggestions = document.getElementById('aiSuggestions');

  let msgCount = 0;
  let chatHistory = []; // { role: 'user'|'assistant', content: string }
  let sending = false;

  function addMessage(text, type) {
    const div = document.createElement('div');
    div.className = `msg msg-${type}`;
    div.textContent = text;
    messages.appendChild(div);
    messages.scrollTop = messages.scrollHeight;
    msgCount++;

    // After 3 user messages, show lead capture prompt
    if (type === 'user' && msgCount >= 6) {
      const prompt = document.createElement('div');
      prompt.className = 'ai-lead-prompt';
      prompt.innerHTML = `💡 Interested in working with Victor? <a href="/contact">Start a project inquiry →</a>`;
      messages.appendChild(prompt);
      messages.scrollTop = messages.scrollHeight;
    }
  }

  function showTyping() {
    const div = document.createElement('div');
    div.className = 'msg msg-bot msg-typing';
    div.id = 'typingIndicator';
    div.innerHTML = '<span></span><span></span><span></span>';
    messages.appendChild(div);
    messages.scrollTop = messages.scrollHeight;
  }

  function hideTyping() {
    document.getElementById('typingIndicator')?.remove();
  }

  async function fetchAssistantReply(text) {
    const res = await fetch('/api/assistant/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message: text, history: chatHistory }),
    });
    if (!res.ok) throw new Error(`Assistant request failed: ${res.status}`);
    const data = await res.json();
    return data.reply;
  }

  function handleSend(text) {
    text = text.trim();
    if (!text || sending) return;
    sending = true;
    send.disabled = true;
    addMessage(text, 'user');
    chatHistory.push({ role: 'user', content: text });
    input.value = '';
    showTyping();

    fetchAssistantReply(text)
      .then(reply => {
        hideTyping();
        addMessage(reply, 'bot');
        chatHistory.push({ role: 'assistant', content: reply });
      })
      .catch(() => {
        hideTyping();
        addMessage(RESPONSES_FALLBACK, 'bot');
      })
      .finally(() => {
        sending = false;
        send.disabled = false;
      });
  }

  toggle.addEventListener('click', () => {
    const isOpen = chat.classList.toggle('open');
    toggle.innerHTML = isOpen ? '<div class="pulse-ring"></div>✕' : '<div class="pulse-ring"></div>🤖';
    if (isOpen && messages.children.length === 0) {
      setTimeout(() => {
        addMessage(`Hi! I'm Victor's AI assistant. I can tell you about his projects, services, technologies, and availability. What would you like to know?`, 'bot');
      }, 300);
    }
  });

  closeBtn.addEventListener('click', () => {
    chat.classList.remove('open');
    toggle.innerHTML = '<div class="pulse-ring"></div>🤖';
  });

  send.addEventListener('click', () => handleSend(input.value));
  input.addEventListener('keydown', e => { if (e.key === 'Enter') handleSend(input.value); });

  suggestions.querySelectorAll('.suggestion').forEach(btn => {
    btn.addEventListener('click', () => {
      handleSend(btn.textContent);
      suggestions.style.display = 'none';
    });
  });
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', createAssistant);
} else {
  createAssistant();
}
