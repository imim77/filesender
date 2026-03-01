<script lang="ts">
	import { cn } from "$lib/utils.js";
	import * as Item from "$lib/components/ui/item/index.js";
	import * as Avatar from "$lib/components/ui/avatar/index.js";
	import * as Empty from "$lib/components/ui/empty/index.js";	
 	import { Spinner } from "$lib/components/ui/spinner/index.js";
	import { buttonVariants } from "$lib/components/ui/button/index.js";
	import laptopDog from "../../assets/laptopdog.png";
	import mobileDog from "../../assets/mobiledog.png";
	import { Badge } from "$lib/components/ui/badge/index.js";
	import * as Resizable from "$lib/components/ui/resizable/index.js";
	import { Separator } from "$lib/components/ui/separator/index.js";

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

	const fileInputId = (peerId: string): string =>
		`peer-file-${peerId.replace(/[^a-zA-Z0-9_-]/g, "")}`;
</script>

<main data-slot="hero" class={cn("flex flex-1 flex-col items-center bg-background", className)}>
	
	<div class="w-full px-4 pt-[150px] pb-10 sm:px-6 sm:pt-[170px] md:px-8 md:pb-14">
		<h1 class="mx-auto max-w-5xl text-center font-semibold tracking-tight text-foreground">
			<span class="block text-3xl leading-[1.1] sm:text-4xl md:text-5xl lg:text-6xl">You are known as</span>
			<span class="mt-2 block">
				<code
					class="inline-flex max-w-full items-center rounded-md border border-border/60 bg-muted px-3 py-2 text-xl leading-tight font-semibold tracking-normal sm:px-4 sm:py-2.5 sm:text-2xl md:text-3xl lg:text-4xl"
				>
					<span class="max-w-full truncate">@{alias}</span>
				</code>
			</span>
		</h1>
		
		<small class="mt-4 block w-full px-4 py-2 text-center text-sm leading-none font-medium text-muted-foreground">
			open this site in <span class="underline underline-offset-4">another browser</span>
		</small>
	</div>
	
	<div class="flex w-full flex-1 items-center justify-center px-4 pb-10 sm:px-6 md:px-8">
	<Resizable.PaneGroup
		direction="vertical"
		class="!h-[460px] w-full max-w-screen-lg rounded-md border bg-card"
	>
		<Resizable.Pane defaultSize={24} minSize={18}>
			<div class="flex h-full flex-col">
				<Item.Header class="px-6 py-4">
					<h2 class="text-base font-semibold tracking-tight">Available clients</h2>
					<Badge
						class="h-8 min-w-8 rounded-md bg-[oklch(0.70_0.19_48)] px-3 text-sm font-semibold tabular-nums text-black shadow-xs"
					>
						{peers.length}
					</Badge>
				</Item.Header>
				<Separator />
			</div>
		</Resizable.Pane>
		<Resizable.Handle />
		<Resizable.Pane defaultSize={76} minSize={30}>
			<div class="h-full overflow-y-auto p-4 ">
				{#if peers.length === 0}
					<div class="flex min-h-full items-center justify-center">
						<Empty.Root class="mx-auto w-full max-w-md md:p-6">
							<Empty.Header>
								<Empty.Media variant="icon">
									<Spinner />
								</Empty.Media>
								<Empty.Title>Waiting for other clients to join</Empty.Title>
							</Empty.Header>
						</Empty.Root>
					</div>
				{:else}
					<Item.Group class="w-full">
						{#each peers as peer, index (peer.id)}
							<Item.Root variant="outline" class="w-full">
								<Item.Media>
									<Avatar.Root class="size-12 sm:size-14 md:size-16">
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
										id={fileInputId(peer.id)}
										type="file"
										multiple
										class="sr-only"
										onchange={(event) => handleFileChange(peer.id, event)}
									/>
									<label
										for={fileInputId(peer.id)}
										class={cn(buttonVariants({ variant: "outline", size: "sm" }), "cursor-pointer")}
									>
										Browse files
									</label>
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
	</div>

	
</main>
