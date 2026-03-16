# Game Development Guidelines

## Game Naming Strategy

GameID consists of 3 parts:

```
<provider_prefix[2]>-<game_name>-<RTP>
```

**Components**:
- `<provider_prefix[2]>` - Two-letter prefix for game provider (SwipeGames: `sg`)
- `<game_name>` - Game name in lowercase with hyphens instead of spaces
- `<RTP>` - RTP value of game (01-99)

**Example**: `sg-catch-95` (SwipeGames Catch game with 95% RTP)

## Game Launch

### Base URL
All games started from same base URL defined in `GAMES_BASE_URL` environment variable for core service.

### Launch URL Format
```
<GAMES_BASE_URL>/<game_id>
```

### Parameter Precedence
When same parameters passed through Core API and launch URL:
- Launch URL parameters have precedence
- **Cannot override**: `demo`, `currency`, `locale` (handled by Core API)

## Demo Mode

### Behavior
- Start new game session per request
- No level progression between sessions
- No persistent data between sessions

### User Identification
- API allows integration to pass `userID` for demo mode
- If `userID` provided, may have progression (not cleaned on every start)
- Without `userID`, strictly no persistent data

### Testing Demo Games
- Public API requires API tokens (not suitable for testing)
- Use internal endpoint to start game session with demo parameters
- Avoids exposing API tokens

## UGID (User Global ID)

Clearly identifies user across different aggregators and casinos.

### Format
```
<CID>-<ExtCID>-<UID>
```

**Components**:
- `<CID>` - Client ID
- `<ExtCID>` - External Client ID
- `<UID>` - User ID

**Note**: Demo mode and currency already provided in game service (not in UGID).

## Game Config and Round Concepts

### Active Game Config
- **Only one** active game config per game session
- Used to generate random values for game round
- Rotate config on bet or close round
- Rotation creates new config and closes old one (set as `rotated`)
- Always single `active` config per session

### Round
- Related to specific game round
- Created when user starts playing (`open round` call)
- Closed when user finishes round (`close round` or `bet` call)
- Related to specific game config active during round
- **1:1 relation** between round and game config

### Open/Closed Rounds

**Open but not closed**:
- Happens when round opened but not bet placed
- Config might be `rotated` but round not `closed`

**Identifying played rounds**:
- Use only `closed` rounds for history and statistics
- Open but not closed rounds cleaned by background job

### Security in Async Games

For games requiring multiple actions to finish round:
- Be careful not to save sensitive data in round (outcomes)
- Store round info for async games in Redis (not database)
- Prevents cheating based on round data

## Game Session

### Duration
- Default: 4 hours (configurable up to 24 hours via environment)

### Creation
- Every `create-new-game` API call creates new session
- Even if external session ID same
- Improves speed (no search for existing session)

### Management
- Fully managed by core service
- Other services use `get-game-session-info` private API endpoint

### Game Settings
- Session contains game-specific settings (min/max bets)
- Prevents changing game settings during session

## Game Settings

### Purpose
Define rules and parameters: min/max bets, allowed currencies, game-specific configs.

### Currency-Based Settings
- Settings defined against specific currency
- If currency not set, game cannot be played
- No base currency conversion (direct currency-based setup)

### Lookup Algorithm

When game settings requested:

1. Check exact match: `CID`, `ExtCID`, `GameID`, `currency`
2. If not found: `CID`, `ExtCID`, `currency` with `GameID = null`
3. If not found: `CID`, `currency` with `ExtCID = null` and `GameID = null`
4. If not found: `currency` only with `CID`, `ExtCID`, `GameID = null`
5. If not found: game cannot be played (no settings)

### Wildcard Configurations
- **GameID-wide**: `GameID = null` (applies to all games)
- **ExtCID-wide**: `ExtCID = null` (applies to all external CIDs)
- **CID-wide**: `CID = null` (applies to all CIDs)

## Free Rounds

See [SwipeGames Public API](https://swipegames.github.io/public-api/free-rounds) for details.

### Bet Lines Lookup Algorithm

When bet lines requested:

1. Check exact match: `CID`, `ExtCID`, `GameID`, `currency`
2. If not found: `CID`, `ExtCID`, `currency` with `GameID = null` (wildcard game)
3. If not found: `CID`, `currency` with `ExtCID = null` and `GameID = null` (wildcard ExtCID)
4. If not found: bet lines not defined, free rounds cannot be played

**Note**:
- Currency required parameter (currency-based config)
- When `ExtCID` not provided, don't match against `GameID`
- Don't create configs with `ExtCID = null` but `GameID` provided

### Game Integration with Free Rounds

**Process**:
1. Game performs bet or accept/decline free rounds
2. Pulls from core and persists in game repository
3. If admin cancels free rounds, user still finishes them
4. Once finished, process updates balances (bonus to real money)
5. Process started/located from game's side

## Bonus Balance Processing

### Demo Mode
1. Core calls `ledgerCl.ApplyBonusBalance()` with account ID and transaction ID
2. Ledger atomically transfers bonus to real balance
3. No external casino API calls
4. Completes synchronously

### Production Mode (Non-Demo)

**Synchronous Phase** (ApplyBonus usecase):
1. Check bonus balance exists in ledger
2. Create outbox record with `response_tx_id = NULL` (pending)
3. Return current balance to caller

**Asynchronous Phase** (ApplyBonusProcessingService):
1. Background service runs periodically (default: 10 seconds)
2. Pick pending outbox records (`response_tx_id IS NULL`)
3. Execute workflow:
   - Call `ledgerCl.WithdrawBonusBalanceStart()` (lock funds)
   - Call `integrationCl.Win()` (credit via external API)
   - Call `ledgerCl.WithdrawBonusBalanceFinish()` (complete transfer)
4. Update outbox with result

**Error Handling**:
- Failed attempts remain pending (`response_tx_id IS NULL`)
- Auto-retry on next cycle
- Records older than 3 months excluded from retry
- Errors logged in `response_error` field
