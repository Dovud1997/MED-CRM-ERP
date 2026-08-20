CREATE TABLE diagnosis_catalog(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  specialty_id uuid NOT NULL REFERENCES specialty_catalog(id) ON DELETE CASCADE,
  name text NOT NULL,
  icd10_code text,
  description text,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(organization_id,specialty_id,name),
  CHECK(length(trim(name)) BETWEEN 2 AND 240)
);
ALTER TABLE patient_diagnoses ADD COLUMN catalog_id uuid REFERENCES diagnosis_catalog(id);
CREATE INDEX idx_diagnosis_catalog_specialty ON diagnosis_catalog(organization_id,specialty_id,is_active,name);

INSERT INTO diagnosis_catalog(organization_id,specialty_id,name,icd10_code)
SELECT sc.organization_id,sc.id,d.name,d.code FROM specialty_catalog sc JOIN (VALUES
 ('педиатр','Острая инфекция верхних дыхательных путей','J06.9'),('педиатр','Острый бронхит','J20.9'),('педиатр','Атопический дерматит','L20.9'),('педиатр','Железодефицитная анемия','D50.9'),
 ('гинеколог','Воспалительные заболевания органов малого таза','N73.9'),('гинеколог','Нарушение менструального цикла','N92.6'),('гинеколог','Кандидоз вульвы и вагины','B37.3'),
 ('кардиолог','Артериальная гипертензия','I10'),('кардиолог','Ишемическая болезнь сердца','I25.9'),('кардиолог','Нарушение сердечного ритма','I49.9'),
 ('невролог','Мигрень','G43.9'),('невролог','Головная боль напряжения','G44.2'),('невролог','Радикулопатия','M54.1'),
 ('эндокринолог','Сахарный диабет 2 типа','E11.9'),('эндокринолог','Гипотиреоз','E03.9'),('эндокринолог','Ожирение','E66.9'),
 ('стоматолог','Кариес зубов','K02.9'),('стоматолог','Пульпит','K04.0'),('стоматолог','Гингивит','K05.1'),
 ('отоларинголог','Острый тонзиллит','J03.9'),('отоларинголог','Острый средний отит','H66.9'),('отоларинголог','Острый синусит','J01.9'),
 ('офтальмолог','Конъюнктивит','H10.9'),('офтальмолог','Миопия','H52.1'),('офтальмолог','Астигматизм','H52.2'),
 ('дерматолог','Атопический дерматит','L20.9'),('дерматолог','Контактный дерматит','L23.9'),('дерматолог','Акне','L70.9'),
 ('гастроэнтеролог','Гастрит','K29.7'),('гастроэнтеролог','Гастроэзофагеальная рефлюксная болезнь','K21.9'),('гастроэнтеролог','Синдром раздражённого кишечника','K58.9'),
 ('уролог','Цистит','N30.9'),('уролог','Мочекаменная болезнь','N20.9'),('уролог','Простатит','N41.9'),
 ('хирург','Острый аппендицит','K35.8'),('хирург','Пупочная грыжа','K42.9'),('хирург','Паховая грыжа','K40.9'),
 ('терапевт','Острая респираторная инфекция','J06.9'),('терапевт','Артериальная гипертензия','I10'),('терапевт','Железодефицитная анемия','D50.9')
) AS d(pattern,name,code) ON lower(sc.name) LIKE '%'||d.pattern||'%'
ON CONFLICT DO NOTHING;
