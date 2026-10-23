@echo off
rem ==============================================================================
rem   NATIONAL HIGHWAYS INFRA TRUST (NHIT) - ROOT HARNESS WINDOWS BATCH ENTRYPOINT
rem ==============================================================================
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0harness\run-harness.ps1" %*
exit /b %ERRORLEVEL%
