"use client";

import Image from "next/image";
import { FormEvent, useEffect, useRef, useState } from "react";
import { Locale, messages } from "@/lib/i18n";
import { AdminManagement } from "@/components/admin-management";
import { PatientManagement } from "@/components/patient-management";
import { AppointmentManagement } from "@/components/appointment-management";
import { DoctorScheduleManagement } from "@/components/doctor-schedule-management";
import { MessageCenter } from "@/components/message-center";
import { MessageNotifier } from "@/components/message-notifier";
import { MessageArchive } from "@/components/message-archive";
import { ServiceManagement } from "@/components/service-management";
import { SpecialistManagement } from "@/components/specialist-management";
import { InpatientManagement } from "@/components/inpatient-management";
import { ReportManagement } from "@/components/report-management";
import { DiagnosisCatalogManagement } from "@/components/diagnosis-catalog-management";
import { LaboratoryManagement } from "@/components/laboratory-management";
import { InventoryManagement } from "@/components/inventory-management";
import { CashDeskManagement } from "@/components/cash-management";
import { AccountingManagement } from "@/components/accounting-management";
import { DocumentTemplateBuilder } from "@/components/document-template-builder";
import { QueueManagement } from "@/components/queue-management";

const Icon = ({
  name,
}: {
  name:
    | "home"
    | "calendar"
    | "users"
    | "doctor"
    | "chart"
    | "settings"
    | "building"
    | "bell"
    | "plus"
    | "search"
    | "chevron";
}) => {
  const paths = {
    home: (
      <>
        <path d="M3 10.5 12 3l9 7.5" />
        <path d="M5.5 9.5V21h13V9.5M9 21v-7h6v7" />
      </>
    ),
    calendar: (
      <>
        <rect x="3" y="5" width="18" height="16" rx="3" />
        <path d="M8 3v4m8-4v4M3 10h18" />
      </>
    ),
    users: (
      <>
        <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" />
        <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
      </>
    ),
    doctor: (
      <>
        <circle cx="12" cy="7" r="4" />
        <path d="M5 21a7 7 0 0 1 14 0M9 7h6M12 4v6" />
      </>
    ),
    chart: (
      <>
        <path d="M4 20V10m6 10V4m6 16v-7m5 7H2" />
      </>
    ),
    settings: (
      <>
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.7 1.7 0 0 0 .34 1.88l.06.06-2.83 2.83-.06-.06A1.7 1.7 0 0 0 15 19.4a1.7 1.7 0 0 0-1 .6 1.7 1.7 0 0 0-.4 1.1V21h-4v-.1A1.7 1.7 0 0 0 8.6 19.4a1.7 1.7 0 0 0-1.88.34l-.06.06-2.83-2.83.06-.06A1.7 1.7 0 0 0 4.6 15a1.7 1.7 0 0 0-.6-1 1.7 1.7 0 0 0-1.1-.4H3v-4h.1A1.7 1.7 0 0 0 4.6 8.6a1.7 1.7 0 0 0-.34-1.88l-.06-.06 2.83-2.83.06.06A1.7 1.7 0 0 0 9 4.6a1.7 1.7 0 0 0 1-.6 1.7 1.7 0 0 0 .4-1.1V3h4v.1A1.7 1.7 0 0 0 15.4 4.6a1.7 1.7 0 0 0 1.88-.34l.06-.06 2.83 2.83-.06.06A1.7 1.7 0 0 0 19.4 9c.15.36.2.75.2 1.1V10h1.4v4h-.1a1.7 1.7 0 0 0-1.5 1Z" />
      </>
    ),
    building: (
      <>
        <path d="M4 21V7l8-4 8 4v14M2 21h20M8 9h2m4 0h2M8 13h2m4 0h2M10 21v-4h4v4" />
      </>
    ),
    bell: (
      <>
        <path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4" />
      </>
    ),
    plus: <path d="M12 5v14M5 12h14" />,
    search: (
      <>
        <circle cx="11" cy="11" r="7" />
        <path d="m20 20-4-4" />
      </>
    ),
    chevron: <path d="m9 18 6-6-6-6" />,
  };
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      {paths[name]}
    </svg>
  );
};

type Branch = {
  id: string;
  name: string;
  address: string | null;
  timezone: string;
  isActive: boolean;
};
type Employee = {
  id: string;
  login: string;
  firstName: string;
  lastName: string;
  position: string;
  branch: string | null;
  isActive: boolean;
};
type DashboardAppointment = {
  id: string;
  startsAt: string;
  status: string;
  reason?: string | null;
  patient: string;
  phone: string;
  secondPhone?: string | null;
  telegram?: string | null;
  employee: string;
  employeeUserId: string;
  employeePhone?: string | null;
  employeeEmail?: string | null;
  employeeTelegram?: string | null;
  position: string;
  specialty?: string | null;
  branch: string;
};
type DashboardData = {
  today: number;
  waiting: number;
  doctorsOnDuty: number;
  completed: number;
  upcoming: DashboardAppointment[];
};
// Legacy form kept temporarily while employee and branch management use dedicated screens.
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function Management({
  kind,
  t,
}: {
  kind: "branches" | "employees";
  t: (typeof messages)[Locale];
}) {
  const [items, setItems] = useState<Branch[] | Employee[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);
  const load = async () => {
    setLoading(true);
    const r = await fetch(`/api/clinic/${kind}`, { cache: "no-store" });
    if (!r.ok) {
      setError(t.loadError);
      setLoading(false);
      return;
    }
    const data = await r.json();
    setItems(data.items ?? []);
    setLoading(false);
  };
  useEffect(() => {
    void load();
  }, [kind]);
  const create = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    const f = new FormData(e.currentTarget);
    const r = await fetch("/api/clinic/branches", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        name: f.get("name"),
        address: f.get("address"),
        timezone: "Asia/Tashkent",
      }),
    });
    if (!r.ok) {
      setError(t.branchCreateError);
      return;
    }
    setShowForm(false);
    e.currentTarget.reset();
    await load();
  };
  return (
    <section className="management">
      <div className="management-head">
        <div>
          <p>{t.settings}</p>
          <h1>{kind === "branches" ? t.branches : t.employees}</h1>
          <span>
            {kind === "branches" ? t.branchesSubtitle : t.employeesSubtitle}
          </span>
        </div>
        {kind === "branches" && (
          <button className="primary" onClick={() => setShowForm((v) => !v)}>
            <Icon name="plus" />
            {t.addBranch}
          </button>
        )}
      </div>
      {showForm && (
        <form className="inline-form" onSubmit={create}>
          <label>
            {t.branchName}
            <input name="name" required minLength={2} />
          </label>
          <label>
            {t.address}
            <input name="address" />
          </label>
          <button className="primary">{t.save}</button>
        </form>
      )}
      {error && <div className="management-error">{error}</div>}
      <div className="management-panel">
        {loading ? (
          <div className="table-empty">{t.loading}</div>
        ) : items.length === 0 ? (
          <div className="table-empty">{t.noItems}</div>
        ) : (
          <div className="data-table">
            <div className="table-row table-head">
              <span>{t.name}</span>
              <span>{kind === "branches" ? t.address : t.position}</span>
              <span>{kind === "branches" ? t.timezone : t.branch}</span>
              <span>{t.status}</span>
            </div>
            {kind === "branches"
              ? (items as Branch[]).map((x) => (
                  <div className="table-row" key={x.id}>
                    <strong>{x.name}</strong>
                    <span>{x.address ?? "—"}</span>
                    <span>{x.timezone}</span>
                    <span className="status-active">
                      {x.isActive ? t.active : t.inactive}
                    </span>
                  </div>
                ))
              : (items as Employee[]).map((x) => (
                  <div className="table-row" key={x.id}>
                    <strong>
                      {x.firstName} {x.lastName}
                      <small>@{x.login}</small>
                    </strong>
                    <span>{x.position}</span>
                    <span>{x.branch ?? "—"}</span>
                    <span className="status-active">
                      {x.isActive ? t.active : t.inactive}
                    </span>
                  </div>
                ))}
          </div>
        )}
      </div>
    </section>
  );
}

export default function Dashboard() {
  const [locale, setLocale] = useState<Locale>("ru");
  const [dark, setDark] = useState(false);
  const [session, setSession] = useState<"loading" | "guest" | "authenticated">(
    "loading",
  );
  const [loginError, setLoginError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [unreadMessages, setUnreadMessages] = useState(0);
  const [isOwner, setIsOwner] = useState(false);
  const [dashboard, setDashboard] = useState<DashboardData>({
    today: 0,
    waiting: 0,
    doctorsOnDuty: 0,
    completed: 0,
    upcoming: [],
  });
  const [dashboardLoading, setDashboardLoading] = useState(true);
  const [section, setSection] = useState<
    | "overview"
    | "appointments"
    | "schedules"
    | "branches"
    | "employees"
    | "roles"
    | "patients"
    | "messages"
    | "messageArchive"
    | "services"
    | "specialists"
    | "diagnosisCatalog"
    | "laboratory"
    | "inventory"
    | "cash"
    | "accounting"
    | "documentBuilder"
    | "inpatient"
    | "reports"
    | "queue"
  >("overview");
  const sectionReady = useRef(false);
  const t = messages[locale];
  const dateLocale = locale === "uz" ? "uz-UZ" : locale;

  useEffect(() => {
    const allowed = new Set(["overview", "appointments", "schedules", "branches", "employees", "roles", "patients", "messages", "messageArchive", "services", "specialists", "diagnosisCatalog", "laboratory", "inventory", "cash", "accounting", "documentBuilder", "inpatient", "reports", "queue"]);
    const restore = () => {
      const value = window.location.hash.slice(1);
      if (allowed.has(value)) setSection(value as typeof section);
    };
    restore();
    sectionReady.current = true;
    window.addEventListener("hashchange", restore);
    return () => window.removeEventListener("hashchange", restore);
  }, []);

  useEffect(() => {
    if (!sectionReady.current) return;
    const nextHash = `#${section}`;
    if (window.location.hash !== nextHash) window.history.replaceState(null, "", `${window.location.pathname}${window.location.search}${nextHash}`);
  }, [section]);

  useEffect(() => {
    if (session === "authenticated" && section === "messageArchive" && !isOwner) setSection("overview");
  }, [session, section, isOwner]);

  useEffect(() => {
    const saved = localStorage.getItem("ona-va-bola-theme");
    const prefersDark = window.matchMedia(
      "(prefers-color-scheme: dark)",
    ).matches;
    setDark(saved ? saved === "dark" : prefersDark);
  }, []);

  useEffect(() => {
    fetch("/api/session/me", { cache: "no-store" })
      .then(async (r) => { if(!r.ok){setSession("guest");return} const data=await r.json();setIsOwner(Boolean(data.user?.permissions?.["*"]));setSession("authenticated") })
      .catch(() => setSession("guest"));
  }, []);

  useEffect(() => {
    document.documentElement.dataset.theme = dark ? "dark" : "light";
    document.documentElement.style.colorScheme = dark ? "dark" : "light";
  }, [dark]);

  useEffect(() => {
    if (session !== "authenticated" || section !== "overview") return;
    const loadDashboard = async () => {
      const response = await fetch("/api/clinic/appointments/dashboard", {
        cache: "no-store",
      });
      if (response.ok) setDashboard(await response.json());
      setDashboardLoading(false);
    };
    void loadDashboard();
    const timer = window.setInterval(() => void loadDashboard(), 5000);
    return () => window.clearInterval(timer);
  }, [session, section]);

  const toggleTheme = () => {
    const next = !dark;
    setDark(next);
    localStorage.setItem("ona-va-bola-theme", next ? "dark" : "light");
  };

  const login = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSubmitting(true);
    setLoginError("");
    const form = new FormData(event.currentTarget);
    const response = await fetch("/api/session/login", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        login: form.get("login"),
        password: form.get("password"),
      }),
    }).catch(() => null);
    setSubmitting(false);
    if (!response?.ok) {
      setLoginError(t.loginError);
      return;
    }
    const me = await fetch("/api/session/me", { cache: "no-store" });
    if (me.ok) {
      const data = await me.json();
      setIsOwner(Boolean(data.user?.permissions?.["*"]));
    }
    setSession("authenticated");
  };
  const logout = async () => {
    await fetch("/api/session/logout", { method: "POST" });
    setSession("guest");
  };

  if (session === "loading")
    return (
      <div className="session-splash">
        <Image src="/brand-mark.png" alt="" width={54} height={54} />
        <span>ONA VA BOLA KLINIKASI</span>
      </div>
    );
  if (session === "guest")
    return (
      <main className="login-page">
        <section className="login-brand">
          <div className="login-brand-content">
            <div className="login-logo">
              <Image src="/brand-mark.png" alt="" width={72} height={72} />
            </div>
            <p>ONA VA BOLA</p>
            <h1>KLINIKASI</h1>
            <span>{t.loginPromise}</span>
          </div>
          <div className="brand-orb one" />
          <div className="brand-orb two" />
        </section>
        <section className="login-card-wrap">
          <div className="login-controls">
            <select
              value={locale}
              onChange={(e) => setLocale(e.target.value as Locale)}
            >
              <option value="ru">RU</option>
              <option value="uz">UZ</option>
              <option value="en">EN</option>
            </select>
            <button
              onClick={toggleTheme}
              className="icon-button"
              aria-label={dark ? t.lightTheme : t.darkTheme}
            >
              {dark ? "☀" : "☾"}
            </button>
          </div>
          <form className="login-card" onSubmit={login}>
            <span className="login-kicker">ONA VA BOLA KLINIKASI</span>
            <h2>{t.signIn}</h2>
            <p>{t.signInSubtitle}</p>
            <label>
              {t.loginLabel}
              <input
                name="login"
                autoComplete="username"
                required
                placeholder="owner"
              />
            </label>
            <label>
              {t.passwordLabel}
              <input
                name="password"
                type="password"
                autoComplete="current-password"
                required
                minLength={8}
                placeholder="••••••••••••"
              />
            </label>
            {loginError && (
              <div className="login-error" role="alert">
                {loginError}
              </div>
            )}
            <button className="primary login-submit" disabled={submitting}>
              {submitting ? t.signingIn : t.signInButton}
            </button>
            <small>{t.secureSession}</small>
          </form>
        </section>
      </main>
    );

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">
            <Image
              src="/brand-mark.png"
              alt=""
              width={42}
              height={42}
              priority
            />
          </div>
          <div>
            <strong>ONA VA BOLA</strong>
            <span>KLINIKASI</span>
          </div>
        </div>
        <div className="workspace">
          <span className="workspace-label">{t.branch}</span>
          <button>
            <span className="clinic-dot" />
            Markaziy klinika
            <Icon name="chevron" />
          </button>
        </div>
        <nav className="nav-list" aria-label={t.navigation}>
          <a
            className={section === "overview" ? "active" : ""}
            href="#"
            onClick={(e) => {
              e.preventDefault();
              setSection("overview");
            }}
          >
            <Icon name="home" />
            <span>{t.overview}</span>
          </a>
          <a
            className={section === "appointments" ? "active" : ""}
            href="#"
            onClick={(e) => {
              e.preventDefault();
              setSection("appointments");
            }}
          >
            <Icon name="calendar" />
            <span>{t.appointments}</span>
          </a>
          <a
            className={section === "schedules" ? "active" : ""}
            href="#"
            onClick={(e) => {
              e.preventDefault();
              setSection("schedules");
            }}
          >
            <Icon name="doctor" />
            <span>
              {locale === "ru"
                ? "График врачей"
                : locale === "uz"
                  ? "Shifokorlar jadvali"
                  : "Doctor schedules"}
            </span>
          </a>
          <a className={section === "services" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("services"); }}>
            <Icon name="chart" />
            <span>{locale === "ru" ? "Услуги и цены" : locale === "uz" ? "Xizmatlar va narxlar" : "Services and prices"}</span>
          </a>
          <a className={section === "specialists" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("specialists"); }}>
            <Icon name="doctor" />
            <span>{locale === "ru" ? "Специалисты" : locale === "uz" ? "Mutaxassislar" : "Specialists"}</span>
          </a>
          <a className={section === "diagnosisCatalog" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("diagnosisCatalog"); }}>
            <Icon name="chart" />
            <span>{locale === "ru" ? "Диагнозы" : locale === "uz" ? "Tashxislar" : "Diagnoses"}</span>
          </a>
          <a className={section === "laboratory" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("laboratory"); }}>
            <Icon name="chart" />
            <span>{locale === "ru" ? "Лаборатория" : locale === "uz" ? "Laboratoriya" : "Laboratory"}</span>
          </a>
          <a className={section === "inventory" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("inventory"); }}>
            <Icon name="building" />
            <span>{locale === "ru" ? "Склад" : locale === "uz" ? "Ombor" : "Inventory"}</span>
          </a>
          <a className={section === "queue" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("queue"); }}>
            <Icon name="calendar" />
            <span>{locale === "ru" ? "Электронная очередь" : locale === "uz" ? "Elektron navbat" : "Electronic queue"}</span>
          </a>
          <a className={section === "cash" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("cash"); }}>
            <Icon name="chart" />
            <span>{locale === "ru" ? "Касса" : locale === "uz" ? "Kassa" : "Cash desk"}</span>
          </a>
          <a className={section === "accounting" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("accounting"); }}>
            <Icon name="chart" />
            <span>{locale === "ru" ? "Бухгалтерия" : locale === "uz" ? "Buxgalteriya" : "Accounting"}</span>
          </a>
          <a className={section === "documentBuilder" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("documentBuilder"); }}>
            <Icon name="settings" />
            <span>{locale === "ru" ? "Конструктор документов" : locale === "uz" ? "Hujjatlar konstruktori" : "Document builder"}</span>
          </a>
          <a className={section === "inpatient" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("inpatient"); }}>
            <Icon name="building" />
            <span>{locale === "ru" ? "Стационар" : locale === "uz" ? "Statsionar" : "Inpatient"}</span>
          </a>
          <a className={section === "messages" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("messages"); }}>
            <Icon name="bell" />
            {unreadMessages > 0 && <b className="nav-unread">{unreadMessages > 99 ? "99+" : unreadMessages}</b>}
            <span>{locale === "ru" ? "Сообщения" : locale === "uz" ? "Xabarlar" : "Messages"}</span>
          </a>
          {isOwner && <a className={section === "messageArchive" ? "active" : ""} href="#" onClick={(e)=>{e.preventDefault();setSection("messageArchive")}}><Icon name="chart"/><span>{locale==="ru"?"Архив чатов":locale==="uz"?"Chat arxivi":"Chat archive"}</span></a>}
          <a
            className={section === "patients" ? "active" : ""}
            href="#"
            onClick={(e) => {
              e.preventDefault();
              setSection("patients");
            }}
          >
            <Icon name="users" />
            <span>{t.patients}</span>
          </a>
          <a
            className={section === "employees" ? "active" : ""}
            href="#"
            onClick={(e) => {
              e.preventDefault();
              setSection("employees");
            }}
          >
            <Icon name="doctor" />
            <span>{t.employees}</span>
          </a>
          <a className={section === "reports" ? "active" : ""} href="#" onClick={(e) => { e.preventDefault(); setSection("reports"); }}>
            <Icon name="chart" />
            <span>{t.reports}</span>
          </a>
        </nav>
        <nav className="nav-list bottom">
          <a
            className={section === "roles" ? "active" : ""}
            href="#"
            onClick={(e) => {
              e.preventDefault();
              setSection("roles");
            }}
          >
            <Icon name="users" />
            <span>
              {locale === "ru" ? "Роли" : locale === "uz" ? "Rollar" : "Roles"}
            </span>
          </a>
          <a
            className={section === "branches" ? "active" : ""}
            href="#"
            onClick={(e) => {
              e.preventDefault();
              setSection("branches");
            }}
          >
            <Icon name="building" />
            <span>{t.branches}</span>
          </a>
        </nav>
        <div className="sidebar-help">
          <div className="help-heart">♥</div>
          <strong>{t.needHelp}</strong>
          <span>{t.supportText}</span>
          <button>{t.contact}</button>
        </div>
      </aside>

      <main className="content">
        <header className="app-header">
          <div className="mobile-brand">
            <Image
              src="/brand-mark.png"
              alt="ONA VA BOLA KLINIKASI"
              width={38}
              height={38}
            />
          </div>
          <div className="search">
            <Icon name="search" />
            <input aria-label={t.search} placeholder={t.search} />
            <kbd>⌘ K</kbd>
          </div>
          <div className="header-actions">
            <select
              value={locale}
              onChange={(e) => setLocale(e.target.value as Locale)}
              aria-label="Language"
            >
              <option value="ru">RU</option>
              <option value="uz">UZ</option>
              <option value="en">EN</option>
            </select>
            <button
              className="icon-button theme-button"
              onClick={toggleTheme}
              aria-label={dark ? t.lightTheme : t.darkTheme}
              title={dark ? t.lightTheme : t.darkTheme}
            >
              {dark ? "☀" : "☾"}
            </button>
            <button className="icon-button" aria-label={t.notifications}>
              <Icon name="bell" />
              <i />
            </button>
            <button className="profile" onClick={logout} title={t.logout}>
              <span>ШХ</span>
              <div>
                <strong>Шохрух Хаджаев</strong>
                <small>{t.owner}</small>
              </div>
              <Icon name="chevron" />
            </button>
          </div>
        </header>

        <div className="page">
          {section === "appointments" ? (
            <AppointmentManagement locale={locale} onMessageDoctor={() => setSection("messages")} />
          ) : section === "messages" ? (
            <MessageCenter locale={locale} />
          ) : section === "messageArchive" ? (
            <MessageArchive locale={locale} />
          ) : section === "services" ? (
            <ServiceManagement locale={locale} />
          ) : section === "specialists" ? (
            <SpecialistManagement locale={locale} onServices={() => setSection("services")} />
          ) : section === "diagnosisCatalog" ? (
            <DiagnosisCatalogManagement locale={locale} />
          ) : section === "laboratory" ? (
            <LaboratoryManagement locale={locale} />
          ) : section === "inventory" ? (
            <InventoryManagement locale={locale} />
          ) : section === "queue" ? (
            <QueueManagement locale={locale} />
          ) : section === "cash" ? (
            <CashDeskManagement locale={locale} />
          ) : section === "accounting" ? (
            <AccountingManagement locale={locale} />
          ) : section === "documentBuilder" ? (
            <DocumentTemplateBuilder locale={locale} />
          ) : section === "inpatient" ? (
            <InpatientManagement locale={locale} />
          ) : section === "reports" ? (
            <ReportManagement locale={locale} />
          ) : section === "schedules" ? (
            <DoctorScheduleManagement locale={locale} />
          ) : section === "patients" ? (
            <PatientManagement locale={locale} />
          ) : section !== "overview" ? (
            <AdminManagement kind={section} locale={locale} />
          ) : (
            <>
              <section className="welcome">
                <div>
                  <p>{t.today}</p>
                  <h1>{t.greeting}</h1>
                  <span>
                    {new Intl.DateTimeFormat(dateLocale, {
                      weekday: "long",
                      day: "numeric",
                      month: "long",
                      year: "numeric",
                    }).format(new Date())}
                  </span>
                </div>
                <button
                  className="primary"
                  onClick={() => setSection("appointments")}
                >
                  <Icon name="plus" />
                  {t.newAppointment}
                </button>
              </section>

              <section className="stats-grid">
                <article className="stat">
                  <div className="stat-icon rose">
                    <Icon name="calendar" />
                  </div>
                  <div>
                    <span>{t.todayAppointments}</span>
                    <strong>{dashboardLoading ? "…" : dashboard.today}</strong>
                    <small>
                      {locale === "ru"
                        ? "Данные за текущий день"
                        : locale === "uz"
                          ? "Bugungi ma’lumotlar"
                          : "Current day data"}
                    </small>
                  </div>
                </article>
                <article className="stat">
                  <div className="stat-icon amber">
                    <Icon name="users" />
                  </div>
                  <div>
                    <span>{t.waiting}</span>
                    <strong>
                      {dashboardLoading ? "…" : dashboard.waiting}
                    </strong>
                    <small>{t.liveStatus}</small>
                  </div>
                </article>
                <article className="stat">
                  <div className="stat-icon teal">
                    <Icon name="doctor" />
                  </div>
                  <div>
                    <span>{t.doctorsOnDuty}</span>
                    <strong>
                      {dashboardLoading ? "…" : dashboard.doctorsOnDuty}
                    </strong>
                    <small>{t.scheduleData}</small>
                  </div>
                </article>
                <article className="stat">
                  <div className="stat-icon violet">
                    <Icon name="chart" />
                  </div>
                  <div>
                    <span>{t.completed}</span>
                    <strong>
                      {dashboardLoading ? "…" : dashboard.completed}
                    </strong>
                    <small>{t.currentDay}</small>
                  </div>
                </article>
              </section>

              <section className="dashboard-grid">
                <article className="panel schedule-panel">
                  <div className="panel-head">
                    <div>
                      <h2>
                        {locale === "ru"
                          ? "Ближайшие записи"
                          : locale === "uz"
                            ? "Yaqin qabullar"
                            : "Upcoming appointments"}
                      </h2>
                      <p>
                        {locale === "ru"
                          ? "Сегодня и следующие рабочие дни"
                          : locale === "uz"
                            ? "Bugun va keyingi ish kunlari"
                            : "Today and the next working days"}
                      </p>
                    </div>
                    <button
                      className="text-button"
                      onClick={() => setSection("appointments")}
                    >
                      {t.openCalendar}
                      <Icon name="chevron" />
                    </button>
                  </div>
                  {dashboard.upcoming.length === 0 ? (
                    <div className="empty-state">
                      <div className="empty-illustration">
                        <Icon name="calendar" />
                        <span>+</span>
                      </div>
                      <h3>{t.noAppointments}</h3>
                      <p>{t.noData}</p>
                      <button
                        className="secondary"
                        onClick={() => setSection("appointments")}
                      >
                        <Icon name="plus" />
                        {t.createAppointment}
                      </button>
                    </div>
                  ) : (
                    <div className="dashboard-appointments">
                      {dashboard.upcoming.map((item) => (
                        <article
                          key={item.id}
                        >
                          <time>
                            <strong>
                              {new Date(item.startsAt).toLocaleTimeString(
                                dateLocale,
                                { hour: "2-digit", minute: "2-digit" },
                              )}
                            </strong>
                            <span>
                              {new Date(item.startsAt).toLocaleDateString(
                                dateLocale,
                                { day: "2-digit", month: "short" },
                              )}
                            </span>
                          </time>
                          <div>
                            <div className="dashboard-patient-heading">
                              <strong>{item.patient}</strong>
                              <ContactDetails
                                phone={item.phone}
                                secondPhone={item.secondPhone}
                                telegram={item.telegram}
                                locale={locale}
                              />
                            </div>
                            <span>{item.reason || "—"}</span>
                          </div>
                          <div>
                            <div className="dashboard-doctor-heading">
                              <strong>{item.employee}</strong>
                              <DashboardDoctorActions
                                item={item}
                                locale={locale}
                                onMessage={() => setSection("messages")}
                              />
                            </div>
                            <span>
                              {item.specialty || item.position} · {item.branch}
                            </span>
                          </div>
                        </article>
                      ))}
                    </div>
                  )}
                </article>
                <aside className="side-stack">
                  <article className="panel quick-panel">
                    <div className="panel-head">
                      <div>
                        <h2>{t.quickActions}</h2>
                        <p>{t.frequentActions}</p>
                      </div>
                    </div>
                    <div className="quick-grid">
                      <button onClick={() => setSection("appointments")}>
                        <span className="quick-icon rose">
                          <Icon name="plus" />
                        </span>
                        {t.newAppointment}
                      </button>
                      <button onClick={() => setSection("patients")}>
                        <span className="quick-icon blue">
                          <Icon name="users" />
                        </span>
                        {t.addPatient}
                      </button>
                      <button onClick={() => setSection("schedules")}>
                        <span className="quick-icon mint">
                          <Icon name="calendar" />
                        </span>
                        {t.doctorSchedule}
                      </button>
                      <button>
                        <span className="quick-icon lilac">
                          <Icon name="chart" />
                        </span>
                        {t.openReport}
                      </button>
                    </div>
                  </article>
                  <article className="panel load-panel">
                    <div className="panel-head">
                      <div>
                        <h2>{t.clinicLoad}</h2>
                        <p>{t.loadSubtitle}</p>
                      </div>
                      <span className="live-badge">
                        <i />
                        {t.live}
                      </span>
                    </div>
                    <div className="load-empty">
                      <div className="load-bars">
                        <i />
                        <i />
                        <i />
                        <i />
                        <i />
                        <i />
                        <i />
                      </div>
                      <p>{t.connectData}</p>
                    </div>
                  </article>
                </aside>
              </section>
            </>
          )}
        </div>
      </main>
      <MessageNotifier locale={locale} onOpen={() => setSection("messages")} onOpenAppointments={() => setSection("appointments")} onUnreadChange={setUnreadMessages} />
    </div>
  );
}

function ContactDetails({ phone, secondPhone, telegram, locale }: { phone: string; secondPhone: string | null | undefined; telegram: string | null | undefined; locale: Locale }) {
  const detailsRef = useRef<HTMLDetailsElement>(null);
  useEffect(() => {
    const closeOutside = (event: PointerEvent) => {
      if (detailsRef.current?.open && !detailsRef.current.contains(event.target as Node)) detailsRef.current.open = false;
    };
    const closeEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape" && detailsRef.current) detailsRef.current.open = false;
    };
    document.addEventListener("pointerdown", closeOutside);
    document.addEventListener("keydown", closeEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOutside);
      document.removeEventListener("keydown", closeEscape);
    };
  }, []);
  const label = locale === "ru" ? "Показать контакты" : locale === "uz" ? "Kontaktlarni ko‘rsatish" : "Show contacts";
  return (
    <details ref={detailsRef} className="contact-details">
      <summary aria-label={label} title={label}>☎</summary>
      <div>
        <strong>{locale === "ru" ? "Связь с пациентом" : locale === "uz" ? "Bemor bilan aloqa" : "Patient contact"}</strong>
        <a href={`tel:${phone}`}>☎ {phone}</a>
        {secondPhone && <a href={`tel:${secondPhone}`}>☎ {secondPhone}</a>}
        {telegram && <span>Telegram: {telegram}</span>}
      </div>
    </details>
  );
}

function DashboardDoctorActions({ item, locale, onMessage }: { item: DashboardAppointment; locale: Locale; onMessage: () => void }) {
  const detailsRef = useRef<HTMLDetailsElement>(null);
  useEffect(() => {
    const closeOutside = (event: PointerEvent) => {
      if (detailsRef.current?.open && !detailsRef.current.contains(event.target as Node)) detailsRef.current.open = false;
    };
    const closeEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape" && detailsRef.current) detailsRef.current.open = false;
    };
    document.addEventListener("pointerdown", closeOutside);
    document.addEventListener("keydown", closeEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOutside);
      document.removeEventListener("keydown", closeEscape);
    };
  }, []);
  const contactLabel = locale === "ru" ? "Связь с врачом" : locale === "uz" ? "Shifokor bilan aloqa" : "Doctor contact";
  const messageLabel = locale === "ru" ? "Написать врачу" : locale === "uz" ? "Shifokorga yozish" : "Message doctor";
  return (
    <span className="dashboard-doctor-actions">
      <details ref={detailsRef} className="contact-details dashboard-doctor-contact">
        <summary aria-label={contactLabel} title={contactLabel}>☎</summary>
        <div>
          <strong>{contactLabel}</strong>
          {item.employeePhone && <a href={`tel:${item.employeePhone}`}>☎ {item.employeePhone}</a>}
          {item.employeeEmail && <a href={`mailto:${item.employeeEmail}`}>✉ {item.employeeEmail}</a>}
          {item.employeeTelegram && <span>Telegram: {item.employeeTelegram}</span>}
          {!item.employeePhone && !item.employeeEmail && !item.employeeTelegram && <span>{locale === "ru" ? "Контакты не указаны" : locale === "uz" ? "Kontaktlar ko‘rsatilmagan" : "No contacts specified"}</span>}
        </div>
      </details>
      <button type="button" aria-label={messageLabel} title={messageLabel} onClick={onMessage}>✉</button>
    </span>
  );
}
