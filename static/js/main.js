/* Victor Akor Portfolio v3 — Main JS */
'use strict';

// ---- Navbar scroll effect ----
const navbar = document.querySelector('.navbar');
if (navbar) {
  window.addEventListener('scroll', () => {
    navbar.classList.toggle('scrolled', window.scrollY > 40);
  }, { passive: true });
}

// ---- Mobile menu ----
const hamburger = document.querySelector('.hamburger');
const navLinks = document.querySelector('.navbar-links');
if (hamburger && navLinks) {
  hamburger.addEventListener('click', () => {
    const open = navLinks.classList.toggle('mobile-open');
    hamburger.setAttribute('aria-expanded', open);
    navLinks.style.cssText = open
      ? 'display:flex;flex-direction:column;position:fixed;top:70px;left:0;right:0;background:rgba(5,10,20,0.97);backdrop-filter:blur(20px);padding:24px;gap:20px;border-bottom:1px solid rgba(255,255,255,0.06);z-index:999;'
      : '';
  });
}

// ---- Hero Canvas: Particles + Grid + Neural Lines + Orbs ----
(function initHeroCanvas() {
  const canvas = document.getElementById('heroCanvas');
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  let W, H, particles = [], nodes = [], animId;

  function resize() {
    W = canvas.width = canvas.offsetWidth;
    H = canvas.height = canvas.offsetHeight;
  }
  resize();
  window.addEventListener('resize', resize, { passive: true });

  // Particles
  function createParticle() {
    return {
      x: Math.random() * W,
      y: H + 10,
      vx: (Math.random() - 0.5) * 0.5,
      vy: -(Math.random() * 0.8 + 0.3),
      size: Math.random() * 1.5 + 0.5,
      alpha: Math.random() * 0.5 + 0.2,
      life: 0,
      maxLife: Math.random() * 300 + 200,
    };
  }
  for (let i = 0; i < 60; i++) {
    const p = createParticle();
    p.y = Math.random() * H;
    p.life = Math.random() * p.maxLife;
    particles.push(p);
  }

  // Neural nodes
  for (let i = 0; i < 12; i++) {
    nodes.push({ x: Math.random() * W, y: Math.random() * H, vx: (Math.random()-0.5)*0.3, vy: (Math.random()-0.5)*0.3 });
  }

  let mouse = { x: W/2, y: H/2 };
  canvas.addEventListener('mousemove', e => {
    const r = canvas.getBoundingClientRect();
    mouse.x = e.clientX - r.left;
    mouse.y = e.clientY - r.top;
  }, { passive: true });

  function drawGrid() {
    ctx.strokeStyle = 'rgba(255,255,255,0.025)';
    ctx.lineWidth = 1;
    const size = 60;
    for (let x = 0; x < W; x += size) {
      ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, H); ctx.stroke();
    }
    for (let y = 0; y < H; y += size) {
      ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(W, y); ctx.stroke();
    }
  }

  function drawNeuralLines() {
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const dx = nodes[i].x - nodes[j].x;
        const dy = nodes[i].y - nodes[j].y;
        const dist = Math.sqrt(dx*dx + dy*dy);
        if (dist < 200) {
          const alpha = (1 - dist/200) * 0.15;
          ctx.strokeStyle = `rgba(37,99,235,${alpha})`;
          ctx.lineWidth = 1;
          ctx.beginPath();
          ctx.moveTo(nodes[i].x, nodes[i].y);
          ctx.lineTo(nodes[j].x, nodes[j].y);
          ctx.stroke();
        }
      }
      // Mouse connection
      const dx = nodes[i].x - mouse.x;
      const dy = nodes[i].y - mouse.y;
      const dist = Math.sqrt(dx*dx + dy*dy);
      if (dist < 180) {
        const alpha = (1 - dist/180) * 0.3;
        ctx.strokeStyle = `rgba(6,182,212,${alpha})`;
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(nodes[i].x, nodes[i].y);
        ctx.lineTo(mouse.x, mouse.y);
        ctx.stroke();
      }
    }
  }

  function drawNodes() {
    nodes.forEach(n => {
      n.x += n.vx; n.y += n.vy;
      if (n.x < 0 || n.x > W) n.vx *= -1;
      if (n.y < 0 || n.y > H) n.vy *= -1;
      ctx.beginPath();
      ctx.arc(n.x, n.y, 2, 0, Math.PI*2);
      ctx.fillStyle = 'rgba(37,99,235,0.4)';
      ctx.fill();
    });
  }

  function drawParticles() {
    particles.forEach((p, i) => {
      p.x += p.vx; p.y += p.vy; p.life++;
      const progress = p.life / p.maxLife;
      const alpha = p.alpha * (progress < 0.1 ? progress/0.1 : progress > 0.9 ? (1-progress)/0.1 : 1);
      ctx.beginPath();
      ctx.arc(p.x, p.y, p.size, 0, Math.PI*2);
      ctx.fillStyle = `rgba(37,99,235,${alpha})`;
      ctx.fill();
      if (p.life >= p.maxLife) particles[i] = createParticle();
    });
  }

  function animate() {
    ctx.clearRect(0, 0, W, H);
    drawGrid();
    drawNeuralLines();
    drawNodes();
    drawParticles();
    animId = requestAnimationFrame(animate);
  }
  animate();
})();

// ---- Animated Counters ----
function animateCounter(el) {
  const target = parseInt(el.dataset.target, 10);
  const duration = 2000;
  const start = performance.now();
  function update(now) {
    const elapsed = now - start;
    const progress = Math.min(elapsed / duration, 1);
    const eased = 1 - Math.pow(1 - progress, 3);
    el.textContent = Math.round(eased * target);
    if (progress < 1) requestAnimationFrame(update);
    else el.textContent = target;
  }
  requestAnimationFrame(update);
}

// ---- Intersection Observer for reveals + counters ----
const revealObserver = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      entry.target.classList.add('revealed');
      revealObserver.unobserve(entry.target);
    }
  });
}, { threshold: 0.1, rootMargin: '0px 0px -40px 0px' });

const counterObserver = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      animateCounter(entry.target);
      counterObserver.unobserve(entry.target);
    }
  });
}, { threshold: 0.5 });

document.querySelectorAll('[data-reveal]').forEach(el => revealObserver.observe(el));
document.querySelectorAll('.counter[data-target]').forEach(el => counterObserver.observe(el));

// ---- Parallax on hero ----
window.addEventListener('scroll', () => {
  const hero = document.querySelector('.hero-content');
  if (hero) {
    hero.style.transform = `translateY(${window.scrollY * 0.15}px)`;
  }
}, { passive: true });

// ---- Smooth scroll for anchor links ----
document.querySelectorAll('a[href^="#"]').forEach(a => {
  a.addEventListener('click', e => {
    const target = document.querySelector(a.getAttribute('href'));
    if (target) {
      e.preventDefault();
      target.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  });
});

// ---- Analytics tracking ----
function trackEvent(type, page) {
  const sessionId = sessionStorage.getItem('sid') || (() => {
    const id = Math.random().toString(36).slice(2);
    sessionStorage.setItem('sid', id);
    return id;
  })();
  fetch('/api/track', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ event_type: type, page: page || location.pathname, session_id: sessionId }),
  }).catch(() => {});
}
trackEvent('pageview');

// ---- Blog preview loader ----
async function loadBlogPreview() {
  const grid = document.getElementById('blogPreviewGrid');
  if (!grid) return;
  try {
    const res = await fetch('/api/blog');
    const posts = await res.json();
    if (!posts || !posts.length) return;
    grid.innerHTML = posts.slice(0, 3).map(p => `
      <article class="blog-card" data-reveal>
        <div class="blog-card-img">${getCategoryEmoji(p.category)}</div>
        <div class="blog-card-body">
          <div class="blog-card-category">${p.category || 'Engineering'}</div>
          <h3 class="blog-card-title">${p.title}</h3>
          <p class="blog-card-excerpt">${p.excerpt || ''}</p>
        </div>
        <div class="blog-card-footer">
          <span class="blog-card-date">${formatDate(p.created_at)}</span>
          <a href="/blog/${p.slug}" class="case-card-link">Read →</a>
        </div>
      </article>
    `).join('');
    grid.querySelectorAll('[data-reveal]').forEach(el => revealObserver.observe(el));
  } catch {}
}

function getCategoryEmoji(cat) {
  const map = { Go: '🐹', AI: '🤖', 'Computer Vision': '👁️', Backend: '⚙️', 'System Design': '🏗️', Career: '🚀' };
  return map[cat] || '📝';
}
function formatDate(d) {
  return d ? new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : '';
}

loadBlogPreview();
