<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Button from '$lib/components/ui/Button.svelte';

  export let name: string;
  export let os: string;
  export let status: 'connected' | 'offline' | 'busy' = 'offline';
  export let id: string;
  export let connectionType: string = 'usb';

  const dispatch = createEventDispatcher<{ select: { id: string } }>();

  function handleSelect() {
    dispatch('select', { id });
  }
</script>

<div class="relative group cursor-pointer" on:click={handleSelect} role="button" tabindex="0" on:keydown={(e) => e.key === 'Enter' && handleSelect()}>
  <!-- Solid Plate Background (Layer) -->
  <div class="absolute inset-0 bg-[#B829EA] border-4 border-white translate-x-[8px] translate-y-[8px] group-hover:translate-x-[4px] group-hover:translate-y-[4px] transition-all duration-300 ease-spring"></div>
  
  <!-- Main Card Surface -->
  <div class="relative bg-[#101010] border-4 border-white p-6 flex flex-col gap-6 shadow-none group-hover:-translate-x-1 group-hover:-translate-y-1 transition-all duration-300 ease-spring h-full">
    
    <!-- Badges Row -->
    <div class="flex items-center justify-between">
      <span class="text-xs font-mono font-bold uppercase tracking-wider px-2 py-0.5 bg-[#202020] border-2 border-white/60 text-[#00F0FF]">
        {connectionType}
      </span>
      <div class="flex items-center gap-2 bg-[#050505] border-2 border-white px-2 py-1">
        <div class="w-3 h-3 rounded-none
          {status === 'connected' ? 'bg-[#00F0FF] animate-pulse' : status === 'busy' ? 'bg-[#FF9900]' : 'bg-[#FF0055]'}"
        ></div>
        <span class="text-xs font-black uppercase tracking-wider">{status}</span>
      </div>
    </div>

    <div>
      <h3 class="text-2xl font-black mb-1 text-white group-hover:text-[#00F0FF] transition-colors">{name}</h3>
      <p class="text-base opacity-80 font-bold">{os}</p>
      <p class="text-xs opacity-50 font-mono mt-1 truncate">ID: {id}</p>
    </div>

    <div class="mt-auto pt-4 border-t-4 border-white border-dashed">
      <Button class="w-full" on:click={(e) => { e.stopPropagation(); handleSelect(); }}>
        Inspect Device
      </Button>
    </div>
  </div>
</div>
