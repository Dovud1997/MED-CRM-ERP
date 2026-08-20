"use client";

import { FormEvent, useEffect, useState } from "react";
import { Locale } from "@/lib/i18n";

type Kind = "branches" | "employees" | "roles";
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
  middleName?: string | null;
  position: string;
  specialty?: string | null;
  publicEmail?: string | null;
  telegramUrl?: string | null;
  passport: string;
  permanentAddress: string;
  phoneLocal: string;
  branchId: string | null;
  roleIds: string[];
  branch: string | null;
  isActive: boolean;
  isOwner: boolean;
  hasPhoto?: boolean;
};

async function employeePhotoPayload(file:File){if(file.size>5*1024*1024)throw new Error("photo_too_large");if(!["image/jpeg","image/png","image/webp"].includes(file.type))throw new Error("invalid_photo");const bytes=new Uint8Array(await file.arrayBuffer());let binary="";for(let i=0;i<bytes.length;i+=32768)binary+=String.fromCharCode(...bytes.subarray(i,i+32768));return{contentType:file.type,data:btoa(binary)}}
function EmployeePhotoInput({employee}:{employee?:Employee|null}){const[preview,setPreview]=useState(employee?.hasPhoto?`/api/clinic/employees/${employee.id}/photo`:"");return <label className="profile-photo-field"><span className="profile-photo-preview" style={preview?{backgroundImage:`url(${preview})`}:undefined}>{!preview&&"Фото"}</span><span><strong>Фотография сотрудника</strong><small>JPG, PNG или WEBP · до 5 МБ</small><input type="file" name="profilePhoto" accept="image/jpeg,image/png,image/webp" onChange={e=>{const file=e.target.files?.[0];if(file)setPreview(URL.createObjectURL(file))}}/></span></label>}
type Role = {
  id: string;
  code: string;
  name: string;
  isSystem: boolean;
  employeeCount: number;
  permissions: string[];
};
type Permission = { id: string; code: string; description: string };

const permissionNames: Record<Locale, Record<string, string>> = {
  ru: {
    "*": "Все права",
    "appointments:read": "Просмотр записей",
    "appointments:write": "Управление записями",
    "audit:read": "Просмотр журнала действий",
    "branches:read": "Просмотр филиалов",
    "branches:write": "Управление филиалами",
    "employees:read": "Просмотр сотрудников",
    "employees:write": "Управление сотрудниками",
    "finance:read": "Просмотр финансов",
    "finance:write": "Управление финансами",
    "patients:read": "Просмотр пациентов",
    "patients:write": "Управление пациентами",
    "roles:read": "Просмотр ролей",
    "roles:write": "Управление ролями",
  },
  uz: {
    "*": "Barcha huquqlar",
    "appointments:read": "Qabullarni ko‘rish",
    "appointments:write": "Qabullarni boshqarish",
    "audit:read": "Harakatlar jurnalini ko‘rish",
    "branches:read": "Filiallarni ko‘rish",
    "branches:write": "Filiallarni boshqarish",
    "employees:read": "Xodimlarni ko‘rish",
    "employees:write": "Xodimlarni boshqarish",
    "finance:read": "Moliyani ko‘rish",
    "finance:write": "Moliyani boshqarish",
    "patients:read": "Bemorlarni ko‘rish",
    "patients:write": "Bemorlarni boshqarish",
    "roles:read": "Rollarni ko‘rish",
    "roles:write": "Rollarni boshqarish",
  },
  en: {
    "*": "All permissions",
    "appointments:read": "Read appointments",
    "appointments:write": "Manage appointments",
    "audit:read": "Read audit trail",
    "branches:read": "Read branches",
    "branches:write": "Manage branches",
    "employees:read": "Read employees",
    "employees:write": "Manage employees",
    "finance:read": "Read finance",
    "finance:write": "Manage finance",
    "patients:read": "Read patient directory",
    "patients:write": "Manage patients",
    "roles:read": "Read roles",
    "roles:write": "Manage roles",
  },
};

const copy = {
  ru: {
    branches: "Филиалы",
    employees: "Сотрудники",
    roles: "Роли и права",
    branchText: "Добавляйте филиалы без ограничений",
    employeeText:
      "Создавайте аккаунты сотрудников и назначайте несколько ролей",
    roleText: "Создавайте собственные роли с нужными правами",
    addBranch: "Добавить филиал",
    addEmployee: "Добавить сотрудника",
    addRole: "Создать роль",
    name: "Название",
    address: "Адрес",
    first: "Имя",
    last: "Фамилия",
    middle: "Отчество",
    passport: "Серия и номер паспорта",
    residence: "Постоянное место жительства",
    phone: "Номер телефона",
    email: "Публичная рабочая почта",
    telegram: "Ссылка Telegram",
    login: "Логин",
    password: "Временный пароль (от 12 символов)",
    position: "Должность",
    specialty: "Специальность",
    branch: "Филиал",
    rolesLabel: "Роли сотрудника",
    permissions: "Права доступа",
    save: "Сохранить",
    cancel: "Отмена",
    status: "Статус",
    active: "Активен",
    empty: "Записей пока нет",
    loading: "Загрузка…",
    error: "Проверьте правильность заполнения полей",
    delete: "Удалить",
    edit: "Изменить",
    people: "Сотрудников",
    system: "Защищена",
  },
  uz: {
    branches: "Filiallar",
    employees: "Xodimlar",
    roles: "Rollar va huquqlar",
    branchText: "Istalgancha filial qo‘shing",
    employeeText: "Xodim hisoblarini yarating va bir nechta rol bering",
    roleText: "Kerakli huquqlar bilan o‘z rollaringizni yarating",
    addBranch: "Filial qo‘shish",
    addEmployee: "Xodim qo‘shish",
    addRole: "Rol yaratish",
    name: "Nomi",
    address: "Manzil",
    first: "Ism",
    last: "Familiya",
    middle: "Otasining ismi",
    passport: "Pasport seriyasi va raqami",
    residence: "Doimiy yashash manzili",
    phone: "Telefon raqami",
    email: "Ommaviy ish elektron pochtasi",
    telegram: "Telegram havolasi",
    login: "Login",
    password: "Vaqtinchalik parol (12+ belgi)",
    position: "Lavozim",
    specialty: "Mutaxassislik",
    branch: "Filial",
    rolesLabel: "Xodim rollari",
    permissions: "Kirish huquqlari",
    save: "Saqlash",
    cancel: "Bekor qilish",
    status: "Holat",
    active: "Faol",
    empty: "Hozircha yozuv yo‘q",
    loading: "Yuklanmoqda…",
    error: "Maydonlarni to‘g‘ri to‘ldiring",
    delete: "O‘chirish",
    edit: "O‘zgartirish",
    people: "Xodimlar",
    system: "Himoyalangan",
  },
  en: {
    branches: "Branches",
    employees: "Employees",
    roles: "Roles & access",
    branchText: "Add as many branches as your clinic needs",
    employeeText: "Create staff accounts and assign multiple roles",
    roleText: "Create your own roles with selected permissions",
    addBranch: "Add branch",
    addEmployee: "Add employee",
    addRole: "Create role",
    name: "Name",
    address: "Address",
    first: "First name",
    last: "Last name",
    middle: "Middle name",
    passport: "Passport series and number",
    residence: "Permanent residence",
    phone: "Phone number",
    email: "Public work email",
    telegram: "Telegram link",
    login: "Login",
    password: "Temporary password (12+ characters)",
    position: "Position",
    specialty: "Specialty",
    branch: "Branch",
    rolesLabel: "Employee roles",
    permissions: "Access permissions",
    save: "Save",
    cancel: "Cancel",
    status: "Status",
    active: "Active",
    empty: "No records yet",
    loading: "Loading…",
    error: "Check the entered fields",
    delete: "Delete",
    edit: "Edit",
    people: "Employees",
    system: "Protected",
  },
} as const;

export function AdminManagement({
  kind,
  locale,
}: {
  kind: Kind;
  locale: Locale;
}) {
  const t = copy[locale];
  const [items, setItems] = useState<(Branch | Employee | Role)[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [editRole, setEditRole] = useState<Role | null>(null);
  const [editBranch, setEditBranch] = useState<Branch | null>(null);
  const [editEmployee, setEditEmployee] = useState<Employee | null>(null);
  const [error, setError] = useState("");
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const load = async () => {
    setLoading(true);
    setError("");
    try {
      const requests = [fetch(`/api/clinic/${kind}`, { cache: "no-store" })];
      if (kind === "employees")
        requests.push(
          fetch("/api/clinic/branches"),
          fetch("/api/clinic/roles"),
        );
      if (kind === "roles") requests.push(fetch("/api/clinic/permissions"));
      const data = await Promise.all(requests);
      if (data.some((x) => !x.ok)) throw new Error();
      const json = await Promise.all(data.map((x) => x.json()));
      const loadedItems = json[0].items ?? [];
      setItems(
        kind === "roles"
          ? loadedItems.map((role: Role) => ({
              ...role,
              permissions: Array.isArray(role.permissions)
                ? role.permissions
                : [],
            }))
          : loadedItems,
      );
      if (kind === "employees") {
        setBranches(json[1].items ?? []);
        setRoles(json[2].items ?? []);
      }
      if (kind === "roles") setPermissions(json[1].items ?? []);
    } catch {
      setError(t.error);
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => {
    void load();
  }, [kind]);
  const create = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    setFieldErrors({});
    const form = e.currentTarget;
    const f = new FormData(form);
    const profilePhoto=f.get("profilePhoto");
    let body: Record<string, unknown>;
    if (kind === "branches")
      body = {
        name: f.get("name"),
        address: f.get("address"),
        timezone: "Asia/Tashkent",
      };
    else if (kind === "roles")
      body = { name: f.get("name"), permissionCodes: f.getAll("permissions") };
    else
      body = {
        firstName: f.get("firstName"),
        lastName: f.get("lastName"),
        middleName: f.get("middleName"),
        passport: f.get("passport"),
        permanentAddress: f.get("permanentAddress"),
        phoneLocal: f.get("phoneLocal"),
        publicEmail: f.get("publicEmail"),
        telegramUrl: f.get("telegramUrl"),
        login: f.get("login"),
        password: f.get("password"),
        position: f.get("position"),
        specialty: f.get("specialty"),
        branchId: f.get("branchId"),
        roleIds: f.getAll("roleIds"),
      };
    const editId =
      kind === "roles"
        ? editRole?.id
        : kind === "branches"
          ? editBranch?.id
          : editEmployee?.id;
    const r = await fetch(
      editId ? `/api/clinic/${kind}/${editId}` : `/api/clinic/${kind}`,
      {
        method: editId ? "PATCH" : "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(body),
      },
    );
    if (!r.ok) {
      const data = await r.json().catch(() => null);
      const fields = data?.error?.fields as Record<string, string> | undefined;
      if (fields) setFieldErrors(fields);
      if (data?.error?.code === "login_exists") setFieldErrors({ login: "login_exists" });
      setError(t.error);
      return;
    }
    const saved=await r.json().catch(()=>({}));
    const photoTargetId=editId||saved.id;
    if(kind==="employees"&&profilePhoto instanceof File&&profilePhoto.size&&photoTargetId){try{const photoResponse=await fetch(`/api/clinic/employees/${photoTargetId}/photo`,{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify(await employeePhotoPayload(profilePhoto))});if(!photoResponse.ok)throw new Error()}catch{setError(locale==="ru"?"Сотрудник сохранён, но фотографию загрузить не удалось":"Photo upload failed");return}}
    form.reset();
    setOpen(false);
    setEditRole(null);
    setEditBranch(null);
    setEditEmployee(null);
    await load();
  };
  const remove = async (id: string) => {
    if (!confirm(`${t.delete}?`)) return;
    const r = await fetch(`/api/clinic/roles/${id}`, { method: "DELETE" });
    if (!r.ok) {
      setError(t.error);
      return;
    }
    await load();
  };
  const removeBranch = async (id: string) => {
    if (!confirm(`${t.delete}?`)) return;
    const r = await fetch(`/api/clinic/branches/${id}`, { method: "DELETE" });
    if (!r.ok) {
      const data = await r.json().catch(() => null);
      setError(
        data?.error?.code === "branch_has_employees"
          ? locale === "ru"
            ? "Сначала переведите сотрудников в другой филиал"
            : locale === "uz"
              ? "Avval xodimlarni boshqa filialga o‘tkazing"
              : "Move employees to another branch first"
          : t.error,
      );
      return;
    }
    await load();
  };
  const changeEmployeeStatus = async (employee: Employee) => {
    setError("");
    const r = await fetch(`/api/clinic/employees/${employee.id}/status`, { method: "PATCH", headers: { "content-type": "application/json" }, body: JSON.stringify({ isActive: !employee.isActive }) });
    if (!r.ok) { const data = await r.json().catch(() => null); setError(data?.error?.code === "protected_owner" ? (locale === "ru" ? "Профиль владельца нельзя отключить" : locale === "uz" ? "Egasi profilini o‘chirib bo‘lmaydi" : "The owner profile cannot be disabled") : t.error); return; }
    await load();
  };
  const removeEmployee = async (employee: Employee) => {
    if (!confirm(locale === "ru" ? "Удалить профиль сотрудника? Связанные записи сохранятся в архиве." : `${t.delete}?`)) return;
    const r = await fetch(`/api/clinic/employees/${employee.id}`, { method: "DELETE" });
    if (!r.ok) { const data = await r.json().catch(() => null); setError(data?.error?.code === "protected_owner" ? (locale === "ru" ? "Профиль владельца нельзя удалить" : locale === "uz" ? "Egasi profilini o‘chirib bo‘lmaydi" : "The owner profile cannot be deleted") : t.error); return; }
    await load();
  };
  const title =
    kind === "branches"
      ? t.branches
      : kind === "employees"
        ? t.employees
        : t.roles;
  const subtitle =
    kind === "branches"
      ? t.branchText
      : kind === "employees"
        ? t.employeeText
        : t.roleText;
  const add =
    kind === "branches"
      ? t.addBranch
      : kind === "employees"
        ? t.addEmployee
        : t.addRole;
  return (
    <section className="management">
      <div className="management-head">
        <div>
          <p>
            {locale === "ru"
              ? "Управление"
              : locale === "uz"
                ? "Boshqaruv"
                : "Management"}
          </p>
          <h1>{title}</h1>
          <span>{subtitle}</span>
        </div>
        <button
          className="primary"
          onClick={() => {
            setEditRole(null);
            setEditBranch(null);
            setEditEmployee(null);
            setOpen(!open);
          }}
        >
          + {add}
        </button>
      </div>
      {open && (
        <form className="admin-form" onSubmit={create} noValidate>
          {kind === "branches" ? (
            <>
              <Field
                name="name"
                label={t.name}
                defaultValue={editBranch?.name}
              />
              <Field
                name="address"
                label={t.address}
                required={false}
                defaultValue={editBranch?.address || ""}
              />
            </>
          ) : kind === "roles" ? (
            <>
              <Field name="name" label={t.name} defaultValue={editRole?.name} />
              <Choice
                name="permissions"
                label={t.permissions}
                selected={editRole?.permissions}
                items={permissions.map((x) => ({
                  id: x.code,
                  name: permissionNames[locale][x.code] || x.code,
                }))}
              />
            </>
          ) : (
            <>
              <EmployeePhotoInput employee={editEmployee} />
              <Field name="firstName" label={t.first} defaultValue={editEmployee?.firstName} error={fieldErrors.firstName} locale={locale} />
              <Field name="lastName" label={t.last} defaultValue={editEmployee?.lastName} error={fieldErrors.lastName} locale={locale} />
              <Field name="middleName" label={t.middle} defaultValue={editEmployee?.middleName || ""} error={fieldErrors.middleName} locale={locale} />
              <Field
                name="passport"
                label={t.passport}
                pattern="[A-Za-z]{2} ?[0-9]{7}"
                placeholder="AA 1234567"
                error={fieldErrors.passport}
                locale={locale}
                defaultValue={editEmployee?.passport}
              />
              <Field name="permanentAddress" label={t.residence} defaultValue={editEmployee?.permanentAddress} minLength={5} error={fieldErrors.permanentAddress} locale={locale} />
              <PhoneField label={t.phone} defaultValue={editEmployee?.phoneLocal} error={fieldErrors.phoneLocal} locale={locale} />
              <Field
                name="publicEmail"
                label={t.email}
                type="email"
                placeholder="doctor@clinic.uz"
                required={false}
                error={fieldErrors.publicEmail}
                locale={locale}
                defaultValue={editEmployee?.publicEmail || ""}
              />
              <Field
                name="telegramUrl"
                label={t.telegram}
                placeholder="@username или https://t.me/username"
                required={false}
                error={fieldErrors.telegramUrl}
                locale={locale}
                defaultValue={editEmployee?.telegramUrl || ""}
              />
              <Field name="login" label={t.login} defaultValue={editEmployee?.login} error={fieldErrors.login} locale={locale} />
              <Field
                name="password"
                label={t.password}
                type="password"
                minLength={12}
                error={fieldErrors.password}
                locale={locale}
                required={!editEmployee}
              />
              <Field name="position" label={t.position} defaultValue={editEmployee?.position} error={fieldErrors.position} locale={locale} />
              <Field name="specialty" label={t.specialty} defaultValue={editEmployee?.specialty || ""} required={false} />
              <label className={fieldErrors.branchId ? "field-invalid" : undefined}>
                {t.branch}
                <select name="branchId" required defaultValue={editEmployee?.branchId || ""}>
                  <option value="">—</option>
                  {branches.map((x) => (
                    <option key={x.id} value={x.id}>
                      {x.name}
                    </option>
                  ))}
                </select>
                {fieldErrors.branchId && <small className="field-error">{fieldErrorText(fieldErrors.branchId, locale)}</small>}
              </label>
              <Choice
                name="roleIds"
                label={t.rolesLabel}
                required
                error={fieldErrors.roleIds}
                locale={locale}
                selected={editEmployee?.roleIds}
                items={roles.map((x) => ({ id: x.id, name: x.name }))}
              />
            </>
          )}
          <div className="form-actions">
            <button
              type="button"
              className="secondary"
              onClick={() => {
                setOpen(false);
                setEditRole(null);
                setEditBranch(null);
                setEditEmployee(null);
              }}
            >
              {t.cancel}
            </button>
            <button className="primary">{t.save}</button>
          </div>
        </form>
      )}
      {error && <div className="management-error">{error}</div>}
      <div className="management-panel">
        {loading ? (
          <div className="table-empty">{t.loading}</div>
        ) : items.length === 0 ? (
          <div className="table-empty">{t.empty}</div>
        ) : (
          <div className="admin-list">
            {kind === "branches"
              ? (items as Branch[]).map((x) => (
                  <article key={x.id}>
                    <div>
                      <strong>{x.name}</strong>
                      <small>
                        {x.address || "—"} · {x.timezone}
                      </small>
                    </div>
                    <div className="role-actions">
                      <button
                        className="secondary"
                        onClick={() => {
                          setEditBranch(x);
                          setOpen(true);
                          window.scrollTo({ top: 0, behavior: "smooth" });
                        }}
                      >
                        {t.edit}
                      </button>
                      <button
                        className="danger-button"
                        onClick={() => void removeBranch(x.id)}
                      >
                        {t.delete}
                      </button>
                    </div>
                  </article>
                ))
              : kind === "employees"
                ? (items as Employee[]).map((x) => (
                    <article key={x.id}>
                      <div className="profile-list-main">
                        <span className="profile-list-avatar" style={x.hasPhoto?{backgroundImage:`url(/api/clinic/employees/${x.id}/photo)`}:undefined}>{!x.hasPhoto&&`${x.firstName[0]||""}${x.lastName[0]||""}`}</span>
                        <div>
                        <strong>
                          {x.firstName} {x.lastName}
                        </strong>
                        <small>
                          @{x.login} · {x.position}
                          {x.specialty ? ` · ${x.specialty}` : ""} ·{" "}
                          {x.branch || "—"}
                        </small>
                        </div>
                      </div>
                      <div className="role-actions">
                        <span className={x.isActive ? "status-active" : "system-badge"}>{x.isActive ? t.active : (locale === "ru" ? "Неактивен" : locale === "uz" ? "Nofaol" : "Inactive")}</span>
                        <button className="secondary" onClick={() => { setEditEmployee(x); setOpen(true); window.scrollTo({ top: 0, behavior: "smooth" }); }}>{t.edit}</button>
                        {!x.isOwner && <button className="secondary" onClick={() => void changeEmployeeStatus(x)}>{x.isActive ? (locale === "ru" ? "Отключить" : locale === "uz" ? "O‘chirish" : "Disable") : (locale === "ru" ? "Активировать" : locale === "uz" ? "Faollashtirish" : "Activate")}</button>}
                        {!x.isOwner && <button className="danger-button" onClick={() => void removeEmployee(x)}>{t.delete}</button>}
                      </div>
                    </article>
                  ))
                : (items as Role[]).map((x) => (
                    <article key={x.id}>
                      <div>
                        <strong>{x.name}</strong>
                        <small>
                          {x.employeeCount} {t.people} ·{" "}
                          {(Array.isArray(x.permissions) ? x.permissions : [])
                            .map(
                              (code) => permissionNames[locale][code] || code,
                            )
                            .join(", ") || "—"}
                        </small>
                      </div>
                      <div className="role-actions">
                        {x.code === "OWNER" ? (
                          <span className="system-badge">{t.system}</span>
                        ) : (
                          <>
                            <button
                              className="secondary"
                              onClick={() => {
                                setEditRole(x);
                                setOpen(true);
                                window.scrollTo({ top: 0, behavior: "smooth" });
                              }}
                            >
                              {t.edit}
                            </button>
                            <button
                              className="danger-button"
                              onClick={() => void remove(x.id)}
                            >
                              {t.delete}
                            </button>
                          </>
                        )}
                      </div>
                    </article>
                  ))}
          </div>
        )}
      </div>
    </section>
  );
}

function Field({
  name,
  label,
  type = "text",
  required = true,
  minLength = 2,
  defaultValue,
  pattern,
  placeholder,
  error,
  locale = "ru",
}: {
  name: string;
  label: string;
  type?: string;
  required?: boolean;
  minLength?: number;
  defaultValue?: string | undefined;
  pattern?: string | undefined;
  placeholder?: string | undefined;
  error?: string | undefined;
  locale?: Locale;
}) {
  return (
    <label className={error ? "field-invalid" : undefined}>
      {label}
      <input
        name={name}
        type={type}
        required={required}
        minLength={minLength}
        defaultValue={defaultValue}
        pattern={pattern}
        placeholder={placeholder}
        aria-invalid={Boolean(error)}
      />
      {error && <small className="field-error">{fieldErrorText(error, locale)}</small>}
    </label>
  );
}
function PhoneField({ label, error, locale, defaultValue }: { label: string; error?: string | undefined; locale: Locale; defaultValue?: string | undefined }) {
  return (
    <label className={error ? "field-invalid" : undefined}>
      {label}
      <span className="phone-input">
        <span aria-hidden="true">+998</span>
        <input
          name="phoneLocal"
          type="tel"
          inputMode="numeric"
          pattern="[0-9]{9}"
          maxLength={9}
          placeholder="901234567"
          required
          defaultValue={defaultValue}
        />
      </span>
      {error && <small className="field-error">{fieldErrorText(error, locale)}</small>}
    </label>
  );
}
function Choice({
  name,
  label,
  items,
  selected = [],
  error,
  locale = "ru",
}: {
  name: string;
  label: string;
  items: { id: string; name: string }[];
  required?: boolean;
  selected?: string[] | undefined;
  error?: string | undefined;
  locale?: Locale;
}) {
  return (
    <fieldset className="choice-field">
      <legend>{label}</legend>
      <div>
        {items.map((x) => (
          <label key={x.id}>
            <input
              type="checkbox"
              name={name}
              value={x.id}
              defaultChecked={Array.isArray(selected) && selected.includes(x.id)}
            />
            <span>{x.name}</span>
          </label>
        ))}
      </div>
      {error && <small className="field-error">{fieldErrorText(error, locale)}</small>}
    </fieldset>
  );
}

function fieldErrorText(code: string, locale: Locale) {
  const messages: Record<Locale, Record<string, string>> = {
    ru: { required: "Заполните это поле", too_short: "Значение слишком короткое", invalid_login: "Логин: 3–64 латинских букв, цифр или . _ -", login_exists: "Этот логин уже занят", invalid_passport: "Формат паспорта: AA 1234567", invalid_phone: "Введите 9 цифр после +998", invalid_email: "Введите корректный адрес почты", invalid_telegram: "Введите @username или https://t.me/username" },
    uz: { required: "Maydonni to‘ldiring", too_short: "Qiymat juda qisqa", invalid_login: "Login: 3–64 lotin harfi, raqam yoki . _ -", login_exists: "Bu login band", invalid_passport: "Pasport formati: AA 1234567", invalid_phone: "+998 dan keyin 9 raqam kiriting", invalid_email: "To‘g‘ri email kiriting", invalid_telegram: "@username yoki https://t.me/username kiriting" },
    en: { required: "This field is required", too_short: "Value is too short", invalid_login: "Use 3–64 letters, numbers, or . _ -", login_exists: "This login is already used", invalid_passport: "Passport format: AA 1234567", invalid_phone: "Enter 9 digits after +998", invalid_email: "Enter a valid email", invalid_telegram: "Enter @username or https://t.me/username" },
  };
  return messages[locale][code] || messages[locale].required;
}
