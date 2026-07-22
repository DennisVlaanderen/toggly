<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { apiRequest } from '$lib/client/api';
	import LocaleSwitcher from '$lib/components/LocaleSwitcher.svelte';
	import { localizeHref } from '$lib/paraglide/runtime';
	import { m } from '$lib/paraglide/messages.js';
	import type { Pathname } from '$app/types';

	let isSubmitting = $state(false);
	let errorMessage = $state('');

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		const formEl = event.currentTarget as HTMLFormElement;
		const data = new FormData(formEl);

		isSubmitting = true;
		errorMessage = '';

		const result = await apiRequest('/bff/login', {
			method: 'POST',
			body: JSON.stringify({
				username: (data.get('username') ?? '').toString(),
				password: (data.get('password') ?? '').toString()
			})
		});

		if (result.error) {
			errorMessage = result.error;
			isSubmitting = false;
			return;
		}

		await goto(resolve(localizeHref('/dashboard') as Pathname));
	}
</script>

<svelte:head>
	<title>Login • Toggly</title>
</svelte:head>

<div class="fixed top-6 right-6 z-20">
	<LocaleSwitcher />
</div>

<div
	class="grid min-h-screen place-items-center bg-gradient-to-br from-brand-50 to-accent-50 p-8 font-sans"
>
	<div class="w-full max-w-md rounded-3xl border border-brand-100 bg-white p-8 shadow-xl">
		<div class="mb-6">
			<p class="mb-1 text-xs font-bold tracking-widest text-brand-600 uppercase">
				{m.login_eyebrow()}
			</p>
			<h1 class="text-3xl font-bold text-brand-900">{m.login_title()}</h1>
			<p class="mt-1 text-accent-900/70">{m.login_subtitle()}</p>
		</div>

		<form method="POST" class="grid gap-4" onsubmit={handleSubmit}>
			<label class="grid gap-1.5 font-semibold text-brand-800">
				<span>{m.login_username_label()}</span>
				<div class="relative">
					<span
						class="icon-[lucide--user] absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-brand-400"
						aria-hidden="true"
					></span>
					<input
						name="username"
						type="text"
						autocomplete="username"
						required
						class="w-full rounded-2xl border border-brand-200 bg-accent-50/40 py-3 pr-4 pl-10 text-base focus:border-brand-500 focus:ring-2 focus:ring-brand-500 focus:outline-none"
					/>
				</div>
			</label>

			<label class="grid gap-1.5 font-semibold text-brand-800">
				<span>{m.login_password_label()}</span>
				<div class="relative">
					<span
						class="icon-[lucide--lock] absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-brand-400"
						aria-hidden="true"
					></span>
					<input
						name="password"
						type="password"
						autocomplete="current-password"
						required
						class="w-full rounded-2xl border border-brand-200 bg-accent-50/40 py-3 pr-4 pl-10 text-base focus:border-brand-500 focus:ring-2 focus:ring-brand-500 focus:outline-none"
					/>
				</div>
			</label>

			{#if errorMessage}
				<p class="flex items-center gap-2 text-sm text-error-600">
					<span class="icon-[lucide--circle-alert] size-4 shrink-0" aria-hidden="true"></span>
					{errorMessage}
				</p>
			{/if}

			<button
				type="submit"
				disabled={isSubmitting}
				class="cursor-pointer rounded-full bg-gradient-to-br from-brand-500 to-accent-500 px-4 py-3.5 font-bold text-white disabled:cursor-wait disabled:opacity-70"
			>
				{isSubmitting ? m.login_submitting() : m.login_submit()}
			</button>
		</form>

		<div class="mt-5 border-t border-brand-100 pt-4 text-sm text-accent-900/70">
			<p>{m.login_demo_hint()}</p>
			<p>{m.login_demo_admin()}</p>
		</div>
	</div>
</div>
