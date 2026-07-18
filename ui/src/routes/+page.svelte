<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		void goto('/login');
	});
</script>

<div class="redirecting">Redirecting to the sign-in page…</div>

<style>
	.redirecting {
		display: grid;
		place-items: center;
		min-height: 100vh;
		font-size: 1.1rem;
		font-weight: 600;
		color: oklch(0.35 0.06 300);
	}
</style>
