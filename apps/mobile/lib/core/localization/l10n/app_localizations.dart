import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_ru.dart';
import 'app_localizations_uz.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
      : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
    delegate,
    GlobalMaterialLocalizations.delegate,
    GlobalCupertinoLocalizations.delegate,
    GlobalWidgetsLocalizations.delegate,
  ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('ru'),
    Locale('uz'),
    Locale('en')
  ];

  /// No description provided for @appTitle.
  ///
  /// In ru, this message translates to:
  /// **'Клиника'**
  String get appTitle;

  /// No description provided for @loginTitle.
  ///
  /// In ru, this message translates to:
  /// **'Вход'**
  String get loginTitle;

  /// No description provided for @loginSubtitle.
  ///
  /// In ru, this message translates to:
  /// **'Войдите в аккаунт клиники'**
  String get loginSubtitle;

  /// No description provided for @organizationId.
  ///
  /// In ru, this message translates to:
  /// **'ID организации'**
  String get organizationId;

  /// No description provided for @login.
  ///
  /// In ru, this message translates to:
  /// **'Логин'**
  String get login;

  /// No description provided for @password.
  ///
  /// In ru, this message translates to:
  /// **'Пароль'**
  String get password;

  /// No description provided for @signIn.
  ///
  /// In ru, this message translates to:
  /// **'Войти'**
  String get signIn;

  /// No description provided for @signOut.
  ///
  /// In ru, this message translates to:
  /// **'Выйти'**
  String get signOut;

  /// No description provided for @navHome.
  ///
  /// In ru, this message translates to:
  /// **'Главная'**
  String get navHome;

  /// No description provided for @navDoctors.
  ///
  /// In ru, this message translates to:
  /// **'Врачи'**
  String get navDoctors;

  /// No description provided for @navFavorites.
  ///
  /// In ru, this message translates to:
  /// **'Избранное'**
  String get navFavorites;

  /// No description provided for @navSearch.
  ///
  /// In ru, this message translates to:
  /// **'Поиск'**
  String get navSearch;

  /// No description provided for @navAppointments.
  ///
  /// In ru, this message translates to:
  /// **'Записи'**
  String get navAppointments;

  /// No description provided for @navMedical.
  ///
  /// In ru, this message translates to:
  /// **'Медицина'**
  String get navMedical;

  /// No description provided for @navProfile.
  ///
  /// In ru, this message translates to:
  /// **'Профиль'**
  String get navProfile;

  /// No description provided for @navToday.
  ///
  /// In ru, this message translates to:
  /// **'Сегодня'**
  String get navToday;

  /// No description provided for @navSchedule.
  ///
  /// In ru, this message translates to:
  /// **'Расписание'**
  String get navSchedule;

  /// No description provided for @navPatients.
  ///
  /// In ru, this message translates to:
  /// **'Пациенты'**
  String get navPatients;

  /// No description provided for @navMessages.
  ///
  /// In ru, this message translates to:
  /// **'Сообщения'**
  String get navMessages;

  /// No description provided for @navDashboard.
  ///
  /// In ru, this message translates to:
  /// **'Панель'**
  String get navDashboard;

  /// No description provided for @navAnalytics.
  ///
  /// In ru, this message translates to:
  /// **'Аналитика'**
  String get navAnalytics;

  /// No description provided for @navClinic.
  ///
  /// In ru, this message translates to:
  /// **'Клиника'**
  String get navClinic;

  /// No description provided for @navFinance.
  ///
  /// In ru, this message translates to:
  /// **'Финансы'**
  String get navFinance;

  /// No description provided for @navOperations.
  ///
  /// In ru, this message translates to:
  /// **'Операции'**
  String get navOperations;

  /// No description provided for @navReports.
  ///
  /// In ru, this message translates to:
  /// **'Отчёты'**
  String get navReports;

  /// No description provided for @navDebts.
  ///
  /// In ru, this message translates to:
  /// **'Задолженности'**
  String get navDebts;

  /// No description provided for @welcome.
  ///
  /// In ru, this message translates to:
  /// **'Здравствуйте'**
  String get welcome;

  /// No description provided for @hello.
  ///
  /// In ru, this message translates to:
  /// **'Привет'**
  String get hello;

  /// No description provided for @bookAppointment.
  ///
  /// In ru, this message translates to:
  /// **'Записаться к врачу'**
  String get bookAppointment;

  /// No description provided for @myMedicalRecord.
  ///
  /// In ru, this message translates to:
  /// **'Моя медицинская карта'**
  String get myMedicalRecord;

  /// No description provided for @chatWithDoctor.
  ///
  /// In ru, this message translates to:
  /// **'Чат с врачом'**
  String get chatWithDoctor;

  /// No description provided for @upcomingAppointment.
  ///
  /// In ru, this message translates to:
  /// **'Ближайшая запись'**
  String get upcomingAppointment;

  /// No description provided for @noUpcomingAppointment.
  ///
  /// In ru, this message translates to:
  /// **'Нет ближайших записей'**
  String get noUpcomingAppointment;

  /// No description provided for @searchDoctorHint.
  ///
  /// In ru, this message translates to:
  /// **'Найдите подходящего врача'**
  String get searchDoctorHint;

  /// No description provided for @myAppointments.
  ///
  /// In ru, this message translates to:
  /// **'Мои записи'**
  String get myAppointments;

  /// No description provided for @seeAll.
  ///
  /// In ru, this message translates to:
  /// **'Все'**
  String get seeAll;

  /// No description provided for @doctorSpecialty.
  ///
  /// In ru, this message translates to:
  /// **'Специальности'**
  String get doctorSpecialty;

  /// No description provided for @popularDoctors.
  ///
  /// In ru, this message translates to:
  /// **'Популярные врачи'**
  String get popularDoctors;

  /// No description provided for @doctor.
  ///
  /// In ru, this message translates to:
  /// **'Врач'**
  String get doctor;

  /// No description provided for @myProfile.
  ///
  /// In ru, this message translates to:
  /// **'Мой профиль'**
  String get myProfile;

  /// No description provided for @accountSettings.
  ///
  /// In ru, this message translates to:
  /// **'Настройки аккаунта'**
  String get accountSettings;

  /// No description provided for @personalInformation.
  ///
  /// In ru, this message translates to:
  /// **'Личные данные'**
  String get personalInformation;

  /// No description provided for @bookingHistory.
  ///
  /// In ru, this message translates to:
  /// **'История записей'**
  String get bookingHistory;

  /// No description provided for @myCards.
  ///
  /// In ru, this message translates to:
  /// **'Мои карты'**
  String get myCards;

  /// No description provided for @myTests.
  ///
  /// In ru, this message translates to:
  /// **'Мои анализы'**
  String get myTests;

  /// No description provided for @notifications.
  ///
  /// In ru, this message translates to:
  /// **'Уведомления'**
  String get notifications;

  /// No description provided for @security.
  ///
  /// In ru, this message translates to:
  /// **'Безопасность'**
  String get security;

  /// No description provided for @helpCenter.
  ///
  /// In ru, this message translates to:
  /// **'Справка'**
  String get helpCenter;

  /// No description provided for @language.
  ///
  /// In ru, this message translates to:
  /// **'Язык'**
  String get language;

  /// No description provided for @languageRussian.
  ///
  /// In ru, this message translates to:
  /// **'Русский'**
  String get languageRussian;

  /// No description provided for @languageUzbek.
  ///
  /// In ru, this message translates to:
  /// **'Узбекский'**
  String get languageUzbek;

  /// No description provided for @languageEnglish.
  ///
  /// In ru, this message translates to:
  /// **'Английский'**
  String get languageEnglish;

  /// No description provided for @uiPreview.
  ///
  /// In ru, this message translates to:
  /// **'Превью UI (временно)'**
  String get uiPreview;

  /// No description provided for @specialtyNeurologist.
  ///
  /// In ru, this message translates to:
  /// **'Невролог'**
  String get specialtyNeurologist;

  /// No description provided for @specialtyPediatric.
  ///
  /// In ru, this message translates to:
  /// **'Педиатр'**
  String get specialtyPediatric;

  /// No description provided for @specialtyCardiologist.
  ///
  /// In ru, this message translates to:
  /// **'Кардиолог'**
  String get specialtyCardiologist;

  /// No description provided for @specialtyDentist.
  ///
  /// In ru, this message translates to:
  /// **'Стоматолог'**
  String get specialtyDentist;

  /// No description provided for @specialtyTherapist.
  ///
  /// In ru, this message translates to:
  /// **'Терапевт'**
  String get specialtyTherapist;

  /// No description provided for @rolePatient.
  ///
  /// In ru, this message translates to:
  /// **'Пациент'**
  String get rolePatient;

  /// No description provided for @roleDoctor.
  ///
  /// In ru, this message translates to:
  /// **'Врач'**
  String get roleDoctor;

  /// No description provided for @roleOwner.
  ///
  /// In ru, this message translates to:
  /// **'Владелец'**
  String get roleOwner;

  /// No description provided for @roleAccountant.
  ///
  /// In ru, this message translates to:
  /// **'Бухгалтер'**
  String get roleAccountant;

  /// No description provided for @roleOther.
  ///
  /// In ru, this message translates to:
  /// **'Другое'**
  String get roleOther;

  /// No description provided for @errorNetwork.
  ///
  /// In ru, this message translates to:
  /// **'Нет подключения к интернету'**
  String get errorNetwork;

  /// No description provided for @errorServer.
  ///
  /// In ru, this message translates to:
  /// **'Сервер недоступен'**
  String get errorServer;

  /// No description provided for @errorUnauthorized.
  ///
  /// In ru, this message translates to:
  /// **'Сессия истекла. Войдите снова.'**
  String get errorUnauthorized;

  /// No description provided for @errorForbidden.
  ///
  /// In ru, this message translates to:
  /// **'Недостаточно прав'**
  String get errorForbidden;

  /// No description provided for @errorNotFound.
  ///
  /// In ru, this message translates to:
  /// **'Не найдено'**
  String get errorNotFound;

  /// No description provided for @errorValidation.
  ///
  /// In ru, this message translates to:
  /// **'Проверьте введённые данные'**
  String get errorValidation;

  /// No description provided for @errorTimeout.
  ///
  /// In ru, this message translates to:
  /// **'Превышено время ожидания'**
  String get errorTimeout;

  /// No description provided for @errorGeneric.
  ///
  /// In ru, this message translates to:
  /// **'Что-то пошло не так'**
  String get errorGeneric;

  /// No description provided for @loading.
  ///
  /// In ru, this message translates to:
  /// **'Загрузка…'**
  String get loading;

  /// No description provided for @empty.
  ///
  /// In ru, this message translates to:
  /// **'Пока пусто'**
  String get empty;

  /// No description provided for @retry.
  ///
  /// In ru, this message translates to:
  /// **'Повторить'**
  String get retry;

  /// No description provided for @roleUnsupported.
  ///
  /// In ru, this message translates to:
  /// **'Эта роль пока недоступна в мобильном приложении'**
  String get roleUnsupported;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'ru', 'uz'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'ru':
      return AppLocalizationsRu();
    case 'uz':
      return AppLocalizationsUz();
  }

  throw FlutterError(
      'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
      'an issue with the localizations generation tool. Please file an issue '
      'on GitHub with a reproducible sample app and the gen-l10n configuration '
      'that was used.');
}
