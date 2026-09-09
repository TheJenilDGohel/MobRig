# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-09-09

### Added
- **Core Engine Architecture**: Rebuilt MobRig from the ground up as a high-performance Go application compiled into a single, self-contained native binary with zero CGO dependencies.
- **Official Go MCP Server**: Built-in Model Context Protocol server using the official Go SDK over `stdio` with strict `stderr` stream hygiene. Implements 38 dual-aliased tools (`tap`/`mobile_tap`, `get_ui_tree`, `get_screenshot`, `swipe`, etc.) for AI agents like Claude Code, Cursor, and Windsurf.
- **Desktop Hub & Device Inspector**: Interactive graphical user interface built with Wails v3 and Svelte 5. Features real-time device gallery, hardware telemetry cards, phone bezel view, and quick hardware action triggers.
- **HTTP REST API Daemon**: RFC 7807 compliant HTTP service serving on port `8686` with complete CORS middleware and dynamic port binding for test runners.
- **Android Dual-Path Engine**: Optimized execution strategy prioritizing Google's official `android` CLI tool when available, with automatic degradation to standard `adb`.
- **Embedded SQLite Persistence**: CGO-free database powered by `modernc.org/sqlite` and embedded Goose migrations. Stores element fingerprints, OEM quirks, test run logs, and failure patterns.
- **Automated Environment Setup (`mobrig setup`)**: Built-in diagnostic CLI tool that verifies host toolchain availability and automatically updates configuration files for Cursor, Claude Code, and Claude Desktop.
- **Cross-Platform CI/CD**: Automated GitHub Actions workflows building and testing across Windows AMD64, Linux AMD64, macOS ARM64 (Apple Silicon), and macOS AMD64 (Intel).
- **Comprehensive Test Suite**: 73+ unit, integration, and end-to-end tests covering all platform adapters, database queries, and MCP tool endpoints.
