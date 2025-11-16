find and kill port 

if windows, use powershell

netstat -ano | findstr :8080

result will be: 
TCP    0.0.0.0:8080     0.0.0.0:0     LISTENING     12345

then 
taskkill /PID 12345 /F
