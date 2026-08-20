"use client";
import { FormEvent, useEffect, useState } from "react";
import { Locale } from "@/lib/i18n";
type Branch = { id: string; name: string };
type Patient = {
  id: string;
  firstName: string;
  lastName: string;
  middleName?: string | null;
  birthDate: string;
  gender: "female" | "male";
  branch?: string | null;
  hasPhoto?: boolean;
};
type PatientDetails = Patient & {
  phoneLocal: string;
  passport?: string | null;
  permanentAddress: string;
  guardianName?: string | null;
  guardianPhoneLocal?: string | null;
  telegramUrl?: string | null;
  homeBranchId: string;
  notes?: string | null;
};

async function profilePhotoPayload(file:File){
  if(file.size>5*1024*1024)throw new Error("Фотография больше 5 МБ");
  if(!["image/jpeg","image/png","image/webp"].includes(file.type))throw new Error("Разрешены JPG, PNG или WEBP");
  const bytes=new Uint8Array(await file.arrayBuffer());let binary="";for(let i=0;i<bytes.length;i+=32768)binary+=String.fromCharCode(...bytes.subarray(i,i+32768));
  return{contentType:file.type,data:btoa(binary)};
}
function ProfilePhotoInput({existingUrl}:{existingUrl?:string|undefined}){const[preview,setPreview]=useState(existingUrl||"");return <label className="profile-photo-field"><span className="profile-photo-preview" style={preview?{backgroundImage:`url(${preview})`}:undefined}>{!preview&&"Фото"}</span><span><strong>Фотография пациента</strong><small>JPG, PNG или WEBP · до 5 МБ</small><input type="file" name="profilePhoto" accept="image/jpeg,image/png,image/webp" onChange={e=>{const file=e.target.files?.[0];if(file)setPreview(URL.createObjectURL(file))}}/></span></label>}
type Diagnosis={id:string;name:string;icd10Code?:string|null;diagnosisType:"PRIMARY"|"SECONDARY";certainty:"PRELIMINARY"|"CONFIRMED";status:"ACTIVE"|"RESOLVED"|"RULED_OUT";diagnosedOn:string;notes?:string|null;author?:string|null;branch?:string|null};
type ClinicalOrder={id:string;diagnosisId:string;orderType:"MEDICATION"|"LAB"|"IMAGING"|"REFERRAL"|"PROCEDURE"|"FOLLOW_UP";title:string;dosage?:string|null;frequency?:string|null;durationDays?:number|null;instructions:string;startOn:string;endOn?:string|null;status:"ACTIVE"|"COMPLETED"|"CANCELLED";author?:string|null};
type Clinical={bloodGroup:string;heightCm?:number|null;weightKg?:number|null;allergies:{id:string;allergen:string;reaction?:string|null;severity:string;isActive:boolean}[];vaccinations:{id:string;name:string;administeredOn:string;dose?:string|null;batchNumber?:string|null;notes?:string|null}[];labResults:{id:string;type:"blood"|"urine";collectedOn:string;testName:string;result:string;unit?:string|null;referenceRange?:string|null;notes?:string|null;attachmentId?:string|null;attachmentName?:string|null}[];imagingStudies:{id:string;modality:"MRI"|"CT"|"XRAY";performedOn:string;bodyArea:string;diagnosis:string;conclusion?:string|null;notes?:string|null;fileName:string;author?:string|null}[];diagnoses:Diagnosis[];orders:ClinicalOrder[];history:{id:string;type:string;occurredAt:string;complaints?:string|null;diagnosis?:string|null;treatment?:string|null;notes?:string|null;author:string;branch?:string|null}[]};

function PatientClinicalOverview({patient,clinical,locale}:{patient:Patient;clinical:Clinical;locale:Locale}){
  const age=Math.max(0,new Date().getFullYear()-new Date(patient.birthDate).getFullYear());
  const latestLabs=clinical.labResults.slice(0,5),visits=clinical.history.slice(0,4);
  const diagnoses=(clinical.diagnoses??[]).filter(x=>x.status==="ACTIVE").slice(0,4);
  const orders=(clinical.orders??[]).filter(x=>x.status==="ACTIVE").slice(0,5);
  const scrollTo=(selector:string)=>document.querySelector(selector)?.scrollIntoView({behavior:"smooth",block:"start"});
  const date=(value:string)=>new Intl.DateTimeFormat(locale==="uz"?"uz-UZ":locale,{day:"2-digit",month:"short",year:"numeric"}).format(new Date(value));
  return <div className="patient-overview-grid">
    <aside className="patient-overview-column">
      <article className="patient-summary-card patient-identity-card">
        <div className="patient-avatar" style={patient.hasPhoto?{backgroundImage:`url(/api/clinic/patients/${patient.id}/photo)`,backgroundSize:"cover",backgroundPosition:"center"}:undefined}>{!patient.hasPhoto&&<>{patient.firstName.slice(0,1)}{patient.lastName.slice(0,1)}</>}</div>
        <div><small>ПАЦИЕНТ</small><h3>{patient.lastName} {patient.firstName}</h3><p>{patient.middleName||""}</p></div>
        <button aria-label="Изменить показатели" onClick={()=>scrollTo(".measurements-form")}>•••</button>
      </article>
      <article className="patient-summary-card"><header><h3>Основная информация</h3></header><div className="patient-info-grid"><Info label="Пол" value={patient.gender==="female"?"Женский":"Мужской"}/><Info label="Дата рождения" value={`${date(patient.birthDate)} · ${age} лет`}/><Info label="Филиал" value={patient.branch||"Не указан"}/><Info label="Группа крови" value={clinical.bloodGroup==="unknown"?"Не указана":clinical.bloodGroup}/></div></article>
      <article className="patient-summary-card"><header><h3>Показатели</h3><button onClick={()=>scrollTo(".measurements-form")}>Изменить</button></header><div className="vitals-grid"><Metric icon="↕" label="Рост" value={clinical.heightCm?`${clinical.heightCm} см`:"—"}/><Metric icon="◉" label="Вес" value={clinical.weightKg?`${clinical.weightKg} кг`:"—"}/><Metric icon="♥" label="Аллергии" value={String(clinical.allergies.filter(x=>x.isActive).length)}/><Metric icon="✚" label="Прививки" value={String(clinical.vaccinations.length)}/></div></article>
      <article className={`patient-summary-card allergy-summary ${clinical.allergies.some(x=>x.severity==="severe"&&x.isActive)?"has-severe":""}`}><header><h3>Аллергии</h3><button onClick={()=>scrollTo(".allergy-form")}>Добавить</button></header>{clinical.allergies.filter(x=>x.isActive).slice(0,4).map(x=><div className="overview-list-row" key={x.id}><strong>{x.allergen}</strong><span>{x.reaction||"Реакция не указана"}</span></div>)}{!clinical.allergies.some(x=>x.isActive)&&<p className="overview-empty">Аллергии не зарегистрированы</p>}</article>
    </aside>
    <div className="patient-overview-column patient-overview-main">
      <article className="patient-summary-card"><header><div><small>РЕЗУЛЬТАТЫ</small><h3>Последние анализы</h3></div><button onClick={()=>scrollTo(".blood-panel")}>Все анализы ↓</button></header>{latestLabs.map(x=><div className="overview-list-row overview-lab" key={x.id}><div><strong>{x.testName}</strong><span>{date(x.collectedOn)} · {x.type==="blood"?"Кровь":"Моча"}</span></div><b>{x.result} {x.unit||""}</b>{x.attachmentId&&<a href={`/api/clinic/lab-attachments/${x.attachmentId}`} target="_blank" rel="noreferrer">↗</a>}</div>)}{!latestLabs.length&&<p className="overview-empty">Результатов анализов пока нет</p>}</article>
      <article className="patient-summary-card visits-summary"><header><div><small>ХРОНОЛОГИЯ</small><h3>История посещений</h3></div><button onClick={()=>scrollTo(".history-list")}>Вся история</button></header>{visits.map(x=><div className="visit-row" key={x.id}><time>{date(x.occurredAt)}</time><div><strong>{x.diagnosis||x.complaints||"Посещение"}</strong><span>{x.complaints||x.treatment||"Без дополнительного описания"}</span></div><small>{x.author}{x.branch?` · ${x.branch}`:""}</small></div>)}{!visits.length&&<p className="overview-empty">Посещений пока нет</p>}</article>
    </div>
    <aside className="patient-overview-column">
      <article className="patient-summary-card diagnosis-summary"><header><div><small>МЕДИЦИНСКАЯ ИСТОРИЯ</small><h3>Активные диагнозы</h3></div><span className="overview-count">{diagnoses.length}</span></header>{diagnoses.map(x=><div className="overview-list-row" key={x.id}><div><strong>{x.name}</strong><span>{x.icd10Code||"Без кода МКБ"} · {date(x.diagnosedOn)}</span></div><b>{x.certainty==="CONFIRMED"?"Подтверждён":"Предварительный"}</b></div>)}{!diagnoses.length&&<p className="overview-empty">Активных диагнозов нет</p>}</article>
      <article className="patient-summary-card treatment-summary"><header><div><small>ПЛАН ЛЕЧЕНИЯ</small><h3>Текущие назначения</h3></div><span className="overview-count">{orders.length}</span></header>{orders.map(x=><div className="overview-list-row" key={x.id}><div><strong>{x.title}</strong><span>{[x.dosage,x.frequency,x.durationDays?`${x.durationDays} дней`:null].filter(Boolean).join(" · ")||x.instructions}</span></div><b>{x.orderType==="MEDICATION"?"Rx":"→"}</b></div>)}{!orders.length&&<p className="overview-empty">Активных назначений нет</p>}</article>
      <article className="patient-summary-card"><header><h3>Прививки</h3><button onClick={()=>scrollTo(".record-form")}>Добавить</button></header>{clinical.vaccinations.slice(0,3).map(x=><div className="overview-list-row" key={x.id}><strong>{x.name}</strong><span>{date(x.administeredOn)}{x.dose?` · ${x.dose}`:""}</span></div>)}{!clinical.vaccinations.length&&<p className="overview-empty">Прививки не добавлены</p>}</article>
    </aside>
  </div>
}
function Info({label,value}:{label:string;value:string}){return <div><small>{label}</small><strong>{value}</strong></div>}
function Metric({icon,label,value}:{icon:string;label:string;value:string}){return <div className="vital-card"><i>{icon}</i><small>{label}</small><strong>{value}</strong></div>}
const urineTemplate=[
  ["Цвет мочи","Светлый соломенно-жёлтый"],["Прозрачность","Прозрачная"],["Плотность мочи","1010–1022 г/л"],["Реакция pH","Кислая, 4–7"],["Запах","Слабо выраженный, нерезкий"],["Белок","До 0,033 г/л"],["Глюкоза","До 0,8 ммоль/л"],["Кетоновые тела","Отсутствуют"],["Билирубин","Не обнаруживается"],["Гемоглобин","Не обнаруживается"],["Лейкоциты","До 3 у мужчин, до 6 у женщин в поле зрения"],["Эритроциты","У мужчин нет, у женщин до 2–3 в поле зрения"],["Эпителий","До 8–10"],["Цилиндры","Не обнаружены"],["Соли","Отсутствуют"],["Бактерии и нитраты","Не обнаруживаются"],["Грибок","Отсутствует"],
] as const;
const bloodTemplate=[
  ["Гемоглобин","Hb","г/л","130–160","120–140"],["Гематокрит","HCT","","0,40–0,48","0,36–0,42"],["Эритроциты","RBC","×10¹²/л","4,0–5,1","3,7–4,7"],["Цветовой показатель","MCHC","","0,86–1,05","0,86–1,05"],["Ретикулоциты","RTC","%","0,2–1,2","0,2–1,2"],["СОЭ","ESR","мм/ч","1–10","2–16"],["Лейкоциты","WBC","×10⁹/л","4,0–8,8","4,0–8,8"],["Палочкоядерные","","","1–6%; 0,04–0,3 ×10⁹/л","1–6%; 0,04–0,3 ×10⁹/л"],["Сегментоядерные","","","47–72%; 2,0–5,5 ×10⁹/л","47–72%; 2,0–5,5 ×10⁹/л"],["Эозинофилы","EOS","","0,5–5%; 0,02–0,3 ×10⁹/л","0,5–5%; 0,02–0,3 ×10⁹/л"],["Базофилы","BAS","","0–1%; 0,00–0,065 ×10⁹/л","0–1%; 0,00–0,065 ×10⁹/л"],["Лимфоциты","LYM","","19–87%; 1,2–3,0 ×10⁹/л","19–87%; 1,2–3,0 ×10⁹/л"],["Моноциты","MON","","3–11%; 0,09–0,6 ×10⁹/л","3–11%; 0,09–0,6 ×10⁹/л"],["Тромбоциты","PLT","","Не указано в шаблоне","Не указано в шаблоне"],
] as const;

async function prepareLabAttachment(file:File){
  if(file.size>10*1024*1024)throw new Error("Файл больше 10 МБ");
  if(!["image/jpeg","image/png","image/webp","application/pdf"].includes(file.type))throw new Error("Разрешены JPG, PNG, WEBP или PDF");
  let quality="not_checked";
  if(file.type.startsWith("image/")){
    const bitmap=await createImageBitmap(file);if(bitmap.width<1000||bitmap.height<700){bitmap.close();throw new Error("Фото слишком маленькое. Сфотографируйте анализ заново ближе и чётче");}
    const canvas=document.createElement("canvas"),size=192;canvas.width=size;canvas.height=size;const ctx=canvas.getContext("2d",{willReadFrequently:true});if(!ctx){bitmap.close();throw new Error("Не удалось проверить фотографию");}
    ctx.drawImage(bitmap,0,0,size,size);bitmap.close();const px=ctx.getImageData(0,0,size,size).data;let light=0,detail=0,count=0;const gray=new Float32Array(size*size);
    for(let i=0;i<gray.length;i++){const value=.299*(px[i*4]??0)+.587*(px[i*4+1]??0)+.114*(px[i*4+2]??0);gray[i]=value;light+=value;}
    light/=gray.length;for(let y=1;y<size-1;y++)for(let x=1;x<size-1;x++){const i=y*size+x;detail+=Math.abs(4*(gray[i]??0)-(gray[i-1]??0)-(gray[i+1]??0)-(gray[i-size]??0)-(gray[i+size]??0));count++;}
    detail/=count;if(light<42)throw new Error("Фото слишком тёмное. Сфотографируйте анализ заново при хорошем освещении");if(light>238)throw new Error("Фото пересвечено. Сфотографируйте анализ заново без бликов");if(detail<10)throw new Error("Фото нечёткое или размытое. Сфотографируйте анализ заново");quality="accepted";
  }
  const bytes=new Uint8Array(await file.arrayBuffer());let binary="";for(let i=0;i<bytes.length;i+=32768)binary+=String.fromCharCode(...bytes.subarray(i,i+32768));
  return{name:file.name,contentType:file.type,data:btoa(binary),quality};
}

async function prepareImagingAttachment(file:File){
  const isDicom=file.name.toLowerCase().endsWith(".dcm")||file.type==="application/dicom";if(file.size>25*1024*1024)throw new Error("Файл больше 25 МБ");
  if(isDicom){const bytes=new Uint8Array(await file.arrayBuffer());let binary="";for(let i=0;i<bytes.length;i+=32768)binary+=String.fromCharCode(...bytes.subarray(i,i+32768));return{name:file.name,contentType:"application/dicom",data:btoa(binary),quality:"not_checked"};}
  return prepareLabAttachment(file);
}

function AnalysisFileInput(){return <label className="analysis-file"><strong>Файл или фотография анализа *</strong><input type="file" name="attachment" accept="image/jpeg,image/png,image/webp,application/pdf" required capture="environment"/><span>JPG, PNG, WEBP или PDF · до 10 МБ. Фотография автоматически проверяется на чёткость и освещение.</span></label>}
const c = {
  ru: {
    title: "Пациенты",
    sub: "Единый защищённый реестр пациентов клиники",
    add: "Добавить пациента",
    search: "Поиск по ФИО",
    first: "Имя",
    last: "Фамилия",
    middle: "Отчество",
    birth: "Дата рождения",
    gender: "Пол",
    female: "Женский",
    male: "Мужской",
    phone: "Номер телефона",
    passport: "Паспорт (необязательно)",
    address: "Постоянное место жительства",
    guardian: "Родитель или опекун",
    guardianPhone: "Телефон опекуна",
    telegram: "Telegram пациента",
    branch: "Основной филиал",
    notes: "Примечание",
    save: "Сохранить",
    cancel: "Отмена",
    empty: "Пациентов пока нет",
    error:
      "Проверьте данные. Возможно, пациент с таким телефоном уже существует.",
    invalid: "Проверьте формат полей. Паспорт: 2 латинские буквы и 7 цифр, например AD1234567.",
    duplicate: "Пациент с таким номером телефона уже существует.",
    backup: "Скачать резервную копию",
    backupPassword: "Пароль резервной копии (от 12 символов)",
    backupHint: "Этот пароль потребуется для восстановления. Он не сохраняется в системе.",
  },
  uz: {
    title: "Bemorlar",
    sub: "Klinikaning himoyalangan bemorlar reyestri",
    add: "Bemor qo‘shish",
    search: "F.I.Sh. bo‘yicha qidirish",
    first: "Ism",
    last: "Familiya",
    middle: "Otasining ismi",
    birth: "Tug‘ilgan sana",
    gender: "Jinsi",
    female: "Ayol",
    male: "Erkak",
    phone: "Telefon raqami",
    passport: "Pasport (ixtiyoriy)",
    address: "Doimiy yashash manzili",
    guardian: "Ota-ona yoki vasiy",
    guardianPhone: "Vasiy telefoni",
    telegram: "Bemor Telegram’i",
    branch: "Asosiy filial",
    notes: "Izoh",
    save: "Saqlash",
    cancel: "Bekor qilish",
    empty: "Bemorlar yo‘q",
    error:
      "Ma’lumotlarni tekshiring. Bu telefon bilan bemor mavjud bo‘lishi mumkin.",
    invalid: "Maydonlar formatini tekshiring. Pasport: 2 ta lotin harfi va 7 ta raqam, masalan AD1234567.",
    duplicate: "Bu telefon raqami bilan bemor allaqachon mavjud.",
    backup: "Zaxira nusxasini yuklab olish",
    backupPassword: "Zaxira nusxasi paroli (12+ belgi)",
    backupHint: "Tiklash uchun ushbu parol kerak. U tizimda saqlanmaydi.",
  },
  en: {
    title: "Patients",
    sub: "Secure unified patient registry",
    add: "Add patient",
    search: "Search by name",
    first: "First name",
    last: "Last name",
    middle: "Middle name",
    birth: "Date of birth",
    gender: "Gender",
    female: "Female",
    male: "Male",
    phone: "Phone number",
    passport: "Passport (optional)",
    address: "Permanent residence",
    guardian: "Parent or guardian",
    guardianPhone: "Guardian phone",
    telegram: "Patient Telegram",
    branch: "Home branch",
    notes: "Notes",
    save: "Save",
    cancel: "Cancel",
    empty: "No patients yet",
    error: "Check the data. A patient with this phone may already exist.",
    invalid: "Check the field formats. Passport: 2 Latin letters and 7 digits, for example AD1234567.",
    duplicate: "A patient with this phone number already exists.",
    backup: "Download encrypted backup",
    backupPassword: "Backup password (12+ characters)",
    backupHint: "You will need this password to restore. It is not stored by the system.",
  },
} as const;
export function PatientManagement({ locale }: { locale: Locale }) {
  const t = c[locale];
  const ct = {
    ru: { open: "Открыть карту", card: "Медицинская карта", blood: "Группа крови", allergies: "Аллергия на препараты", allergen: "Препарат или аллерген", reaction: "Реакция", severity: "Тяжесть", unknown: "Не указана", mild: "Лёгкая", moderate: "Средняя", severe: "Тяжёлая", addAllergy: "Добавить аллергию", visit: "Запись", date: "Дата", complaints: "Жалобы", diagnosis: "Диагноз", treatment: "Лечение и назначения", visitNotes: "Дополнительная запись", addVisit: "Сохранить", history: "История болезни", noAllergies: "Аллергии не зарегистрированы", noHistory: "Записей пока нет", close: "Закрыть" },
    uz: { open: "Kartani ochish", card: "Tibbiy karta", blood: "Qon guruhi", allergies: "Dorilarga allergiya", allergen: "Dori yoki allergen", reaction: "Reaksiya", severity: "Og‘irlik", unknown: "Ko‘rsatilmagan", mild: "Yengil", moderate: "O‘rtacha", severe: "Og‘ir", addAllergy: "Allergiya qo‘shish", visit: "Yangi tashrif", date: "Sana va vaqt", complaints: "Shikoyatlar", diagnosis: "Tashxis", treatment: "Davolash va tavsiyalar", visitNotes: "Qo‘shimcha qayd", addVisit: "Tashrifni saqlash", history: "Kasallik tarixi", noAllergies: "Allergiya qayd etilmagan", noHistory: "Tashriflar yo‘q", close: "Yopish" },
    en: { open: "Open record", card: "Medical record", blood: "Blood group", allergies: "Drug allergies", allergen: "Drug or allergen", reaction: "Reaction", severity: "Severity", unknown: "Unknown", mild: "Mild", moderate: "Moderate", severe: "Severe", addAllergy: "Add allergy", visit: "New visit", date: "Date and time", complaints: "Complaints", diagnosis: "Diagnosis", treatment: "Treatment and prescriptions", visitNotes: "Additional note", addVisit: "Save visit", history: "Medical history", noAllergies: "No allergies recorded", noHistory: "No visits yet", close: "Close" },
  }[locale];
  const mt = {
    ru:{height:"Рост (см)",weight:"Вес (кг)",saveProfile:"Сохранить показатели",vaccinations:"Прививки",vaccine:"Название вакцины",date:"Дата",dose:"Доза",batch:"Серия / партия",addVaccine:"Добавить прививку",noVaccines:"Прививки не добавлены",bloodTests:"Анализы крови",urineTests:"Анализы мочи",testName:"Название анализа",result:"Результат",unit:"Единица",reference:"Норма",addTest:"Добавить анализ",noTests:"Анализы не добавлены",saved:"Данные сохранены"},
    uz:{height:"Bo‘y (sm)",weight:"Vazn (kg)",saveProfile:"Ko‘rsatkichlarni saqlash",vaccinations:"Emlashlar",vaccine:"Vaksina nomi",date:"Sana",dose:"Doza",batch:"Seriya / partiya",addVaccine:"Emlash qo‘shish",noVaccines:"Emlashlar yo‘q",bloodTests:"Qon tahlillari",urineTests:"Siydik tahlillari",testName:"Tahlil nomi",result:"Natija",unit:"Birlik",reference:"Me’yor",addTest:"Tahlil qo‘shish",noTests:"Tahlillar yo‘q",saved:"Ma’lumotlar saqlandi"},
    en:{height:"Height (cm)",weight:"Weight (kg)",saveProfile:"Save measurements",vaccinations:"Vaccinations",vaccine:"Vaccine name",date:"Date",dose:"Dose",batch:"Batch number",addVaccine:"Add vaccination",noVaccines:"No vaccinations",bloodTests:"Blood tests",urineTests:"Urine tests",testName:"Test name",result:"Result",unit:"Unit",reference:"Reference range",addTest:"Add test",noTests:"No tests",saved:"Saved"}
  }[locale];
  const [items, setItems] = useState<Patient[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [error, setError] = useState("");
  const [backupOpen, setBackupOpen] = useState(false);
  const [selected,setSelected]=useState<Patient|null>(null);
  const [clinical,setClinical]=useState<Clinical|null>(null);
  const [bloodChoice,setBloodChoice]=useState("unknown");
  const [editing,setEditing]=useState<PatientDetails|null>(null);
  const load = async (q = "") => {
    const [r, b] = await Promise.all([
      fetch(`/api/clinic/patients?q=${encodeURIComponent(q)}`, {
        cache: "no-store",
      }),
      fetch("/api/clinic/branches", { cache: "no-store" }),
    ]);
    if (r.ok) setItems((await r.json()).items ?? []);
    if (b.ok) setBranches((await b.json()).items ?? []);
  };
  useEffect(() => {
    void load();
  }, []);
  const submit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    const form = e.currentTarget,
      f = new FormData(form);
    const profilePhoto=f.get("profilePhoto");
    const body = {
      firstName: f.get("firstName"),
      lastName: f.get("lastName"),
      middleName: f.get("middleName"),
      birthDate: f.get("birthDate"),
      gender: f.get("gender"),
      phoneLocal: f.get("phoneLocal"),
      passport: f.get("passport"),
      permanentAddress: f.get("permanentAddress"),
      guardianName: f.get("guardianName"),
      guardianPhoneLocal: f.get("guardianPhoneLocal"),
      telegramUrl: f.get("telegramUrl"),
      homeBranchId: f.get("homeBranchId"),
      notes: f.get("notes"),
    };
    const r = await fetch(editing?`/api/clinic/patients/${editing.id}`:"/api/clinic/patients", {
      method: editing?"PATCH":"POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!r.ok) {
      const response = await r.json().catch(() => null);
      setError(response?.error?.code === "patient_exists" ? t.duplicate : t.invalid);
      return;
    }
    const created = await r.json();
    const patientId=editing?.id??created.id;
    if(profilePhoto instanceof File&&profilePhoto.size){const photoResponse=await fetch(`/api/clinic/patients/${patientId}/photo`,{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify(await profilePhotoPayload(profilePhoto))});if(!photoResponse.ok){setError("Пациент сохранён, но фотографию загрузить не удалось");return}}
    const createdPatient: Patient = {
      id: patientId,
      firstName: String(body.firstName),
      lastName: String(body.lastName),
      middleName: body.middleName ? String(body.middleName) : null,
      birthDate: String(body.birthDate),
      gender: body.gender as "female" | "male",
      branch: branches.find((x) => x.id === String(body.homeBranchId))?.name ?? null,
      hasPhoto: editing?.hasPhoto||profilePhoto instanceof File&&profilePhoto.size>0,
    };
    form.reset();
    setOpen(false);
    setEditing(null);
    await load(query);
    await openClinical(createdPatient);
  };
  const startCreate=()=>{setEditing(null);setOpen(true);setError("")};
  const startEdit=async(patient:Patient)=>{setError("");const r=await fetch(`/api/clinic/patients/${patient.id}`,{cache:"no-store"});if(!r.ok){setError("Не удалось загрузить данные пациента");return};setEditing(await r.json());setOpen(true);requestAnimationFrame(()=>document.querySelector(".patient-management .admin-form")?.scrollIntoView({behavior:"smooth",block:"start"}))};
  const removePatient=async(patient:Patient)=>{if(!window.confirm(`Удалить пациента «${patient.lastName} ${patient.firstName}»?\n\nПациент исчезнет из рабочего реестра. Медицинские данные останутся в защищённом архиве согласно сроку хранения.`))return;setError("");const r=await fetch(`/api/clinic/patients/${patient.id}`,{method:"DELETE"});if(!r.ok){setError("Не удалось удалить пациента");return};if(selected?.id===patient.id){setSelected(null);setClinical(null)};await load(query)};
  const downloadBackup = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    const password = String(new FormData(e.currentTarget).get("backupPassword") ?? "");
    const response = await fetch("/api/clinic/backups/export", {method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify({password})});
    if(!response.ok){setError(t.error);return}
    const blob=await response.blob();const disposition=response.headers.get("content-disposition")??"";const match=disposition.match(/filename="([^"]+)"/);const link=document.createElement("a");link.href=URL.createObjectURL(blob);link.download=match?.[1]??"ona-va-bola-backup.ovbk";link.click();URL.revokeObjectURL(link.href);setBackupOpen(false);
  };
  const openClinical=async(patient:Patient)=>{setSelected(patient);const r=await fetch(`/api/clinic/patients/${patient.id}/clinical`,{cache:"no-store"});if(r.ok){const data=await r.json();setClinical(data);setBloodChoice(data.bloodGroup??"unknown")}};
  const saveBloodChoice=async(value:string)=>{if(!selected)return;setBloodChoice(value);const r=await fetch(`/api/clinic/patients/${selected.id}/blood-group`,{method:"PUT",headers:{"content-type":"application/json"},body:JSON.stringify({bloodGroup:value})});if(r.ok)setClinical(x=>x?{...x,bloodGroup:value}:x)};
  const saveProfile=async(e:FormEvent<HTMLFormElement>)=>{e.preventDefault();if(!selected)return;const f=new FormData(e.currentTarget);const r=await fetch(`/api/clinic/patients/${selected.id}/clinical-profile`,{method:"PUT",headers:{"content-type":"application/json"},body:JSON.stringify({bloodGroup:bloodChoice,heightCm:Number(f.get("heightCm")),weightKg:Number(f.get("weightKg"))})});if(r.ok){const data=await r.json();setClinical(x=>x?{...x,...data}:x)}};
  const addAllergy=async(e:FormEvent<HTMLFormElement>)=>{e.preventDefault();if(!selected)return;const form=e.currentTarget;const f=new FormData(form);const r=await fetch(`/api/clinic/patients/${selected.id}/allergies`,{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify({allergen:f.get("allergen"),reaction:f.get("reaction"),severity:f.get("severity")})});if(r.ok){const item=await r.json();setClinical(x=>x?{...x,allergies:[item,...x.allergies]}:x);form.reset()}};
  const addVaccination=async(e:FormEvent<HTMLFormElement>)=>{e.preventDefault();if(!selected)return;const form=e.currentTarget;const f=new FormData(form);const r=await fetch(`/api/clinic/patients/${selected.id}/vaccinations`,{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify({name:f.get("name"),administeredOn:f.get("administeredOn"),dose:f.get("dose"),batchNumber:f.get("batchNumber"),notes:f.get("notes")})});if(r.ok){const item=await r.json();setClinical(x=>x?{...x,vaccinations:[item,...x.vaccinations]}:x);form.reset()}};
  const addBloodPanel=async(e:FormEvent<HTMLFormElement>)=>{e.preventDefault();if(!selected)return;setError("");const form=e.currentTarget;const f=new FormData(form);try{const file=f.get("attachment");if(!(file instanceof File)||!file.size)throw new Error("Добавьте файл или фотографию анализа");const attachment=await prepareLabAttachment(file);const items=bloodTemplate.map(([name,abbr,unit,male,female],index)=>({testName:abbr?`${name} (${abbr})`:name,result:String(f.get(`blood-${index}`)??"").trim(),unit,referenceRange:selected.gender==="male"?male:female})).filter(x=>x.result);const r=await fetch(`/api/clinic/patients/${selected.id}/blood-panel`,{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify({collectedOn:f.get("collectedOn"),items,attachment})});if(!r.ok)throw new Error("Не удалось сохранить анализ. Проверьте дату, результаты и файл");const data=await r.json();setClinical(x=>x?{...x,labResults:[...data.items,...x.labResults]}:x);form.reset()}catch(reason){setError(reason instanceof Error?reason.message:"Не удалось проверить файл")}};
  const addUrinePanel=async(e:FormEvent<HTMLFormElement>)=>{e.preventDefault();if(!selected)return;setError("");const form=e.currentTarget;const f=new FormData(form);try{const file=f.get("attachment");if(!(file instanceof File)||!file.size)throw new Error("Добавьте файл или фотографию анализа");const attachment=await prepareLabAttachment(file);const items=urineTemplate.map(([testName,referenceRange],index)=>({testName,result:String(f.get(`urine-${index}`)??"").trim(),unit:"",referenceRange})).filter(x=>x.result);const r=await fetch(`/api/clinic/patients/${selected.id}/urine-panel`,{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify({collectedOn:f.get("collectedOn"),items,attachment})});if(!r.ok)throw new Error("Не удалось сохранить анализ. Проверьте дату, результаты и файл");const data=await r.json();setClinical(x=>x?{...x,labResults:[...data.items,...x.labResults]}:x);form.reset()}catch(reason){setError(reason instanceof Error?reason.message:"Не удалось проверить файл")}};
  const addImagingStudy=async(e:FormEvent<HTMLFormElement>)=>{e.preventDefault();if(!selected)return;setError("");const form=e.currentTarget;const f=new FormData(form);try{const file=f.get("attachment");if(!(file instanceof File)||!file.size)throw new Error("Прикрепите файл исследования");const attachment=await prepareImagingAttachment(file);const body={modality:f.get("modality"),performedOn:f.get("performedOn"),bodyArea:f.get("bodyArea"),diagnosis:f.get("diagnosis"),conclusion:f.get("conclusion"),notes:f.get("notes"),attachment};const r=await fetch(`/api/clinic/patients/${selected.id}/imaging`,{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify(body)});if(!r.ok)throw new Error("Не удалось сохранить исследование. Проверьте обязательные поля и файл");const item=await r.json();setClinical(x=>x?{...x,imagingStudies:[item,...(x.imagingStudies??[])]}:x);form.reset()}catch(reason){setError(reason instanceof Error?reason.message:"Не удалось обработать файл исследования")}};
  const addDiagnosis=async(e:FormEvent<HTMLFormElement>)=>{e.preventDefault();if(!selected)return;setError("");const form=e.currentTarget,f=new FormData(form);const r=await fetch(`/api/clinic/patients/${selected.id}/diagnoses`,{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify({name:f.get("name"),icd10Code:f.get("icd10Code"),diagnosisType:f.get("diagnosisType"),certainty:f.get("certainty"),diagnosedOn:f.get("diagnosedOn"),notes:f.get("notes"),branchId:f.get("branchId")})});if(!r.ok){setError("Не удалось сохранить диагноз. Проверьте обязательные поля");return};const item=await r.json();setClinical(x=>x?{...x,diagnoses:[item,...(x.diagnoses??[])]}:x);form.reset()};
  const changeDiagnosisStatus=async(id:string,status:Diagnosis["status"])=>{const r=await fetch(`/api/clinic/diagnoses/${id}/status`,{method:"PATCH",headers:{"content-type":"application/json"},body:JSON.stringify({status})});if(r.ok)setClinical(x=>x?{...x,diagnoses:(x.diagnoses??[]).map(d=>d.id===id?{...d,status}:d)}:x)};
  const addClinicalOrder=async(e:FormEvent<HTMLFormElement>)=>{e.preventDefault();if(!selected)return;setError("");const form=e.currentTarget,f=new FormData(form);const duration=String(f.get("durationDays")??"");const r=await fetch(`/api/clinic/patients/${selected.id}/orders`,{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify({diagnosisId:f.get("diagnosisId"),branchId:f.get("branchId"),orderType:f.get("orderType"),title:f.get("title"),dosage:f.get("dosage"),frequency:f.get("frequency"),durationDays:duration?Number(duration):null,instructions:f.get("instructions"),startOn:f.get("startOn"),endOn:f.get("endOn")})});if(!r.ok){setError("Не удалось сохранить назначение. Выберите диагноз и проверьте поля");return};const item=await r.json();setClinical(x=>x?{...x,orders:[item,...(x.orders??[])]}:x);form.reset()};
  const changeOrderStatus=async(id:string,status:ClinicalOrder["status"])=>{const r=await fetch(`/api/clinic/orders/${id}/status`,{method:"PATCH",headers:{"content-type":"application/json"},body:JSON.stringify({status})});if(r.ok)setClinical(x=>x?{...x,orders:(x.orders??[]).map(o=>o.id===id?{...o,status}:o)}:x)};
  return (
    <section className="management patient-management">
      <div className="management-head">
        <div>
          <p>
            {locale === "ru"
              ? "Реестр"
              : locale === "uz"
                ? "Reyestr"
                : "Registry"}
          </p>
          <h1>{t.title}</h1>
          <span>{t.sub}</span>
        </div>
        <div className="head-buttons"><button className="secondary" onClick={() => setBackupOpen(!backupOpen)}>{t.backup}</button><button className="primary" onClick={startCreate}>+ {t.add}</button></div>
      </div>
      {backupOpen&&<form className="backup-form" onSubmit={downloadBackup}><label>{t.backupPassword}<input name="backupPassword" type="password" minLength={12} required autoComplete="new-password"/></label><small>{t.backupHint}</small><button className="primary">{t.backup}</button></form>}
      <div className="patient-search">
        <input
          value={query}
          placeholder={t.search}
          onChange={(e) => {
            setQuery(e.target.value);
            void load(e.target.value);
          }}
        />
      </div>
      {open && (
        <form key={editing?.id??"new"} className="admin-form" onSubmit={submit}>
          <ProfilePhotoInput existingUrl={editing?.hasPhoto?`/api/clinic/patients/${editing.id}/photo`:undefined}/>
          <Field name="firstName" label={t.first} defaultValue={editing?.firstName}/>
          <Field name="lastName" label={t.last} defaultValue={editing?.lastName}/>
          <Field name="middleName" label={t.middle} required={false} defaultValue={editing?.middleName??""}/>
          <Field name="birthDate" label={t.birth} type="date" defaultValue={editing?.birthDate}/>
          <label>
            {t.gender}
            <select name="gender" required defaultValue={editing?.gender??"female"}>
              <option value="female">{t.female}</option>
              <option value="male">{t.male}</option>
            </select>
          </label>
          <Phone name="phoneLocal" label={t.phone} defaultValue={editing?.phoneLocal}/>
          <Field
            name="passport"
            label={t.passport}
            required={false}
            placeholder="AD1234567"
            pattern="[A-Za-z]{2}[0-9]{7}"
            title={locale === "ru" ? "2 латинские буквы и 7 цифр" : "2 Latin letters and 7 digits"}
            defaultValue={editing?.passport??""}
          />
          <Field name="permanentAddress" label={t.address} defaultValue={editing?.permanentAddress}/>
          <Field name="guardianName" label={t.guardian} required={false} defaultValue={editing?.guardianName??""}/>
          <Phone
            name="guardianPhoneLocal"
            label={t.guardianPhone}
            required={false}
            defaultValue={editing?.guardianPhoneLocal??""}
          />
          <Field
            name="telegramUrl"
            label={t.telegram}
            required={false}
            placeholder="@username или https://t.me/username"
            defaultValue={editing?.telegramUrl??""}
          />
          <label>
            {t.branch}
            <select name="homeBranchId" required defaultValue={editing?.homeBranchId??""}>
              <option value="">—</option>
              {branches.map((x) => (
                <option key={x.id} value={x.id}>
                  {x.name}
                </option>
              ))}
            </select>
          </label>
          <Field name="notes" label={t.notes} required={false} defaultValue={editing?.notes??""}/>
          <div className="form-actions">
            <button
              type="button"
              className="secondary"
              onClick={() => {setOpen(false);setEditing(null)}}
            >
              {t.cancel}
            </button>
            <button className="primary">{editing?"Сохранить изменения":t.save}</button>
          </div>
        </form>
      )}
      {error && <div className="management-error">{error}</div>}
      {selected && clinical && (
        <section className="clinical-panel">
          <div className="clinical-title"><div><small>{ct.card}</small><h2>{selected.lastName} {selected.firstName} {selected.middleName || ""}</h2><span className="clinical-subtitle">Единый обзор состояния пациента и медицинской истории</span></div><button className="secondary" onClick={() => { setSelected(null); setClinical(null); }}>{ct.close}</button></div>
          <PatientClinicalOverview patient={selected} clinical={clinical} locale={locale}/>
          <div className="clinical-details-title"><small>ПОДРОБНЫЕ ДАННЫЕ И РЕДАКТИРОВАНИЕ</small><h2>Медицинские сведения</h2></div>
          <div className="clinical-grid">
            <div className="clinical-card clinical-wide"><h3>{ct.blood}</h3><div className="blood-options"><button type="button" className={bloodChoice==="unknown"?"active":""} onClick={()=>void saveBloodChoice("unknown")}>{ct.unknown}</button>{["O+","O-","A+","A-","B+","B-","AB+","AB-"].map(x=><button type="button" key={x} className={bloodChoice===x?"active":""} onClick={()=>void saveBloodChoice(x)}>{x}</button>)}</div></div>
            <div className="clinical-card clinical-wide"><h3>{mt.height} · {mt.weight}</h3><form className="profile-form measurements-form" onSubmit={saveProfile}><Field name="heightCm" label={mt.height} type="number" defaultValue={clinical.heightCm??""} min="30" max="250" step="0.1"/><Field name="weightKg" label={mt.weight} type="number" defaultValue={clinical.weightKg??""} min="0.5" max="500" step="0.1"/><button className="primary">{mt.saveProfile}</button></form></div>
            <div className="clinical-card clinical-wide"><h3>{ct.allergies}</h3><form className="clinical-form allergy-form" onSubmit={addAllergy}><Field name="allergen" label={ct.allergen} /><Field name="reaction" label={ct.reaction} required={false} /><label>{ct.severity}<select name="severity" defaultValue="unknown"><option value="unknown">{ct.unknown}</option><option value="mild">{ct.mild}</option><option value="moderate">{ct.moderate}</option><option value="severe">{ct.severe}</option></select></label><button className="primary">{ct.addAllergy}</button></form><div className="allergy-list">{clinical.allergies.length ? clinical.allergies.map(x => <span key={x.id} className={`allergy-${x.severity}`}><strong>{x.allergen}</strong>{x.reaction ? ` — ${x.reaction}` : ""}</span>) : <small>{ct.noAllergies}</small>}</div></div>
            <div className="clinical-card clinical-wide"><h3>{mt.vaccinations}</h3><form className="clinical-form record-form" onSubmit={addVaccination}><Field name="name" label={mt.vaccine}/><Field name="administeredOn" label={mt.date} type="date"/><Field name="dose" label={mt.dose} required={false}/><Field name="batchNumber" label={mt.batch} required={false}/><Field name="notes" label={t.notes} required={false}/><button className="primary">{mt.addVaccine}</button></form><div className="medical-record-list">{clinical.vaccinations.length?clinical.vaccinations.map(x=><article key={x.id}><strong>{x.name}</strong><span>{x.administeredOn}{x.dose?` · ${x.dose}`:""}{x.batchNumber?` · ${x.batchNumber}`:""}</span>{x.notes&&<small>{x.notes}</small>}</article>):<small>{mt.noVaccines}</small>}</div></div>
            <div className="clinical-card clinical-wide blood-panel"><h3>{mt.bloodTests}</h3><form onSubmit={addBloodPanel}><label className="urine-date">{mt.date}<input type="date" name="collectedOn" required/></label><div className="blood-table"><div className="blood-head"><strong>Показатель</strong><strong>Обозначение</strong><strong>Единицы</strong><strong>Результат пациента</strong><strong>Норма: мужчины</strong><strong>Норма: женщины</strong></div>{bloodTemplate.map(([name,abbr,unit,male,female],index)=><div className="blood-row" key={name}><strong>{name}</strong><span>{abbr||"—"}</span><span>{unit||"—"}</span><input name={`blood-${index}`} placeholder="Результат"/><span>{male}</span><span>{female}</span></div>)}</div><div className="analysis-save-row"><AnalysisFileInput/><button className="primary panel-save">Сохранить анализ крови</button></div></form><div className="medical-record-list">{clinical.labResults.filter(x=>x.type==="blood").length?clinical.labResults.filter(x=>x.type==="blood").map(x=><article key={x.id}><strong>{x.testName}: {x.result} {x.unit||""}</strong><span>{x.collectedOn}{x.referenceRange?` · ${mt.reference}: ${x.referenceRange}`:""}</span>{x.attachmentId&&<a className="analysis-document" href={`/api/clinic/lab-attachments/${x.attachmentId}`} target="_blank" rel="noreferrer">Открыть файл: {x.attachmentName||"анализ"}</a>}</article>):<small>{mt.noTests}</small>}</div></div>
            <div className="clinical-card clinical-wide urine-panel"><h3>{mt.urineTests}</h3><form onSubmit={addUrinePanel}><label className="urine-date">{mt.date}<input type="date" name="collectedOn" required/></label><div className="urine-table"><div className="urine-head"><strong>Показатель</strong><strong>Результат пациента</strong><strong>Расшифровка / ориентир</strong></div>{urineTemplate.map(([name,reference],index)=><div className="urine-row" key={name}><strong>{name}</strong><input name={`urine-${index}`} placeholder="Введите результат"/><span>{reference}</span></div>)}</div><div className="analysis-save-row"><AnalysisFileInput/><button className="primary urine-save">Сохранить анализ мочи</button></div></form><div className="medical-record-list">{clinical.labResults.filter(x=>x.type==="urine").length?clinical.labResults.filter(x=>x.type==="urine").map(x=><article key={x.id}><strong>{x.testName}: {x.result}</strong><span>{x.collectedOn}{x.referenceRange?` · ${x.referenceRange}`:""}</span>{x.attachmentId&&<a className="analysis-document" href={`/api/clinic/lab-attachments/${x.attachmentId}`} target="_blank" rel="noreferrer">Открыть файл: {x.attachmentName||"анализ"}</a>}</article>):<small>{mt.noTests}</small>}</div></div>
            <div className="clinical-card clinical-wide diagnosis-panel"><h3>Диагнозы</h3><form className="diagnosis-form" onSubmit={addDiagnosis}><Field name="name" label="Название диагноза"/><Field name="icd10Code" label="Код МКБ-10" required={false} placeholder="Например: J06.9"/><label>Тип диагноза<select name="diagnosisType" defaultValue="PRIMARY"><option value="PRIMARY">Основной</option><option value="SECONDARY">Сопутствующий</option></select></label><label>Подтверждение<select name="certainty" defaultValue="PRELIMINARY"><option value="PRELIMINARY">Предварительный</option><option value="CONFIRMED">Подтверждённый</option></select></label><Field name="diagnosedOn" label="Дата постановки" type="date"/><label>Филиал<select name="branchId"><option value="">—</option>{branches.map(x=><option key={x.id} value={x.id}>{x.name}</option>)}</select></label><Field name="notes" label="Комментарий врача" required={false}/><button className="primary">Добавить диагноз</button></form><div className="diagnosis-list">{(clinical.diagnoses??[]).length?(clinical.diagnoses??[]).map(d=><article key={d.id}><div><strong>{d.name}{d.icd10Code?` · ${d.icd10Code}`:""}</strong><span>{d.diagnosedOn} · {d.diagnosisType==="PRIMARY"?"Основной":"Сопутствующий"} · {d.certainty==="CONFIRMED"?"Подтверждён":"Предварительный"}</span>{d.notes&&<p>{d.notes}</p>}<small>{d.author||"—"}{d.branch?` · ${d.branch}`:""}</small></div><select value={d.status} onChange={e=>void changeDiagnosisStatus(d.id,e.target.value as Diagnosis["status"])}><option value="ACTIVE">Активный</option><option value="RESOLVED">Завершён</option><option value="RULED_OUT">Исключён</option></select></article>):<small>Диагнозы пока не добавлены</small>}</div></div>
            <div className="clinical-card clinical-wide order-panel"><h3>Назначения</h3><form className="order-form" onSubmit={addClinicalOrder}><label>Связанный диагноз<select name="diagnosisId" required defaultValue=""><option value="" disabled>Выберите диагноз</option>{(clinical.diagnoses??[]).filter(d=>d.status==="ACTIVE").map(d=><option key={d.id} value={d.id}>{d.name}{d.icd10Code?` (${d.icd10Code})`:""}</option>)}</select></label><label>Тип назначения<select name="orderType" defaultValue="MEDICATION"><option value="MEDICATION">Лекарство</option><option value="LAB">Лабораторный анализ</option><option value="IMAGING">МРТ / КТ / рентген</option><option value="REFERRAL">Консультация специалиста</option><option value="PROCEDURE">Процедура</option><option value="FOLLOW_UP">Повторный приём</option></select></label><Field name="title" label="Название назначения"/><Field name="dosage" label="Дозировка" required={false}/><Field name="frequency" label="Частота применения" required={false}/><Field name="durationDays" label="Длительность, дней" type="number" required={false} min="1" max="3650"/><Field name="startOn" label="Дата начала" type="date"/><Field name="endOn" label="Дата окончания" type="date" required={false}/><label>Филиал<select name="branchId"><option value="">—</option>{branches.map(x=><option key={x.id} value={x.id}>{x.name}</option>)}</select></label><Field name="instructions" label="Инструкция и рекомендации"/><button className="primary">Добавить назначение</button></form><div className="order-list">{(clinical.orders??[]).length?(clinical.orders??[]).map(o=>{const d=(clinical.diagnoses??[]).find(x=>x.id===o.diagnosisId);return <article key={o.id}><div className="order-kind">{{MEDICATION:"Лекарство",LAB:"Анализ",IMAGING:"Исследование",REFERRAL:"Специалист",PROCEDURE:"Процедура",FOLLOW_UP:"Повторный приём"}[o.orderType]}</div><div><strong>{o.title}</strong><span>Диагноз: {d?.name||"—"} · с {o.startOn}{o.endOn?` до ${o.endOn}`:""}</span>{(o.dosage||o.frequency||o.durationDays)&&<p>{[o.dosage,o.frequency,o.durationDays?`${o.durationDays} дн.`:""].filter(Boolean).join(" · ")}</p>}<p>{o.instructions}</p><small>{o.author||"—"}</small></div><select value={o.status} onChange={e=>void changeOrderStatus(o.id,e.target.value as ClinicalOrder["status"])}><option value="ACTIVE">Активно</option><option value="COMPLETED">Завершено</option><option value="CANCELLED">Отменено</option></select></article>}):<small>Назначения пока не добавлены</small>}</div></div>
            <div className="clinical-card clinical-wide imaging-panel"><h3>МРТ, КТ и рентген</h3><form className="imaging-form" onSubmit={addImagingStudy}><label>Тип исследования<select name="modality" required defaultValue="MRI"><option value="MRI">МРТ</option><option value="CT">КТ</option><option value="XRAY">Рентген</option></select></label><Field name="performedOn" label="Дата исследования" type="date"/><Field name="bodyArea" label="Область исследования" placeholder="Например: головной мозг"/><Field name="diagnosis" label="Диагноз, к которому относится исследование"/><Field name="conclusion" label="Заключение" required={false}/><Field name="notes" label="Примечание врача" required={false}/><label className="imaging-file"><strong>Файл исследования *</strong><input type="file" name="attachment" accept="image/jpeg,image/png,image/webp,application/pdf,.dcm,application/dicom" required/><span>PDF, JPG, PNG, WEBP или DICOM (.dcm). Файл обязателен.</span></label><button className="primary">Сохранить исследование</button></form><div className="imaging-list">{(clinical.imagingStudies??[]).length?(clinical.imagingStudies??[]).map(x=><article key={x.id}><div className={`imaging-kind ${x.modality.toLowerCase()}`}>{x.modality==="MRI"?"МРТ":x.modality==="CT"?"КТ":"РЕНТГЕН"}</div><div><strong>{x.bodyArea}</strong><span>{x.performedOn} · Диагноз: {x.diagnosis}</span>{x.conclusion&&<p>Заключение: {x.conclusion}</p>}{x.notes&&<small>{x.notes}</small>}<a href={`/api/clinic/imaging/${x.id}/file`} target="_blank" rel="noreferrer">Открыть файл: {x.fileName}</a></div></article>):<small>Исследования пока не добавлены</small>}</div></div>
            <div className="clinical-card clinical-wide"><h3>{ct.history}</h3><div className="history-list">{clinical.history.length ? clinical.history.map(x => <article key={x.id}><time>{new Date(x.occurredAt).toLocaleString(locale)}</time><strong>{x.diagnosis || x.complaints || ct.visit}</strong>{x.complaints && <p>{ct.complaints}: {x.complaints}</p>}{x.treatment && <p>{ct.treatment}: {x.treatment}</p>}{x.notes && <p>{x.notes}</p>}<small>{x.author || "—"}{x.branch ? ` · ${x.branch}` : ""}</small></article>) : <small>{ct.noHistory}</small>}</div></div>
          </div>
        </section>
      )}
      <div className="management-panel">
        {items.length === 0 ? (
          <div className="table-empty patient-empty">
            <div className="empty-medical-icon">♡</div>
            <strong>{t.empty}</strong>
            <span>
              {locale === "ru"
                ? "Сначала добавьте пациента. После сохранения автоматически откроется медицинская карта с группой крови, аллергиями и историей визитов."
                : locale === "uz"
                  ? "Avval bemorni qo‘shing. Saqlangandan keyin qon guruhi, allergiyalar va tashriflar tarixi bilan tibbiy karta avtomatik ochiladi."
                  : "Add a patient first. Their medical record with blood group, allergies, and visit history will open automatically."}
            </span>
            <button className="primary" onClick={() => setOpen(true)}>
              + {t.add}
            </button>
          </div>
        ) : (
          <div className="admin-list">
            {items.map((x) => (
              <article key={x.id}>
                <div className="profile-list-main">
                  <span className="profile-list-avatar" style={x.hasPhoto?{backgroundImage:`url(/api/clinic/patients/${x.id}/photo)`}:undefined}>{!x.hasPhoto&&`${x.firstName[0]||""}${x.lastName[0]||""}`}</span>
                  <div>
                  <strong>
                    {x.lastName} {x.firstName} {x.middleName || ""}
                  </strong>
                  <small>
                    {x.birthDate} · {x.gender === "female" ? t.female : t.male}{" "}
                    · {x.branch || "—"}
                  </small>
                  </div>
                </div>
                <div className="patient-row-actions"><button className="secondary" onClick={() => void openClinical(x)}>{ct.open}</button><button className="secondary" onClick={()=>void startEdit(x)}>Изменить</button><button className="danger-button" onClick={()=>void removePatient(x)}>Удалить</button></div>
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
  placeholder,
  pattern,
  title,
  defaultValue,
  min,
  max,
  step,
}: {
  name: string;
  label: string;
  type?: string;
  required?: boolean;
  placeholder?: string;
  pattern?: string;
  title?: string;
  defaultValue?: string | number | undefined;
  min?: string;
  max?: string;
  step?: string;
}) {
  return (
    <label>
      {label}
      <input
        name={name}
        type={type}
        required={required}
        placeholder={placeholder}
        pattern={pattern}
        title={title}
        defaultValue={defaultValue}
        min={min}
        max={max}
        step={step}
      />
    </label>
  );
}
function Phone({
  name,
  label,
  required = true,
  defaultValue,
}: {
  name: string;
  label: string;
  required?: boolean;
  defaultValue?: string | undefined;
}) {
  return (
    <label>
      {label}
      <span className="phone-input">
        <span>+998</span>
        <input
          name={name}
          type="tel"
          inputMode="numeric"
          pattern="[0-9]{9}"
          maxLength={9}
          placeholder="901234567"
          required={required}
          defaultValue={defaultValue}
        />
      </span>
    </label>
  );
}
