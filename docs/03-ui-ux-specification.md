# 03 - UI/UX Specification & Interaction Design

**Product:** Mindora — Smart Student Productivity Application  
**Version:** 1.0 (Prototype Redesign)  
**Visual Style:** Nature Inspired + Balanced Information Density

---

## 1. Overview & Experience Architecture

Mindora separates its primary experience into two distinct environments:
1. **User Workspace (`/` & `/login`):** An empathetic, calm, student-focused productivity hub designed to structure daily academic tasks, foster focused study intervals, and provide proactive AI guidance.
2. **Admin Console (`/admin`):** A high-level administrative interface for student organization leaders or educators to manage member access, workspace allocations, and aggregated activity without infringing on individual student privacy.

---

## 2. Page Specifications

### 2.1 Login Page (`/login`)
- **Reference:** `Login.pdf`
- **Layout:** Split-card layout inside a responsive, centered container.
- **Left Column (Nature Hero & Branding):**
  - **Brand Header:** Official Mindora gradient brain logo + wordmark.
  - **Brand Tagline:** `PLAN · FOCUS · ACHIEVE` in uppercase tracking.
  - **Core Value Proposition:** *"Make every minute feel meaningful."* accompanied by supportive copy.
  - **Social Proof / Empathetic Touch:** Testimonial card highlighting incremental progress (*"You don't have to be perfect every day..."* — Alicia, Product Designer).
  - **Visual Decoration:** Subtle botanical pine tree and rolling hill silhouette at the bottom.
- **Right Column (Sign-In Interface):**
  - **Workspace Role Tabs:** Dual tab selector between **User Workspace** and **Admin Console** allowing test users to simulate both roles immediately.
  - **Form Fields:** Email input with left mail icon; Password input with lock icon and interactive show/hide toggle.
  - **Form Utilities:** Remember me checkbox + Forgot password action.
  - **Primary CTA:** Full-width primary teal button (`#168F7A`) reading *"Sign in to workspace >"* or *"Sign in to console >"*.
  - **Footer:** *"New to Mindora? Create an account"* (eliminating all legacy Timeyfi branding).
  - **Prototype Footnote:** Clarifies that authentication will integrate with backend endpoints in subsequent phases.

### 2.2 User Dashboard (`/`)
- **Reference:** `Dashboard User.pdf`
- **Layout:** 3-Column Balanced Information Density Architecture.
- **Column 1 — Navigation Sidebar (Left):**
  - Mindora logo with direct home routing.
  - Primary navigation: Dashboard (active), List, Calendar, Messages.
  - **My List Section:** Inbox, Work, Study, Health, Holiday trip with colored circular category badges.
  - **Workspace Section:** UI Design, A-team Class with category iconography.
  - Bottom user profile pill with quick logout trigger and shortcut to Admin console.
- **Column 2 — Main Activity Feed (Center, 8 cols):**
  - **User Hero Greeting:** Personalized banner (*"Good Morning, John"*) displaying the count of today's pending tasks alongside an integrated *"Up next"* schedule widget (e.g. Design Team Meeting at 11:30 AM).
  - **AI Planning Assistant:** Card featuring quick planning chips (*Plan my day*, *Prioritize tasks*, *Create focus plan*) and an interactive natural-language prompt field.
  - **Quick Add Task Bar:** Single-line task creation input with schedule chips (*Now*, *Tomorrow*, *Next week*, *Custom*) and category selector.
  - **Today's Tasks List:** Interactive task rows with completion checkmarks, category indicator dots, timestamps, and detail chevrons.
- **Column 3 — Utilities & Focus Rail (Right, 4 cols):**
  - **Calendar Widget:** Interactive monthly calendar view (December 2026) with event indicators and day selection.
  - **Focus Timer Widget:** Dedicated Pomodoro/study timer displaying selected task, `-` and `+` duration adjusters (default 30 mins), Play/Pause controls, and grounded with a calm botanical pine tree landscape illustration.

### 2.3 Admin Console Dashboard (`/admin`)
- **Reference:** `Dashbboard Admin.pdf`
- **Layout:** Standard Enterprise / SaaS Admin Architecture.
- **Sidebar (Left):**
  - Logo with prominent "ADMIN CONSOLE" badge.
  - Sections: **Workspace** (Overview, Members, Workspaces, Reports) and **Management** (Roles & access, Settings).
  - Bottom action: *"Back to main app"* returning seamlessly to the user workspace.
- **Main Viewport:**
  - **Admin Overview Header:** Date range filter dropdown (*"This week"*).
  - **KPI Metric Row (4 Cards):** Total members (248, +12.5%), Active tasks (1,429, +8.2%), Completion rate (84.6%, +4.1%), Active workspaces (16, +2 new).
  - **Analytics Grid:**
    - *Task Productivity Chart:* Visual dual-bar comparison of Completed vs Created tasks from Monday to Sunday.
    - *Recent Activity Feed:* Chronological audit log of organizational events (member joined, workspace created, role updated).
  - **Team Members Management Table:** Searchable list of members displaying Name, Email, Role (Admin, Manager, Member), Aggregated Tasks Completed, Status (Active, Away), and Action dropdowns.
  - **Privacy Enforcement:** Admin interface strictly restricts visibility to aggregate metrics and administrative flags; private student tasks, journal notes, and personal productivity entries are completely inaccessible to administrative roles.

---

## 3. UI States & Interaction Guidelines

1. **Empty States:** When lists or search queries return no results, render soft mint illustrations with gentle guidance (e.g. *"No pending tasks for today. Take a breath or plan ahead."*).
2. **Hover & Focus Feedback:** All interactive buttons and inputs utilize a `2px` focus ring in `primary/20` and smooth `transition-all 150ms`.
3. **Responsive Breakpoints:**
   - On screens `< 1024px`, sidebar converts to a smooth drawer triggered by the top navigation hamburger button.
   - On screens `< 768px`, the right rail widgets (Calendar & Focus timer) stack cleanly below the main feed.
