CREATE TABLE document_templates(
 id uuid PRIMARY KEY,
 organization_id uuid NOT NULL REFERENCES organizations(id),
 name text NOT NULL,
 document_type text NOT NULL CHECK(document_type IN ('ASSIGNMENT','PRESCRIPTION','RECEIPT','OTHER')),
 page_size text NOT NULL CHECK(page_size IN ('A4','A5','RECEIPT_80','CUSTOM')),
 width_mm numeric(7,2) NOT NULL CHECK(width_mm BETWEEN 40 AND 500),
 height_mm numeric(7,2) NOT NULL CHECK(height_mm BETWEEN 40 AND 1000),
 layout jsonb NOT NULL DEFAULT '{"elements":[]}'::jsonb,
 is_default boolean NOT NULL DEFAULT false,
 created_by uuid REFERENCES users(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX document_templates_org_idx ON document_templates(organization_id,document_type,updated_at DESC);
CREATE UNIQUE INDEX document_templates_default_idx ON document_templates(organization_id,document_type) WHERE is_default;

INSERT INTO document_templates(id,organization_id,name,document_type,page_size,width_mm,height_mm,is_default,layout)
SELECT gen_random_uuid(),o.id,'Тестовый рецепт','PRESCRIPTION','A5',148,210,true,
jsonb_build_object('elements',jsonb_build_array(
 jsonb_build_object('id','logo','type','image','x',6,'y',5,'width',15,'height',11,'src','/brand-mark.png'),
 jsonb_build_object('id','title','type','text','x',25,'y',6,'width',68,'height',8,'text','ONA VA BOLA KLINIKASI','fontSize',18,'fontFamily','Manrope','fontWeight',800,'color','#17202a','align','center'),
 jsonb_build_object('id','line','type','shape','shape','line','x',6,'y',19,'width',88,'height',1,'fill','transparent','stroke','#e80d4f','strokeWidth',2),
 jsonb_build_object('id','patient','type','text','x',7,'y',24,'width',60,'height',6,'text','Пациент: {{patient.fullName}}','fontSize',13,'fontFamily','Manrope','fontWeight',600,'color','#17202a','align','left'),
 jsonb_build_object('id','date','type','text','x',70,'y',24,'width',23,'height',6,'text','Дата: {{date}}','fontSize',12,'fontFamily','Manrope','fontWeight',500,'color','#17202a','align','right'),
 jsonb_build_object('id','heading','type','text','x',7,'y',34,'width',86,'height',8,'text','РЕЦЕПТ И НАЗНАЧЕНИЯ','fontSize',17,'fontFamily','Manrope','fontWeight',800,'color','#e80d4f','align','center'),
 jsonb_build_object('id','body','type','text','x',9,'y',47,'width',82,'height',31,'text','{{assignment}}','fontSize',14,'fontFamily','Manrope','fontWeight',500,'color','#17202a','align','left'),
 jsonb_build_object('id','doctor','type','text','x',52,'y',86,'width',40,'height',7,'text','Врач: {{doctor.fullName}}','fontSize',12,'fontFamily','Manrope','fontWeight',600,'color','#17202a','align','right')
)) FROM organizations o;
