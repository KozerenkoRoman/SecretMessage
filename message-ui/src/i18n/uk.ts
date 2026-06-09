// src/locales/uk.ts
export const uk = {
  // ==== Language switcher =============================================
  language: {
    switcherAriaLabel: "Перемикач мови",
    uk: "Українська",
    en: "English",
  },

  // ==== Common / shared ===============================================
  common: {
    you: "Ви",
    guest: "Гість",
    player: "Гравець",
    opponent: "Опонент",
    nobody: "Ніхто",
    avatarAlt: "Аватар гравця",
    cardBackAlt: "Сорочка карти",
    chipAlt: "Фішка перемоги",
    myChipsTitle: "Ваші фішки перемоги",
  },

  // ==== Admin =========================================================
  admin: {
    title: "Панель Адміністратора",
    subtitle: "Керування користувачами та активними сесіями хабу",
    refresh: "Оновити список ↻",
    backToDesktop: "← На робочий стіл",
    table: {
      title: "Усі зареєстровані користувачі",
      id: "ID / UUID",
      username: "Нікнейм",
      email: "Email",
      role: "Роль",
      action: "Дія",
    },
    block: "Заблокувати",
    blocked: "недоторканний",
    emptyUsers: "Користувачів не знайдено або завантаження...",
    blockConfirm: {
      title: "Блокування користувача",
      message:
        "Ви впевнені, що хочете заблокувати користувача {username}? Ця дія обмежить доступ гравця до ігрового хабу.",
      confirm: "Заблокувати",
      cancel: "Скасувати",
    },
    errors: {
      noAccess: "У вас немає прав доступу до панелі адміністратора.",
      loadFailed: "Не вдалося завантажити список користувачів.",
      blockFailed: "Не вдалося заблокувати користувача {username}",
    },
    blockSuccess: "Користувача {username} успішно заблоковано на бекенді.",
  },

  // ==== Confirm modal ================================================
  modal: {
    confirm: {
      title: "Підтвердження",
      message: "Ви впевнені, що хочете виконати цю дію?",
      ok: "Підтвердити",
      cancel: "Скасувати",
    },
  },

  // ==== Auth =========================================================
  auth: {
    title: "Вхід до гри",
    username: "Логін / Нікнейм",
    password: "Пароль",
    submit: "Увійти до гри",
    registerPrompt: "Вже маєте акаунт?",
    registerLink: "Зареєструватися",
    errors: {
      invalidCredentials: "Неправильний логін або пароль",
      noToken: "Сервер не повернув JWT-токен доступу",
    },
  },

  // ==== Register ======================================================
  register: {
    title: "Реєстрація",
    username: "Логін / Нікнейм",
    email: "Email-адреса",
    emailPlaceholder: "example@domain.com",
    password: "Пароль",
    submit: "Зареєструватися",
    loginPrompt: "Вже маєте акаунт?",
    loginLink: "Увійти",
    errors: {
      registrationFailed: "Помилка реєстрації нового користувача",
      noToken: "Сервер успішно створив акаунт, але не надіслав токен авторизації",
    },
  },

  // ==== Desktop =======================================================
  desktop: {
    brand: "Secret Message",
    welcome: "Вітаємо, {username}",
    profileTitle: "Налаштування профілю",
    avatarChangeHint: "Змінити",
    myAvatarAlt: "Мій аватар",
    hostAvatarAlt: "Аватар хоста",
    adminLink: "Панель адміна ⚙",
    logout: "Вийти",
    createRoom: "+ Створити кімнату",
    lobbyHeader: "Доступні лобі (в реальному часі):",
    noRooms: "Активних кімнат немає або вони вже розпочали гру. Створіть першу!",
    lobbyEmpty: "Користувачів не знайдено або завантаження...",
    roomLabel: 'Кімната "{id}"',
    playersCount: "Гравців:{count}/{max}",
    participants: "Учасники:{names}",
    join: "Приєднатися",
    errors: {
      sessionExpired: "Сесія застаріла або неавторизована. Будь ласка, перезайдіть.",
      loadRoomsFailed: "Не вдалося завантажити список кімнат",
      noRightsToCreate: "Немає прав для створення кімнати (401)",
      createFailed: "Не вдалося створити кімнату",
    },
  },

  // ==== Board =========================================================
  board: {
    leave: "Вийти",
    room: "Кімната:",
    time: "Час:",
    timeUnit: "с",
    start: "Почати гру",
    waiting: "Очікування...",
    yourTurn: "Ваш хід!",
    currentTurn: "Ходить:",
    protection: "Захист",
    discardPile: "Відбій:",
    table: "Стіл відбою",
    deck: "Колода",
    chancellorBadge: "Вибір Канцлера!",
    chancellorPickOwn: "Обрати карту собі",
    waitingForCards: "Очікування карт...",
    discardTooltip: "{name}— Скинув {owner}",
    protectionTooltip: "{name}— Активний захист",
    leaveConfirm: {
      title: "Вихід з гри",
      message:
        "Ви впевнені, що хочете покинути поточну гру та повернутися в десктоп лобі?",
      confirm: "Вийти",
      cancel: "Залишитись",
    },
  },

  // ==== Action modal ==================================================
  action: {
    title: "Розіграш карти:",
    targetSelect: "Виберіть карту опонента:",
    noTargets: "Немає доступних цілей!",
    noTargetsHint:
      "Усі інші гравці захищені ефектом Служниці або вибули. Карта буде скинута в загальний відбій без застосування ефекту.",
    autoApply: "Ця карта застосовується автоматично на вас або скидається в стіл.",
    guardGuess: "Виберіть карту опонента:",
    chosen: "Обрано",
    youSuffix: " (Ви)",
    cancel: "Скасувати",
    submit: "Підтвердити Хід",
  },

  // ==== Card reveal modal =============================================
  reveal: {
    duelBaron: "Дуель Барона",
    priestEffect: "Ефект Священника",
    close: "Закрити та продовжити",
  },

  // ==== Chancellor modal ===============================================
  chancellor: {
    title: "Ефект Канцлера",
    step1: "Крок 1:Оберіть 1 карту, яку хочете ЗАЛИШИТИ у себе в руці",
    step2: "Крок 2:Оберіть послідовність карт для відправки на дно колоди",
    keepBadge: "В Руку",
    bottomOrderTitle: "Порядок карт на дно:",
    reset: "Скинути вибір",
    submit: "Підтвердити хід",
  },

  // ==== User profile modal ============================================
  profile: {
    title: "Налаштування профілю",
    subtitle: "Змініть свій ігровий аватар, ім'я або пароль",
    avatarChange: "Випадковий аватар",
    avatarAlt: "Аватар користувача",
    usernameLabel: "Ім'я користувача (Username)",
    usernamePlaceholder: "Введіть нікнейм",
    passwordSectionDivider: "Зміна пароля",
    currentPassword: "Поточний пароль",
    currentPasswordPlaceholder: "Необхідно для зміни пароля",
    newPassword: "Новий пароль",
    newPasswordPlaceholder: "Введіть новий пароль",
    save: "Зберегти все",
    saving: "Збереження...",
    cancel: "Скасувати",
    errors: {
      currentRequired:
        "Будь ласка, вкажіть ваш поточний пароль для встановлення нового.",
      saveFailed: "Не вдалося зберегти зміни профілю.",
      networkError: "Помилка з'єднання з сервером",
    },
  },

  // ==== Room manager ==================================================
  room: {
    connecting: "Підключення до ігрової кімнати",
    cancel: "Скасувати підключення",
  },

  // ==== Game end modal ================================================
  gameEnd: {
    final: "👑 ФІНАЛ ПАРТІЇ 👑",
    roundEnd: "⚔️ КІНЕЦЬ РАУНДУ ⚔️",
    winner: "Ви абсолютний чемпіон!",
    loser: "Ви виграли раунд!",
    gameOver: "Гру завершено",
    roundOver: "Раунд закінчено",
    winnerLabel: "Переможець:",
    scoreboardTitle: "Поточний рахунок у кімнаті:",
    scoreOutOf: "/ 7",
    waitingForPlayers: "Очікування гравців...",
    waitingForHost: "Очікуємо,поки власник кімнати почне нову гру...",
    nextRound: "Наступний раунд",
    leaveLobby: "Вийти в лобі",
    restart: "Грати знову",
  },

  // ==== Game error modal ==============================================
  gameError: {
    title: "Помилка ігрового ходу",
    server: "Сервер відхилив вашу дію",
    codeLabel: "код:",
    ok: "Зрозуміло",
    leave: "Вийти в лобі",
  },

  // ==== Status strings ================================================
  status: {
    protected: "Захист",
    out: "Вибув",
    inGame: "У грі",
    waiting: "Очікування...",
  },

  // ==== Cards (fallbacks) =============================================
  cards: {
    unknown: "Невідома карта",
    noDescription: "Опис відсутній",
    serverHandled: "Ефект оброблюється сервером",
  },

  // ==== Текстові повідомлення, що вже в i18n (error dictionary) =====
  // (перенесено в файл i18n/errorMessages.ts – залиште без змін)
};
