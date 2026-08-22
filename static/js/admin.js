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
let postsCache = [];

async function loadBlogPosts() {
  const res = await apiFetch('/api/admin/blog');
  if (!res) return;
  const posts = await res.json();
  postsCache = posts || [];
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
  const res = await apiFetch(`/api/admin/blog/${id}`, { method: 'DELETE' });
  if (!res) return;
  if (!res.ok) { showAlert('Could not delete that post.', 'error'); return; }
  showAlert('Post deleted.');
  loadBlogPosts();
}

function openPostModal(post = null) {
  const f = id => document.getElementById(id);
  f('postId').value = post ? post.id : '';
  f('postTitle').value = post ? post.title || '' : '';
  f('postCategory').value = post ? post.category || '' : '';
  f('postExcerpt').value = post ? post.excerpt || '' : '';
  f('postContent').value = post ? post.content || '' : '';
  f('postCoverUrl').value = post ? post.cover_image || '' : '';
  f('postPublished').checked = post ? !!post.published : false;
  f('postCoverFile').value = '';
  setPreview('postCoverPreview', post ? post.cover_image : '');
  f('postModalTitle').textContent = post ? 'Edit Post' : 'New Post';
  openModal('postModal');
}

function editPost(id) {
  const post = postsCache.find(p => p.id === id);
  if (!post) { showAlert('Could not load that post.', 'error'); return; }
  openPostModal(post);
}

async function savePost() {
  const f = id => document.getElementById(id);
  const btn = f('postSaveBtn');
  const title = f('postTitle').value.trim();
  if (!title) { showAlert('Title is required.', 'error'); return; }

  btn.disabled = true;
  try {
    let cover = f('postCoverUrl').value;
    if (f('postCoverFile').files.length) {
      btn.textContent = 'Uploading...';
      cover = (await uploadImage(f('postCoverFile'))) || cover;
    }
    btn.textContent = 'Saving...';

    const id = f('postId').value;
    const body = JSON.stringify({
      title,
      excerpt: f('postExcerpt').value,
      content: f('postContent').value,
      cover_image: cover,
      category: f('postCategory').value,
      published: f('postPublished').checked,
    });

    const res = id
      ? await apiFetch(`/api/admin/blog/${id}`, { method: 'PUT', body })
      : await apiFetch('/api/admin/blog/create', { method: 'POST', body });
    if (!res) return;
    if (!res.ok) throw new Error((await res.text()).trim() || 'Save failed');

    closeModal('postModal');
    showAlert(id ? 'Post updated.' : 'Post created.');
    loadBlogPosts();
  } catch (err) {
    showAlert(err.message || 'Save failed.', 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Save';
  }
}

// ---- Projects management ----
let projectsCache = [];

async function loadProjects() {
  const res = await apiFetch('/api/admin/projects');
  if (!res) return;
  const projects = await res.json();
  projectsCache = projects || [];
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
  const res = await apiFetch(`/api/admin/projects/${id}`, { method: 'DELETE' });
  if (!res) return;
  if (!res.ok) { showAlert('Could not delete that project.', 'error'); return; }
  showAlert('Project deleted.');
  loadProjects();
}

function openProjectModal(project = null) {
  const f = id => document.getElementById(id);
  f('projectId').value = project ? project.id : '';
  f('projectTitle').value = project ? project.title || '' : '';
  f('projectDescription').value = project ? project.description || '' : '';
  f('projectGithub').value = project ? project.github_url || '' : '';
  f('projectLive').value = project ? project.live_url || '' : '';
  f('projectImageUrl').value = project ? project.image_url || '' : '';
  f('projectFeatured').checked = project ? !!project.featured : false;
  f('projectImageFile').value = '';
  setPreview('projectImagePreview', project ? project.image_url : '');
  f('projectModalTitle').textContent = project ? 'Edit Project' : 'New Project';
  openModal('projectModal');
}

function editProject(id) {
  const project = projectsCache.find(p => p.id === id);
  if (!project) { showAlert('Could not load that project.', 'error'); return; }
  openProjectModal(project);
}

async function saveProject() {
  const f = id => document.getElementById(id);
  const btn = f('projectSaveBtn');
  const title = f('projectTitle').value.trim();
  if (!title) { showAlert('Title is required.', 'error'); return; }

  btn.disabled = true;
  try {
    let image = f('projectImageUrl').value;
    if (f('projectImageFile').files.length) {
      btn.textContent = 'Uploading...';
      image = (await uploadImage(f('projectImageFile'))) || image;
    }
    btn.textContent = 'Saving...';

    const id = f('projectId').value;
    const body = JSON.stringify({
      title,
      description: f('projectDescription').value,
      github_url: f('projectGithub').value,
      live_url: f('projectLive').value,
      image_url: image,
      featured: f('projectFeatured').checked,
      archived: false,
    });

    const res = id
      ? await apiFetch(`/api/admin/projects/${id}`, { method: 'PUT', body })
      : await apiFetch('/api/admin/projects/create', { method: 'POST', body });
    if (!res) return;
    if (!res.ok) throw new Error((await res.text()).trim() || 'Save failed');

    closeModal('projectModal');
    showAlert(id ? 'Project updated.' : 'Project created.');
    loadProjects();
  } catch (err) {
    showAlert(err.message || 'Save failed.', 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Save';
  }
}

// ---- Image upload ----
async function uploadImage(fileInput) {
  const file = fileInput.files && fileInput.files[0];
  if (!file) return null;

  const fd = new FormData();
  fd.append('file', file);

  // Deliberately NOT apiFetch: that helper forces Content-Type: application/json,
  // which would override the multipart boundary the browser must set itself.
  const res = await fetch('/api/admin/upload', { method: 'POST', body: fd });
  if (res.status === 401) { location.href = '/admin/login'; return null; }
  if (!res.ok) throw new Error((await res.text()).trim() || 'Upload failed');

  const data = await res.json();
  return data.url;
}

// Show the chosen file immediately, before it is uploaded.
function previewLocal(input, imgId) {
  const file = input.files && input.files[0];
  if (file) setPreview(imgId, URL.createObjectURL(file));
}

function setPreview(imgId, url) {
  const img = document.getElementById(imgId);
  if (!img) return;
  if (url) { img.src = url; img.style.display = 'block'; }
  else { img.removeAttribute('src'); img.style.display = 'none'; }
}

// ---- Modals ----
function openModal(id) { const el = document.getElementById(id); if (el) el.classList.add('open'); }
function closeModal(id) { const el = document.getElementById(id); if (el) el.classList.remove('open'); }

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
