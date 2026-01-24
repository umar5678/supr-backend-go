# PowerShell + wscat Troubleshooting

## Your Error Explained

```
wscat -c "wss://api.pittapizzahusrev/go/ws?token=$TOKEN" > {"type":"message:send","data":{"rideId":"ride-123","content":"Hello!","messageType":"text"}}

Unexpected token ':"message:send"'
```

### What Went Wrong

In PowerShell, the `>` character is a **redirection operator** (like `|` in bash). You were trying to:
1. Connect to wscat
2. Redirect output (`>`) to a file named `{"type":"message:send"...`
3. PowerShell couldn't parse the JSON as a filename

### The Fix

**wscat is an interactive tool.** You need to:

1. **Connect first** (opens a prompt)
2. **Type messages interactively** (or pipe them)
3. **The `>` prompt appears automatically**

## Correct Workflow

### Method 1: Interactive (Manual)

```powershell
# Terminal 1: Connect
$TOKEN = "your_jwt_token"
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"

# Wait for: Connected (press CTRL+C to quit)
#           >

# Terminal continues - now type at the > prompt:
{"type":"message:send","data":{"rideId":"ride-123","content":"Hello!","messageType":"text"}}

# Press Enter - see response
# Type more messages or Ctrl+C to exit
```

### Method 2: Pipe Commands (Batch)

```powershell
# All messages at once
@"
{"type":"message:send","data":{"rideId":"ride-123","content":"Message 1","messageType":"text"}}
{"type":"message:typing","data":{"rideId":"ride-123","isTyping":true}}
{"type":"presence:online","data":{"rideId":"ride-123"}}
"@ | wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"
```

### Method 3: From File

```powershell
# Create messages.json
@"
{"type":"message:send","data":{"rideId":"ride-123","content":"Hello!","messageType":"text"}}
{"type":"message:read","data":{"messageId":"msg-456","rideId":"ride-123"}}
"@ | Out-File messages.json -Encoding UTF8

# Send from file
Get-Content messages.json | wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"
```

## Common Mistakes & Fixes

### ❌ Wrong: Using redirect `>`
```powershell
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN" > message.json
```
**Error:** Tries to redirect wscat output to file

**✅ Right:** Use Tee-Object to log
```powershell
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN" | Tee-Object -FilePath output.txt
```

### ❌ Wrong: Putting JSON on command line
```powershell
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN" {"type":"message:send"...}
```
**Error:** PowerShell tries to interpret `{` as a hash table

**✅ Right:** Send after connecting (interactively or piped)
```powershell
# Interactive: connect first, then type at >
# Piped: pipe to wscat
echo '{"type":"message:send",...}' | wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"
```

### ❌ Wrong: Token with `$` characters
```powershell
$TOKEN = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWI6$..."
# $ in double quotes triggers variable expansion
```

**✅ Right: Use single quotes or escape**
```powershell
# Single quotes (no expansion)
$TOKEN = 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWI$...'

# Or escape with backtick
$TOKEN = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWI`$..."
```

## Step-by-Step Video Guide (Text Format)

### Scenario: Send a message from PowerShell

**Step 1: Open PowerShell**
```powershell
PS C:\Users\YourUser>
```

**Step 2: Get your JWT token**
```powershell
# Login to get token (save it)
$TOKEN = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Step 3: Connect to wscat**
```powershell
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"
```

**Step 4: See connection confirmation**
```
Connected (press CTRL+C to quit)
>
```

**Step 5: Type the message (at the `>` prompt)**
```
> {"type":"message:send","data":{"rideId":"ride-123","content":"Hello!","messageType":"text"}}
```

**Step 6: Press Enter**

**Step 7: See the server response**
```
{"type":"message:new","data":{"message":{...},"rideId":"ride-123"}}
```

**Step 8: Send more or exit**
```
> {"type":"presence:online","data":{"rideId":"ride-123"}}
# Or Ctrl+C to disconnect
```

## Testing with Multiple Clients

### Terminal 1 (Client A - Driver)
```powershell
$TOKEN_A = "driver_token"
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN_A"
# Wait for connection...
# (you'll see messages from Client B here)
```

### Terminal 2 (Client B - Rider)
```powershell
$TOKEN_B = "rider_token"
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN_B"

# Once connected, send a message:
> {"type":"message:send","data":{"rideId":"ride-123","content":"Where are you?","messageType":"text"}}
```

### Terminal 1 (You'll see the message automatically)
```
{"type":"message:new","data":{"message":{...,"content":"Where are you?",...}}}
```

## Debugging Commands

### Check wscat installation
```powershell
npm list -g wscat
# or
where.exe wscat
```

### Check token format
```powershell
# Tokens are 3 parts separated by dots
$TOKEN = "part1.part2.part3"
$TOKEN.Split('.').Count  # Should be 3
```

### Test localhost connection
```powershell
Test-NetConnection localhost -Port 8080
```

### Test WebSocket URL (without token)
```powershell
# This should fail (no auth):
wscat -c "ws://localhost:8080/ws/connect"

# Should show: error - authentication failed
```

### View server logs
```powershell
# In another terminal, watch server output:
Get-Content logs.txt -Wait
```

## Quick Reference: Common Commands

| Task | Command |
|------|---------|
| Install wscat | `npm install -g wscat` |
| Connect | `wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"` |
| Send message | `{"type":"message:send","data":{"rideId":"...","content":"...","messageType":"text"}}` |
| Mark as read | `{"type":"message:read","data":{"messageId":"...","rideId":"..."}}` |
| Show typing | `{"type":"message:typing","data":{"rideId":"...","isTyping":true}}` |
| Go online | `{"type":"presence:online","data":{"rideId":"..."}}` |
| Go offline | `{"type":"presence:offline","data":{"rideId":"..."}}` |
| Exit wscat | `Ctrl+C` |

## Success Indicators

✅ **Connection successful:**
```
Connected (press CTRL+C to quit)
>
```

✅ **Message sent successfully:**
```
(You see the same message echoed back)
{"type":"message:new","data":{"message":{...}}}
```

✅ **Multiple clients in sync:**
```
Client A sends message -> Client B sees it immediately
Client B types indicator -> Client A sees "typing..."
```

## Support

If you still have issues:

1. Check `POWERSHELL-WEBSOCKET-GUIDE.md` for detailed explanations
2. Check `WEBSOCKET-TESTING-GUIDE.md` for general WebSocket help
3. Check `QUICKSTART-MESSAGING.md` for quick setup
4. Check server logs: `Get-Content server.log -Wait`

**Key Takeaway:**
> wscat is interactive. You connect first (`wscat -c ...`), then type messages at the `>` prompt. The `>` is NOT typed by you - it appears automatically!

Good luck! 🚀
