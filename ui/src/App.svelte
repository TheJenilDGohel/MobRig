<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import DeviceCard from '$lib/components/DeviceCard.svelte';
  import DeviceDetail from '$lib/components/DeviceDetail.svelte';
  import SettingsModal from '$lib/components/SettingsModal.svelte';
  import { getDevices, getHealth, getConfig, type DeviceListItem, type HealthInfo, type ConfigInfo } from '$lib/wails';

  let devices: DeviceListItem[] = [];
  let health: HealthInfo | null = null;
  let config: ConfigInfo | null = null;
  let loading = true;
  let refreshing = false;

  // View routing state
  let currentView: 'list' | 'detail' = 'list';
  let selectedDeviceId: string | null = null;
  let showSettings = false;

  async function loadData() {
    refreshing = true;
    try {
      const [devList, healthInfo, cfgInfo] = await Promise.all([
        getDevices(),
        getHealth(),
        getConfig(),
      ]);
      devices = devList;
      health = healthInfo;
      config = cfgInfo;
    } catch (err) {
      console.error('Failed to load MobRig data:', err);
    } finally {
      loading = false;
      refreshing = false;
    }
  }

  function handleSelectDevice(id: string) {
    selectedDeviceId = id;
    currentView = 'detail';
  }

  function handleBackToList() {
    selectedDeviceId = null;
    currentView = 'list';
    loadData();
  }

  onMount(() => {
    loadData();
  });

  $: activeCount = devices.filter((d) => d.status === 'connected').length;
</script>

<main class="min-h-screen bg-transparent text-white flex flex-col font-sans relative">
  <!-- Header -->
  <header class="sticky top-0 z-40 border-b-4 border-white bg-[#050505]/90 backdrop-blur-md p-6 flex justify-between items-center">
    <div class="flex items-center gap-4 cursor-pointer" on:click={handleBackToList} role="button" tabindex="0" on:keydown={(e) => e.key === 'Enter' && handleBackToList()}>
      <div class="w-12 h-12 bg-[#B829EA] border-4 border-white flex items-center justify-center font-black text-2xl shadow-shadow">
        M
      </div>
      <div>
        <div class="flex items-center gap-3">
          <h1 class="text-3xl font-black tracking-tight">MobRig</h1>
          {#if health}
            <span class="text-xs font-mono font-bold bg-[#141414] border-2 border-white px-2 py-0.5 text-[#00F0FF]">
              v{health.version}
            </span>
          {/if}
        </div>
        <p class="text-sm font-bold opacity-80">
          Phantom Tester Hub
          {#if health}
            <span class="opacity-50">• Port :{health.http_port} • Up {health.uptime}</span>
          {/if}
        </p>
      </div>
    </div>
    
    <div class="flex gap-4 items-center">
      {#if health?.status === 'ok'}
        <div class="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-[#00F0FF]/10 border-2 border-[#00F0FF] text-[#00F0FF] text-xs font-black uppercase">
          <div class="w-2 h-2 bg-[#00F0FF] animate-pulse"></div>
          Daemon Active
        </div>
      {/if}
      <Button variant="secondary" on:click={() => showSettings = true}>
        ⚙ Settings
      </Button>
      {#if currentView === 'list'}
        <Button on:click={loadData} disabled={refreshing}>
          {refreshing ? 'Scanning...' : 'Refresh Devices'}
        </Button>
      {/if}
    </div>
  </header>

  <!-- Content -->
  <div class="flex-1 p-8 flex flex-col gap-8 max-w-7xl w-full mx-auto">
    {#if currentView === 'detail' && selectedDeviceId}
      <DeviceDetail 
        deviceId={selectedDeviceId} 
        on:back={handleBackToList}
      />
    {:else}
      <!-- Device List View -->
      <div class="flex justify-between items-end border-b-4 border-white pb-4">
        <div>
          <h2 class="text-4xl font-black">Connected Devices</h2>
          <p class="text-sm opacity-60 font-medium mt-1">Real-time mobile test devices managed by MobRig</p>
        </div>
        <span class="bg-white text-black font-black px-3 py-1 border-4 border-black">
          {activeCount} Active
        </span>
      </div>

      {#if loading}
        <div class="flex flex-col items-center justify-center py-24 gap-4">
          <div class="w-12 h-12 border-4 border-white border-t-[#00F0FF] animate-spin"></div>
          <p class="font-mono text-sm uppercase tracking-widest text-white/70">Connecting to MobRig daemon...</p>
        </div>
      {:else if devices.length === 0}
        <!-- Empty State -->
        <div class="border-4 border-white border-dashed bg-[#101010]/60 p-12 flex flex-col items-center justify-center text-center gap-6 max-w-2xl mx-auto my-12 shadow-shadow">
          <div class="w-16 h-16 bg-[#202020] border-4 border-white flex items-center justify-center text-3xl font-black">
            📱
          </div>
          <div class="flex flex-col gap-2">
            <h3 class="text-2xl font-black uppercase tracking-wider">No Devices Connected</h3>
            <p class="text-sm opacity-70 max-w-md leading-relaxed">
              Connect an Android device with USB debugging enabled, or start an Android Virtual Device (AVD). MobRig will automatically discover it.
            </p>
          </div>
          <div class="flex gap-4">
            <Button on:click={loadData} disabled={refreshing}>
              {refreshing ? 'Scanning...' : 'Scan for Devices'}
            </Button>
          </div>
        </div>
      {:else}
        <!-- Device Grid -->
        <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-8">
          {#each devices as device (device.id)}
            <DeviceCard 
              name={device.name} 
              os={device.os} 
              status={device.status} 
              id={device.id} 
              connectionType={device.connection_type}
              on:select={(e) => handleSelectDevice(e.detail.id)}
            />
          {/each}
        </div>
      {/if}
    {/if}
  </div>

  <!-- Settings Modal -->
  {#if showSettings}
    <SettingsModal 
      {health} 
      {config} 
      on:close={() => showSettings = false} 
    />
  {/if}
</main>
