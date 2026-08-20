import 'package:flutter/widgets.dart';

/// Hand-maintained strings. Primary: RU, secondary: UZ, also EN.
class AppLocalizations {
  AppLocalizations(this.locale);

  final Locale locale;

  static const supportedLocales = [
    Locale('ru'),
    Locale('uz'),
    Locale('en'),
  ];

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations) ??
        AppLocalizations(const Locale('ru'));
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _LocalizationsDelegate();

  String get _lang => locale.languageCode;

  String _t(String ru, String uz, String en) => switch (_lang) {
        'uz' => uz,
        'en' => en,
        _ => ru,
      };

  String get appTitle => _t('Клиника', 'Klinika', 'Clinic');
  String get clinicBrand => _t('ONA VA BOLA', 'ONA VA BOLA', 'ONA VA BOLA');
  String get loginTitle => _t('Вход', 'Kirish', 'Sign in');
  String get loginSubtitle => _t(
        'Войдите в аккаунт клиники',
        'Klinika hisobingizga kiring',
        'Access your clinic account',
      );
  String get organizationId =>
      _t('ID организации', 'Tashkilot ID', 'Organization ID');
  String get login => _t('Логин', 'Login', 'Login');
  String get password => _t('Пароль', 'Parol', 'Password');
  String get signIn => _t('Войти', 'Kirish', 'Sign in');
  String get signOut => _t('Выйти', 'Chiqish', 'Sign out');
  String get save => _t('Сохранить', 'Saqlash', 'Save');
  String get cancel => _t('Отменить', 'Bekor qilish', 'Cancel');
  String get confirm => _t('Подтвердить', 'Tasdiqlash', 'Confirm');
  String get back => _t('Назад', 'Orqaga', 'Back');
  String get done => _t('Готово', 'Tayyor', 'Done');
  String get add => _t('Добавить', 'Qo‘shish', 'Add');
  String get edit => _t('Изменить', 'Tahrirlash', 'Edit');
  String get delete => _t('Удалить', 'O‘chirish', 'Delete');
  String get comingSoon =>
      _t('Скоро будет доступно', 'Tez orada', 'Coming soon');

  String get navHome => _t('Главная', 'Bosh sahifa', 'Home');
  String get navDoctors => _t('Врачи', 'Shifokorlar', 'Doctors');
  String get navFavorites => _t('Избранное', 'Sevimlilar', 'Favorites');
  String get navSearch => _t('Поиск', 'Qidiruv', 'Search');
  String get navAppointments => _t('Записи', 'Yozuvlar', 'Appointments');
  String get navMedical => _t('Медицина', 'Tibbiyot', 'Medical');
  String get navProfile => _t('Профиль', 'Profil', 'Profile');
  String get navToday => _t('Сегодня', 'Bugun', 'Today');
  String get navSchedule => _t('Расписание', 'Jadval', 'Schedule');
  String get navPatients => _t('Пациенты', 'Bemorlar', 'Patients');
  String get navMessages => _t('Сообщения', 'Xabarlar', 'Messages');
  String get navDashboard => _t('Панель', 'Boshqaruv', 'Dashboard');
  String get navAnalytics => _t('Аналитика', 'Tahlil', 'Analytics');
  String get navClinic => _t('Клиника', 'Klinika', 'Clinic');
  String get navFinance => _t('Финансы', 'Moliya', 'Finance');
  String get navOperations => _t('Операции', 'Operatsiyalar', 'Operations');
  String get navReports => _t('Отчёты', 'Hisobotlar', 'Reports');
  String get navDebts => _t('Задолженности', 'Qarzlar', 'Debts');

  String get welcome => _t('Здравствуйте', 'Xush kelibsiz', 'Welcome');
  String get hello => _t('Привет', 'Salom', 'Hello');
  String get bookAppointment =>
      _t('Записаться к врачу', 'Shifokorga yozilish', 'Book a doctor');
  String get myMedicalRecord =>
      _t('Моя медицинская карта', 'Tibbiy kartam', 'My medical record');
  String get chatWithDoctor =>
      _t('Чат с врачом', 'Shifokor bilan chat', 'Chat with doctor');
  String get upcomingAppointment =>
      _t('Ближайшая запись', 'Keyingi qabul', 'Next appointment');
  String get noUpcomingAppointment => _t(
        'Нет ближайших записей',
        'Yaqin qabullar yo‘q',
        'No upcoming appointments',
      );

  String get searchDoctorHint => _t(
        'Найдите подходящего врача',
        'O‘zingizga mos shifokorni toping',
        'Find the right doctor for you',
      );
  String get myAppointments =>
      _t('Мои записи', 'Mening qabullarim', 'My Appointments');
  String get seeAll => _t('Все', 'Hammasi', 'See All');
  String get doctorSpecialty =>
      _t('Специальности', 'Mutaxassisliklar', 'Doctor Specialty');
  String get popularDoctors =>
      _t('Популярные врачи', 'Mashhur shifokorlar', 'Popular Doctors');
  String get doctor => _t('Врач', 'Shifokor', 'Doctor');
  String get aboutDoctor => _t('О враче', 'Shifokor haqida', 'About');
  String get availability =>
      _t('Расписание', 'Ish jadvali', 'Availability');
  String get rating => _t('Рейтинг', 'Reyting', 'Rating');
  String get experience => _t('Опыт', 'Tajriba', 'Experience');
  String get patientsCount => _t('Пациенты', 'Bemorlar', 'Patients');
  String get yearsShort => _t('лет', 'yil', 'yrs');
  String get notWorkingToday =>
      _t('Сегодня не принимает', 'Bugun qabul yo‘q', 'Not available today');
  String get workingToday =>
      _t('Сегодня принимает', 'Bugun qabul bor', 'Available today');
  String get breakLabel => _t('перерыв', 'tanaffus', 'break');
  String get selectDate => _t('Выберите дату', 'Sanani tanlang', 'Select date');
  String get selectTime =>
      _t('Выберите время', 'Vaqtni tanlang', 'Select time');
  String get noSlots =>
      _t('Нет свободных слотов', 'Bo‘sh vaqt yo‘q', 'No free slots');
  String get confirmBooking =>
      _t('Подтвердить запись', 'Yozuvni tasdiqlash', 'Confirm booking');
  String get bookingSuccess =>
      _t('Запись создана', 'Yozuv yaratildi', 'Appointment booked');
  String get selectPatient =>
      _t('Выберите пациента', 'Bemorni tanlang', 'Select patient');
  String get searchPatientHint =>
      _t('Поиск пациента', 'Bemor qidirish', 'Search patient');

  String get tabUpcoming => _t('Предстоящие', 'Kelgusi', 'Upcoming');
  String get tabPast => _t('Прошедшие', 'O‘tgan', 'Past');
  String get tabCancelled => _t('Отменённые', 'Bekor qilingan', 'Cancelled');
  String get appointmentDetails =>
      _t('Детали записи', 'Yozuv tafsilotlari', 'Appointment details');
  String get statusLabel => _t('Статус', 'Holat', 'Status');
  String statusText(String status) {
    final s = status.toLowerCase();
    return switch (s) {
      'scheduled' || 'confirmed' || 'booked' =>
        _t('Запланировано', 'Rejalashtirilgan', 'Scheduled'),
      'completed' || 'done' => _t('Завершено', 'Yakunlangan', 'Completed'),
      'cancelled' || 'canceled' =>
        _t('Отменено', 'Bekor qilingan', 'Cancelled'),
      'no_show' => _t('Не явился', 'Kelmadi', 'No-show'),
      'waiting' || 'in_queue' => _t('Ожидает', 'Kutmoqda', 'Waiting'),
      'in_progress' => _t('На приёме', 'Qabulda', 'In progress'),
      _ => status,
    };
  }

  String get cancelAppointment =>
      _t('Отменить запись', 'Yozuvni bekor qilish', 'Cancel appointment');
  String get noFavorites => _t(
        'Пока нет избранных врачей',
        'Sevimli shifokorlar yo‘q',
        'No favorite doctors yet',
      );
  String get addToFavorites =>
      _t('В избранное', 'Sevimlilarga', 'Add to favorites');
  String get removeFromFavorites =>
      _t('Убрать из избранного', 'Sevimlilardan olib tashlash', 'Remove');

  String get myProfile => _t('Мой профиль', 'Mening profilim', 'My Profile');
  String get accountSettings =>
      _t('Настройки аккаунта', 'Hisob sozlamalari', 'Account Settings');
  String get personalInformation =>
      _t('Личные данные', 'Shaxsiy ma’lumotlar', 'Personal Information');
  String get bookingHistory =>
      _t('История записей', 'Yozuvlar tarixi', 'Booking History');
  String get myCards => _t('Мои карты', 'Kartalarim', 'My Cards');
  String get myTests => _t('Мои анализы', 'Tahlillarim', 'My Tests');
  String get notifications =>
      _t('Уведомления', 'Bildirishnomalar', 'Notifications');
  String get security => _t('Безопасность', 'Xavfsizlik', 'Security');
  String get helpCenter =>
      _t('Справка', 'Yordam markazi', 'Help Center');
  String get language => _t('Язык', 'Til', 'Language');
  String get languageRussian => _t('Русский', 'Ruscha', 'Russian');
  String get languageUzbek => _t('Узбекский', 'O‘zbekcha', 'Uzbek');
  String get languageEnglish => _t('Английский', 'Inglizcha', 'English');
  String get uiPreview =>
      _t('Превью UI (временно)', 'UI ko‘rinish (vaqtincha)', 'UI preview (temp)');
  String get fullName => _t('ФИО', 'F.I.Sh.', 'Full name');
  String get phone => _t('Телефон', 'Telefon', 'Phone');
  String get email => _t('Email', 'Email', 'Email');
  String get dateOfBirth =>
      _t('Дата рождения', 'Tug‘ilgan sana', 'Date of birth');
  String get noCards =>
      _t('Карты не добавлены', 'Kartalar yo‘q', 'No cards added');
  String get pushNotifications =>
      _t('Push-уведомления', 'Push-bildirishnomalar', 'Push notifications');
  String get appointmentReminders =>
      _t('Напоминания о записи', 'Qabul eslatmalari', 'Appointment reminders');
  String get changePassword =>
      _t('Сменить пароль', 'Parolni almashtirish', 'Change password');
  String get faq => _t('Частые вопросы', 'Ko‘p so‘raladigan', 'FAQ');
  String get contactSupport =>
      _t('Связаться с поддержкой', 'Yordam bilan bog‘lanish', 'Contact support');
  String get noNotifications =>
      _t('Нет уведомлений', 'Bildirishnomalar yo‘q', 'No notifications');

  String get medicalTimeline =>
      _t('История', 'Tarix', 'Timeline');
  String get medicalLabs => _t('Анализы', 'Tahlillar', 'Labs');
  String get medicalPrescriptions =>
      _t('Назначения', 'Retseptlar', 'Prescriptions');
  String get medicalOverview => _t('Обзор', 'Umumiy', 'Overview');
  String get bloodGroup => _t('Группа крови', 'Qon guruhi', 'Blood group');
  String get height => _t('Рост', 'Bo‘y', 'Height');
  String get weight => _t('Вес', 'Vazn', 'Weight');
  String get allergies => _t('Аллергии', 'Allergiyalar', 'Allergies');
  String get noAllergies =>
      _t('Не указаны', 'Ko‘rsatilmagan', 'Not specified');
  String get diagnoses => _t('Диагнозы', 'Tashxislar', 'Diagnoses');
  String get searchPatient =>
      _t('Найти пациента', 'Bemorni topish', 'Find patient');
  String get selectPatientFirst => _t(
        'Выберите пациента для просмотра карты',
        'Kartani ko‘rish uchun bemorni tanlang',
        'Select a patient to view the record',
      );
  String get typeReply =>
      _t('Написать ответ…', 'Javob yozing…', 'Type a reply…');
  String get send => _t('Отправить', 'Yuborish', 'Send');
  String get bookingConfirmTitle =>
      _t('Подтверждение записи', 'Yozuvni tasdiqlash', 'Confirm appointment');
  String get notifyApptReminderBody =>
      _t('Завтра 10:00 · Давидова Е.', 'Ertaga 10:00 · Davidova E.', 'Tomorrow 10:00 · E. Davidova');
  String get notifyLabReadyBody => _t(
        'Готов результат: Гемоглобин',
        'Natija tayyor: Gemoglobin',
        'Result ready: Hemoglobin',
      );
  String get notifyNewSlotBody => _t(
        'Новый слот у невролога',
        'Nevrologda yangi vaqt',
        'New slot with neurologist',
      );

  String get specialtyNeurologist =>
      _t('Невролог', 'Nevrolog', 'Neurologist');
  String get specialtyPediatric =>
      _t('Педиатр', 'Pediatr', 'Pediatric');
  String get specialtyCardiologist =>
      _t('Кардиолог', 'Kardiolog', 'Cardiologist');
  String get specialtyDentist =>
      _t('Стоматолог', 'Stomatolog', 'Dentist');
  String get specialtyTherapist =>
      _t('Терапевт', 'Terapevt', 'Therapist');

  String get rolePatient => _t('Пациент', 'Bemor', 'Patient');
  String get roleDoctor => _t('Врач', 'Shifokor', 'Doctor');
  String get roleOwner => _t('Владелец', 'Egasi', 'Owner');
  String get roleAccountant => _t('Бухгалтер', 'Buxgalter', 'Accountant');
  String get roleOther => _t('Другое', 'Boshqa', 'Other');

  String get todayAppointments =>
      _t('Записи на сегодня', 'Bugungi qabullar', 'Today’s appointments');
  String get revenue => _t('Выручка', 'Tushum', 'Revenue');
  String get debts => _t('Долги', 'Qarzlar', 'Debts');
  String get operations => _t('Операции', 'Operatsiyalar', 'Operations');
  String get noMessages =>
      _t('Сообщений пока нет', 'Xabarlar yo‘q', 'No messages yet');
  String get clinicOverview =>
      _t('Обзор клиники', 'Klinika ko‘rinishi', 'Clinic overview');
  String get financeOverview =>
      _t('Финансовый обзор', 'Moliyaviy ko‘rinish', 'Finance overview');

  String get errorNetwork =>
      _t('Нет подключения к интернету', 'Internet yo‘q', 'No internet connection');
  String get errorServer =>
      _t('Сервер недоступен', 'Server mavjud emas', 'Server unavailable');
  String get errorUnauthorized => _t(
        'Сессия истекла. Войдите снова.',
        'Sessiya tugadi. Qayta kiring.',
        'Session expired. Please sign in again.',
      );
  String get errorForbidden =>
      _t('Недостаточно прав', 'Ruxsat yo‘q', 'You do not have access');
  String get errorNotFound => _t('Не найдено', 'Topilmadi', 'Not found');
  String get errorValidation => _t(
        'Проверьте введённые данные',
        'Ma’lumotlarni tekshiring',
        'Please check the entered data',
      );
  String get errorTimeout =>
      _t('Превышено время ожидания', 'Vaqt tugadi', 'Request timed out');
  String get errorGeneric =>
      _t('Что-то пошло не так', 'Xatolik yuz berdi', 'Something went wrong');
  String get loading => _t('Загрузка…', 'Yuklanmoqda…', 'Loading…');
  String get empty => _t('Пока пусто', 'Hozircha bo‘sh', 'Nothing here yet');
  String get retry => _t('Повторить', 'Qayta urinish', 'Retry');
  String get roleUnsupported => _t(
        'Эта роль пока недоступна в мобильном приложении',
        'Bu rol mobil ilovada hali mavjud emas',
        'This role is not available in the mobile app yet',
      );

  List<String> get fallbackSpecialties => [
        specialtyNeurologist,
        specialtyPediatric,
        specialtyCardiologist,
        specialtyDentist,
        specialtyTherapist,
      ];
}

class _LocalizationsDelegate extends LocalizationsDelegate<AppLocalizations> {
  const _LocalizationsDelegate();

  @override
  bool isSupported(Locale locale) =>
      ['ru', 'uz', 'en'].contains(locale.languageCode);

  @override
  Future<AppLocalizations> load(Locale locale) async =>
      AppLocalizations(locale);

  @override
  bool shouldReload(covariant LocalizationsDelegate<AppLocalizations> old) =>
      false;
}
