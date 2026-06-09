// src/locales/en.ts
export const en = {
  // ==== Language switcher =============================================
  language: {
    switcherAriaLabel: "Language switcher",
    uk: "Ukrainian",
    en: "English",
  },

  // ==== Common / shared ===============================================
  common: {
    you: "You",
    guest: "Guest",
    player: "Player",
    opponent: "Opponent",
    nobody: "Nobody",
    avatarAlt: "Player avatar",
    cardBackAlt: "Card back",
    chipAlt: "Victory chip",
    myChipsTitle: "Your victory chips",
  },

  // ==== Admin =========================================================
  admin: {
    title: "Administrator Panel",
    subtitle: "Manage users and active hub sessions",
    refresh: "Refresh list ↻",
    backToDesktop: "← Back to desktop",
    table: {
      title: "All registered users",
      id: "ID / UUID",
      username: "Username",
      email: "Email",
      role: "Role",
      action: "Action",
    },
    block: "Block",
    blocked: "untouchable",
    emptyUsers: "No users found or loading...",
    blockConfirm: {
      title: "Block user",
      message:
        "Are you sure you want to block user {username}? This will restrict the player's access to the game hub.",
      confirm: "Block",
      cancel: "Cancel",
    },
    errors: {
      noAccess: "You do not have access to the administrator panel.",
      loadFailed: "Failed to load the list of users.",
      blockFailed: "Failed to block user {username}",
    },
    blockSuccess: "User {username} successfully blocked on the backend.",
  },

  // ==== Confirm modal ================================================
  modal: {
    confirm: {
      title: "Confirmation",
      message: "Are you sure you want to perform this action?",
      ok: "Confirm",
      cancel: "Cancel",
    },
  },

  // ==== Auth =========================================================
  auth: {
    title: "Sign in to play",
    username: "Login / Username",
    password: "Password",
    submit: "Sign in",
    registerPrompt: "Don't have an account yet?",
    registerLink: "Register",
    errors: {
      invalidCredentials: "Invalid login or password",
      noToken: "Server did not return a JWT access token",
    },
  },

  // ==== Register ======================================================
  register: {
    title: "Registration",
    username: "Login / Username",
    email: "Email address",
    emailPlaceholder: "user{'@'}localhost.com",
    password: "Password",
    submit: "Register",
    loginPrompt: "Already have an account?",
    loginLink: "Sign in",
    errors: {
      registrationFailed: "Failed to register a new user",
      noToken: "Server created the account but did not send an authorization token",
    },
  },

  // ==== Desktop =======================================================
  desktop: {
    brand: "Secret Message",
    welcome: "Welcome, {username}",
    profileTitle: "Profile settings",
    avatarChangeHint: "Change",
    myAvatarAlt: "My avatar",
    hostAvatarAlt: "Host avatar",
    adminLink: "Admin panel ⚙",
    logout: "Log out",
    createRoom: "+ Create room",
    lobbyHeader: "Available lobbies (real-time):",
    noRooms: "No active rooms or they already started a game. Create the first one!",
    lobbyEmpty: "No users found or loading...",
    roomLabel: 'Room "{id}"',
    playersCount: "Players:{count}/{max}",
    participants: "Participants:{names}",
    join: "Join",
    errors: {
      sessionExpired: "Session expired or unauthorized. Please sign in again.",
      loadRoomsFailed: "Failed to load the list of rooms",
      noRightsToCreate: "No rights to create a room (401)",
      createFailed: "Failed to create a room",
    },
  },

  // ==== Board =========================================================
  board: {
    leave: "Leave",
    room: "Room:",
    time: "Time:",
    timeUnit: "s",
    start: "Start game",
    waiting: "Waiting...",
    yourTurn: "Your turn!",
    currentTurn: "Turn:",
    protection: "Protection",
    discardPile: "Discard:",
    table: "Discard table",
    deck: "Deck",
    chancellorBadge: "Chancellor's choice!",
    chancellorPickOwn: "Pick a card to keep",
    waitingForCards: "Waiting for cards...",
    discardTooltip: "{name}— Discarded by {owner}",
    protectionTooltip: "{name}— Active protection",
    leaveConfirm: {
      title: "Leave the game",
      message:
        "Are you sure you want to leave the current game and return to the desktop lobby?",
      confirm: "Leave",
      cancel: "Stay",
    },
  },

  // ==== Action modal ==================================================
  action: {
    title: "Playing card:",
    targetSelect: "Pick an opponent's card:",
    noTargets: "No available targets!",
    noTargetsHint:
      "All other players are protected by the Handmaid's effect or are out. The card will be discarded without applying its effect.",
    autoApply: "This card is automatically applied to you or discarded to the table.",
    guardGuess: "Pick an opponent's card:",
    chosen: "Chosen",
    youSuffix: " (You)",
    cancel: "Cancel",
    submit: "Confirm move",
  },

  // ==== Card reveal modal =============================================
  reveal: {
    duelBaron: "Baron's Duel",
    priestEffect: "Priest's Effect",
    close: "Close and continue",
  },

  // ==== Chancellor modal ===============================================
  chancellor: {
    title: "Chancellor's Effect",
    step1: "Step 1: Pick 1 card you want to KEEP in your hand",
    step2: "Step 2: Pick the order of cards to send to the bottom of the deck",
    keepBadge: "Keep",
    bottomOrderTitle: "Order of cards to the bottom:",
    reset: "Reset selection",
    submit: "Confirm move",
  },

  // ==== User profile modal ============================================
  profile: {
    title: "Profile settings",
    subtitle: "Change your in-game avatar, name or password",
    avatarChange: "Random avatar",
    avatarAlt: "User avatar",
    usernameLabel: "Username",
    usernamePlaceholder: "Enter a nickname",
    passwordSectionDivider: "Password change",
    currentPassword: "Current password",
    currentPasswordPlaceholder: "Required to change the password",
    newPassword: "New password",
    newPasswordPlaceholder: "Enter the new password",
    save: "Save all",
    saving: "Saving...",
    cancel: "Cancel",
    errors: {
      currentRequired: "Please provide your current password to set a new one.",
      saveFailed: "Failed to save profile changes.",
      networkError: "Connection error with the server",
    },
  },

  // ==== Room manager ==================================================
  room: {
    connecting: "Connecting to the game room",
    cancel: "Cancel connection",
  },

  // ==== Game end modal ================================================
  gameEnd: {
    final: "👑 GAME FINAL 👑",
    roundEnd: "⚔️ END OF ROUND ⚔️",
    winner: "You are the absolute champion!",
    loser: "You won the round!",
    gameOver: "Game over",
    roundOver: "Round over",
    winnerLabel: "Winner:",
    scoreboardTitle: "Current score in the room:",
    scoreOutOf: "/ 7",
    waitingForPlayers: "Waiting for players...",
    waitingForHost: "Waiting for the room host to start a new game...",
    nextRound: "Next round",
    leaveLobby: "Back to lobby",
    restart: "Play again",
  },

  // ==== Game error modal ==============================================
  gameError: {
    title: "Game move error",
    server: "The server rejected your action",
    codeLabel: "code:",
    ok: "Got it",
    leave: "Back to lobby",
  },

  // ==== Status strings ================================================
  status: {
    protected: "Protected",
    out: "Out",
    inGame: "In game",
    waiting: "Waiting...",
  },

  // ==== Cards (fallbacks) =============================================
  cards: {
    unknown: "Unknown card",
    noDescription: "No description",
    serverHandled: "Effect handled on the server",
    SPY: { name: "Spy", desc: "Score one point at the end of the round." },
    GUARD: { name: "Guard", desc: "Guess another player's card." },
    PRIEST: { name: "Priest", desc: "Look at another player's hand." },
    BARON: { name: "Baron", desc: "Compare hands with another player." },
    HANDMAID: { name: "Handmaid", desc: "Protected from all effects until your next turn." },
    PRINCE: { name: "Prince", desc: "Choose any player (including yourself) to discard their hand." },
    CHANCELLOR: { name: "Chancellor", desc: "Draw 2 cards, keep 1, place the others on the bottom of the deck." },
    KING: { name: "King", desc: "Trade hands with another player." },
    COUNTESS: { name: "Countess", desc: "Must be discarded if you also hold the Prince or King." },
    PRINCESS: { name: "Princess", desc: "If you discard this card, you are out of the round." },
  },
};
