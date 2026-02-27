<script lang="ts">
	import './app.css';
	import { onDestroy, onMount } from 'svelte';
	import { WebRTCController } from './services/webrtccontroller.svelte';
	import { generateName, getAgentInfo } from './utilis/uaNames';
	import Footer from './lib/Footer.svelte';
	import { Button } from "$lib/components/ui/button/index.js";
	import Navigation from '$lib/components/Navigation.svelte';
	import slika from './assets/svelte.svg'

	const localAlias = generateName();
	const localDevice = getAgentInfo(navigator.userAgent);
	const controller = new WebRTCController(localAlias, localDevice);

	function sendFiles(peerId: string, event: Event): void {
		const input = event.currentTarget as HTMLInputElement;

		if (!input.files || input.files.length === 0) return;

		controller.sendFiles(peerId, input.files);
		input.value = '';
	}

	onDestroy(() => {
		controller.destroy();
	});
</script>

<main>
	<Navigation logoSrc={slika} />
	<Button>Click me</Button>
	<h1>FileSender</h1>
	<p>Status: {controller.connectionStatus}</p>
	<h2>I am known as {controller.myName || localAlias}</h2>
	<h2>Peers ({controller.peers.length})</h2>

	{#if controller.peers.length === 0}
		<p>Waiting for peers to join...</p>
	{:else}
		<ul>
			{#each controller.peers as peer}
				<li>
					<div>
						<strong>{peer.alias || 'Unnamed device'}</strong>
						<span>{peer.deviceModel || peer.deviceType || 'Unknown device'}</span>
						<span>{controller.connectionLabel(peer.id)}</span>
					</div>

					<div>
						<input
							type="file"
							multiple
							disabled={!controller.isPeerConnected(peer.id)}
							on:change={(event) => sendFiles(peer.id, event)}
						/>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</main>

<Footer />


