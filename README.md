# kal

kal is a lightweight command-line tool that helps you track how much time you spend on your real-life activities (touching grass for example).
You can start and stop timers for different activities, view stats, and list active sessions — all from the terminal.

### Features

- Track time spent on various activities
- Start, stop, and resume activity tracking
- View detailed statistics for each activity
- List all activities or only those currently active
- Output data in JSON for scripting or integration with your WM or whatever

### Installation
1. Clone this repository
```bash
git clone https://github.com/deservehumanity/kal.git
cd kal
```
2. Build the project (ensure golang is installed) (optionally add to PATH)
```bash
go build -o kal ./cmd/kal
sudo mv kal /usr/local/bin/
```
or (if not afraid that I might steal your precious crypto)
```bash
sudo ./install.sh
```
3. Check installation
```bash
kal --help
```

### Example Usage
```bash
kal new-activity coding
kal start coding
# ... a few minutes later ...
kal stop coding

kal stat coding
# Activity: coding
# Total time spent (formatted): 00h 45m 32s
# Total time spent (hours): 0.758889
```

##### Planned Features

- SQLite-based backend for faster queries
- Custom hooks for integrations (e.g. desktop widgets, status bars, WMs)
- Enhanced reporting and summaries

##### License
MIT License © 2025 deservehumanity
