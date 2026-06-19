# DEV.MD MASTER SPECIFICATION

This is the specification 

---

# PROJECT NAME

```text
Victor Akor Portfolio v3
```

---

# PROJECT PURPOSE

Build a premium software engineering portfolio and client acquisition platform for Victor Akor.

The platform must establish authority in:

* Software Engineering
* AI Engineering
* Computer Vision
* Backend Engineering
* Go Development
* System Architecture

The platform should create the perception of:

```text
Senior Engineer
Consultant
Technical Founder
AI Specialist
```

rather than:

```text
Freelancer
Junior Developer
Student
```

---

# TECHNOLOGY STACK

Frontend:

```text
HTML5
CSS3
Vanilla JavaScript
```

Backend:

```text
Go
```

Database:

```text
PostgreSQL
```

Hosting:

```text
Railway
```

Email:

```text
Resend
```

Storage:

```text
Cloudinary
```

Analytics:

```text
Google Analytics
```

Authentication:

```text
JWT
```

---

# DESIGN PHILOSOPHY

The design must feel like a combination of:

```text
Apple
Stripe
Linear
Framer
Vercel
```

Characteristics:

```text
Clean
Premium
Modern
Fast
Professional
Minimal
Trustworthy
```

Avoid:

```text
Bright colors
Cheap gradients
Overloaded UI
Cartoonish designs
```

---

# COLOR SYSTEM

Primary:

```css
#2563eb
```

Secondary:

```css
#06b6d4
```

Background:

```css
#050a14
```

Cards:

```css
#0d1628
```

Text:

```css
#e2e8f0
```

Muted:

```css
#64748b
```

Success:

```css
#10b981
```

Danger:

```css
#ef4444
```

---

# WEBSITE STRUCTURE

---

## HOME

Route:

```text
/
```

Contains:

```text
Hero
Trust Indicators
Services
Case Studies
Process
Skills
Testimonials
Blog Preview
Contact CTA
```

---

# HERO SECTION

Full viewport.

Contains:

### Animated background

Features:

```text
Particles
Grid
Neural lines
Floating orbs
Parallax
```

---

### Headline

```text
Building Intelligent Software
That Solves Real Business Problems
```

---

### Subheadline

```text
I design and engineer scalable web platforms,
AI-powered systems, and high-performance backend
systems for startups, businesses, and organizations.
```

---

### Buttons

Primary:

```text
Book Consultation
```

Secondary:

```text
View Case Studies
```

---

### Metrics

Animated counters.

```text
Projects Built

AI Systems

Years Coding

Technologies Mastered
```

---

# TRUST SECTION

Purpose:

Build credibility.

Contains:

```text
AI Engineer
Backend Engineer
Computer Vision Specialist
Go Developer
System Architect
```

Displayed as floating achievement cards.

---

# ABOUT SECTION

Contains:

Professional portrait.

Biography.

Story.

Mission.

Vision.

Timeline.

Career journey.

---

# SERVICES SECTION

Four major services.

---

## AI Engineering

Includes:

```text
Face Recognition
Object Detection
Computer Vision
Deep Learning
NLP
```

---

## Backend Engineering

Includes:

```text
REST APIs
Authentication
Microservices
Database Design
Performance Optimization
```

---

## Full Stack Development

Includes:

```text
Business Applications
Dashboards
Admin Panels
CMS
SaaS Platforms
```

---

## Technical Consulting

Includes:

```text
Architecture
Code Review
Scaling
Optimization
Technology Decisions
```

---

# CASE STUDIES

Not project cards.

Actual case studies.

Each project becomes its own page.

---

## Mall Surveillance System

URL:

```text
/projects/mall-surveillance-system
```

Contains:

```text
Problem
Challenges
Architecture
Implementation
Technologies
Results
Lessons Learned
```

---

## AI Face Recognition System

URL:

```text
/projects/face-recognition-system
```

Contains:

```text
Problem
Dataset
Training
Deployment
Results
```

---

## Eye Disease Detection

Contains:

```text
Problem
CNN Design
Training
Evaluation
Accuracy
```

---

## Hackerthon

Contains:

```text
Platform Design
Architecture
Backend Design
Scalability
```

---

## Text Analyzer

Contains:

```text
NLP Pipeline
Analysis Engine
Architecture
```

---

## CLI Calculator

Contains:

```text
Parser Design
Go Concepts Used
```

---

## Gwinks Hub

Contains:

```text
Authentication
Database Design
Real Time Features
```

---

# SKILLS SECTION

Interactive.

No progress bars.

Skills represented as technology nodes.

Technologies:

```text
Go
Python
JavaScript
HTML
CSS
PostgreSQL
Docker
OpenCV
TensorFlow
Git
Linux
REST APIs
```

Hovering reveals:

```text
Experience
Projects
Use Cases
```

---

# BLOG SECTION

Route:

```text
/blog
```

Purpose:

SEO.

Authority.

Traffic generation.

---

Blog Categories:

```text
Go
AI
Computer Vision
Backend
System Design
Career
```

---

# CONTACT FUNNEL

This is critical.

Not a normal form.

Multi-step form.

---

Step 1

Project Type.

```text
Web Application
AI System
Backend API
Consulting
Other
```

---

Step 2

Budget.

```text
$500-$1000
$1000-$5000
$5000-$10000
$10000+
```

---

Step 3

Timeline.

```text
ASAP
1 Month
3 Months
Flexible
```

---

Step 4

Client Information.

```text
Name
Email
Phone
Company
```

---

Step 5

Project Description.

---

Step 6

Submit.

---

# ADMIN SYSTEM

Route:

```text
/admin
```

---

# AUTHENTICATION

Do NOT hardcode credentials.

The current portfolio does this conceptually.

Production system must not.

Store:

```text
email
password_hash
```

inside PostgreSQL.

Use bcrypt.

---

# ADMIN DASHBOARD

Displays:

```text
Total Leads
New Leads
Contacted
Negotiation
Won
Lost
```

---

# LEAD MANAGEMENT

Table:

```text
Name
Email
Phone
Company
Budget
Project Type
Status
Created Date
```

Statuses:

```text
New
Contacted
Meeting Scheduled
Proposal Sent
Negotiation
Won
Lost
```

---

# LEAD DETAILS PAGE

Contains:

```text
Lead Information
Message
Timeline
Budget
Notes
Communication History
```

---

# EMAIL CENTER

Admin can:

```text
Send Email
Reply
Follow Up
Schedule Meeting
```

through dashboard.

---

# BLOG MANAGEMENT

Create.

Edit.

Delete.

Publish.

Unpublish.

---

# PROJECT MANAGEMENT

Create.

Edit.

Delete.

Feature.

Archive.

---

# ANALYTICS

Track:

```text
Visitors
Sessions
Contact Submissions
Conversion Rate
Project Views
Blog Views
```

---

# DATABASE SCHEMA

## users

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

---

## leads

```sql
CREATE TABLE leads (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT,
    company TEXT,
    project_type TEXT,
    budget TEXT,
    timeline TEXT,
    message TEXT,
    status TEXT DEFAULT 'new',
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

---

## lead_notes

```sql
CREATE TABLE lead_notes (
    id UUID PRIMARY KEY,
    lead_id UUID REFERENCES leads(id),
    note TEXT,
    created_at TIMESTAMP
);
```

---

## projects

```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    title TEXT,
    slug TEXT,
    description TEXT,
    github_url TEXT,
    live_url TEXT,
    featured BOOLEAN,
    created_at TIMESTAMP
);
```

---

## blog_posts

```sql
CREATE TABLE blog_posts (
    id UUID PRIMARY KEY,
    title TEXT,
    slug TEXT,
    excerpt TEXT,
    content TEXT,
    cover_image TEXT,
    published BOOLEAN,
    created_at TIMESTAMP
);
```

---

## testimonials

```sql
CREATE TABLE testimonials (
    id UUID PRIMARY KEY,
    name TEXT,
    company TEXT,
    role TEXT,
    rating INTEGER,
    content TEXT
);
```

---

# GO PROJECT STRUCTURE

```text
portfolio/

├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── auth/
│   ├── leads/
│   ├── projects/
│   ├── blog/
│   ├── dashboard/
│   ├── analytics/
│   ├── middleware/
│   ├── templates/
│   └── database/
│
├── migrations/
│
├── static/
│   ├── css/
│   ├── js/
│   ├── images/
│   └── icons/
│
├── templates/
│   ├── public/
│   └── admin/
│
├── docs/
│
├── railway.json
│
├── Dockerfile
│
└── go.mod
```

---

# PERFORMANCE REQUIREMENTS

Lighthouse:

```text
Performance > 95

Accessibility > 95

SEO > 95

Best Practices > 95
```

---

# MOBILE REQUIREMENTS

Must support:

```text
320px
375px
425px
768px
1024px
1440px
1920px
```

No horizontal scrolling.

No layout shifts.

---

# UNIQUE FEATURE

The single most valuable feature:

## AI SALES ASSISTANT

Floating assistant trained on:

* your projects
* services
* technologies
* case studies

Visitors can ask:

```text
Can Victor build an AI surveillance system?

Can Victor build a SaaS platform?

How much experience does Victor have with Go?

What technologies does Victor use?
```

The assistant answers using portfolio content and can convert the conversation into a lead form.