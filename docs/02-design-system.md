# 02 - Mindora Unified Design System

**Visual Direction:** NATURE INSPIRED + BALANCED INFORMATION DENSITY  
**Brand Identity:** Mindora · PLAN · FOCUS · ACHIEVE  
**Target Platform:** Web (Desktop, Tablet, Mobile)

---

## 1. Visual Language & Core Principles

The Mindora interface provides a calm, supportive, and focused atmosphere tailored for students balancing academic coursework, self-improvement, and personal well-being.

### Core Principles
1. **Calm & Grounded:** Soft botanical tones (sage, mint, forest teal) replace harsh high-contrast saturated colors, reducing cognitive fatigue during prolonged study sessions.
2. **Balanced Information Density:** Crucial information (upcoming meetings, priority tasks, focus timer) is immediately scannable without visual crowding or excessive blank space.
3. **Purposeful Nature Accents:** Botanical curves and landscape motifs are used subtly in banners, empty states, and focus timers rather than cluttering functional workspaces.
4. **Accessible Hierarchy:** Explicit typographic scale and WCAG AA contrast compliance between primary text (`#183B35`) and surfaces (`#FFFFFF`, `#F5F8F5`).

---

## 2. Design Tokens

### 2.1 Color Palette

| Token Name | Hex Code | Tailwind Token | Intended Usage |
|---|---|---|---|
| **Primary Teal** | `#168F7A` | `primary.DEFAULT` | Primary brand actions, active navigation, key highlights |
| **Primary Hover** | `#127564` | `primary.hover` | Hover states for primary buttons and interactive elements |
| **Secondary Aqua** | `#20B8A5` | `secondary.DEFAULT` | Secondary accents, progress bars, active toggle tabs |
| **Nature Green** | `#6F9F78` | `nature.green` | Botanical illustrations, organic chart series |
| **Sage Green** | `#A8C5A5` | `nature.sage` | Decorative accents, subtle badges, chart secondary series |
| **Soft Mint** | `#E5F3E9` | `nature.mint` | Light container fills, badge backgrounds, input focus rings |
| **Background Tint**| `#F5F8F5` | `surface.muted` | Application body background, sidebar hover states |
| **Surface (Card)** | `#FFFFFF` | `surface.DEFAULT` | Content cards, modal sheets, dropdown menus |
| **Primary Text** | `#183B35` | `text.primary` | Headings, labels, high-emphasis text |
| **Secondary Text** | `#71827B` | `text.secondary` | Subtitles, helper text, timestamps, placeholder text |
| **Border Neutral** | `#E2EAE4` | `border.DEFAULT` | Card borders, dividers, table row separators |
| **Subtle Border** | `#EEF4F0` | `border.subtle` | Soft internal card dividers, pill borders |

### 2.2 Semantic Status Colors

| State | Background | Text | Border | Usage |
|---|---|---|---|---|
| **Success / Active** | `#ECFDF5` (`emerald-50`) | `#047857` (`emerald-700`) | `#A7F3D0` (`emerald-200`) | Running status, active members, task completed |
| **Warning / Away** | `#FFFBEB` (`amber-50`) | `#B45309` (`amber-700`) | `#FDE68A` (`amber-200`) | Away status, approaching deadline |
| **Info** | `#EFF6FF` (`blue-50`) | `#1D4ED8` (`blue-700`) | `#BFDBFE` (`blue-200`) | Workspaces, system announcements |
| **Neutral / Muted** | `#F8FAFC` (`slate-50`) | `#475569` (`slate-600`) | `#E2E8F0` (`slate-200`) | Inactive tabs, disabled states |

---

## 3. Typography System

**Font Family:** Inter, system-ui, -apple-system, sans-serif

| Style Name | Size | Line Height | Weight | Tailwind Class | Usage |
|---|---|---|---|---|---|
| **Display Heading** | 30px (1.875rem) | 36px | 700 (Bold) | `text-3xl font-bold` | Login hero statement, top dashboard greeting |
| **Page Heading** | 24px (1.5rem) | 32px | 700 (Bold) | `text-2xl font-bold` | Admin overview title, section titles |
| **Card Title** | 18px (1.125rem) | 28px | 600 (Semibold)| `text-lg font-semibold` | Widget headers (AI Assistant, Focus Timer) |
| **Section Label** | 12px (0.75rem) | 16px | 600 (Semibold)| `text-xs font-semibold uppercase tracking-wider` | Category headers, sidebar section labels |
| **Body Large** | 16px (1rem) | 24px | 400 (Regular) | `text-base` | Login inputs, primary button text |
| **Body Default** | 14px (0.875rem) | 20px | 500 (Medium) | `text-sm font-medium` | Task titles, table row items, navigation links |
| **Caption / Meta** | 12px (0.75rem) | 16px | 500 (Medium) | `text-xs font-medium` | Timestamps, tags, badge labels |

---

## 4. Spacing, Radii & Elevation

### 4.1 Spacing Scale
- Micro spacing: `gap-1.5` (6px), `gap-2` (8px), `gap-3` (12px)
- Component internal padding: `p-4` (16px), `p-6` (24px), `p-8` (32px)
- Layout margins: `space-y-6` (24px), `gap-6` (24px)

### 4.2 Border Radius
- Pill elements: `rounded-full` (badges, quick chips, avatars)
- Buttons & inputs: `rounded-xl` (12px)
- Cards & modals: `rounded-2xl` (16px)
- Large hero containers: `rounded-3xl` (24px)

### 4.3 Elevation & Shadows
- **Card Default:** `shadow-sm` (`box-shadow: 0 1px 2px 0 rgba(22, 143, 122, 0.04)`)
- **Card Hover:** `hover:shadow-md` (`box-shadow: 0 4px 6px -1px rgba(22, 143, 122, 0.08)`)
- **Dropdown & Modal:** `shadow-lg` (`box-shadow: 0 10px 15px -3px rgba(24, 59, 53, 0.1)`)

---

## 5. Responsive Grid & Breakpoints

| Breakpoint | Minimum Width | Layout Behavior |
|---|---|---|
| **Desktop Wide** | 1440px | Full 3-column layout: 260px Sidebar + Main Feed + 340px Right Rail |
| **Desktop Standard**| 1366px | Balanced 3-column layout: 240px Sidebar + Main Feed + 300px Right Rail |
| **Tablet** | 768px - 1024px | Collapsible sidebar, 2-column feed (main task feed + stacked widgets) |
| **Mobile Large** | 390px - 430px | Single-column fluid view, drawer navigation, scrollable tabs |
| **Mobile Compact**| 375px | Single-column condensed view, touch targets >= 44px |

---

## 6. Atomic Component Architecture

Mindora components are implemented according to Atomic Design:

```text
src/components/
├── atoms/          # Primitive components (Button, Input, Checkbox, Badge, Heading, Text, PulseDot, Icons)
├── molecules/      # Composed functional units (StatusBadge, HeaderBlock, SearchInput, TaskItem, MetricCard)
├── organisms/      # Complex autonomous widgets (Sidebar, TopNavbar, AIAssistant, FocusTimer, MembersTable)
└── templates/      # Page layout skeletons (AuthTemplate, DashboardTemplate)
```
