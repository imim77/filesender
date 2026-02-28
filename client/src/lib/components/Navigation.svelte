<script lang="ts">
	import { cn } from "$lib/utils.js";

	export let logoSrc = "";
	export let logoAlt = "Ollama";
	export let aboutHref = "#";
	export let className = "";
</script>

<header
	data-slot="navigation"
	class={cn("sticky top-0 z-50 w-full  bg-background", className)}
>
	<nav class="flex w-full items-center justify-start px-6 py-4">
		<div class="flex items-center gap-6">
			{#if logoSrc}
				<img src={logoSrc} alt={logoAlt} class="h-10" />
			{/if}
			<a
				href={aboutHref}
				class="focus-visible:border-ring focus-visible:ring-ring/50 rounded-sm text-base font-medium text-foreground transition-colors outline-none hover:text-muted-foreground focus-visible:ring-[3px]"
			>
				About project
			</a>
		</div>
	</nav>
</header>
