/* Victor Akor Portfolio v3 — Multi-step Contact Form */
'use strict';

(function initContactForm() {
  const form = document.getElementById('contactForm');
  if (!form) return;

  let currentStep = 1;
  const totalSteps = 6;
  const data = {};

  function showStep(n) {
    document.querySelectorAll('.form-step').forEach(s => s.classList.remove('active'));
    const step = document.getElementById(`step${n}`);
    if (step) step.classList.add('active');

    document.querySelectorAll('.step-dot').forEach((dot, i) => {
      dot.classList.remove('active', 'done');
      if (i + 1 < n) dot.classList.add('done');
      else if (i + 1 === n) dot.classList.add('active');
    });
    document.querySelectorAll('.step-line').forEach((line, i) => {
      line.classList.toggle('done', i + 1 < n);
    });

    currentStep = n;
  }

  function nextStep() {
    if (!validateStep(currentStep)) return;
    collectStep(currentStep);
    if (currentStep < totalSteps) showStep(currentStep + 1);
  }

  function prevStep() {
    if (currentStep > 1) showStep(currentStep - 1);
  }

  function validateStep(n) {
    if (n === 1) {
      const selected = document.querySelector('#step1 .option-card.selected');
      if (!selected) { showError('Please select a project type.'); return false; }
    }
    if (n === 2) {
      const selected = document.querySelector('#step2 .option-card.selected');
      if (!selected) { showError('Please select a budget range.'); return false; }
    }
    if (n === 3) {
      const selected = document.querySelector('#step3 .option-card.selected');
      if (!selected) { showError('Please select a timeline.'); return false; }
    }
    if (n === 4) {
      const name = document.getElementById('clientName');
      const email = document.getElementById('clientEmail');
      if (!name?.value.trim()) { showError('Name is required.'); return false; }
      if (!email?.value.trim() || !email.value.includes('@')) { showError('Valid email is required.'); return false; }
    }
    if (n === 5) {
      const desc = document.getElementById('projectDesc');
      if (!desc?.value.trim() || desc.value.trim().length < 20) {
        showError('Please describe your project (at least 20 characters).'); return false;
      }
    }
    return true;
  }

  function collectStep(n) {
    if (n === 1) data.project_type = document.querySelector('#step1 .option-card.selected')?.dataset.value;
    if (n === 2) data.budget = document.querySelector('#step2 .option-card.selected')?.dataset.value;
    if (n === 3) data.timeline = document.querySelector('#step3 .option-card.selected')?.dataset.value;
    if (n === 4) {
      data.name = document.getElementById('clientName')?.value.trim();
      data.email = document.getElementById('clientEmail')?.value.trim();
      data.phone = document.getElementById('clientPhone')?.value.trim();
      data.company = document.getElementById('clientCompany')?.value.trim();
    }
    if (n === 5) data.message = document.getElementById('projectDesc')?.value.trim();
  }

  // Option card selection
  document.querySelectorAll('.option-card').forEach(card => {
    card.addEventListener('click', () => {
      const group = card.closest('.option-grid');
      group.querySelectorAll('.option-card').forEach(c => c.classList.remove('selected'));
      card.classList.add('selected');
    });
  });

  // Navigation buttons
  document.querySelectorAll('[data-next]').forEach(btn => btn.addEventListener('click', nextStep));
  document.querySelectorAll('[data-prev]').forEach(btn => btn.addEventListener('click', prevStep));

  // Submit
  const submitBtn = document.getElementById('submitBtn');
  if (submitBtn) {
    submitBtn.addEventListener('click', async () => {
      if (!validateStep(5)) return;
      collectStep(5);

      submitBtn.disabled = true;
      submitBtn.textContent = 'Sending...';

      try {
        const res = await fetch('/api/contact', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(data),
        });

        if (res.ok) {
          showStep(6);
          document.getElementById('step6').innerHTML = `
            <div style="text-align:center;padding:40px 0;">
              <div style="font-size:4rem;margin-bottom:24px;">🎉</div>
              <h2 style="margin-bottom:12px;">Message Received!</h2>
              <p style="color:var(--muted);max-width:400px;margin:0 auto 32px;">
                Thank you, ${data.name}. I'll review your project and get back to you within 24 hours.
              </p>
              <a href="/" class="btn btn-primary btn-lg">Back to Home</a>
            </div>
          `;
        } else {
          showError('Something went wrong. Please try again.');
          submitBtn.disabled = false;
          submitBtn.textContent = 'Submit Project';
        }
      } catch {
        showError('Network error. Please check your connection.');
        submitBtn.disabled = false;
        submitBtn.textContent = 'Submit Project';
      }
    });
  }

  function showError(msg) {
    let err = document.getElementById('formError');
    if (!err) {
      err = document.createElement('div');
      err.id = 'formError';
      err.className = 'alert alert-error';
      form.prepend(err);
    }
    err.textContent = msg;
    err.style.display = 'block';
    setTimeout(() => { err.style.display = 'none'; }, 4000);
  }

  showStep(1);
})();
