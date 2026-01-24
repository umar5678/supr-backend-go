# WebSocket Testing with wscat - PowerShell Guide

## Quick Fix for Your Error

The error you got:
```
Unexpected token ':"message:send"' in expression or statement.
```

**Root Cause:** You tried to use `>` as a PowerShell redirect operator. In wscat, the `>` is the **prompt** (appears automatically), not a command.

## Correct Usage

### Step 1: Connect to wscat

```powershell
$TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI4NDJlOWVmZi0wMjNiLTQ0N2YtODgzYS0zMjBlYTM1ZDJmNjMiLCJyb2xlIjoicmlkZXIiLCJleHAiOjE3Njk0ODEyMDYsIm5iZiI6MTc2OTEwMzIwNiwiaWF0IjoxNzY5MTAzMjA2fQ.hwl66YA1240vuuYq8r4ensscZGGYAkVS-y8A03kZnww"
wscat -c "wss://api.pittapizzahusrev.be/go/ws?token=$TOKEN"
```

### Step 2: Wait for Connection

You should see:
```
Connected (press CTRL+C to quit)
>
```

### Step 3: Send JSON (in the wscat prompt)

The `>` appears automatically. Just type/paste the JSON and press Enter:

```json
{"type":"message:send","data":{"rideId":"7733547e-9338-4dd9-9b97-8c888e36cc0a","content":"Hello!","messageType":"text"}}
```

### Step 4: See Response

The server will broadcast back:

```json
{"type":"message:new","data":{"message":{...},"rideId":"ride-123"}}
```

## Complete PowerShell Session Example

```powershell
# Step 1: Set your JWT token
$TOKEN = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."

# Step 2: Connect
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"

# You see: Connected (press CTRL+C to quit)
# Prompt shows: >

# Step 3: Send a message (paste this in the prompt)
{"type":"message:send","data":{"rideId":"ride-123","content":"Hello from PowerShell!","messageType":"text"}}

# Step 4: See the broadcast response
# You should see the message echoed back

# Step 5: Send more commands
{"type":"message:typing","data":{"rideId":"ride-123","isTyping":true}}

# Step 6: Exit
# Press Ctrl+C
```

## Multiple Commands (Batch Mode)

If you want to send multiple messages automatically, use a here-string:

```powershell
$TOKEN = "your_jwt_token_here"
$commands = @"
{"type":"message:send","data":{"rideId":"ride-123","content":"Message 1","messageType":"text"}}
{"type":"message:typing","data":{"rideId":"ride-123","isTyping":true}}
{"type":"presence:online","data":{"rideId":"ride-123"}}
"@

$commands | wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"
```

## Save Commands to File

Create `messages.txt`:
```json
{"type":"message:send","data":{"rideId":"ride-123","content":"Hello!","messageType":"text"}}
{"type":"message:read","data":{"messageId":"msg-456","rideId":"ride-123"}}
{"type":"message:typing","data":{"rideId":"ride-123","isTyping":true}}
{"type":"presence:online","data":{"rideId":"ride-123"}}
```

Then run:
```powershell
$TOKEN = "your_token"
Get-Content messages.txt | wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"
```

## Interactive Script (Save as `test-ws.ps1`)

```powershell
#!/usr/bin/env pwsh

# Get token from environment or ask user
$TOKEN = $env:WS_TOKEN
if (-not $TOKEN) {
    $TOKEN = Read-Host "Enter your JWT token"
}

$rideId = Read-Host "Enter ride ID (default: ride-123)" 
if (-not $rideId) { $rideId = "ride-123" }

Write-Host "Connecting to ws://localhost:8080/ws/connect?token=$($TOKEN.Substring(0,20))..." -ForegroundColor Cyan

# Build commands
$commands = @(
    @{ type = "message:send"; data = @{ rideId = $rideId; content = "Hello from PowerShell!"; messageType = "text" } } | ConvertTo-Json -Compress,
    @{ type = "message:typing"; data = @{ rideId = $rideId; isTyping = $true } } | ConvertTo-Json -Compress,
    @{ type = "presence:online"; data = @{ rideId = $rideId } } | ConvertTo-Json -Compress
)

Write-Host "Sending $($commands.Count) commands..." -ForegroundColor Green

$commands | wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"
```

Run it:
```powershell
.\test-ws.ps1
```

## Common PowerShell Issues & Fixes

### Issue 1: Token Contains `$` Characters

If your token has `$` in it, escape it:

```powershell
# Wrong:
$TOKEN = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWI..."

# Right (use single quotes):
$TOKEN = 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWI...'

# Or escape with backtick:
$TOKEN = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWI`$..."
```

### Issue 2: `>` in JSON Causes Redirect

**Wrong:**
```powershell
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN" > output.json
```

This tries to redirect wscat output to a file. Instead:

**Right (for logging):**
```powershell
# Use Tee-Object to both display and log
$commands | wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN" | Tee-Object -FilePath output.txt
```

### Issue 3: JSON with Double Quotes

PowerShell might interpret `"` characters. Use `@'...'@` (here-string):

```powershell
$command = @'
{"type":"message:send","data":{"rideId":"ride-123","content":"Hello!","messageType":"text"}}
'@

$command | wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"
```

## Recommended: Create a PowerShell Function

Add to your PowerShell profile (`$PROFILE`):

```powershell
function Send-WebSocketMessage {
    param(
        [Parameter(Mandatory=$true)]
        [string]$Token,
        
        [Parameter(Mandatory=$true)]
        [string]$RideId,
        
        [string]$Content = "Test message",
        [string]$MessageType = "text",
        [string]$Type = "message:send"
    )
    
    $message = @{
        type = $Type
        data = @{
            rideId = $RideId
            content = $Content
            messageType = $MessageType
        }
    } | ConvertTo-Json -Compress
    
    Write-Host "Sending: $message" -ForegroundColor Yellow
    
    $message | wscat -c "ws://localhost:8080/ws/connect?token=$Token"
}

# Usage:
# Send-WebSocketMessage -Token $TOKEN -RideId "ride-123" -Content "Hello!"
```

## Environment Variable Method (Recommended)

```powershell
# Set token as environment variable (one time)
$env:WS_TOKEN = "your_jwt_token_here"

# Now use it anytime
wscat -c "ws://localhost:8080/ws/connect?token=$env:WS_TOKEN"
```

## Debugging

### Check if wscat is installed
```powershell
where.exe wscat
```

### Test connection without messages
```powershell
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"
# Just wait to see if connection succeeds
# Ctrl+C to exit
```

### Check token validity
```powershell
# Decode JWT (requires System.IdentityModel.Tokens.Jwt)
[System.Reflection.Assembly]::LoadWithPartialName("System.IdentityModel.Tokens.Jwt") | Out-Null

$token = "your_token_here"
$parts = $token.Split('.')
$payload = $parts[1] + '=' * (4 - $parts[1].Length % 4)
$decoded = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($payload))
$decoded | ConvertFrom-Json
```

## Summary

✅ **Remember:**
- Connect first: `wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"`
- Wait for prompt: `>`
- The `>` is automatic, don't type it
- Paste JSON and press Enter
- Use single quotes for tokens with special characters
- Ctrl+C to disconnect

Now you're ready to test WebSocket messaging from PowerShell! 🎉
