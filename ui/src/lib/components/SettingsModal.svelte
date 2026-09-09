<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Button from '$lib/components/ui/Button.svelte';
  import type { HealthInfo, ConfigInfo } from '$lib/wails';

  export let health: HealthInfo | null = null;
  export let config: ConfigInfo | null = null;

  const dispatch = createEventDispatcher<{ close: void }>();

  let copiedTab = '';

  function copyText(text: string, tab: string) {
    navigator.clipboard.writeText(text);
    copiedTab = tab;
    setTimeout(() => {
      copiedTab = '';
    }, 2000);
  }

  const cursorJson = JSON.stringify({
    mcpServers: {
      mobrig: {
        command: "mobrig",
        args: ["mcp"]
      }
    }
  }, null, 2);

  const claudeCliCommand = "claude mcp add mobrig mobrig mcp";
</script>

<!-- Backdrop (closes only when backdrop itself is clicked) -->
<div 
  class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm" 
  on:click|self={() => dispatch('close')} 
  on:keydown={(e) => e.key === 'Escape' && dispatch('close')}
  role="dialog" 
  aria-modal="true"
  tabindex="0"
>
  <!-- Modal Card -->
  <div 
    class="bg-[#101010] border-4 border-white p-8 max-w-2xl w-full flex flex-col gap-6 shadow-shadow relative" 
    role="document"
  >
    <!-- Header -->
    <div class="flex justify-between items-center border-b-4 border-white pb-4">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 bg-[#00F0FF] border-2 border-white flex items-center justify-center font-black text-black">
          ⚙
        </div>
        <h3 class="text-2xl font-black">MobRig Settings</h3>
      </div>
      <button 
        class="text-xl font-mono font-bold hover:text-[#00F0FF] transition-colors px-2 py-1 border-2 border-transparent hover:border-white"
        on:click={() => dispatch('close')}
        aria-label="Close Settings"
      >
        ✕
      </button>
    </div>

    <!-- Daemon Status & Config -->
    <div class="grid grid-cols-2 sm:grid-cols-3 gap-3 text-xs font-mono">
      <div class="bg-[#181818] p-3 border-2 border-white/20">
        <span class="opacity-50 block uppercase">Version</span>
        <span class="text-sm font-bold text-[#00F0FF]">v{health?.version || '0.1.0-dev'}</span>
      </div>
      <div class="bg-[#181818] p-3 border-2 border-white/20">
        <span class="opacity-50 block uppercase">Daemon Port</span>
        <span class="text-sm font-bold">:{health?.http_port || 8686}</span>
      </div>
      <div class="bg-[#181818] p-3 border-2 border-white/20">
        <span class="opacity-50 block uppercase">Uptime</span>
        <span class="text-sm font-bold">{health?.uptime || '0s'}</span>
      </div>
      <div class="col-span-2 sm:col-span-3 bg-[#181818] p-3 border-2 border-white/20">
        <span class="opacity-50 block uppercase">Data Directory</span>
        <span class="text-xs font-mono break-all">{config?.data_dir || '~/.mobrig'}</span>
      </div>
    </div>

    <!-- Agent MCP Integration Snippets -->
    <div class="flex flex-col gap-4 border-t-2 border-white/20 pt-4">
      <h4 class="text-sm font-black uppercase tracking-wider text-[#B829EA]">
        AI Agent Configuration
      </h4>

      <!-- Claude Code CLI -->
      <div class="flex flex-col gap-2">
        <div class="flex justify-between items-center text-xs font-mono">
          <span class="opacity-70">Claude Code Command:</span>
          <button 
            class="text-[#00F0FF] hover:underline font-bold"
            on:click={() => copyText(claudeCliCommand, 'claude')}
          >
            {copiedTab === 'claude' ? '✓ Copied!' : 'Copy Command'}
          </button>
        </div>
        <div class="bg-[#050505] p-3 border-2 border-white/40 font-mono text-xs text-[#00F0FF]">
          {claudeCliCommand}
        </div>
      </div>

      <!-- Cursor mcp.json -->
      <div class="flex flex-col gap-2">
        <div class="flex justify-between items-center text-xs font-mono">
          <span class="opacity-70">Cursor (.cursor/mcp.json):</span>
          <button 
            class="text-[#00F0FF] hover:underline font-bold"
            on:click={() => copyText(cursorJson, 'cursor')}
          >
            {copiedTab === 'cursor' ? '✓ Copied!' : 'Copy JSON'}
          </button>
        </div>
        <pre class="bg-[#050505] p-3 border-2 border-white/40 font-mono text-xs text-white/90 overflow-x-auto">{cursorJson}</pre>
      </div>
    </div>

    <div class="flex justify-end pt-2">
      <Button on:click={() => dispatch('close')}>
        Done
      </Button>
    </div>
  </div>
</div>
