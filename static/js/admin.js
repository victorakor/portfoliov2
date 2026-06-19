/* Victor Akor Portfolio v3 — Admin Dashboard JS */
'use strict';

// ---- Auth helpers ----
function getToken() { return document.cookie.split(';').find(c => c.trim().startsWith('admin_token='))?.split('=')[1]; }

async function apiFetch(url, opts = {}) {
  const res = await fetch(url, { ...opts, headers: { 'Content-Type': 'application/json', ...opts.headers } });
  if (res.status === 401) { location.href = '/admin/login'; return null; }
  return res;
}

// ---- Login form ----
const loginForm = document.getElementById('loginForm');
if (loginForm) {
  loginForm.addEventListener('submit', async e => {
    e.preventDefault();
    const btn = loginForm.querySelector('button[type=submit]');
    btn.disabled = true; btn.textContent = 'Signing in...';
    const body = { email: loginForm.email.value, password: loginForm.password.value };
    try {
      const res = await fetch('/admin/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      if (res.ok) { location.href = '/admin'; }
      else {
        showAlert('Invalid email or password.', 'error');
        btn.disabled = false; btn.textContent = 'Sign In';
      }
    } catch { showAlert('Network error.', 'error'); btn.disabled = false; btn.textContent = 'Sign In'; }
  });
}

// ---- Dashboard stats ----
async function loadStats() {
  const res = await apiFetch('/api/admin/stats');
  if (!res) return;
  const stats = await res.json();
  Object.entries(stats).forEach(([key, val]) => {
    const el = document.getElementById(`stat_${key}`);
    if (el) el.textContent = val;
  });
}

// ---- Leads table ----
async function loadLeads(status = '') {
  const url = status ? `/api/admin/leads?status=${status}` : '/api/admin/leads';
  const res = await apiFetch(url);
  if (!res) return;
  const leads = await res.json();
  const tbody = document.getElementById('leadsTableBody');
  if (!tbody) return;

  tbody.innerHTML = (leads || []).map(l => `
    <tr>
      <td><strong>${escHtml(l.name)}</strong></td>
      <td>${escHtml(l.email)}</td>
      <td>${escHtml(l.company || '—')}</td>
      <td>${escHtml(l.project_type || '—')}</td>
      <td>${escHtml(l.budget || '—')}</td>
      <td>
        <select class="status-select" data-id="${l.id}" onchange="updateLeadStatus(this)">
          ${['new','contacted','meeting_scheduled','proposal_sent','negotiation','won','lost'].map(s =>
            `<option value="${s}" ${l.status===s?'selected':''}>${s.replace('_',' ')}</option>`
          ).join('')}
        </select>
      </td>
      <td>${formatDate(l.created_at)}</td>
      <td><a href="/admin/leads/${l.id}" class="btn btn-sm btn-secondary">View</a></td>
    </tr>
  `).join('') || '<tr><td colspan="8" style="text-align:center;color:var(--muted);padding:40px">No leads yet.</td></tr>';
}

async function updateLeadStatus(select) {
  const id = select.dataset.id;
  const status = select.value;
  await apiFetch(`/api/admin/leads/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) });
}

// ---- Blog management ----
async function loadBlogPosts() {
  const res = await apiFetch('/api/admin/blog');
  if (!res) return;
  const posts = await res.json();
  const tbody = document.getElementById('blogTableBody');
  if (!tbody) return;

  tbody.innerHTML = (posts || []).map(p => `
    <tr>
      <td><strong>${escHtml(p.title)}</strong></td>
      <td>${escHtml(p.category || '—')}</td>
      <td><span class="badge ${p.published ? 'badge-success' : 'badge-warning'}">${p.published ? 'Published' : 'Draft'}</span></td>
      <td>${formatDate(p.created_at)}</td>
      <td>
        <button class="btn btn-sm btn-secondary" onclick="editPost('${p.id}')">Edit</button>
        <button class="btn btn-sm btn-danger" onclick="deletePost('${p.id}')">Delete</button>
      </td>
    </tr>
  `).join('') || '<tr><td colspan="5" style="text-align:center;color:var(--muted);padding:40px">No posts yet.</td></tr>';
}

async function deletePost(id) {
  if (!confirm('Delete this post?')) return;
  await apiFetch(`/api/admin/blog/${id}`, { method: 'DELETE' });
  loadBlogPosts();
}

// ---- Projects management ----
async function loadProjects() {
  const res = await apiFetch('/api/admin/projects');
  if (!res) return;
  const projects = await res.json();
  const tbody = document.getElementById('projectsTableBody');
  if (!tbody) return;

  tbody.innerHTML = (projects || []).map(p => `
    <tr>
      <td><strong>${escHtml(p.title)}</strong></td>
      <td>${escHtml(p.slug)}</td>
      <td><span class="badge ${p.featured ? 'badge-primary' : 'badge-cyan'}">${p.featured ? 'Featured' : 'Normal'}</span></td>
      <td>${formatDate(p.created_at)}</td>
      <td>
        <button class="btn btn-sm btn-secondary" onclick="editProject('${p.id}')">Edit</button>
        <button class="btn btn-sm btn-danger" onclick="deleteProject('${p.id}')">Delete</button>
      </td>
    </tr>
  `).join('') || '<tr><td colspan="5" style="text-align:center;color:var(--muted);padding:40px">No projects yet.</td></tr>';
}

async function deleteProject(id) {
  if (!confirm('Delete this project?')) return;
  await apiFetch(`/api/admin/projects/${id}`, { method: 'DELETE' });
  loadProjects();
}

// ---- Utilities ----
function escHtml(s) { return String(s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }
function formatDate(d) { return d ? new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : '—'; }
function showAlert(msg, type = 'success') {
  const el = document.createElement('div');
  el.className = `alert alert-${type}`;
  el.textContent = msg;
  document.body.prepend(el);
  setTimeout(() => el.remove(), 4000);
}

// ---- Sidebar active link ----
document.querySelectorAll('.sidebar-link').forEach(link => {
  if (link.href === location.href) link.classList.add('active');
});

// ---- Auto-init based on page ----
if (document.getElementById('stat_total_leads')) loadStats();
if (document.getElementById('leadsTableBody')) loadLeads();
if (document.getElementById('blogTableBody')) loadBlogPosts();
if (document.getElementById('projectsTableBody')) loadProjects();
