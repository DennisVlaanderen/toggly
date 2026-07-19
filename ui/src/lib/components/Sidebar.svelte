<script lang="ts">
	import type { Pathname } from '$app/types';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { localizeHref } from '$lib/paraglide/runtime';
	import { m } from '$lib/paraglide/messages.js';
	import LocaleSwitcher from './LocaleSwitcher.svelte';
	import type { FlagSummary } from '$lib/server/flags';

	let { flags, username }: { flags: FlagSummary[]; username: string } = $props();

	let collapsed = $state(false);

	function isActive(pathname: string) {
		return page.url.pathname === pathname || page.url.pathname.startsWith(`${pathname}/`);
	}
</script>

<aside
	class="flex h-full flex-shrink-0 flex-col gap-5 border-r border-brand-100 bg-white p-5 transition-[width] duration-200 {collapsed
		? 'w-[4.5rem]'
		: 'w-64'}"
>
	<div class="flex items-center gap-2.5">
		<button
			type="button"
			class="flex h-9 w-9 flex-shrink-0 cursor-pointer items-center justify-center rounded-xl border border-brand-200 bg-accent-50 text-brand-800"
			onclick={() => (collapsed = !collapsed)}
			aria-label={collapsed ? m.sidebar_expand() : m.sidebar_collapse()}
		>
			{#if collapsed}
				<span class="icon-[lucide--panel-left-open] size-4.5" aria-hidden="true"></span>
			{:else}
				<span class="icon-[lucide--panel-left-close] size-4.5" aria-hidden="true"></span>
			{/if}
		</button>
		{#if !collapsed}
			<span class="truncate font-bold text-brand-900">{username}</span>
		{/if}
	</div>

	<nav class="flex flex-1 flex-col gap-1 overflow-y-auto">
		<a
			class="flex items-center gap-2.5 truncate rounded-2xl px-2.5 py-2.5 font-semibold text-brand-800 no-underline hover:bg-accent-50 {isActive(
				'/dashboard'
			)
				? 'bg-accent-100 text-brand-700'
				: ''}"
			href={resolve(localizeHref('/dashboard') as Pathname)}
		>
			<span class="flex w-6 flex-shrink-0 justify-center" aria-hidden="true">
				<span class="icon-[lucide--layout-dashboard] size-4"></span>
			</span>
			{#if !collapsed}<span>{m.nav_dashboard()}</span>{/if}
		</a>

		{#if !collapsed}
			<p
				class="mt-3 mb-0.5 flex items-center gap-1.5 text-xs font-bold tracking-wider text-accent-900/70 uppercase"
			>
				<span class="icon-[lucide--flag] size-3.5" aria-hidden="true"></span>
				{m.nav_flags()}
			</p>
		{/if}

		{#if flags.length === 0 && !collapsed}
			<p class="text-sm text-accent-900/60">{m.nav_no_flags()}</p>
		{:else}
			{#each flags as flag (flag.key)}
				<a
					class="flex items-center gap-2.5 truncate rounded-2xl px-2.5 py-2.5 font-semibold text-brand-800 no-underline hover:bg-accent-50 {isActive(
						`/dashboard/flags/${flag.key}`
					)
						? 'bg-accent-100 text-brand-700'
						: ''}"
					href={resolve(localizeHref(`/dashboard/flags/${flag.key}`) as Pathname)}
				>
					<span class="flex w-6 flex-shrink-0 justify-center" aria-hidden="true">
						{#if flag.enabled}
							<span class="icon-[lucide--toggle-right] size-4 text-success-500"></span>
						{:else}
							<span class="icon-[lucide--toggle-left] size-4 text-accent-900/40"></span>
						{/if}
					</span>
					{#if !collapsed}<span>{flag.key}</span>{/if}
				</a>
			{/each}
		{/if}
	</nav>

	<div class="flex flex-col gap-3 border-t border-brand-100 pt-4">
		<LocaleSwitcher compact={collapsed} />
		<form method="POST" action="/logout">
			<button
				type="submit"
				class="flex w-full cursor-pointer items-center gap-2.5 truncate rounded-2xl bg-error-50 px-2.5 py-2.5 font-bold text-error-700"
			>
				<span class="flex w-6 flex-shrink-0 justify-center" aria-hidden="true">
					<span class="icon-[lucide--log-out] size-4"></span>
				</span>
				{#if !collapsed}<span>{m.logout_button()}</span>{/if}
			</button>
		</form>
	</div>
</aside>
