@echo off
REM Aether Development Script for Windows
REM This script helps with local development using docker-compose

REM Check if docker is installed
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker is not installed. Please install Docker Desktop first.
    exit /b 1
)

REM Check if docker-compose is installed
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] docker-compose is not installed. Please install Docker Desktop first.
    exit /b 1
)

REM Main script logic
if "%1"=="" goto help
if "%1"=="start" goto start
if "%1"=="stop" goto stop
if "%1"=="restart" goto restart
if "%1"=="logs" goto logs
if "%1"=="status" goto status
if "%1"=="test" goto test
goto help

:help
echo Aether Development Script for Windows
echo Usage: %0 {start|stop|restart|logs|status|test}
echo.
echo Commands:
echo   start   - Build and start all services
echo   stop    - Stop all services
echo   restart - Restart all services
echo   logs    - View logs for all services
echo   status  - Show status of all services
echo   test    - Run tests for all services
exit /b 0

:start
echo [STATUS] Starting Aether services...

REM Check if .env file exists in synaptic-core
if not exist ".\synaptic-core\.env" (
    echo [WARNING] No .env file found in synaptic-core.
    if exist ".\synaptic-core\.env.example" (
        copy ".\synaptic-core\.env.example" ".\synaptic-core\.env" >nul
        echo [STATUS] Created .env file from .env.example
    ) else (
        type nul > ".\synaptic-core\.env"
        echo [WARNING] Created empty .env file. Please add your GEMINI_API_KEY.
    )
)

REM Build and start services
docker-compose up --build -d

echo [STATUS] Services started successfully!
echo [STATUS] Frontend: http://localhost:3000
echo [STATUS] Backend API: http://localhost:8080
echo [STATUS] Backend WebSocket: ws://localhost:8080/ws
exit /b 0

:stop
echo [STATUS] Stopping Aether services...
docker-compose down
echo [STATUS] Services stopped successfully!
exit /b 0

:restart
call :stop
call :start
exit /b 0

:logs
echo [STATUS] Viewing logs for all services...
docker-compose logs -f
exit /b 0

:status
echo [STATUS] Showing service status...
docker-compose ps
exit /b 0

:test
echo [STATUS] Running tests...
REM Run backend tests
if exist ".\synaptic-core" (
    echo [STATUS] Running backend tests...
    docker-compose exec synaptic-core go test ./...
)

REM Run frontend tests
if exist ".\interactive-carapace" (
    echo [STATUS] Running frontend tests...
    docker-compose exec interactive-carapace pnpm test
)
exit /b 0