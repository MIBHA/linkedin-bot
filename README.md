# LinkedIn Automation Bot

A professional LinkedIn automation tool built with Go and the Rod library, featuring advanced anti-detection techniques for educational purposes.

> **DISCLAIMER**: This tool is for educational purposes only. Automated interaction with LinkedIn may violate their Terms of Service. Use at your own risk.

## Demonstration Video
[![Watch the Demonstration](https://github.com/MIBHA/linkedin-bot/blob/main/pkg/utils/assets/Video%20Demonstration.png?raw=true)](https://www.loom.com/share/e644e883e47b4f5ca18de6deeba4d851)

## Features

### Advanced Stealth Techniques
- **Bézier Curve Mouse Movements**: Natural, curved mouse paths with randomized control points
- **Human-like Typing**: Variable keystroke delays with occasional typos and corrections
- **Fingerprint Masking**: Removes automation indicators and randomizes browser fingerprints
- **Random Delays**: Realistic pauses between actions

### Session Management
- Cookie persistence across sessions
- Automatic session validation
- 2FA detection and manual completion support

### State Tracking
- SQLite database for connection request history
- Duplicate prevention
- Statistics and analytics
- Profile visit tracking

### Automation Features
- People search with filters
- Bulk connection requests
- Personalized connection messages
- Rate limiting and daily quotas

## Project Structure

```
linkedin-bot/
├── cmd/
│   └── linkedin-bot/
│       └── main.go              # Application entry point
├── pkg/
│   ├── stealth/
│   │   ├── mouse.go             # Bézier curve mouse movements
│   │   ├── typing.go            # Human-like typing
│   │   └── fingerprint.go       # Browser fingerprint masking
│   ├── browser/
│   │   ├── browser.go           # Rod browser initialization
│   │   └── session.go           # Cookie/session management
│   └── linkedin/
│       ├── login.go             # Login automation
│       ├── search.go            # Profile search
│       └── connect.go           # Connection requests
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   └── storage/
│       ├── database.go          # Database operations
│       └── models.go            # GORM models
├── .env.example                 # Environment variables template
├── .gitignore
├── go.mod
└── README.md
```

## Installation

### Prerequisites
- Go 1.21 or higher
- Chrome/Chromium browser (automatically managed by Rod)

### Setup

1. **Clone the repository**
   ```bash
   cd linkedin-bot
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment variables**
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env` and add your LinkedIn credentials:
   ```env
   LINKEDIN_EMAIL=your.email@example.com
   LINKEDIN_PASSWORD=your_password_here
   ```

4. **Build the application**
   ```bash
   go build -o linkedin-bot.exe ./cmd/linkedin-bot
   ```

## Usage

### Run the bot
```bash
./linkedin-bot.exe
```

Or run directly with Go:
```bash
go run ./cmd/linkedin-bot/main.go
```

### Interactive Menu
The bot provides an interactive CLI menu:
1. **Search and Connect** - Search for profiles and send connection requests
2. **View Statistics** - Display connection request statistics
3. **Exit** - Save session and exit

### Configuration Options

Edit `.env` to customize behavior:

```env
# Browser settings
HEADLESS=false                    # Set to true for headless mode
SLOW_MOTION=0                     # Add delay (ms) for debugging

# Rate limiting
MAX_CONNECTIONS_PER_DAY=50        # Maximum connections per session
DELAY_BETWEEN_ACTIONS=30          # Seconds between actions

# Search settings
MAX_RESULTS=100                   # Maximum search results to process
```

## Anti-Detection Features

### Bézier Curve Mouse Movement
```go
// Natural mouse movement with randomized path
stealth.MoveMouseBezier(page, targetX, targetY, opts)
```

### Human-like Typing
```go
// Types with variable delays and occasional typos
stealth.TypeHumanLike(page, "Hello World", opts)
```

### Fingerprint Masking
- Removes `navigator.webdriver` property
- Randomizes User-Agent
- Varies viewport dimensions
- Adds canvas noise
- Masks WebGL fingerprints

## Database Schema

The bot uses SQLite to track:
- **Connection Requests**: Profile URL, name, message, status, timestamp
- **Message History**: Recipient, message text, conversation ID
- **Search History**: Query, filters, result count
- **Profile Visits**: Visit count and timestamps

## Safety Features

- **Duplicate Prevention**: Checks database before sending requests
- **Rate Limiting**: Configurable delays between actions
- **Daily Quotas**: Limits connections per session
- **Graceful Shutdown**: Saves session on Ctrl+C

## Development

### Run with verbose logging
```bash
go run ./cmd/linkedin-bot/main.go
```

### Run tests
```bash
go test ./...
```

### View database
```bash
sqlite3 linkedin_bot.db
.tables
SELECT * FROM connection_requests;
```

## Troubleshooting

### Login Issues
- Ensure credentials are correct in `.env`
- If 2FA is enabled, the bot will pause for manual completion
- Check if LinkedIn requires CAPTCHA (manual intervention needed)

### Detection Issues
- Increase `DELAY_BETWEEN_ACTIONS` for more conservative behavior
- Reduce `MAX_CONNECTIONS_PER_DAY` to avoid rate limits
- Use `HEADLESS=false` to monitor browser behavior

### Browser Issues
- Rod automatically downloads Chrome if not found
- Ensure you have internet connection for first run
- Check firewall settings if browser fails to launch

## Legal & Ethical Considerations

**Important**: This tool is provided for educational purposes only. Automated interaction with LinkedIn may:
- Violate LinkedIn's Terms of Service
- Result in account suspension or ban
- Be considered unauthorized access in some jurisdictions

**Use responsibly and at your own risk.**

## License

This project is for educational purposes only. Use at your own risk.

## Contributing

This is an educational project. Feel free to fork and modify for learning purposes.

## Acknowledgments

- [Rod](https://github.com/go-rod/rod) - Browser automation library
- [GORM](https://gorm.io/) - ORM library
- [godotenv](https://github.com/joho/godotenv) - Environment variable management
