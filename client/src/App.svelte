<script lang="ts">
	import './app.css';
	import { onDestroy } from 'svelte';
	import { WebRTCController } from './services/webrtccontroller.svelte';
	import { generateName, getAgentInfo } from './utilis/uaNames';
	import Footer from './lib/Footer.svelte';
	import Hero from '$lib/components/Hero.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import slika from './assets/svelte.svg';
	import imageHello from './assets/hellodog.png'
    import Something from '$lib/components/something.svelte';

	const localAlias = generateName();
	const localDevice = getAgentInfo(navigator.userAgent);
	const controller = new WebRTCController(localAlias, localDevice);

	function sendFiles(peerId: string, files: FileList): void {
		controller.sendFiles(peerId, files);
	}

	onDestroy(() => {
		controller.destroy();
	});
</script>

<Navigation logoSrc={slika} />

<Hero
	alias={controller.myName || localAlias}
	animationSrc={imageHello}
	peers={controller.peers.filter((peer) => controller.isPeerConnected(peer.id))}
	connectionStatus={controller.connectionStatus}
	connectionLabel={(peerId) => controller.connectionLabel(peerId)}
	onSendFiles={sendFiles}
/>



