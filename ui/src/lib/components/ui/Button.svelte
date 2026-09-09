<script lang="ts">
  import { cn } from '$lib/utils';
  
  export let className: string = '';
  export let type: 'button' | 'submit' | 'reset' = 'button';
  export let variant: 'default' | 'secondary' | 'ghost' | 'destructive' = 'default';
  
  const baseClasses = 'inline-flex items-center justify-center font-bold transition-all duration-300 ease-spring gap-2 px-6 py-3 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-main focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-50 border-4 border-white shadow-shadow hover:translate-x-boxShadowX hover:translate-y-boxShadowY hover:shadow-none active:scale-95 text-lg';
  
  const variantClasses = {
    default: 'bg-main text-main-foreground',
    secondary: 'bg-secondary-background text-foreground',
    ghost: 'border-transparent shadow-none hover:bg-main/20 hover:text-main-foreground hover:border-border hover:shadow-shadow hover:translate-x-0 hover:translate-y-0',
    destructive: 'bg-destructive text-destructive-foreground',
  };
</script>

<button
  {type}
  class={cn(baseClasses, variantClasses[variant], className)}
  on:click
  {...$$restProps}
>
  <slot />
</button>
