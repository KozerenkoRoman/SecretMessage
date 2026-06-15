// src/locales/en.ts
export const en = {
  // ==== Language switcher =============================================
  language: {
    switcherAriaLabel: "Language switcher",
    uk: "Українська",
    en: "English",
  },

  titles: {
    auth: 'Authorization | Secret Message',
    register: 'Registration | Secret Message',
    desktop: 'Game Desktop | Secret Message',
    admin: 'Admin Panel | Secret Message',
    room: 'Room | Secret Message',
    default: 'Secret Message'
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
    chipAlt: "Token of affection",
    myChipsTitle: "Your victory tokens",
  },

  // ==== Admin =========================================================
  admin: {
    title: "Admin Panel",
    subtitle: "Manage users and active hub sessions",
    refresh: "Refresh list ↻",
    backToDesktop: "← Back to Desktop",
    table: {
      title: "All Registered Users",
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
      title: "Block User",
      message:
        "Are you sure you want to block user {username}? This action will restrict the player's access to the game hub.",
      confirm: "Block",
      cancel: "Cancel",
    },
    errors: {
      noAccess: "You do not have permission to access the admin panel.",
      loadFailed: "Failed to load the user list.",
      blockFailed: "Failed to block user {username}",
    },
    blockSuccess: "User {username} has been successfully blocked on the backend.",
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
    title: "Sign In",
    username: "Login / Username",
    password: "Password",
    submit: "Enter Game",
    registerPrompt: "Don't have an account?",
    registerLink: "Register",
    errors: {
      invalidCredentials: "Incorrect username or password",
      noToken: "Server did not return an access JWT token",
    },
  },

  // ==== Register ======================================================
  register: {
    title: "Registration",
    username: "Login / Username",
    email: "Email Address",
    emailPlaceholder: "user{'@'}localhost.com",
    password: "Password",
    submit: "Register",
    loginPrompt: "Already have an account?",
    loginLink: "Log In",
    errors: {
      registrationFailed: "Failed to register a new user",
      noToken: "Account created successfully, but the server failed to send an auth token",
    },
  },

  // ==== Desktop =======================================================
  desktop: {
    brand: "Secret Message",
    welcome: "Welcome, {username}",
    profileTitle: "Profile Settings",
    avatarChangeHint: "Change",
    myAvatarAlt: "My avatar",
    hostAvatarAlt: "Host avatar",
    adminLink: "Admin Panel ⚙",
    logout: "Log Out",
    createRoom: "+ Create Room",
    lobbyHeader: "Available Lobbies (Live):",
    noRooms: "There are no active rooms, or they have already started. Create the first one!",
    lobbyEmpty: "No users found or loading...",
    roomLabel: 'Room "{id}"',
    playersCount: "Players: {count} / {max}",
    participants: "Participants: {names}",
    join: "Join",
    rulesButton: "Game Rules",
    leaderboard: {
      title: "Players Rating",
      loading: "Syncing with server...",
      empty: "Database is empty"
    },
    errors: {
      sessionExpired: "Session expired or unauthorized. Please log in again.",
      loadRoomsFailed: "Failed to load the room list",
      noRightsToCreate: "Unauthorized to create a room (401)",
      createFailed: "Failed to create a room",
    },
  },

  // ==== Board =========================================================
  board: {
    leave: "Leave",
    room: "Room:",
    time: "Time:",
    timeUnit: "s",
    start: "Start Game",
    waiting: "Waiting...",
    yourTurn: "Your Turn!",
    currentTurn: "Current turn:",
    protection: "Protected",
    discardPile: "Discard:",
    table: "Discard Pile",
    deck: "Deck",
    chancellorBadge: "Chancellor's Choice!",
    chancellorPickOwn: "Keep card for yourself",
    waitingForCards: "Waiting for cards...",
    discardTooltip: "{name} - Discarded by {owner}",
    protectionTooltip: "{name} - Protection active",
    leaveConfirm: {
      title: "Leave Game",
      message:
        "Are you sure you want to leave the current game and return to the lobby desktop?",
      confirm: "Leave",
      cancel: "Stay",
    },
  },

  // ==== Action modal ==================================================
  action: {
    title: "Play Card:",
    targetSelect: "Select an opponent's card:",
    noTargets: "No targets available!",
    noTargetsHint:
      "All other players are either protected by a Handmaid or eliminated. The card will be discarded into the pile without applying its effect.",
    autoApply: "This card targets yourself automatically or is discarded onto the table.",
    guardGuess: "Guess the opponent's card:",
    chosen: "Selected",
    youSuffix: " (You)",
    cancel: "Cancel",
    submit: "Confirm Move",
  },

  // ==== Card reveal modal =============================================
  reveal: {
    duelBaron: "Baron's Duel",
    priestEffect: "Priest's Effect",
    close: "Close & Continue",
  },

  // ==== Chancellor modal ===============================================
  chancellor: {
    title: "Chancellor's Effect",
    step1: "Step 1: Choose 1 card you want to KEEP in your hand",
    step2: "Step 2: Choose the order of cards to place at the bottom of the deck",
    keepBadge: "To Hand",
    bottomOrderTitle: "Bottom deck order:",
    reset: "Reset Selection",
    submit: "Confirm Move",
  },

  // ==== User profile modal ============================================
  profile: {
    title: "Profile Settings",
    subtitle: "Change your game avatar, username, or password",
    avatarChange: "Random Avatar",
    avatarAlt: "User avatar",
    usernameLabel: "Username",
    usernamePlaceholder: "Enter username",
    passwordSectionDivider: "Change Password",
    currentPassword: "Current Password",
    currentPasswordPlaceholder: "Required to change password",
    newPassword: "New Password",
    newPasswordPlaceholder: "Enter new password",
    save: "Save All",
    saving: "Saving...",
    cancel: "Cancel",
    errors: {
      currentRequired:
        "Please provide your current password to set a new one.",
      saveFailed: "Failed to save profile changes.",
      networkError: "Server connection error",
    },
  },

  // ==== Room manager ==================================================
  room: {
    connecting: "Connecting to the game room",
    cancel: "Cancel Connection",
  },

  // ==== Game end modal ================================================
  gameEnd: {
    final: "👑 GAME OVER - MATCH FINALS 👑",
    roundEnd: "⚔️ ROUND END ⚔️",
    winner: "You are the ultimate champion!",
    loser: "You won the round!",
    gameOver: "Game Over",
    roundOver: "Round Over",
    winnerLabel: "Winner:",
    scoreboardTitle: "Current Room Standings:",
    scoreOutOf: "/ 7",
    waitingForPlayers: "Waiting for players...",
    waitingForHost: "Waiting for the host to start the next game...",
    nextRound: "Next Round",
    leaveLobby: "Leave to Lobby",
    restart: "Play Again",
  },

  // ==== Game error modal ==============================================
  gameError: {
    title: "Invalid Move",
    server: "The server rejected your action",
    codeLabel: "code:",
    ok: "Understood",
    leave: "Leave to Lobby",
  },

  // ==== Status strings ================================================
  status: {
    protected: "Protected",
    out: "Out",
    inGame: "In Game",
    waiting: "Waiting...",
  },

  // ==== Cards (fallbacks) =============================================
  cards: {
    unknown: "Unknown Card",
    noDescription: "No description available",
    serverHandled: "Effect is being processed by the server",
    SPY: { name: "Spy", desc: "Gain one point at the end of the round if no one else played a Spy." },
    GUARD: { name: "Guard", desc: "Guess a non-Guard card in another player's hand." },
    PRIEST: { name: "Priest", desc: "Look at another player's hand." },
    BARON: { name: "Baron", desc: "Compare hands with another player; lower hand is out." },
    HANDMAID: { name: "Handmaid", desc: "You cannot be targeted by card effects until your next turn." },
    PRINCE: { name: "Prince", desc: "Choose any player (including yourself) to discard their hand." },
    CHANCELLOR: { name: "Chancellor", desc: "Draw 2 cards, keep 1, put the rest face down at the bottom of the deck." },
    KING: { name: "King", desc: "Trade hands with another player." },
    COUNTESS: { name: "Countess", desc: "Must be discarded if you hold the Prince or King." },
    PRINCESS: { name: "Princess", desc: "If you discard this card, you are out of the round." },
  },

  // Event reasons
  reasons: {
    guard_hit: "caught by a Guard",
    baron_lost: "defeated in a Baron duel",
    princess_played: "Princess discarded",
    left: "left the game",
    last_standing: "last player standing",
    deck_empty: "deck ran out of cards",
    baron_loss: "lost the duel to a Baron"
  },

  // Event log
  log: {
    baron_result: "💀 Based on the comparison , {winner} wins, and {loser} is eliminated from the round with {loser_card} card!",
    card_drawn: "{player} draws a card from the deck.",
    card_played_targeted: "{player} plays {card} against {target}.",
    card_played: "{player} plays {card}.",
    chancellor_drawn: "{player} played a Chancellor and drew 2 cards from the deck.",
    chancellor_resolved: "{player} kept one card and returned the rest to the bottom of the deck.",
    empty: "Players' moves will be displayed here...",
    guard_hit: "🎯 {player} plays a Guard against {target} and successfully guesses the card {guess}!",
    guard_miss: "💨 {player} plays a Guard against {target}, attempting to guess {guess}, but fails.",
    hands_swapped: "🔄 {player} plays a King and swaps hands with {target}.",
    history: "Game History",
    player_eliminated_with_card: "❌ Player {player} discards {card} and is eliminated from the round! (Reason: {reason}).",
    player_eliminated: "❌ Player {player} is eliminated from the round (Reason: {reason}).",
    player_left: "🚪 Player {player} has left the room.",
    priest_effect: "👁️ {player} uses a Priest to secretly look at {target}'s hand.",
    prince_effect: "👑 {player} forces {target} to discard {discarded_card} and draw a new card from the deck.",
    round_compared: "⚔️ {player} and {target} compare their cards.",
    round_end: "🏆 The round has ended! Winner: {winner} (Condition: {reason}).",
    spy_bonus: "✨ {player} receives bonus points ({points} pts) for the Spy!",
  },

  rules: {
    modalTitle: "Rules of «Secret Message»",
    generalHeader: "General Rules & Turn Order",
    cardsHeader: "Card Reference & Effects",
    nuancesHeader: "Important Nuances",
    strategyHeader: "Strategy Tips",
    strength: "Value",

    turnStep1: "Round Goal: Be the last player standing, or hold the card with the highest value when the deck runs out.",
    turnStep2: "Each card has a value (number) and a text effect (action on players).",
    turnStep3: "At the start of your turn, draw 1 card from the deck (giving you 2 cards in hand).",
    turnStep4: "Choose one of the two cards and play it face up in front of you.",
    turnStep5: "Execute its text effect immediately.",

    nuance1: "All played and discarded cards stay face up in front of players for the rest of the round.",
    nuance2: "If the deck runs out and there is a tie for the highest card, compare the total value of all cards in their respective discard piles. If still tied, all tied players win.",
    nuance3: "To win the match, you must collect a specific number of tokens of affection (depends on the player count).",

    strategy1: "Guard (1) and Baron (3) excel at eliminating rivals",
    strategy2: "Priest (2) provides essential intellect",
    strategy3: "Handmaid (4) grants safety for a round",
    strategy4: "Prince (5) is great for hunting the Princess",
    strategy5: "Chancellor (6) offers elite deck control",
    strategy6: "Spy (0) lets you snatch a bonus token without even winning the round."
  },

};