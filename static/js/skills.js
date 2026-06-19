/* Victor Akor Portfolio v3 — Interactive Skill Nodes */
'use strict';

const SKILLS = [
  { name: 'Go', icon: '🐹', exp: '3+ years', projects: '8+', uses: 'APIs, CLIs, Microservices' },
  { name: 'Python', icon: '🐍', exp: '4+ years', projects: '12+', uses: 'AI, ML, Scripting, CV' },
  { name: 'JavaScript', icon: '⚡', exp: '4+ years', projects: '15+', uses: 'Web Apps, Node.js, APIs' },
  { name: 'HTML', icon: '🌐', exp: '5+ years', projects: '20+', uses: 'Semantic Markup, SEO' },
  { name: 'CSS', icon: '🎨', exp: '5+ years', projects: '20+', uses: 'Animations, Responsive Design' },
  { name: 'PostgreSQL', icon: '🐘', exp: '3+ years', projects: '10+', uses: 'Relational DB, Optimization' },
  { name: 'Docker', icon: '🐳', exp: '2+ years', projects: '6+', uses: 'Containerization, Deployment' },
  { name: 'OpenCV', icon: '👁️', exp: '2+ years', projects: '5+', uses: 'Computer Vision, Image Processing' },
  { name: 'TensorFlow', icon: '🧠', exp: '2+ years', projects: '4+', uses: 'Deep Learning, CNN, NLP' },
  { name: 'Git', icon: '🔀', exp: '5+ years', projects: '30+', uses: 'Version Control, CI/CD' },
  { name: 'Linux', icon: '🐧', exp: '4+ years', projects: '20+', uses: 'Server Admin, Shell Scripting' },
  { name: 'REST APIs', icon: '🔌', exp: '4+ years', projects: '18+', uses: 'API Design, Integration' },
];

function renderSkills() {
  const grid = document.getElementById('skillsGrid');
  if (!grid) return;

  grid.innerHTML = SKILLS.map(s => `
    <div class="skill-node" tabindex="0" role="button" aria-label="${s.name}">
      <span class="skill-icon">${s.icon}</span>
      <span class="skill-name">${s.name}</span>
      <div class="skill-tooltip" role="tooltip">
        <h5>${s.name}</h5>
        <p>⏱ Experience: ${s.exp}</p>
        <p>📦 Projects: ${s.projects}</p>
        <p>🔧 Uses: ${s.uses}</p>
      </div>
    </div>
  `).join('');

  // Keyboard accessibility
  grid.querySelectorAll('.skill-node').forEach(node => {
    node.addEventListener('keydown', e => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        const tooltip = node.querySelector('.skill-tooltip');
        const isVisible = tooltip.style.opacity === '1';
        tooltip.style.opacity = isVisible ? '0' : '1';
      }
    });
  });
}

renderSkills();
