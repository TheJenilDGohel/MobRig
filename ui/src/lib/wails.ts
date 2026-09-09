/**
 * Wails v3 runtime bridge for MobRig.
 * Provides type-safe wrappers around Go backend methods exposed on MobRigService
 * with seamless fallback to localhost:8686 HTTP REST API for browser dev mode.
 */

export interface HealthInfo {
  status: string;
  version: string;
  uptime: string;
  http_port: number;
  debug: boolean;
}

export interface ConfigInfo {
  data_dir: string;
  http_port: number;
  http_host: string;
  db_path: string;
  debug: boolean;
  version: string;
}

export type DeviceStatus = 'connected' | 'offline' | 'busy';
export type DevicePlatform = 'android' | 'ios';

export interface DeviceListItem {
  id: string;
  name: string;
  os: string;
  status: DeviceStatus;
  platform: DevicePlatform;
  connection_type: string;
}

export interface DeviceInfo {
  id: string;
  platform: DevicePlatform;
  model: string;
  manufacturer: string;
  os_version: string;
  api_level?: number;
  screen_width: number;
  screen_height: number;
  screen_density?: number;
  ram_mb?: number;
  status: DeviceStatus;
  connection_type: string;
}

export interface ScreenshotData {
  width: number;
  height: number;
  size_bytes: number;
  format: string;
  data: string; // base64 encoded string
}

export interface ActionResult {
  success: boolean;
  message: string;
  duration?: string;
}

const API_BASE = 'http://127.0.0.1:8686';

/**
 * Returns true if running within the Wails desktop webview environment.
 */
export function isWailsEnvironment(): boolean {
  return typeof window !== 'undefined' && ('_wails' in window || Boolean((window as any).wails));
}

/**
 * Invokes a Go service method exposed via Wails v3 bindings.
 */
async function callBinding<T>(methodName: string, ...args: any[]): Promise<T> {
  const callId = Math.random().toString(36).substring(2, 9);
  const payload = {
    object: 0,
    method: 0,
    args: {
      'call-id': callId,
      methodName,
      args,
    },
  };

  const res = await fetch('/wails/runtime', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    throw new Error(`Wails service call failed: ${methodName}`);
  }
  return (await res.json()) as T;
}

/**
 * Fetches application health status.
 */
export async function getHealth(): Promise<HealthInfo> {
  try {
    return await callBinding<HealthInfo>('main.MobRigService.GetHealth');
  } catch {
    try {
      const res = await fetch(`${API_BASE}/health`);
      if (res.ok) {
        const data = await res.json();
        return {
          status: data.status ?? 'ok',
          version: data.version ?? '0.1.0-dev',
          uptime: data.uptime ?? '0s',
          http_port: 8686,
          debug: false,
        };
      }
    } catch {
      // Offline fallback
    }

    return {
      status: 'offline',
      version: '0.1.0-dev',
      uptime: '0s',
      http_port: 8686,
      debug: false,
    };
  }
}

/**
 * Fetches configuration information.
 */
export async function getConfig(): Promise<ConfigInfo> {
  try {
    return await callBinding<ConfigInfo>('main.MobRigService.GetConfig');
  } catch {
    return {
      data_dir: '~/.mobrig',
      http_port: 8686,
      http_host: '127.0.0.1',
      db_path: '~/.mobrig/mobrig.db',
      debug: false,
      version: '0.1.0-dev',
    };
  }
}

/**
 * Fetches currently connected devices.
 */
export async function getDevices(): Promise<DeviceListItem[]> {
  try {
    return await callBinding<DeviceListItem[]>('main.MobRigService.GetDevices');
  } catch {
    try {
      const res = await fetch(`${API_BASE}/api/v1/devices`);
      if (res.ok) {
        const data = await res.json();
        return (data.devices || []).map((d: any) => ({
          id: d.id,
          name: d.model || d.id,
          os: d.platform ? `${d.platform}` : 'Android',
          status: d.status || 'connected',
          platform: d.platform || 'android',
          connection_type: d.connection_type || 'usb',
        }));
      }
    } catch {
      // Ignored
    }
    return [];
  }
}

/**
 * Fetches detailed specs for a specific device.
 */
export async function getDeviceInfo(id: string): Promise<DeviceInfo | null> {
  try {
    return await callBinding<DeviceInfo>('main.MobRigService.GetDeviceInfo', id);
  } catch {
    try {
      const res = await fetch(`${API_BASE}/api/v1/devices/${encodeURIComponent(id)}`);
      if (res.ok) {
        const data = await res.json();
        return data.device;
      }
    } catch {
      // Ignored
    }
    return null;
  }
}

/**
 * Captures live screenshot for device.
 */
export async function getScreenshot(id: string): Promise<ScreenshotData | null> {
  try {
    const raw = await callBinding<any>('main.MobRigService.GetScreenshot', id);
    if (raw) {
      return {
        width: raw.width,
        height: raw.height,
        size_bytes: raw.size_bytes,
        format: raw.format || 'png',
        data: raw.base64 || '',
      };
    }
  } catch {
    try {
      const res = await fetch(`${API_BASE}/api/v1/devices/${encodeURIComponent(id)}/screenshot`);
      if (res.ok) {
        return await res.json();
      }
    } catch {
      // Ignored
    }
  }
  return null;
}

/**
 * Executes a quick action (home, back, open_url) on the device.
 */
export async function sendQuickAction(id: string, action: string, param = ''): Promise<ActionResult> {
  try {
    return await callBinding<ActionResult>('main.MobRigService.QuickAction', id, action, param);
  } catch {
    try {
      let endpoint = `${API_BASE}/api/v1/devices/${encodeURIComponent(id)}/${action}`;
      let body: any = {};
      if (action === 'back') {
        endpoint = `${API_BASE}/api/v1/devices/${encodeURIComponent(id)}/key`;
        body = { key_code: 4 };
      } else if (action === 'open_url') {
        body = { url: param };
      }

      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      if (res.ok) {
        return await res.json();
      }
    } catch {
      // Ignored
    }
  }
  return { success: false, message: 'Action execution failed' };
}
