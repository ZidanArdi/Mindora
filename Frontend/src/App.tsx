import { useMemo, useState } from "react"
import {
  ArrowUpRight,
  Bell,
  BarChart3,
  Bot,
  BriefcaseBusiness,
  CalendarDays,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  CirclePlus,
  Clock3,
  Download,
  Eye,
  EyeOff,
  Heart,
  House,
  Inbox,
  LayoutDashboard,
  ListTodo,
  LockKeyhole,
  LogOut,
  Mail,
  MoreHorizontal,
  Minus,
  Palette,
  Plane,
  Plus,
  Search,
  Send,
  Settings,
  ShieldCheck,
  Sparkles,
  Timer,
  UserPlus,
  Users,
  WandSparkles,
} from "lucide-react"

const calendarRows = [
  ["31", "30", "1", "2", "3", "4", "5"],
  ["6", "7", "8", "9", "10", "11", "12"],
  ["13", "14", "15", "16", "17", "18", "19"],
  ["20", "21", "22", "23", "24", "25", "26"],
  ["27", "28", "29", "30", "31", "1", "2"],
]

const calendarDots: Record<string, string[]> = {
  "3": ["blue", "orange"],
  "6": ["green", "blue"],
  "10": ["blue", "pink", "green"],
  "13": ["blue", "green"],
  "16": ["green"],
  "20": ["blue", "green"],
  "22": ["blue", "orange"],
  "29": ["green", "blue", "orange"],
  "31": ["green", "blue"],
}

const initialTasks = [
  {
    id: 1,
    title: "Cleaning up desk",
    list: "Inbox",
    time: "08.20 AM",
    color: "blue",
    done: false,
  },
  {
    id: 2,
    title: "Client meeting prep",
    list: "Work",
    time: "10.20 AM",
    color: "green",
    done: false,
  },
  {
    id: 3,
    title: "Small yoga",
    list: "Health",
    time: "02.30 PM",
    color: "orange",
    done: false,
  },
]

function BrandMark() {
  return (
    <span className="brand-mark" aria-hidden="true">
      <img src="/assets/mindora-logo.png" alt="" />
    </span>
  )
}

function LandscapeIllustration({ compact = false }: { compact?: boolean }) {
  const treeXs = compact ? [9, 19, 31, 45, 57, 70, 82, 94] : [18, 38]

  return (
    <svg
      className={`landscape ${compact ? "landscape--forest" : ""}`}
      viewBox="0 0 600 300"
      aria-hidden="true"
    >
      <defs>
        <linearGradient
          id={compact ? "forestSky" : "heroSky"}
          x1="0"
          y1="0"
          x2="0"
          y2="1"
        >
          <stop offset="0" stopColor={compact ? "#c7f0df" : "#a9eadf"} />
          <stop offset="1" stopColor={compact ? "#e6f8d7" : "#dcf7ce"} />
        </linearGradient>
      </defs>
      <rect
        width="600"
        height="300"
        fill={`url(#${compact ? "forestSky" : "heroSky"})`}
      />
      {compact ? (
        <>
          <path
            d="M0 250 L85 128 L150 240 L245 145 L320 250 L410 115 L500 235 L570 150 L600 210 V300 H0Z"
            fill="#e8f8d2"
          />
          <path
            d="M0 260 Q55 210 105 252 T200 250 T300 260 T405 245 T515 252 T600 232 V300 H0Z"
            fill="#77b85b"
          />
          {treeXs.map((x, index) => (
            <g
              key={x}
              transform={`translate(${x * 6 - 24} ${index % 2 ? 86 : 116})`}
            >
              <rect
                x="19"
                y="135"
                width="7"
                height="46"
                rx="3"
                fill="#426651"
              />
              <path
                d="M22 0 L0 58 H14 L-3 103 H16 L-6 150 H52 L29 103 H47 L31 58 H44Z"
                fill={index % 3 === 0 ? "#194b43" : "#245e51"}
              />
            </g>
          ))}
          <circle cx="38" cy="274" r="40" fill="#5d9e4f" />
          <circle cx="78" cy="280" r="43" fill="#a6cf55" />
          <circle cx="550" cy="279" r="48" fill="#94c84e" />
        </>
      ) : (
        <>
          <path
            d="M0 119 Q56 78 112 118 T230 118 T350 112 T470 117 T600 91 V0 H0Z"
            fill="rgba(255,255,255,.76)"
            transform="translate(0 82)"
          />
          <path
            d="M0 218 Q80 160 165 199 T320 208 T450 184 T600 207 V300 H0Z"
            fill="#2b5b59"
          />
          <path
            d="M0 242 Q83 189 165 222 T330 239 T480 213 T600 226 V300 H0Z"
            fill="#7bb991"
          />
          <path
            d="M0 265 Q102 235 196 258 T402 251 T600 260 V300 H0Z"
            fill="#bcdfaa"
          />
          <g transform="translate(58 133)">
            <rect x="22" y="88" width="6" height="69" fill="#976a48" />
            <path
              d="M25 0 C-7 42 -5 100 25 106 C56 99 54 42 25 0Z"
              fill="#a4d477"
            />
          </g>
          <g transform="translate(152 88)">
            <rect x="31" y="121" width="7" height="72" fill="#805d45" />
            <path
              d="M35 0 C-10 54 -10 132 35 142 C81 132 80 55 35 0Z"
              fill="#174e49"
            />
          </g>
          <g
            className="birds"
            fill="none"
            stroke="#183e45"
            strokeWidth="4"
            strokeLinecap="round"
          >
            <path d="M386 66 l8 8 l9 -8" />
            <path d="M444 28 l5 5 l6 -5" />
            <path d="M504 81 l8 8 l9 -8" />
          </g>
        </>
      )}
    </svg>
  )
}

function Sidebar() {
  const [active, setActive] = useState("Dashboard")
  const primary = [
    { label: "Dashboard", icon: House },
    { label: "List", icon: ListTodo },
    { label: "Calendar", icon: CalendarDays },
    { label: "Messages", icon: Mail },
  ]
  const lists = [
    { label: "Inbox", icon: Inbox, color: "blue" },
    { label: "Work", icon: BriefcaseBusiness, color: "green" },
    { label: "Study", icon: Timer, color: "lime" },
    { label: "Health", icon: Heart, color: "orange" },
    { label: "Holiday trip", icon: Plane, color: "pink" },
  ]

  return (
    <aside className="sidebar">
      <div className="brand">
        <BrandMark />
        <strong>Mindora</strong>
      </div>

      <nav className="nav-card" aria-label="Main navigation">
        {primary.map(({ label, icon: Icon }) => (
          <button
            className={active === label ? "nav-item active" : "nav-item"}
            onClick={() => setActive(label)}
            key={label}
          >
            <Icon size={18} strokeWidth={1.8} />
            <span>{label}</span>
          </button>
        ))}
      </nav>

      <div className="section-label">
        <span>My list</span>
        <button aria-label="Add list">
          <Plus size={17} />
        </button>
      </div>
      <nav className="nav-card list-nav" aria-label="My lists">
        {lists.map(({ label, icon: Icon, color }) => (
          <button className="nav-item" key={label}>
            <Icon size={17} strokeWidth={1.7} />
            <span>{label}</span>
            <i className={`status-dot ${color}`} />
          </button>
        ))}
      </nav>

      <div className="section-label workspace-label">
        <span>Workspace</span>
        <button aria-label="Add workspace">
          <Plus size={17} />
        </button>
      </div>
      <div className="workspace-card">
        <button>
          <span className="workspace-icon lime">
            <Palette size={16} />
          </span>
          UI Design
        </button>
        <button>
          <span className="workspace-icon mint">#</span>A-team Class
        </button>
      </div>
    </aside>
  )
}

function Header({
  onOpenAdmin,
  onLogout,
}: {
  onOpenAdmin: () => void
  onLogout: () => void
}) {
  return (
    <header className="topbar">
      <label className="search-box">
        <Search size={19} />
        <input aria-label="Search task" placeholder="Search task" />
      </label>
      <button className="export-button">
        <Download size={17} />
        <span>Export summary</span>
        <ChevronDown size={14} />
      </button>
      <button className="circle-button" aria-label="Timer">
        <Clock3 size={18} />
      </button>
      <button className="circle-button notification" aria-label="Notifications">
        <Bell size={18} />
        <i />
      </button>
      <button className="circle-button" onClick={onLogout} aria-label="Log out">
        <LogOut size={17} />
      </button>
      <button
        className="avatar"
        onClick={onOpenAdmin}
        aria-label="Open admin dashboard"
      >
        <span>J</span>
      </button>
    </header>
  )
}

function HeroCard() {
  return (
    <section className="hero-card">
      <LandscapeIllustration />
      <div className="hero-copy">
        <h1>Good Morning, John</h1>
        <p>You have 6 Task for today</p>
      </div>
      <div className="up-next">
        <strong>Up next</strong>
        <div className="meeting green">
          <p>Design Team Meeting - Discuss new project wireframes</p>
          <small>11:30 AM - 01:00 PM</small>
        </div>
        <div className="meeting blue">
          <p>Client Meeting - Review project updates and seek approval</p>
          <small>03:00 PM - 05:00 PM</small>
        </div>
      </div>
    </section>
  )
}

function CalendarCard() {
  return (
    <section className="calendar-card">
      <div className="calendar-heading">
        <button aria-label="Previous month">
          <ChevronLeft size={18} />
        </button>
        <div>
          <h2>December</h2>
          <span>2023</span>
        </div>
        <button aria-label="Next month">
          <ChevronRight size={18} />
        </button>
      </div>
      <div className="weekdays">
        {["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"].map((day) => (
          <span key={day}>{day}</span>
        ))}
      </div>
      <div className="calendar-grid">
        {calendarRows.flat().map((day, index) => {
          const faded = index < 2 || index > 32
          return (
            <button
              className={`${day === "2" && index === 3 ? "selected " : ""}${
                faded ? "faded" : ""
              }`}
              key={`${day}-${index}`}
            >
              <span>{day}</span>
              {!faded && calendarDots[day] && (
                <i className="calendar-dots">
                  {calendarDots[day].map((color, dot) => (
                    <b className={color} key={`${color}-${dot}`} />
                  ))}
                </i>
              )}
            </button>
          )
        })}
      </div>
    </section>
  )
}

function TaskComposer({ onAdd }: { onAdd: (title: string) => void }) {
  const [value, setValue] = useState("")

  const submit = () => {
    const title = value.trim()
    if (!title) return
    onAdd(title)
    setValue("")
  }

  return (
    <section className="task-composer">
      <div className="composer-input">
        <CirclePlus size={19} />
        <input
          value={value}
          onChange={(event) => setValue(event.target.value)}
          onKeyDown={(event) => event.key === "Enter" && submit()}
          placeholder="Add task for today"
          aria-label="Add task for today"
        />
        <button onClick={submit} aria-label="Submit task">
          <Send size={19} />
        </button>
        <button className="quick-add" onClick={submit} aria-label="Quick add">
          <Plus size={17} />
        </button>
      </div>
      <div className="task-filters">
        <button>
          <i className="status-dot blue" />
          Inbox
          <ChevronDown size={13} />
        </button>
        <button>
          <Clock3 size={16} />
          Now
        </button>
        <button>
          <Clock3 size={16} />
          Tomorrow
        </button>
        <button>
          <Clock3 size={16} />
          Next week
        </button>
        <button>
          <CalendarDays size={16} />
          Custom
        </button>
      </div>
    </section>
  )
}

type Task = typeof initialTasks[number]

function TaskList({
  tasks,
  toggleTask,
}: {
  tasks: Task[]
  toggleTask: (id: number) => void
}) {
  return (
    <section className="tasks-section">
      <h2>Today&apos;s task</h2>
      <div className="tasks-card">
        {tasks.map((task) => (
          <button
            className={task.done ? "task-row done" : "task-row"}
            onClick={() => toggleTask(task.id)}
            key={task.id}
          >
            <span className="task-check">{task.done && "✓"}</span>
            <span className="task-copy">
              <strong>{task.title}</strong>
              <small>
                <i className={`status-dot ${task.color}`} />
                {task.list}
              </small>
            </span>
            <time>{task.time}</time>
            <ChevronRight size={16} />
          </button>
        ))}
      </div>
    </section>
  )
}

function FocusTimer() {
  const [minutes, setMinutes] = useState(30)
  const [running, setRunning] = useState(false)

  return (
    <section className="focus-card">
      <LandscapeIllustration compact />
      <div className="focus-content">
        <h2>Focus timer</h2>
        <button className="focus-task">
          Cleaning up desk <ChevronDown size={12} />
        </button>
        <div className="timer-control">
          <button
            onClick={() => setMinutes((value) => Math.max(5, value - 5))}
            aria-label="Reduce timer"
          >
            <Minus size={18} />
          </button>
          <strong>{minutes} mins</strong>
          <button
            onClick={() => setMinutes((value) => value + 5)}
            aria-label="Increase timer"
          >
            <Plus size={18} />
          </button>
        </div>
        <button
          className={running ? "start-button running" : "start-button"}
          onClick={() => setRunning((value) => !value)}
        >
          <span>{running ? "Ⅱ" : "▶"}</span>
          {running ? "Pause" : "Start"}
        </button>
      </div>
    </section>
  )
}

type ChatMessage = {
  id: number
  role: "assistant" | "user"
  text: string
}

function AiChat() {
  const [value, setValue] = useState("")
  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      id: 1,
      role: "assistant",
      text: "Hi John! I can help you plan your day, break down tasks, or create a focus schedule.",
    },
  ])

  const sendMessage = (prompt = value) => {
    const text = prompt.trim()
    if (!text) return

    setMessages((current) => [
      ...current,
      { id: Date.now(), role: "user", text },
      {
        id: Date.now() + 1,
        role: "assistant",
        text: "Your LLM response will appear here once a model is connected. The chat interface is ready for integration.",
      },
    ])
    setValue("")
  }

  return (
    <section className="ai-assistant-card" aria-label="Timey AI assistant">
      <div className="ai-assistant-orb">
        <WandSparkles size={23} />
      </div>
      <div className="ai-assistant-main">
        <header className="ai-assistant-heading">
          <div>
            <span>
              <i /> AI PLANNING ASSISTANT
            </span>
            <h2>What can I help you finish today?</h2>
          </div>
          <small>Powered by your preferred LLM</small>
        </header>

        {messages.length > 1 && (
          <div className="ai-latest-message">
            <span>
              <Bot size={15} />
            </span>
            <p>{messages[messages.length - 1].text}</p>
          </div>
        )}

        <div className="ai-prompt-row">
          {["Plan my day", "Prioritize tasks", "Create focus plan"].map(
            (prompt) => (
              <button onClick={() => sendMessage(prompt)} key={prompt}>
                {prompt}
              </button>
            ),
          )}
        </div>

        <div className="ai-assistant-input">
          <input
            value={value}
            onChange={(event) => setValue(event.target.value)}
            onKeyDown={(event) => event.key === "Enter" && sendMessage()}
            placeholder="Ask Timey AI about your tasks..."
          />
          <button onClick={() => sendMessage()} aria-label="Send message">
            <Send size={17} />
          </button>
        </div>
      </div>
    </section>
  )
}

type AppView = "login" | "user" | "admin"

function LoginPage({ onLogin }: { onLogin: (view: AppView) => void }) {
  const [role, setRole] = useState<"user" | "admin">("user")
  const [showPassword, setShowPassword] = useState(false)

  return (
    <main className="login-page">
      <section className="login-shell">
        <div className="login-visual">
          <div className="login-brand">
            <BrandMark />
            <strong>Mindora</strong>
          </div>
          <div className="login-visual-copy">
            <span>PLAN · FOCUS · ACHIEVE</span>
            <h1>Make every minute feel meaningful.</h1>
            <p>
              Organize your day, protect your focus, and move your best work
              forward with one calm workspace.
            </p>
          </div>
          <div className="login-art">
            <LandscapeIllustration compact />
          </div>
          <div className="login-quote">
            <span className="member-avatar violet">AK</span>
            <p>
              <strong>
                You don’t have to be perfect every day. Even one small step
                today will bring you closer to your goal.
              </strong>
              <small>Alicia · Product Designer</small>
            </p>
          </div>
        </div>

        <div className="login-form-side">
          <div className="login-mobile-brand">
            <BrandMark />
            <strong>Timeyfi</strong>
          </div>
          <form
            className="login-form"
            onSubmit={(event) => {
              event.preventDefault()
              onLogin(role)
            }}
          >
            <div className="login-heading">
              <span>WELCOME BACK</span>
              <h2>Sign in to Mindora</h2>
              <p>Choose your workspace and continue where you left off.</p>
            </div>

            <div className="role-selector">
              <button
                className={role === "user" ? "active" : ""}
                type="button"
                onClick={() => setRole("user")}
              >
                <Users size={16} />
                User workspace
              </button>
              <button
                className={role === "admin" ? "active" : ""}
                type="button"
                onClick={() => setRole("admin")}
              >
                <ShieldCheck size={16} />
                Admin console
              </button>
            </div>

            <label className="login-field">
              <span>Email address</span>
              <div>
                <Mail size={17} />
                <input
                  type="email"
                  defaultValue="john@timeyfi.io"
                  placeholder="you@company.com"
                  required
                />
              </div>
            </label>

            <label className="login-field">
              <span>Password</span>
              <div>
                <LockKeyhole size={17} />
                <input
                  type={showPassword ? "text" : "password"}
                  defaultValue="timeyfi123"
                  placeholder="Enter your password"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((current) => !current)}
                  aria-label={showPassword ? "Hide password" : "Show password"}
                >
                  {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
            </label>

            <div className="login-options">
              <label>
                <input type="checkbox" defaultChecked />
                <span>Remember me</span>
              </label>
              <button type="button">Forgot password?</button>
            </div>

            <button className="login-submit" type="submit">
              Sign in to {role === "admin" ? "admin console" : "workspace"}
              <ChevronRight size={17} />
            </button>

            <p className="login-signup">
              New to Timeyfi? <button type="button">Create an account</button>
            </p>
          </form>
          <small className="login-legal">
            Frontend preview · Authentication will be enabled after the backend
            is connected.
          </small>
        </div>
      </section>
    </main>
  )
}

const adminStats = [
  {
    label: "Total members",
    value: "248",
    change: "+12.5%",
    icon: Users,
    tone: "mint",
  },
  {
    label: "Active tasks",
    value: "1,429",
    change: "+8.2%",
    icon: ListTodo,
    tone: "blue",
  },
  {
    label: "Completion rate",
    value: "84.6%",
    change: "+4.1%",
    icon: BarChart3,
    tone: "lime",
  },
  {
    label: "Active workspaces",
    value: "16",
    change: "+2 new",
    icon: BriefcaseBusiness,
    tone: "orange",
  },
]

const teamMembers = [
  {
    initials: "AK",
    name: "Alicia Keys",
    email: "alicia@timeyfi.io",
    role: "Admin",
    tasks: 42,
    status: "Active",
    color: "violet",
  },
  {
    initials: "BM",
    name: "Brian Miller",
    email: "brian@timeyfi.io",
    role: "Manager",
    tasks: 36,
    status: "Active",
    color: "blue",
  },
  {
    initials: "CS",
    name: "Clara Smith",
    email: "clara@timeyfi.io",
    role: "Member",
    tasks: 28,
    status: "Active",
    color: "green",
  },
  {
    initials: "DJ",
    name: "Daniel Jones",
    email: "daniel@timeyfi.io",
    role: "Member",
    tasks: 19,
    status: "Away",
    color: "orange",
  },
]

function AdminSidebar({
  onBack,
  onLogout,
}: {
  onBack: () => void
  onLogout: () => void
}) {
  const [active, setActive] = useState("Overview")
  const navigation = [
    { label: "Overview", icon: LayoutDashboard },
    { label: "Members", icon: Users },
    { label: "Workspaces", icon: BriefcaseBusiness },
    { label: "Reports", icon: BarChart3 },
  ]

  return (
    <aside className="admin-sidebar">
      <div className="admin-brand">
        <BrandMark />
        <div>
          <strong>Mindora</strong>
          <small>Admin console</small>
        </div>
      </div>

      <span className="admin-nav-label">Workspace</span>
      <nav className="admin-nav" aria-label="Admin navigation">
        {navigation.map(({ label, icon: Icon }) => (
          <button
            className={active === label ? "active" : ""}
            onClick={() => setActive(label)}
            key={label}
          >
            <Icon size={18} />
            <span>{label}</span>
          </button>
        ))}
      </nav>

      <span className="admin-nav-label">Management</span>
      <nav className="admin-nav">
        <button>
          <ShieldCheck size={18} />
          <span>Roles & access</span>
        </button>
        <button>
          <Settings size={18} />
          <span>Settings</span>
        </button>
      </nav>

      <div className="admin-profile">
        <span className="admin-profile-avatar">JD</span>
        <div>
          <strong>John Doe</strong>
          <small>Super admin</small>
        </div>
        <button onClick={onLogout} aria-label="Sign out">
          <LogOut size={17} />
        </button>
      </div>

      <button className="back-to-app" onClick={onBack}>
        <ChevronLeft size={17} /> Back to main app
      </button>
    </aside>
  )
}

function ProductivityChart() {
  const data = [
    { day: "Mon", completed: 64, created: 40 },
    { day: "Tue", completed: 76, created: 52 },
    { day: "Wed", completed: 55, created: 38 },
    { day: "Thu", completed: 88, created: 60 },
    { day: "Fri", completed: 72, created: 48 },
    { day: "Sat", completed: 45, created: 30 },
    { day: "Sun", completed: 58, created: 35 },
  ]

  return (
    <section className="admin-card productivity-card">
      <div className="admin-card-heading">
        <div>
          <h2>Task productivity</h2>
          <p>Tasks created and completed this week</p>
        </div>
        <div className="chart-legend">
          <span>
            <i className="completed" /> Completed
          </span>
          <span>
            <i className="created" /> Created
          </span>
        </div>
      </div>
      <div className="chart-area">
        <div className="chart-lines" aria-hidden="true">
          <i />
          <i />
          <i />
          <i />
        </div>
        <div className="bar-chart">
          {data.map((item) => (
            <div className="bar-group" key={item.day}>
              <div className="bars">
                <i
                  className="completed"
                  style={{ height: `${item.completed}%` }}
                />
                <i className="created" style={{ height: `${item.created}%` }} />
              </div>
              <span>{item.day}</span>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}

function AdminDashboard({
  onBack,
  onLogout,
}: {
  onBack: () => void
  onLogout: () => void
}) {
  const [period, setPeriod] = useState("This week")

  return (
    <main className="admin-app-shell">
      <div className="admin-dashboard">
        <AdminSidebar onBack={onBack} onLogout={onLogout} />
        <div className="admin-main">
          <header className="admin-topbar">
            <label>
              <Search size={18} />
              <input placeholder="Search members, tasks, workspace..." />
            </label>
            <button
              className="circle-button admin-back-button"
              onClick={onBack}
              aria-label="Open user dashboard"
            >
              <House size={17} />
            </button>
            <button className="circle-button notification">
              <Bell size={18} />
              <i />
            </button>
            <button
              className="circle-button admin-logout-button"
              onClick={onLogout}
              aria-label="Log out"
            >
              <LogOut size={17} />
            </button>
          </header>

          <div className="admin-content">
            <div className="admin-title-row">
              <div>
                <p>ADMIN OVERVIEW</p>
                <h1>Good morning, John</h1>
                <span>Here&apos;s what&apos;s happening across your team.</span>
              </div>
              <div className="period-picker">
                <CalendarDays size={16} />
                <select
                  value={period}
                  onChange={(event) => setPeriod(event.target.value)}
                >
                  <option>This week</option>
                  <option>This month</option>
                  <option>This quarter</option>
                </select>
                <ChevronDown size={14} />
              </div>
            </div>

            <section className="admin-stats">
              {adminStats.map(({ label, value, change, icon: Icon, tone }) => (
                <article className="admin-stat-card" key={label}>
                  <div className={`admin-stat-icon ${tone}`}>
                    <Icon size={20} />
                  </div>
                  <span>{label}</span>
                  <div>
                    <strong>{value}</strong>
                    <small>
                      <ArrowUpRight size={13} /> {change}
                    </small>
                  </div>
                </article>
              ))}
            </section>

            <div className="admin-overview-grid">
              <ProductivityChart />
              <section className="admin-card activity-card">
                <div className="admin-card-heading">
                  <div>
                    <h2>Recent activity</h2>
                    <p>Latest team updates</p>
                  </div>
                  <button>
                    <MoreHorizontal size={19} />
                  </button>
                </div>
                <div className="activity-list">
                  <article>
                    <span className="activity-icon green">
                      <UserPlus size={15} />
                    </span>
                    <p>
                      <strong>Clara joined</strong>
                      <small>UI Design workspace · 8m</small>
                    </p>
                  </article>
                  <article>
                    <span className="activity-icon blue">
                      <ListTodo size={15} />
                    </span>
                    <p>
                      <strong>12 tasks completed</strong>
                      <small>Product Team · 24m</small>
                    </p>
                  </article>
                  <article>
                    <span className="activity-icon orange">
                      <BriefcaseBusiness size={15} />
                    </span>
                    <p>
                      <strong>New workspace created</strong>
                      <small>Marketing 2024 · 1h</small>
                    </p>
                  </article>
                  <article>
                    <span className="activity-icon violet">
                      <ShieldCheck size={15} />
                    </span>
                    <p>
                      <strong>Role permissions updated</strong>
                      <small>Manager access · 2h</small>
                    </p>
                  </article>
                </div>
                <button className="view-activity">View all activity</button>
              </section>
            </div>

            <section className="admin-card members-card">
              <div className="admin-card-heading">
                <div>
                  <h2>Team members</h2>
                  <p>Manage people and workspace access</p>
                </div>
                <button className="view-all-button">
                  View all <ChevronRight size={15} />
                </button>
              </div>
              <div className="members-table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>Member</th>
                      <th>Role</th>
                      <th>Tasks completed</th>
                      <th>Status</th>
                      <th />
                    </tr>
                  </thead>
                  <tbody>
                    {teamMembers.map((member) => (
                      <tr key={member.email}>
                        <td>
                          <span className={`member-avatar ${member.color}`}>
                            {member.initials}
                          </span>
                          <p>
                            <strong>{member.name}</strong>
                            <small>{member.email}</small>
                          </p>
                        </td>
                        <td>
                          <span className="role-badge">{member.role}</span>
                        </td>
                        <td>{member.tasks}</td>
                        <td>
                          <span
                            className={`member-status ${member.status.toLowerCase()}`}
                          >
                            <i /> {member.status}
                          </span>
                        </td>
                        <td>
                          <button aria-label={`Options for ${member.name}`}>
                            <MoreHorizontal size={18} />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>
          </div>
        </div>
      </div>
    </main>
  )
}

export default function App() {
  const [view, setView] = useState<AppView>("login")
  const [tasks, setTasks] = useState(initialTasks)
  const completed = useMemo(
    () => tasks.filter((task) => task.done).length,
    [tasks],
  )

  const addTask = (title: string) => {
    setTasks((current) => [
      ...current,
      {
        id: Date.now(),
        title,
        list: "Inbox",
        time: "Today",
        color: "blue",
        done: false,
      },
    ])
  }

  const toggleTask = (id: number) => {
    setTasks((current) =>
      current.map((task) =>
        task.id === id ? { ...task, done: !task.done } : task,
      ),
    )
  }

  if (view === "login") {
    return <LoginPage onLogin={setView} />
  }

  if (view === "admin") {
    return (
      <AdminDashboard
        onBack={() => setView("user")}
        onLogout={() => setView("login")}
      />
    )
  }

  return (
    <main className="app-shell">
      <div className="dashboard">
        <Sidebar />
        <div className="main-area">
          <Header
            onOpenAdmin={() => setView("admin")}
            onLogout={() => setView("login")}
          />
          <div className="dashboard-grid">
            <div className="center-column">
              <HeroCard />
              <AiChat />
              <TaskComposer onAdd={addTask} />
              <TaskList tasks={tasks} toggleTask={toggleTask} />
            </div>
            <div className="right-column">
              <CalendarCard />
              <FocusTimer />
            </div>
          </div>
          {completed > 0 && (
            <div className="completion-toast">
              <Sparkles size={15} /> {completed} task completed
            </div>
          )}
        </div>
      </div>
    </main>
  )
}
