<script lang="ts">
	import type { Pathname } from '$app/types';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { localizeHref } from '$lib/paraglide/runtime';
	import { m } from '$lib/paraglide/messages.js';
</script>

<svelte:head>
	<title>{m.error_title()} • Aerendil</title>
</svelte:head>

<div class="p-8">
	<div
		class="w-full max-w-2xl rounded-3xl border border-brand-100 bg-white p-8 text-center shadow-xl"
	>
		<span class="icon-[lucide--circle-alert] mx-auto mb-4 size-12 text-error-500" aria-hidden="true"
		></span>
		<p class="mb-1 text-xs font-bold tracking-widest text-brand-600 uppercase">
			{m.error_status_label({ status: page.status })}
		</p>
		<h1 class="text-2xl font-bold text-brand-900">{m.error_title()}</h1>
		<p class="mt-2 text-accent-900/70">{page.error?.message}</p>

		<a
			href={resolve(localizeHref('/dashboard') as Pathname)}
			class="mt-6 inline-block rounded-full bg-gradient-to-br from-brand-500 to-accent-500 px-6 py-3 font-bold text-white no-underline"
		>
			{m.error_back_dashboard()}
		</a>
	</div>
</div>
