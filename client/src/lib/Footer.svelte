<script lang="ts">
	import { onDestroy } from "svelte";
	import { getBrowser } from "../utilis/uaNames";

	const browser = getBrowser(navigator.userAgent);
	const fullCommitHash =
		import.meta.env.VITE_GIT_COMMIT || import.meta.env.PUBLIC_GIT_COMMIT || "local-dev";
	const shortCommitHash = fullCommitHash.slice(0, 7);

	let now = $state(new Date());

	const timer = setInterval(() => {
		now = new Date();
	}, 1000);

	onDestroy(() => {
		clearInterval(timer);
	});

	const timeText = $derived(
		now.toLocaleTimeString([], {
			hour: "2-digit",
			minute: "2-digit",
			second: "2-digit",
		})
	);
</script>

<footer class="w-full  bg-background/80 backdrop-blur-sm">
	<div class="mx-auto grid w-full max-w-screen-xl grid-cols-3 gap-4 px-4 py-3 text-sm">
		<div class="flex flex-col items-center gap-0.5">
			<span class="text-xs text-muted-foreground">Browser:</span>
			<span class="text-foreground">{browser}</span>
		</div>
		<div class="flex flex-col items-center gap-0.5">
			<span class="text-xs text-muted-foreground">Commit:</span>
			<span class="text-foreground">{shortCommitHash}</span>
		</div>
		<div class="flex flex-col items-center gap-0.5">
			<span class="text-xs text-muted-foreground">Time:</span>
			<span class="text-foreground">{timeText}</span>
		</div>
	</div>
</footer>
