-- +goose Up
CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    platform TEXT NOT NULL CHECK(platform IN ('android', 'ios')),
    model TEXT,
    manufacturer TEXT,
    os_version TEXT,
    api_level INTEGER,
    screen_width INTEGER,
    screen_height INTEGER,
    screen_density REAL,
    ram_mb INTEGER,
    features TEXT DEFAULT '[]',  -- JSON array
    first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    total_sessions INTEGER DEFAULT 0,
    connection_type TEXT CHECK(connection_type IN ('usb', 'wifi', 'emulator', 'simulator')),
    nickname TEXT
);

CREATE TABLE IF NOT EXISTS device_quirks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    quirk_type TEXT NOT NULL,
    description TEXT NOT NULL,
    workaround TEXT,
    confidence REAL DEFAULT 0.5 CHECK(confidence >= 0.0 AND confidence <= 1.0),
    discovered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, quirk_type)
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    app_package TEXT,
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ended_at DATETIME,
    actions_count INTEGER DEFAULT 0,
    failures_count INTEGER DEFAULT 0,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS flows (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    app_package TEXT,
    device_platform TEXT,
    steps TEXT NOT NULL DEFAULT '[]',  -- JSON array of actions
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME,
    total_runs INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    avg_duration_ms INTEGER
);

CREATE TABLE IF NOT EXISTS element_fingerprints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_package TEXT NOT NULL,
    screen_name TEXT,
    element_text TEXT,
    element_type TEXT,
    resource_id TEXT,
    content_desc TEXT,
    stable_selector TEXT,
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    reliability REAL DEFAULT 1.0,
    UNIQUE(app_package, stable_selector)
);

CREATE TABLE IF NOT EXISTS failure_patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT REFERENCES devices(id) ON DELETE SET NULL,
    app_package TEXT,
    pattern_type TEXT NOT NULL,
    description TEXT NOT NULL,
    stack_trace TEXT,
    frequency INTEGER DEFAULT 1,
    first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS schema_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT OR IGNORE INTO schema_meta (key, value) VALUES ('version', '0.1.0');
INSERT OR IGNORE INTO schema_meta (key, value) VALUES ('created_at', CURRENT_TIMESTAMP);

CREATE INDEX IF NOT EXISTS idx_sessions_device ON sessions(device_id);
CREATE INDEX IF NOT EXISTS idx_sessions_app ON sessions(app_package);
CREATE INDEX IF NOT EXISTS idx_flows_name ON flows(name);
CREATE INDEX IF NOT EXISTS idx_quirks_device ON device_quirks(device_id);
CREATE INDEX IF NOT EXISTS idx_failures_device ON failure_patterns(device_id);
CREATE INDEX IF NOT EXISTS idx_failures_app ON failure_patterns(app_package);

-- +goose Down
DROP TABLE IF EXISTS schema_meta;
DROP TABLE IF EXISTS failure_patterns;
DROP TABLE IF EXISTS element_fingerprints;
DROP TABLE IF EXISTS flows;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS device_quirks;
DROP TABLE IF EXISTS devices;
