<script lang="ts">
	import { cn } from "$lib/utils.js";
	import * as Item from "$lib/components/ui/item/index.js";
	import * as Avatar from "$lib/components/ui/avatar/index.js";
	import * as Empty from "$lib/components/ui/empty/index.js";	
 	import { Spinner } from "$lib/components/ui/spinner/index.js";
 	import { Button } from "$lib/components/ui/button/index.js";
	import laptopDog from "../../assets/laptopdog.png";
	import mobileDog from "../../assets/mobiledog.png";
	 import * as Resizable from "$lib/components/ui/resizable/index.js";
	      import SearchIcon from "@lucide/svelte/icons/search";
	import * as InputGroup from "$lib/components/ui/input-group/index.js";

	type PeerClient = {
		id: string;
		alias?: string;
		deviceModel?: string;
		deviceType?: string;
	};

	let {
		alias = "Darke Some",
		animationSrc = "",
		animationAlt = "File sender mascot waving",
		className = "",
		peers = [],
		connectionStatus = "connecting",
		connectionLabel = () => "not connected",
		onSendFiles = () => {},
	}: {
		alias?: string;
		animationSrc?: string;
		animationAlt?: string;
		className?: string;
		peers?: PeerClient[];
		connectionStatus?: "disconnected" | "connecting" | "connected";
		connectionLabel?: (peerId: string) => string;
		onSendFiles?: (peerId: string, files: FileList) => void;
	} = $props();

	const peerAvatar = (deviceType?: string): string =>
		deviceType === "Mobile" ? mobileDog : laptopDog;

	const peerName = (peer: PeerClient): string =>
		peer.alias || peer.deviceModel || peer.deviceType || "Unnamed device";

	const handleFileChange = (peerId: string, event: Event): void => {
		const input = event.currentTarget as HTMLInputElement;
		if (!input.files || input.files.length === 0) return;
		onSendFiles(peerId, input.files);
		input.value = "";
	};
</script>

<main data-slot="hero" class={cn("flex flex-1 flex-col items-center bg-background", className)}>
	
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
		
		<small class="mt-4 block w-full px-4 py-2 text-center text-sm leading-none font-medium text-muted-foreground">
			open this site in <span class="underline underline-offset-4">another browser</span>
		</small>
	</div>
	
	<Resizable.PaneGroup
		direction="vertical"
		class="!h-[460px] w-full max-w-screen-md rounded-md border bg-card"
	>
		<Resizable.Pane defaultSize={24} minSize={18}>
			<div class="flex h-full items-center justify-start px-6 py-4">
				<span class="font-semibold">Available clients</span>
			</div>
		</Resizable.Pane>
		<Resizable.Handle />
		<Resizable.Pane defaultSize={76} minSize={30}>
			<div class="h-full overflow-y-auto p-4">
				{#if peers.length === 0}
					<Empty.Root class="w-full max-w-md border md:p-6">
			<Empty.Header>
				<Empty.Media variant="icon">
					<Spinner />
				</Empty.Media>
				<Empty.Title>Waiting for other clients to join</Empty.Title>
				<Empty.Description>
					Please wait while we process your request. Do not refresh the page.
				</Empty.Description>
			</Empty.Header>	
		</Empty.Root>	
				{:else}
					<Item.Group class="w-full">
						{#each peers as peer, index (peer.id)}
							<Item.Root variant="outline" class="w-full">
								<Item.Media>
									<Avatar.Root class="size-10">
										<Avatar.Image src={peerAvatar(peer.deviceType)} alt={peerName(peer)} />
										<Avatar.Fallback>{peerName(peer).charAt(0)}</Avatar.Fallback>
									</Avatar.Root>
								</Item.Media>
								<Item.Content>
									<Item.Title>{peerName(peer)}</Item.Title>
									<Item.Description>{peer.deviceModel || peer.deviceType || "Unknown device"}</Item.Description>
								</Item.Content>
								<Item.Actions>
									<input
										type="file"
										multiple
										onchange={(event) => handleFileChange(peer.id, event)}
									/>
								</Item.Actions>
							</Item.Root>
							{#if index !== peers.length - 1}
								<Item.Separator />
							{/if}
						{/each}
					</Item.Group>
				{/if}
			</div>
		</Resizable.Pane>
	</Resizable.PaneGroup>

	
</main>
