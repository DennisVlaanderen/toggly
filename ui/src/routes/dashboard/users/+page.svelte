<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { apiRequest } from '$lib/client/api';
	import { toast } from '$lib/toast.svelte';
	import { m } from '$lib/paraglide/messages.js';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import type { UserSummary } from '$lib/server/users';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let isCreating = $state(false);
	let createError = $state('');

	let updatingId: string | null = $state(null);
	let updateErrors: Record<string, string> = $state({});

	let deleteModal: { show: () => void } | undefined = $state();
	let pendingDeleteId: string | undefined;
	let pendingDeleteUsername = $state('');

	function requestDelete(user: UserSummary) {
		pendingDeleteId = user.id;
		pendingDeleteUsername = user.username;
		deleteModal?.show();
	}

	async function confirmDelete() {
		const id = pendingDeleteId;
		pendingDeleteId = undefined;
		if (!id) return;

		const result = await apiRequest(`/bff/users/${encodeURIComponent(id)}`, {
			method: 'DELETE'
		});
		if (result.error) {
			toast.show(result.error);
			return;
		}
		await invalidateAll();
	}

	async function handleUpdate(event: SubmitEvent, user: UserSummary) {
		event.preventDefault();
		const data = new FormData(event.currentTarget as HTMLFormElement);

		updatingId = user.id;
		updateErrors[user.id] = '';

		const result = await apiRequest(`/bff/users/${encodeURIComponent(user.id)}`, {
			method: 'PUT',
			body: JSON.stringify({
				username: (data.get('username') ?? '').toString(),
				password: (data.get('password') ?? '').toString(),
				groupIds: data.getAll('groupIds').map(String),
				active: data.get('active') === 'on'
			})
		});

		updatingId = null;
		if (result.error) {
			updateErrors[user.id] = result.error;
			return;
		}
		await invalidateAll();
	}

	async function handleCreate(event: SubmitEvent) {
		event.preventDefault();
		const formEl = event.currentTarget as HTMLFormElement;
		const data = new FormData(formEl);

		isCreating = true;
		createError = '';

		const result = await apiRequest('/bff/users', {
			method: 'POST',
			body: JSON.stringify({
				username: (data.get('username') ?? '').toString(),
				password: (data.get('password') ?? '').toString(),
				groupIds: data.getAll('groupIds').map(String)
			})
		});

		isCreating = false;
		if (result.error) {
			createError = result.error;
			return;
		}
		formEl.reset();
		await invalidateAll();
	}
</script>

<svelte:head>
	<title>{m.users_page_title()} • Toggly</title>
</svelte:head>

<div class="grid gap-6 p-8">
	<div>
		<h1 class="text-2xl font-bold text-brand-900">{m.users_page_title()}</h1>
		<p class="mt-1 text-accent-900/70">{m.users_page_subtitle()}</p>
	</div>

	<div class="rounded-3xl border border-brand-100 bg-white p-6 shadow-xl">
		{#if data.users.length === 0}
			<p class="text-sm text-accent-900/60">{m.users_empty()}</p>
		{:else}
			<div class="grid gap-4">
				{#each data.users as user (user.id)}
					<div class="rounded-2xl border border-brand-100 bg-accent-50/40 p-4">
						<div class="flex items-center justify-between gap-3">
							<div class="flex items-center gap-2">
								<strong class="text-brand-900">{user.username}</strong>
								<span
									class="rounded-full px-2 py-0.5 text-xs font-bold tracking-wide uppercase {user.active
										? 'bg-brand-100 text-brand-700'
										: 'bg-accent-100 text-accent-900/60'}"
								>
									{user.active ? m.users_status_active() : m.users_status_inactive()}
								</span>
							</div>
							<button
								type="button"
								class="cursor-pointer text-sm font-semibold text-error-600 hover:underline"
								onclick={() => requestDelete(user)}
							>
								{m.users_delete_button()}
							</button>
						</div>

						<form onsubmit={(event) => handleUpdate(event, user)} class="mt-3 grid gap-3">
							<div class="grid gap-3 sm:grid-cols-2">
								<label class="grid gap-1.5 text-sm text-brand-800">
									<span class="font-semibold">{m.users_table_username()}</span>
									<input
										name="username"
										value={user.username}
										class="w-full rounded-2xl border border-brand-200 bg-white px-4 py-2 text-sm focus:border-brand-500 focus:ring-2 focus:ring-brand-500 focus:outline-none"
									/>
								</label>
								<label class="grid gap-1.5 text-sm text-brand-800">
									<span class="font-semibold">{m.users_edit_password_label()}</span>
									<input
										name="password"
										type="password"
										placeholder={m.users_edit_password_placeholder()}
										class="w-full rounded-2xl border border-brand-200 bg-white px-4 py-2 text-sm focus:border-brand-500 focus:ring-2 focus:ring-brand-500 focus:outline-none"
									/>
								</label>
							</div>

							{#if data.groups.length > 0}
								<div class="flex flex-wrap gap-3">
									{#each data.groups as group (group.id)}
										<label class="flex items-center gap-1.5 text-sm text-brand-800">
											<input
												type="checkbox"
												name="groupIds"
												value={group.id}
												checked={user.groupIds.includes(group.id)}
											/>
											{group.name}
										</label>
									{/each}
								</div>
							{/if}

							<label class="flex items-center gap-1.5 text-sm text-brand-800">
								<input type="checkbox" name="active" checked={user.active} />
								{m.users_edit_active_label()}
							</label>

							{#if updateErrors[user.id]}
								<p class="flex items-center gap-2 text-sm text-error-600">
									<span class="icon-[lucide--circle-alert] size-4 shrink-0" aria-hidden="true"
									></span>
									{updateErrors[user.id]}
								</p>
							{/if}

							<button
								type="submit"
								disabled={updatingId === user.id}
								class="cursor-pointer justify-self-start rounded-full bg-brand-600 px-4 py-2 text-sm font-bold text-white disabled:cursor-wait disabled:opacity-70"
							>
								{m.users_edit_submit()}
							</button>
						</form>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<div class="rounded-3xl border border-brand-100 bg-white p-6 shadow-xl">
		<h2 class="mb-4 text-lg font-bold text-brand-900">{m.users_create_button()}</h2>
		<form onsubmit={handleCreate} class="grid gap-4">
			<label class="grid gap-1.5 font-semibold text-brand-800">
				<span>{m.users_create_username_label()}</span>
				<input
					name="username"
					type="text"
					required
					class="w-full rounded-2xl border border-brand-200 bg-accent-50/40 px-4 py-2.5 text-base focus:border-brand-500 focus:ring-2 focus:ring-brand-500 focus:outline-none"
				/>
			</label>

			<label class="grid gap-1.5 font-semibold text-brand-800">
				<span>{m.users_create_password_label()}</span>
				<input
					name="password"
					type="password"
					required
					class="w-full rounded-2xl border border-brand-200 bg-accent-50/40 px-4 py-2.5 text-base focus:border-brand-500 focus:ring-2 focus:ring-brand-500 focus:outline-none"
				/>
			</label>

			{#if data.groups.length > 0}
				<fieldset class="grid gap-1.5">
					<legend class="font-semibold text-brand-800">{m.users_create_groups_label()}</legend>
					<div class="flex flex-wrap gap-3">
						{#each data.groups as group (group.id)}
							<label class="flex items-center gap-1.5 text-sm text-brand-800">
								<input type="checkbox" name="groupIds" value={group.id} />
								{group.name}
							</label>
						{/each}
					</div>
				</fieldset>
			{/if}

			{#if createError}
				<p class="flex items-center gap-2 text-sm text-error-600">
					<span class="icon-[lucide--circle-alert] size-4 shrink-0" aria-hidden="true"></span>
					{createError}
				</p>
			{/if}

			<button
				type="submit"
				disabled={isCreating}
				class="cursor-pointer justify-self-start rounded-full bg-gradient-to-br from-brand-500 to-accent-500 px-5 py-2.5 font-bold text-white disabled:cursor-wait disabled:opacity-70"
			>
				{m.users_create_submit()}
			</button>
		</form>
	</div>
</div>

<ConfirmModal
	bind:this={deleteModal}
	title={m.users_delete_confirm_title()}
	description={m.users_delete_confirm_description({ username: pendingDeleteUsername })}
	confirmLabel={m.users_delete_button()}
	cancelLabel={m.modal_cancel()}
	variant="danger"
	onconfirm={confirmDelete}
/>
