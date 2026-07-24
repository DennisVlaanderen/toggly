<script lang="ts">
	import type { Pathname } from '$app/types';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { localizeHref } from '$lib/paraglide/runtime';
	import { m } from '$lib/paraglide/messages.js';

	const REDIRECT_SECONDS = 20;
	let secondsLeft = $state(REDIRECT_SECONDS);

	function goToLogin() {
		goto(resolve(localizeHref('/login') as Pathname));
	}

	$effect(() => {
		const interval = setInterval(() => {
			secondsLeft -= 1;
			if (secondsLeft <= 0) {
				goToLogin();
			}
		}, 1000);

		return () => clearInterval(interval);
	});
</script>

<svelte:head>
	<title>{m.error_title()} • Aerendil</title>
</svelte:head>

<div
	class="grid min-h-screen place-items-center bg-linear-to-br from-brand-50 to-accent-50 p-8 font-sans"
>
	<div
		class="w-full max-w-md rounded-3xl border border-brand-100 bg-white p-8 text-center shadow-xl"
	>
		<span class="icon-[lucide--circle-alert] mx-auto mb-4 size-12 text-error-500" aria-hidden="true"
		></span>
		<p class="mb-1 text-xs font-bold tracking-widest text-brand-600 uppercase">
			{m.error_status_label({ status: page.status })}
		</p>
		<h1 class="text-2xl font-bold text-brand-900">{m.error_title()}</h1>
		<p class="mt-2 text-accent-900/70">{page.error?.message}</p>

		<button
			type="button"
			onclick={goToLogin}
			class="mt-6 w-full cursor-pointer rounded-full bg-linear-to-br from-brand-500 to-accent-500 px-4 py-3 font-bold text-white"
		>
			{m.error_go_login()}
		</button>

		<p class="mt-4 text-sm text-accent-900/60" aria-live="polite">
			{m.error_redirect_countdown({ seconds: secondsLeft })}
		</p>
	</div>
</div>
