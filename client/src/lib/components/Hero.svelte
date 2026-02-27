<script lang="ts">
	import { cn } from "$lib/utils.js";
	import { Button } from "$lib/components/ui/button/index.js";
	import * as Item from "$lib/components/ui/item/index.js";
  	import * as Avatar from "$lib/components/ui/avatar/index.js";
  	import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";	
  	import ChevronDown from "@lucide/svelte/icons/chevron-down";
 
  const people = [
    {
      username: "shadcn",
      avatar: "https://github.com/shadcn.png",
      email: "shadcn@vercel.com"
    },
    {
      username: "maxleiter",
      avatar: "https://github.com/maxleiter.png",
      email: "maxleiter@vercel.com"
    },
    {
      username: "evilrabbit",
      avatar: "https://github.com/evilrabbit.png",
      email: "evilrabbit@vercel.com"
    }
  ];

	export let alias = "Darke Some";
	export let animationSrc = "";
	export let animationAlt = "File sender mascot waving";
	export let className = "";
</script>

<main data-slot="hero" class={cn("flex flex-1 flex-col bg-background", className)}>
	<div class="w-full px-6 pt-8">
		<div class="flex items-center gap-2">
			<div class="h-2 w-2 rounded-full bg-[#ff5a4d]"></div>
			<span class="text-lg font-medium text-foreground md:text-2xl">Send now</span>
		</div>
	</div>

	<div class="w-full px-6 py-10 md:py-14">
		<div class="flex flex-wrap items-center justify-center gap-3 text-center text-3xl font-medium tracking-tight text-foreground md:gap-4 md:text-5xl lg:text-6xl">
			<span>Hi</span>
			<span>there</span>
			<div class="relative flex h-14 w-14 items-center justify-center md:h-20 md:w-20">
				{#if animationSrc}
					<img src={animationSrc} alt={animationAlt} class="h-full w-full object-contain" />
				{/if}
			</div>
			<span>!</span>
			<span>You are</span>
			
		</div>
		<div class="mt-2 flex flex-wrap items-center justify-center gap-3 text-center text-3xl font-medium tracking-tight text-foreground md:gap-4 md:text-5xl lg:text-6xl">
			<span>known</span>
			<span>as</span>
			<span>
				<code class="bg-muted relative rounded px-[0.9rem] py-[0.9rem] font-medium text-sm font-semibold">
					@{alias}
				</code>
			</span>
		</div>
	</div>

	<div class="mx-auto w-full max-w-7xl px-6 pb-12 md:pb-20">
		<div class="relative overflow-hidden rounded-3xl border border-border bg-card p-8 text-card-foreground shadow-sm md:p-12">
			<div
				aria-hidden="true"
				class="pointer-events-none absolute inset-0 bg-[linear-gradient(140deg,rgba(253,248,250,0.92)_0%,rgba(250,247,241,0.9)_48%,rgba(245,249,253,0.92)_100%),radial-gradient(120%_90%_at_0%_0%,rgba(255,255,255,0.55)_0%,transparent_60%),radial-gradient(120%_90%_at_100%_100%,rgba(255,90,77,0.08)_0%,transparent_62%)]"
			></div>
			<div class="relative z-10 flex flex-col items-start justify-between gap-8 md:flex-row md:items-center">
				<div class="flex-1">
					<h2 class="text-2xl font-semibold tracking-tight md:text-3xl">Available clients</h2>
				</div>
				<div class="flex min-h-64 w-full max-w-md flex-col items-center gap-6">
  				<DropdownMenu.Root>
    				<DropdownMenu.Trigger>
      				{#snippet child({ props })}
        				<Button {...props} variant="outline" size="sm" class="w-fit">
          				Select <ChevronDown />
        				</Button>
      				{/snippet}
    				</DropdownMenu.Trigger>
    				<DropdownMenu.Content class="w-72 [--radius:0.65rem]" align="end">
      				{#each people as person (person.username)}
        				<DropdownMenu.Item class="p-0">
          				<Item.Root size="sm" class="w-full p-2">
            				<Item.Media>
              				<Avatar.Root class="size-8">
                				<Avatar.Image src={person.avatar} class="grayscale" />
                				<Avatar.Fallback>{person.username.charAt(0)}</Avatar.Fallback>
              				</Avatar.Root>
            				</Item.Media>
            				<Item.Content class="gap-0.5">
              				<Item.Title>{person.username}</Item.Title>
              				<Item.Description>{person.email}</Item.Description>
            				</Item.Content>
          				</Item.Root>
        				</DropdownMenu.Item>
      				{/each}
    				</DropdownMenu.Content>
  				</DropdownMenu.Root>
				</div>
				<Button class="rounded-full px-8 py-6" type="button">

					<span class="h-2 w-2 rounded-full bg-[#ff5a4d]"></span>
					<span>Select client</span>
				</Button>
			</div>
		</div>
	</div>
</main>
