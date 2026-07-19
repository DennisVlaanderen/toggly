<script lang="ts">
	import { m } from '$lib/paraglide/messages.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<svelte:head>
	<title>{data.flag.key} • Toggly</title>
</svelte:head>

<div class="p-8">
	<div class="w-full max-w-2xl rounded-3xl border border-brand-100 bg-white p-8 shadow-xl">
		<p class="mb-1 text-xs font-bold tracking-widest text-brand-600 uppercase">{m.nav_flags()}</p>
		<h1 class="mb-5 flex items-center gap-3 text-3xl font-bold break-words text-brand-900">
			<span class="icon-[lucide--flag] size-7 shrink-0 text-brand-500" aria-hidden="true"></span>
			{data.flag.key}
		</h1>

		<div class="grid grid-cols-[repeat(auto-fit,minmax(12rem,1fr))] gap-4">
			<div class="grid gap-1.5 rounded-2xl border border-brand-100 bg-accent-50/40 p-4">
				<span class="text-xs font-bold tracking-wider text-accent-900/70 uppercase"
					>{m.flag_detail_status()}</span
				>
				<strong
					class="flex items-center gap-2 {data.flag.enabled
						? 'text-success-600'
						: 'text-brand-900'}"
				>
					{#if data.flag.enabled}
						<span class="icon-[lucide--toggle-right] size-4 shrink-0" aria-hidden="true"></span>
					{:else}
						<span class="icon-[lucide--toggle-left] size-4 shrink-0" aria-hidden="true"></span>
					{/if}
					{data.flag.enabled ? m.flag_enabled() : m.flag_disabled()}
				</strong>
			</div>
			<div class="grid gap-1.5 rounded-2xl border border-brand-100 bg-accent-50/40 p-4">
				<span class="text-xs font-bold tracking-wider text-accent-900/70 uppercase"
					>{m.flag_detail_version()}</span
				>
				<strong class="text-brand-900">{data.flag.version}</strong>
			</div>
			<div
				class="col-span-full grid gap-1.5 rounded-2xl border border-brand-100 bg-accent-50/40 p-4"
			>
				<span class="text-xs font-bold tracking-wider text-accent-900/70 uppercase"
					>{m.flag_detail_value()}</span
				>
				<strong class="text-brand-900">{data.flag.value || '—'}</strong>
			</div>
		</div>
	</div>
</div>
