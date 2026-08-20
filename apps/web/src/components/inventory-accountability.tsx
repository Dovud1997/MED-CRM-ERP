"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";

type Branch = { id: string; name: string };
type Room = { id: string; branchId: string; branch: string; number: string; name: string; floor: number; isActive: boolean };
type Employee = { id: string; firstName: string; lastName: string; isActive: boolean };
type Item = { id: string; name: string; category: string; isActive: boolean };
type Asset = {
  id: string; itemId: string; item: string; branchId: string; branch: string; roomId?: string | null; room?: string | null;
  inventoryNumber: string; serialNumber?: string | null; status: string;
  assignedEmployeeId?: string | null; assignedEmployee?: string | null;
  responsibleEmployeeId?: string | null; responsibleEmployee?: string | null;
  acquiredOn?: string | null; warrantyUntil?: string | null; notes?: string | null;
};
type CountLine = { id: string; item: string; unit: string; lotNumber: string; expiresOn?: string | null; expectedQuantity: number; actualQuantity?: number | null; difference?: number | null };
type Count = { id: string; status: string; branch: string; branchId: string; startedAt: string; completedAt?: string | null; notes?: string | null; lines: CountLine[] };

const statuses: Record<string, string> = { AVAILABLE: "Свободно", IN_USE: "Закреплено", MAINTENANCE: "На обслуживании", RETIRED: "Списано" };

export function InventoryAccountability() {
  const [tab, setTab] = useState<"assets" | "counts">("assets");
  const [assets, setAssets] = useState<Asset[]>([]);
  const [counts, setCounts] = useState<Count[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [rooms, setRooms] = useState<Room[]>([]);
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [items, setItems] = useState<Item[]>([]);
  const [editing, setEditing] = useState<Asset | null>(null);
  const [showAsset, setShowAsset] = useState(false);
  const [showCount, setShowCount] = useState(false);
  const [assetBranch, setAssetBranch] = useState("");
  const [assetQuery, setAssetQuery] = useState("");
  const [error, setError] = useState("");

  const load = async () => {
    const [a, c, b, e, i, r] = await Promise.all([
      fetch("/api/clinic/inventory/assets", { cache: "no-store" }),
      fetch("/api/clinic/inventory/counts", { cache: "no-store" }),
      fetch("/api/clinic/branches", { cache: "no-store" }),
      fetch("/api/clinic/employees", { cache: "no-store" }),
      fetch("/api/clinic/inventory/items", { cache: "no-store" }),
      fetch("/api/clinic/inpatient/rooms", { cache: "no-store" }),
    ]);
    if (a.ok) setAssets((await a.json()).items ?? []);
    if (c.ok) setCounts((await c.json()).items ?? []);
    if (b.ok) setBranches((await b.json()).items ?? []);
    if (e.ok) setEmployees((await e.json()).items ?? []);
    if (i.ok) setItems((await i.json()).items ?? []);
    if (r.ok) setRooms((await r.json()).items ?? []);
  };
  useEffect(() => { void load(); }, []);
  useEffect(() => { if (showAsset) void load(); }, [showAsset]);
  const equipment = useMemo(() => items.filter(x => x.category === "EQUIPMENT" && x.isActive), [items]);
  const activeEmployees = useMemo(() => employees.filter(x => x.isActive), [employees]);
  const availableRooms = useMemo(() => assetBranch ? rooms.filter(x => x.isActive && x.branchId === assetBranch) : [], [rooms, assetBranch]);
  const visibleAssets = useMemo(() => { const q = assetQuery.trim().toLocaleLowerCase("ru"); return q ? assets.filter(x => [x.item, x.inventoryNumber, x.serialNumber, x.branch, x.room, x.assignedEmployee, x.responsibleEmployee].some(v => v?.toLocaleLowerCase("ru").includes(q))) : assets; }, [assets, assetQuery]);

  const saveAsset = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault(); const f = new FormData(event.currentTarget);
    const body = Object.fromEntries(f.entries());
    const response = await fetch(editing ? `/api/clinic/inventory/assets/${editing.id}` : "/api/clinic/inventory/assets", {
      method: editing ? "PATCH" : "POST", headers: { "content-type": "application/json" }, body: JSON.stringify(body),
    });
    if (!response.ok) { setError("Не удалось сохранить оборудование. Проверьте инвентарный номер и обязательные поля."); return; }
    setError(""); setEditing(null); setShowAsset(false); await load();
  };
  const startCount = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault(); const f = new FormData(event.currentTarget);
    const response = await fetch("/api/clinic/inventory/counts", { method: "POST", headers: { "content-type": "application/json" }, body: JSON.stringify({ branchId: f.get("branchId") }) });
    if (!response.ok) { setError("В этом филиале уже идёт инвентаризация или филиал не выбран."); return; }
    setError(""); setShowCount(false); await load();
  };
  const completeCount = async (event: FormEvent<HTMLFormElement>, count: Count) => {
    event.preventDefault(); const f = new FormData(event.currentTarget);
    const lines = count.lines.map(line => ({ lineId: line.id, actualQuantity: Number(f.get(`actual-${line.id}`)) }));
    const response = await fetch(`/api/clinic/inventory/counts/${count.id}/complete`, { method: "POST", headers: { "content-type": "application/json" }, body: JSON.stringify({ lines, notes: f.get("notes") }) });
    if (!response.ok) { setError("Не удалось завершить инвентаризацию. Проверьте фактические остатки."); return; }
    setError(""); await load();
  };

  return <section className="inventory-accountability">
    <div className="inventory-section-head"><div><p>КОНТРОЛЬ И ОТВЕТСТВЕННОСТЬ</p><h2>Оборудование и инвентаризация</h2><span>Закрепляйте имущество за сотрудниками и сверяйте фактические остатки</span></div></div>
    <div className="inventory-tabs inventory-subtabs">
      <button className={tab === "assets" ? "active" : ""} onClick={() => setTab("assets")}>Оборудование</button>
      <button className={tab === "counts" ? "active" : ""} onClick={() => setTab("counts")}>Инвентаризация</button>
    </div>
    {error && <div className="management-error">{error}</div>}
    {tab === "assets" && <>
      <div className="inventory-action-row"><p>Кто пользуется оборудованием, где оно находится и кто за него отвечает</p><button className="primary" onClick={() => { setEditing(null); setAssetBranch(""); setShowAsset(true); }}>+ Добавить оборудование</button></div>
      <div className="asset-search"><input value={assetQuery} onChange={e => setAssetQuery(e.target.value)} placeholder="Найти по названию, инвентарному номеру, палате или сотруднику…"/></div>
      {showAsset && <form className="admin-form asset-form" onSubmit={saveAsset}>
        <label>Оборудование<select name="itemId" required defaultValue={editing?.itemId ?? ""}><option value="">— Выберите —</option>{equipment.map(x => <option key={x.id} value={x.id}>{x.name}</option>)}</select></label>
        <label>Филиал<select name="branchId" required value={assetBranch || editing?.branchId || ""} onChange={e => setAssetBranch(e.target.value)}><option value="">— Выберите —</option>{branches.map(x => <option key={x.id} value={x.id}>{x.name}</option>)}</select></label>
        <label>Местонахождение / палата<select name="roomId" defaultValue={editing?.roomId ?? ""}><option value="">— Без привязки к палате —</option>{availableRooms.map(x => <option key={x.id} value={x.id}>Палата №{x.number}{x.name ? ` — ${x.name}` : ""}, этаж {x.floor}</option>)}</select></label>
        <label>Инвентарный номер<input name="inventoryNumber" required defaultValue={editing?.inventoryNumber}/></label>
        <label>Серийный номер<input name="serialNumber" defaultValue={editing?.serialNumber ?? ""}/></label>
        <label>Состояние<select name="status" defaultValue={editing?.status ?? "AVAILABLE"}>{Object.entries(statuses).map(([id, name]) => <option key={id} value={id}>{name}</option>)}</select></label>
        <label>Кем занято / закреплено<select name="assignedEmployeeId" defaultValue={editing?.assignedEmployeeId ?? ""}><option value="">— Не закреплено —</option>{activeEmployees.map(x => <option key={x.id} value={x.id}>{x.lastName} {x.firstName}</option>)}</select></label>
        <label>Ответственный сотрудник<select name="responsibleEmployeeId" defaultValue={editing?.responsibleEmployeeId ?? ""}><option value="">— Не назначен —</option>{activeEmployees.map(x => <option key={x.id} value={x.id}>{x.lastName} {x.firstName}</option>)}</select></label>
        <label>Дата приобретения<input name="acquiredOn" type="date" defaultValue={editing?.acquiredOn ?? ""}/></label>
        <label>Гарантия до<input name="warrantyUntil" type="date" defaultValue={editing?.warrantyUntil ?? ""}/></label>
        <label className="asset-notes">Примечание<textarea name="notes" rows={3} defaultValue={editing?.notes ?? ""}/></label>
        <div className="form-actions"><button type="button" className="secondary" onClick={() => setShowAsset(false)}>Отмена</button><button className="primary">Сохранить</button></div>
      </form>}
      <div className="asset-list">{visibleAssets.length === 0 ? <div className="table-empty">Оборудование не найдено</div> : visibleAssets.map(x => <article key={x.id}>
        <div><span>{x.inventoryNumber}{x.serialNumber ? ` · S/N ${x.serialNumber}` : ""}</span><h3>{x.item}</h3><small>{x.branch}{x.room ? ` · ${x.room}` : " · местонахождение не указано"}</small></div>
        <div><span>Кем занято</span><strong>{x.assignedEmployee || "Не закреплено"}</strong></div>
        <div><span>Ответственный</span><strong>{x.responsibleEmployee || "Не назначен"}</strong></div>
        <div><span>Статус</span><b className={`asset-status ${x.status.toLowerCase()}`}>{statuses[x.status] ?? x.status}</b></div>
        <button className="secondary" onClick={() => { setEditing(x); setAssetBranch(x.branchId); setShowAsset(true); }}>Изменить</button>
      </article>)}</div>
    </>}
    {tab === "counts" && <>
      <div className="inventory-action-row"><p>Система сравнит учётное и фактическое количество каждой партии</p><button className="primary" onClick={() => setShowCount(true)}>+ Начать инвентаризацию</button></div>
      {showCount && <form className="count-start" onSubmit={startCount}><label>Филиал<select name="branchId" required><option value="">— Выберите филиал —</option>{branches.map(x => <option key={x.id} value={x.id}>{x.name}</option>)}</select></label><div className="form-actions"><button type="button" className="secondary" onClick={() => setShowCount(false)}>Отмена</button><button className="primary">Создать ведомость</button></div></form>}
      <div className="count-list">{counts.length === 0 ? <div className="table-empty">Инвентаризации ещё не проводились</div> : counts.map(count => <form key={count.id} className="count-card" onSubmit={e => completeCount(e, count)}>
        <header><div><span>{count.branch}</span><h3>{count.status === "DRAFT" ? "Инвентаризация в работе" : "Инвентаризация завершена"}</h3><small>{new Date(count.startedAt).toLocaleString("ru-RU")}</small></div><b className={count.status === "DRAFT" ? "draft" : "completed"}>{count.status === "DRAFT" ? "В работе" : "Завершена"}</b></header>
        <div className="count-table"><div className="count-table-head"><span>Позиция / партия</span><span>Срок годности</span><span>По учёту</span><span>Фактически</span><span>Разница</span></div>{count.lines.length === 0 ? <p className="table-empty">В филиале нет партий с остатком</p> : count.lines.map(line => <div key={line.id}><strong>{line.item}<small>Партия {line.lotNumber}</small></strong><span>{line.expiresOn ? new Date(line.expiresOn).toLocaleDateString("ru-RU") : "Не указан"}</span><span>{line.expectedQuantity} {line.unit}</span>{count.status === "DRAFT" ? <input name={`actual-${line.id}`} type="number" min="0" step="0.001" required defaultValue={line.expectedQuantity}/> : <span>{line.actualQuantity} {line.unit}</span>}<b className={(line.difference ?? 0) !== 0 ? "count-difference" : ""}>{count.status === "DRAFT" ? "—" : `${(line.difference ?? 0) > 0 ? "+" : ""}${line.difference ?? 0}`}</b></div>)}</div>
        {count.status === "DRAFT" && <footer><label>Комментарий<textarea name="notes" rows={2}/></label><button className="primary">Завершить и применить остатки</button></footer>}
      </form>)}</div>
    </>}
  </section>;
}
