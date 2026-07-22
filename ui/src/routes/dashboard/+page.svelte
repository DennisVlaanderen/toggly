<script lang="ts">
	import { m } from '$lib/paraglide/messages.js';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let isAdmin = $derived(data.isAdmin);
</script>

<svelte:head>
	<title>Dashboard • Toggly</title>
</svelte:head>

<div class="p-8">
	<div class="w-full max-w-2xl rounded-3xl border border-brand-100 bg-white p-8 shadow-xl">
		<p class="mb-1 text-xs font-bold tracking-widest text-brand-600 uppercase">
			{isAdmin ? m.dashboard_admin_view() : m.dashboard_member_view()}
		</p>
		<h1 class="text-3xl font-bold text-brand-900">
			{m.dashboard_welcome({ username: data.username })}
		</h1>
		<p class="mt-2 mb-5 text-accent-900/70">{m.dashboard_subtext()}</p>

		<div class="grid grid-cols-[repeat(auto-fit,minmax(12rem,1fr))] gap-4">
			<div class="grid gap-1.5 rounded-2xl border border-brand-100 bg-accent-50/40 p-4">
				<span class="text-xs font-bold tracking-wider text-accent-900/70 uppercase"
					>{m.dashboard_access_level()}</span
				>
				<strong class="text-brand-900"
					>{isAdmin ? m.dashboard_full_admin() : m.dashboard_standard_access()}</strong
				>
			</div>
			<div class="grid gap-1.5 rounded-2xl border border-brand-100 bg-accent-50/40 p-4">
				<span class="text-xs font-bold tracking-wider text-accent-900/70 uppercase"
					>{m.dashboard_status()}</span
				>
				<strong class="text-brand-900"
					>{isAdmin ? m.dashboard_ready_manage() : m.dashboard_ready_work()}</strong
				>
			</div>
		</div>
	</div>
</div>
