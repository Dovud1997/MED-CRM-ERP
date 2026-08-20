"use client";
import { FormEvent, useEffect, useRef, useState } from "react";
import { Locale } from "@/lib/i18n";
import { PatientSearch } from "@/components/patient-search";
type Patient = {
  id: string;
  firstName: string;
  lastName: string;
  middleName?: string | null;
};
type Employee = {
  id: string;
  firstName: string;
  lastName: string;
  position: string;
  specialty?: string | null;
  branchId?: string | null;
  branch?: string | null;
  isActive: boolean;
};
type Branch = { id: string; name: string };
type Appointment = {
  id: string;
  startsAt: string;
  endsAt: string;
  status: string;
  reason?: string | null;
  notes?: string | null;
  patient: {
    id: string;
    name: string;
    phone: string;
    secondPhone?: string | null;
    telegram?: string | null;
  };
  employee: {
    id: string;
    userId: string;
    name: string;
    position: string;
    specialty?: string | null;
    phone?: string | null;
    email?: string | null;
    telegram?: string | null;
  };
  branch: { id: string; name: string };
};
const words = {
  ru: {
    title: "Записи и приём",
    sub: "Расписание пациентов и управление статусом приёма",
    add: "Новая запись",
    date: "Дата расписания",
    patient: "Пациент",
    doctor: "Врач или специалист",
    branch: "Филиал",
    start: "Дата и время",
    duration: "Длительность",
    reason: "Причина обращения",
    notes: "Примечание",
    save: "Создать запись",
    cancel: "Отмена",
    empty: "На выбранную дату записей нет",
    active: "Активные записи",
    activeSub: "Предстоящие приёмы независимо от выбранной даты",
    noActive: "Активных записей пока нет",
    conflict: "У врача уже есть запись на это время.",
    invalid: "Проверьте данные записи.",
    outside:
      "Выбранное время находится вне графика врача или попадает на перерыв.",
    past: "Нельзя создать запись на прошедшее время. Выберите будущее время или другую дату.",
    minutes: "мин",
    scheduled: "Запланирован",
    confirmed: "Подтверждён",
    arrived: "Пациент прибыл",
    in_progress: "Идёт приём",
    completed: "Завершён",
    cancelled: "Отменён",
    no_show: "Не пришёл",
  },
  uz: {
    title: "Qabullar",
    sub: "Bemorlar jadvali va qabul holatini boshqarish",
    add: "Yangi qabul",
    date: "Jadval sanasi",
    patient: "Bemor",
    doctor: "Shifokor yoki mutaxassis",
    branch: "Filial",
    start: "Sana va vaqt",
    duration: "Davomiyligi",
    reason: "Murojaat sababi",
    notes: "Izoh",
    save: "Qabul yaratish",
    cancel: "Bekor qilish",
    empty: "Tanlangan sanada qabullar yo‘q",
    active: "Faol qabullar",
    activeSub: "Tanlangan sanadan qat’i nazar, kelgusi qabullar",
    noActive: "Faol qabullar hozircha yo‘q",
    conflict: "Shifokorning bu vaqtda qabuli mavjud.",
    invalid: "Qabul ma’lumotlarini tekshiring.",
    outside:
      "Tanlangan vaqt shifokor jadvalidan tashqarida yoki tanaffusga to‘g‘ri keladi.",
    past: "O‘tgan vaqtga qabul yaratib bo‘lmaydi. Kelajakdagi vaqt yoki boshqa sanani tanlang.",
    minutes: "daq",
    scheduled: "Rejalashtirilgan",
    confirmed: "Tasdiqlangan",
    arrived: "Bemor keldi",
    in_progress: "Qabul davom etmoqda",
    completed: "Yakunlandi",
    cancelled: "Bekor qilindi",
    no_show: "Kelmadi",
  },
  en: {
    title: "Appointments",
    sub: "Patient schedule and appointment status management",
    add: "New appointment",
    date: "Schedule date",
    patient: "Patient",
    doctor: "Doctor or specialist",
    branch: "Branch",
    start: "Date and time",
    duration: "Duration",
    reason: "Reason for visit",
    notes: "Notes",
    save: "Create appointment",
    cancel: "Cancel",
    empty: "No appointments for this date",
    active: "Active appointments",
    activeSub: "Upcoming appointments regardless of selected date",
    noActive: "There are no active appointments yet",
    conflict: "The doctor already has an appointment at this time.",
    invalid: "Check the appointment details.",
    outside:
      "The selected time is outside the doctor's schedule or overlaps a break.",
    past: "An appointment cannot be created in the past. Choose a future time or another date.",
    minutes: "min",
    scheduled: "Scheduled",
    confirmed: "Confirmed",
    arrived: "Arrived",
    in_progress: "In progress",
    completed: "Completed",
    cancelled: "Cancelled",
    no_show: "No show",
  },
} as const;
const today = () => {
  const d = new Date();
  return new Date(d.getTime() - d.getTimezoneOffset() * 60000)
    .toISOString()
    .slice(0, 10);
};
export function AppointmentManagement({ locale, onMessageDoctor }: { locale: Locale; onMessageDoctor?: (userId: string) => void }) {
  const t = words[locale];
  const [date, setDate] = useState(today());
  const [items, setItems] = useState<Appointment[]>([]);
  const [activeItems, setActiveItems] = useState<Appointment[]>([]);
  const [selectedPatient, setSelectedPatient] = useState<Patient | null>(null);
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [selectedBranch, setSelectedBranch] = useState("");
  const [open, setOpen] = useState(false);
  const [error, setError] = useState("");
  const load = async () => {
    const [a, active, e, b] = await Promise.all([
      fetch(`/api/clinic/appointments?date=${date}`, { cache: "no-store" }),
      fetch("/api/clinic/appointments?scope=active", { cache: "no-store" }),
      fetch("/api/clinic/employees", { cache: "no-store" }),
      fetch("/api/clinic/branches", { cache: "no-store" }),
    ]);
    if (a.ok) setItems((await a.json()).items ?? []);
    if (active.ok) setActiveItems((await active.json()).items ?? []);
    if (e.ok)
      setEmployees(
        ((await e.json()).items ?? []).filter((x: Employee) => x.isActive),
      );
    if (b.ok) setBranches((await b.json()).items ?? []);
  };
  useEffect(() => {
    void load();
    const timer = window.setInterval(() => void load(), 3000);
    return () => window.clearInterval(timer);
  }, [date]);
  const create = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    if (!selectedPatient) {
      setError(t.invalid);
      return;
    }
    const form = e.currentTarget,
      f = new FormData(form);
    const start = new Date(String(f.get("startsAt")));
    const r = await fetch("/api/clinic/appointments", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        patientId: selectedPatient.id,
        employeeId: f.get("employeeId"),
        branchId: f.get("branchId"),
        startsAt: start.toISOString(),
        durationMinutes: Number(f.get("duration")),
        reason: f.get("reason"),
        notes: f.get("notes"),
      }),
    });
    if (!r.ok) {
      const x = await r.json().catch(() => null);
      setError(
        x?.error?.code === "schedule_conflict"
          ? t.conflict
          : x?.error?.code === "outside_working_hours"
            ? t.outside
            : x?.error?.code === "appointment_in_past"
              ? t.past
              : t.invalid,
      );
      return;
    }
    form.reset();
    setSelectedPatient(null);
    setSelectedBranch("");
    setOpen(false);
    const createdDate = new Date(
      start.getTime() - start.getTimezoneOffset() * 60000,
    )
      .toISOString()
      .slice(0, 10);
    if (createdDate === date) await load();
    else setDate(createdDate);
  };
  const status = async (id: string, value: string) => {
    const r = await fetch(`/api/clinic/appointments/${id}/status`, {
      method: "PATCH",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ status: value }),
    });
    if (r.ok)
      {
        setItems((x) =>
          x.map((a) => (a.id === id ? { ...a, status: value } : a)),
        );
        setActiveItems((x) =>
          ["completed", "cancelled", "no_show"].includes(value)
            ? x.filter((a) => a.id !== id)
            : x.map((a) => (a.id === id ? { ...a, status: value } : a)),
        );
      }
  };
  return (
    <section className="management appointments">
      <div className="management-head">
        <div>
          <p>
            {locale === "ru"
              ? "РАСПИСАНИЕ"
              : locale === "uz"
                ? "JADVAL"
                : "SCHEDULE"}
          </p>
          <h1>{t.title}</h1>
          <span>{t.sub}</span>
        </div>
        <button className="primary" onClick={() => setOpen((v) => !v)}>
          + {t.add}
        </button>
      </div>
      <div className="appointment-toolbar">
        <label>
          {t.date}
          <input
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
        </label>
        <div className="appointment-summary">
          <strong>{items.length}</strong>
          <span>
            {locale === "ru"
              ? "записей"
              : locale === "uz"
                ? "qabul"
                : "appointments"}
          </span>
        </div>
      </div>
      {open && (
        <form className="appointment-form" onSubmit={create}>
          <label className="appointment-patient-search-field">
            {t.patient}
            <PatientSearch
              locale={locale}
              value={selectedPatient}
              onChange={setSelectedPatient}
            />
          </label>
          <label>
            {t.branch}
            <select
              name="branchId"
              required
              value={selectedBranch}
              onChange={(e) => setSelectedBranch(e.target.value)}
            >
              <option value="">—</option>
              {branches.map((x) => (
                <option key={x.id} value={x.id}>
                  {x.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            {t.doctor}
            <select name="employeeId" required disabled={!selectedBranch}>
              <option value="">—</option>
              {employees
                .filter((x) => x.branchId === selectedBranch)
                .map((x) => (
                  <option key={x.id} value={x.id}>
                    {x.lastName} {x.firstName} · {x.specialty || x.position}
                  </option>
                ))}
            </select>
          </label>
          <label>
            {t.start}
            <input name="startsAt" type="datetime-local" required />
          </label>
          <label>
            {t.duration}
            <select name="duration" defaultValue="30">
              <option value="15">15 {t.minutes}</option>
              <option value="30">30 {t.minutes}</option>
              <option value="45">45 {t.minutes}</option>
              <option value="60">60 {t.minutes}</option>
              <option value="90">90 {t.minutes}</option>
            </select>
          </label>
          <label>
            {t.reason}
            <input name="reason" />
          </label>
          <label className="appointment-notes">
            {t.notes}
            <input name="notes" />
          </label>
          <div className="form-actions">
            <button
              type="button"
              className="secondary"
              onClick={() => setOpen(false)}
            >
              {t.cancel}
            </button>
            <button className="primary">{t.save}</button>
          </div>
        </form>
      )}
      {error && <div className="management-error">{error}</div>}
      <div className="appointment-section-title">
        <div>
          <h2>{t.active}</h2>
          <span>{t.activeSub}</span>
        </div>
        <strong>{activeItems.length}</strong>
      </div>
      <AppointmentList
        items={activeItems}
        locale={locale}
        t={t}
        empty={t.noActive}
        onStatus={status}
        showDate
        onMessageDoctor={onMessageDoctor}
      />
      <div className="appointment-section-title selected-date-title">
        <div>
          <h2>{t.date}</h2>
          <span>{new Intl.DateTimeFormat(locale, { dateStyle: "long" }).format(new Date(`${date}T12:00:00`))}</span>
        </div>
        <strong>{items.length}</strong>
      </div>
      <AppointmentList items={items} locale={locale} t={t} empty={t.empty} onStatus={status} onMessageDoctor={onMessageDoctor} />
    </section>
  );
}

function AppointmentList({ items, locale, t, empty, onStatus, onMessageDoctor, showDate = false }: { items: Appointment[]; locale: Locale; t: (typeof words)[Locale]; empty: string; onStatus: (id: string, value: string) => Promise<void>; onMessageDoctor: ((userId: string) => void) | undefined; showDate?: boolean }) {
  return (
      <div className="appointment-list">
        {items.length === 0 ? (
          <div className="table-empty">{empty}</div>
        ) : (
          items.map((x) => (
            <article
              key={x.id}
              className={`appointment-item status-${x.status}`}
            >
              <time>
                <strong>
                  {new Date(x.startsAt).toLocaleTimeString(locale, {
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </strong>
                <span>
                  {showDate && `${new Date(x.startsAt).toLocaleDateString(locale, { day: "2-digit", month: "short" })} · `}
                  {Math.round(
                    (new Date(x.endsAt).getTime() -
                      new Date(x.startsAt).getTime()) /
                      60000,
                  )}{" "}
                  {t.minutes}
                </span>
              </time>
              <div className="appointment-person">
                <div className="appointment-patient-heading">
                  <strong>{x.patient.name}</strong>
                  <PatientContact patient={x.patient} locale={locale} />
                </div>
                <span>{x.reason || "—"}</span>
              </div>
              <div>
                <div className="doctor-heading"><strong>{x.employee.name}</strong><DoctorActions employee={x.employee} locale={locale} onMessage={onMessageDoctor}/></div>
                <span>
                  {x.employee.specialty || x.employee.position} ·{" "}
                  {x.branch.name}
                </span>
              </div>
              <select
                value={x.status}
                onChange={(e) => void onStatus(x.id, e.target.value)}
                aria-label={`${t.title}: ${x.patient.name}`}
              >
                <option value="scheduled">{t.scheduled}</option>
                <option value="confirmed">{t.confirmed}</option>
                <option value="arrived">{t.arrived}</option>
                <option value="in_progress">{t.in_progress}</option>
                <option value="completed">{t.completed}</option>
                <option value="cancelled">{t.cancelled}</option>
                <option value="no_show">{t.no_show}</option>
              </select>
            </article>
          ))
        )}
      </div>
  );
}

function PatientContact({ patient, locale }: { patient: Appointment["patient"]; locale: Locale }) {
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
    <details ref={detailsRef} className="contact-details appointment-contact-details">
      <summary aria-label={label} title={label}>☎</summary>
      <div>
        <strong>{locale === "ru" ? "Связь с пациентом" : locale === "uz" ? "Bemor bilan aloqa" : "Patient contact"}</strong>
        <a href={`tel:${patient.phone}`}>☎ {patient.phone}</a>
        {patient.secondPhone && <a href={`tel:${patient.secondPhone}`}>☎ {patient.secondPhone}</a>}
        {patient.telegram && <span>Telegram: {patient.telegram}</span>}
      </div>
    </details>
  );
}

function DoctorActions({employee,locale,onMessage}:{employee:Appointment["employee"];locale:Locale;onMessage: ((userId:string)=>void) | undefined}){
 const label=locale==="ru"?"Связь с врачом":locale==="uz"?"Shifokor bilan aloqa":"Doctor contact";
 return <span className="doctor-actions"><details className="contact-details doctor-contact"><summary aria-label={label}>☎</summary><div><strong>{label}</strong>{employee.phone&&<a href={`tel:${employee.phone}`}>☎ {employee.phone}</a>}{employee.email&&<a href={`mailto:${employee.email}`}>✉ {employee.email}</a>}{employee.telegram&&<span>Telegram: {employee.telegram}</span>}{!employee.phone&&!employee.email&&!employee.telegram&&<span>Контакты не указаны</span>}</div></details><button type="button" aria-label={locale==="ru"?"Написать врачу":"Message doctor"} onClick={()=>onMessage?.(employee.userId)}>✉</button></span>
}
