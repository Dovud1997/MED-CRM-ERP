"use client";
import { useEffect, useState } from "react";
import { Locale } from "@/lib/i18n";
type Employee = {
  id: string;
  firstName: string;
  lastName: string;
  position: string;
  specialty?: string | null;
  branch?: string | null;
  isActive: boolean;
};
type Day = {
  weekday: number;
  isWorking: boolean;
  start: string;
  end: string;
  breakFrom: string;
  breakTo: string;
};
const labels = {
  ru: {
    title: "Расписание врачей",
    sub: "Настройте рабочие смены и перерывы каждого специалиста",
    employee: "Врач или специалист",
    choose: "Выберите сотрудника",
    day: "День",
    shift: "Рабочая смена",
    break: "Перерыв",
    off: "Выходной",
    save: "Сохранить расписание",
    saved: "Расписание сохранено",
    invalid: "Проверьте время смен и перерывов",
    notConfigured:
      "Для врача ещё не настроен график. Показан стандартный шаблон.",
    days: [
      "Понедельник",
      "Вторник",
      "Среда",
      "Четверг",
      "Пятница",
      "Суббота",
      "Воскресенье",
    ],
  },
  uz: {
    title: "Shifokorlar jadvali",
    sub: "Har bir mutaxassisning ish va tanaffus vaqtini sozlang",
    employee: "Shifokor yoki mutaxassis",
    choose: "Xodimni tanlang",
    day: "Kun",
    shift: "Ish vaqti",
    break: "Tanaffus",
    off: "Dam olish",
    save: "Jadvalni saqlash",
    saved: "Jadval saqlandi",
    invalid: "Ish va tanaffus vaqtlarini tekshiring",
    notConfigured:
      "Shifokor jadvali hali sozlanmagan. Standart shablon ko‘rsatilgan.",
    days: [
      "Dushanba",
      "Seshanba",
      "Chorshanba",
      "Payshanba",
      "Juma",
      "Shanba",
      "Yakshanba",
    ],
  },
  en: {
    title: "Doctor schedules",
    sub: "Configure working shifts and breaks for each specialist",
    employee: "Doctor or specialist",
    choose: "Select an employee",
    day: "Day",
    shift: "Working shift",
    break: "Break",
    off: "Day off",
    save: "Save schedule",
    saved: "Schedule saved",
    invalid: "Check shift and break times",
    notConfigured:
      "This doctor has no configured schedule yet. A standard template is shown.",
    days: [
      "Monday",
      "Tuesday",
      "Wednesday",
      "Thursday",
      "Friday",
      "Saturday",
      "Sunday",
    ],
  },
} as const;
const defaults = (): Day[] =>
  Array.from({ length: 7 }, (_, index) => ({
    weekday: index + 1,
    isWorking: index < 6,
    start: "09:00",
    end: index === 5 ? "14:00" : "18:00",
    breakFrom: index < 5 ? "13:00" : "",
    breakTo: index < 5 ? "14:00" : "",
  }));
export function DoctorScheduleManagement({ locale }: { locale: Locale }) {
  const t = labels[locale];
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [employeeId, setEmployeeId] = useState("");
  const [days, setDays] = useState<Day[]>(defaults());
  const [configured, setConfigured] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  useEffect(() => {
    void fetch("/api/clinic/employees", { cache: "no-store" }).then(
      async (response) => {
        if (response.ok)
          setEmployees(
            ((await response.json()).items ?? []).filter(
              (item: Employee) => item.isActive && item.branch,
            ),
          );
      },
    );
  }, []);
  useEffect(() => {
    setMessage("");
    setError("");
    if (!employeeId) {
      setDays(defaults());
      setConfigured(false);
      return;
    }
    void fetch(`/api/clinic/doctor-schedules/${employeeId}`, {
      cache: "no-store",
    }).then(async (response) => {
      if (!response.ok) {
        setError(t.invalid);
        return;
      }
      const data = await response.json();
      setConfigured(Boolean(data.configured));
      setDays(data.configured ? data.days : defaults());
    });
  }, [employeeId, t.invalid]);
  const update = (weekday: number, patch: Partial<Day>) =>
    setDays((current) =>
      current.map((day) =>
        day.weekday === weekday ? { ...day, ...patch } : day,
      ),
    );
  const save = async () => {
    setMessage("");
    setError("");
    const payload = {
      days: days.map((day) =>
        day.isWorking
          ? day
          : {
              weekday: day.weekday,
              isWorking: false,
              start: "",
              end: "",
              breakFrom: "",
              breakTo: "",
            },
      ),
    };
    const response = await fetch(`/api/clinic/doctor-schedules/${employeeId}`, {
      method: "PUT",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      setError(t.invalid);
      return;
    }
    setConfigured(true);
    setMessage(t.saved);
  };
  return (
    <section className="management doctor-schedules">
      <div className="management-head">
        <div>
          <p>
            {locale === "ru"
              ? "ПЕРСОНАЛ"
              : locale === "uz"
                ? "XODIMLAR"
                : "STAFF"}
          </p>
          <h1>{t.title}</h1>
          <span>{t.sub}</span>
        </div>
      </div>
      <div className="schedule-doctor-select">
        <label>
          {t.employee}
          <select
            value={employeeId}
            onChange={(event) => setEmployeeId(event.target.value)}
          >
            <option value="">— {t.choose} —</option>
            {employees.map((employee) => (
              <option key={employee.id} value={employee.id}>
                {employee.lastName} {employee.firstName} ·{" "}
                {employee.specialty || employee.position} · {employee.branch}
              </option>
            ))}
          </select>
        </label>
      </div>
      {employeeId && (
        <>
          {!configured && (
            <div className="schedule-info">{t.notConfigured}</div>
          )}
          <div className="schedule-week">
            <div className="schedule-week-head">
              <span>{t.day}</span>
              <span>{t.shift}</span>
              <span>{t.break}</span>
            </div>
            {days.map((day, index) => (
              <div
                className={`schedule-day ${day.isWorking ? "" : "is-off"}`}
                key={day.weekday}
              >
                <label className="schedule-day-toggle">
                  <input
                    type="checkbox"
                    checked={day.isWorking}
                    onChange={(event) =>
                      update(day.weekday, { isWorking: event.target.checked })
                    }
                  />
                  <span>{t.days[index]}</span>
                </label>
                {day.isWorking ? (
                  <>
                    <div className="schedule-time">
                      <input
                        aria-label={`${t.days[index]} ${t.shift} начало`}
                        type="time"
                        value={day.start}
                        onChange={(event) =>
                          update(day.weekday, { start: event.target.value })
                        }
                      />
                      <b>—</b>
                      <input
                        aria-label={`${t.days[index]} ${t.shift} конец`}
                        type="time"
                        value={day.end}
                        onChange={(event) =>
                          update(day.weekday, { end: event.target.value })
                        }
                      />
                    </div>
                    <div className="schedule-time">
                      <input
                        aria-label={`${t.days[index]} ${t.break} начало`}
                        type="time"
                        value={day.breakFrom}
                        onChange={(event) =>
                          update(day.weekday, { breakFrom: event.target.value })
                        }
                      />
                      <b>—</b>
                      <input
                        aria-label={`${t.days[index]} ${t.break} конец`}
                        type="time"
                        value={day.breakTo}
                        onChange={(event) =>
                          update(day.weekday, { breakTo: event.target.value })
                        }
                      />
                    </div>
                  </>
                ) : (
                  <span className="schedule-off">{t.off}</span>
                )}
              </div>
            ))}
          </div>
          {error && <div className="management-error">{error}</div>}
          {message && <div className="schedule-success">{message}</div>}
          <div className="schedule-actions">
            <button className="primary" onClick={() => void save()}>
              {t.save}
            </button>
          </div>
        </>
      )}
    </section>
  );
}
