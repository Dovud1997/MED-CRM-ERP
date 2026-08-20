"use client";

import { useEffect, useRef, useState } from "react";
import { Locale } from "@/lib/i18n";

type Patient = {
  id: string;
  firstName: string;
  lastName: string;
  middleName?: string | null;
  birthDate?: string;
};

const text = {
  ru: { placeholder: "Введите имя, фамилию или отчество", hint: "Введите минимум 2 буквы", empty: "Совпадений не найдено", loading: "Ищем…", clear: "Очистить выбор" },
  uz: { placeholder: "Ism, familiya yoki otasining ismini kiriting", hint: "Kamida 2 ta harf kiriting", empty: "Mos bemor topilmadi", loading: "Qidirilmoqda…", clear: "Tanlovni tozalash" },
  en: { placeholder: "Enter first, last or middle name", hint: "Enter at least 2 letters", empty: "No matches found", loading: "Searching…", clear: "Clear selection" },
} as const;

const patientName = (patient: Patient) =>
  [patient.lastName, patient.firstName, patient.middleName].filter(Boolean).join(" ");

export function PatientSearch({ locale, value, onChange }: { locale: Locale; value: Patient | null; onChange: (patient: Patient | null) => void }) {
  const t = text[locale];
  const [query, setQuery] = useState("");
  const [items, setItems] = useState<Patient[]>([]);
  const [loading, setLoading] = useState(false);
  const [focused, setFocused] = useState(false);
  const request = useRef<AbortController | null>(null);

  useEffect(() => {
    const normalized = query.trim();
    if (value || normalized.length < 2) {
      request.current?.abort();
      setItems([]);
      setLoading(false);
      return;
    }
    const timer = window.setTimeout(async () => {
      request.current?.abort();
      const controller = new AbortController();
      request.current = controller;
      setLoading(true);
      try {
        const response = await fetch(`/api/clinic/patients?q=${encodeURIComponent(normalized)}`, { cache: "no-store", signal: controller.signal });
        if (response.ok) setItems((await response.json()).items ?? []);
      } catch (error) {
        if (!(error instanceof DOMException && error.name === "AbortError")) setItems([]);
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    }, 250);
    return () => window.clearTimeout(timer);
  }, [query, value]);

  const choose = (patient: Patient) => {
    onChange(patient);
    setQuery(patientName(patient));
    setItems([]);
    setFocused(false);
  };

  return (
    <div className="appointment-patient-search">
      <input
        type="search"
        value={value ? patientName(value) : query}
        placeholder={t.placeholder}
        autoComplete="off"
        aria-label={t.placeholder}
        onFocus={() => setFocused(true)}
        onBlur={() => window.setTimeout(() => setFocused(false), 120)}
        onChange={(event) => {
          onChange(null);
          setQuery(event.target.value);
          setFocused(true);
        }}
      />
      {value && <button type="button" className="appointment-patient-search-clear" aria-label={t.clear} onClick={() => { onChange(null); setQuery(""); setFocused(true); }}>×</button>}
      {focused && !value && query.trim().length > 0 && (
        <div className="appointment-patient-search-results" role="listbox">
          {query.trim().length < 2 ? <span>{t.hint}</span> : loading ? <span>{t.loading}</span> : items.length === 0 ? <span>{t.empty}</span> : items.map((patient) => (
            <button type="button" role="option" key={patient.id} onMouseDown={(event) => event.preventDefault()} onClick={() => choose(patient)}>
              <strong>{patientName(patient)}</strong>
              {patient.birthDate && <small>{new Date(`${patient.birthDate}T00:00:00`).toLocaleDateString(locale)}</small>}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
