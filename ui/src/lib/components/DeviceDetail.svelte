<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import { getDeviceInfo, getScreenshot, sendQuickAction, type DeviceInfo, type ScreenshotData } from '$lib/wails';

  export let deviceId: string;

  const dispatch = createEventDispatcher<{ back: void }>();

  let info: DeviceInfo | null = null;
  let screenshot: ScreenshotData | null = null;
  let loading = true;
  let refreshingScreen = false;
  let actionLoading = false;
  let actionMessage = '';
  let urlToOpen = '';

  async function loadDeviceData() {
    loading = true;
    try {
      const [devInfo, shot] = await Promise.all([
        getDeviceInfo(deviceId),
        getScreenshot(deviceId),
      ]);
      info = devInfo;
      screenshot = shot;
    } catch (err) {
      console.error('Failed to load device detail:', err);
    } finally {
      loading = false;
    }
  }

  async function refreshScreen() {
    refreshingScreen = true;
    try {
      const shot = await getScreenshot(deviceId);
      if (shot) {
        screenshot = shot;
      }
    } catch (err) {
      console.error('Failed to capture screenshot:', err);
    } finally {
      refreshingScreen = false;
    }
  }

  async function triggerAction(action: string, param = '') {
    actionLoading = true;
    actionMessage = '';
    try {
      const res = await sendQuickAction(deviceId, action, param);
      actionMessage = res.message || (res.success ? 'Success' : 'Failed');
      setTimeout(() => {
        actionMessage = '';
      }, 4000);
      // Auto-refresh screen after action
      setTimeout(refreshScreen, 500);
    } catch (err) {
      actionMessage = `Error: ${err}`;
    } finally {
      actionLoading = false;
    }
  }

  onMount(() => {
    loadDeviceData();
  });
</script>

<div class="flex flex-col gap-8">
  <!-- Top Navigation & Title -->
  <div class="flex flex-wrap items-center justify-between gap-4 border-b-4 border-white pb-6">
    <div class="flex items-center gap-4">
      <Button variant="secondary" on:click={() => dispatch('back')}>
        ← Back to Devices
      </Button>
      <div>
        <h2 class="text-3xl font-black">{info?.model || deviceId}</h2>
        <p class="text-sm font-mono opacity-60">ID: {deviceId}</p>
      </div>
    </div>

    <div class="flex items-center gap-3">
      {#if actionMessage}
        <span class="text-xs font-mono font-bold bg-[#00F0FF]/10 border-2 border-[#00F0FF] text-[#00F0FF] px-3 py-1 animate-pulse">
          {actionMessage}
        </span>
      {/if}
      <Button on:click={refreshScreen} disabled={refreshingScreen}>
        {refreshingScreen ? 'Capturing...' : '📸 Refresh Screen'}
      </Button>
    </div>
  </div>

  {#if loading}
    <div class="flex flex-col items-center justify-center py-24 gap-4">
      <div class="w-12 h-12 border-4 border-white border-t-[#00F0FF] animate-spin"></div>
      <p class="font-mono text-sm uppercase tracking-widest text-white/70">Inspecting device hardware...</p>
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
      <!-- Left Column: Screen Preview & Frame -->
      <div class="lg:col-span-5 xl:col-span-4 flex flex-col items-center gap-4">
        <div class="relative p-4 bg-[#101010] border-4 border-white shadow-shadow w-full max-w-[340px] flex flex-col items-center">
          <!-- Screen Frame Header -->
          <div class="w-full flex justify-between items-center pb-2 border-b-2 border-white/20 mb-3 text-xs font-mono opacity-70">
            <span>{screenshot?.width || info?.screen_width || 1080} x {screenshot?.height || info?.screen_height || 2400}</span>
            <span class="uppercase font-bold text-[#00F0FF]">Live Preview</span>
          </div>

          <!-- Phone Screen Container -->
          <div class="w-full aspect-[9/19] bg-[#050505] border-2 border-white/40 overflow-hidden flex items-center justify-center relative group">
            {#if screenshot?.data}
              <img 
                src="data:image/png;base64,{screenshot.data}" 
                alt="Device Screen"
                class="w-full h-full object-contain select-none"
              />
            {:else}
              <div class="text-center p-6 flex flex-col items-center gap-3">
                <span class="text-4xl">📱</span>
                <p class="text-xs opacity-60 font-mono">No screenshot captured yet</p>
                <Button size="sm" on:click={refreshScreen} disabled={refreshingScreen}>
                  Capture Now
                </Button>
              </div>
            {/if}
          </div>

          <!-- Device Hardware Buttons Bar -->
          <div class="w-full grid grid-cols-2 gap-2 mt-4">
            <Button 
              variant="secondary" 
              class="text-xs py-2"
              on:click={() => triggerAction('back')}
              disabled={actionLoading}
            >
              ◀ Back
            </Button>
            <Button 
              variant="secondary" 
              class="text-xs py-2"
              on:click={() => triggerAction('home')}
              disabled={actionLoading}
            >
              ● Home
            </Button>
          </div>
        </div>
      </div>

      <!-- Right Column: Specs & Control Center -->
      <div class="lg:col-span-7 xl:col-span-8 flex flex-col gap-6">
        <!-- Specs Panel -->
        <div class="bg-[#101010] border-4 border-white p-6 flex flex-col gap-4 shadow-shadow">
          <h3 class="text-xl font-black uppercase tracking-wider text-[#00F0FF] border-b-2 border-white/20 pb-2">
            Device Specifications
          </h3>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="bg-[#181818] p-3 border-2 border-white/20">
              <span class="text-xs font-mono uppercase opacity-50 block">Manufacturer</span>
              <span class="text-lg font-bold">{info?.manufacturer || 'Android'}</span>
            </div>

            <div class="bg-[#181818] p-3 border-2 border-white/20">
              <span class="text-xs font-mono uppercase opacity-50 block">Model</span>
              <span class="text-lg font-bold">{info?.model || deviceId}</span>
            </div>

            <div class="bg-[#181818] p-3 border-2 border-white/20">
              <span class="text-xs font-mono uppercase opacity-50 block">OS Version</span>
              <span class="text-lg font-bold">Android {info?.os_version || 'Unknown'}</span>
            </div>

            <div class="bg-[#181818] p-3 border-2 border-white/20">
              <span class="text-xs font-mono uppercase opacity-50 block">API Level</span>
              <span class="text-lg font-bold">API {info?.api_level || 'N/A'}</span>
            </div>

            <div class="bg-[#181818] p-3 border-2 border-white/20">
              <span class="text-xs font-mono uppercase opacity-50 block">Screen Resolution</span>
              <span class="text-lg font-bold">{info?.screen_width || 1080} x {info?.screen_height || 2400}</span>
            </div>

            <div class="bg-[#181818] p-3 border-2 border-white/20">
              <span class="text-xs font-mono uppercase opacity-50 block">Connection</span>
              <span class="text-lg font-bold uppercase text-[#B829EA]">{info?.connection_type || 'USB'}</span>
            </div>
          </div>
        </div>

        <!-- Quick Interaction Panel -->
        <div class="bg-[#101010] border-4 border-white p-6 flex flex-col gap-4 shadow-shadow">
          <h3 class="text-xl font-black uppercase tracking-wider text-[#B829EA] border-b-2 border-white/20 pb-2">
            Quick Actions
          </h3>

          <div class="flex flex-col gap-3">
            <label for="url-input" class="text-xs font-mono uppercase opacity-70">Launch URL / Deep Link</label>
            <div class="flex gap-2">
              <input
                id="url-input"
                type="text"
                bind:value={urlToOpen}
                placeholder="https://example.com or myapp://screen"
                class="flex-1 bg-[#181818] border-2 border-white px-3 py-2 text-sm font-mono text-white placeholder-white/40 focus:outline-none focus:border-[#00F0FF]"
              />
              <Button 
                on:click={() => triggerAction('open_url', urlToOpen)}
                disabled={actionLoading || !urlToOpen}
              >
                Open
              </Button>
            </div>
          </div>
        </div>

        <!-- API Reference Snippet -->
        <div class="bg-[#050505] border-2 border-white/40 p-4 font-mono text-xs flex flex-col gap-2">
          <span class="text-[#00F0FF] font-bold uppercase tracking-wider">REST API Endpoints for this device:</span>
          <div class="bg-[#101010] p-2 border border-white/20 text-white/80 space-y-1">
            <p>GET  http://127.0.0.1:8686/api/v1/devices/{deviceId}</p>
            <p>GET  http://127.0.0.1:8686/api/v1/devices/{deviceId}/screenshot</p>
            <p>GET  http://127.0.0.1:8686/api/v1/devices/{deviceId}/ui-tree</p>
            <p>POST http://127.0.0.1:8686/api/v1/devices/{deviceId}/tap</p>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>
