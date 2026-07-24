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
		const formData = new FormData(event.currentTarget as HTMLFormElement);

		updatingId = user.id;
		updateErrors[user.id] = '';

		// groupIds is only included when the checkboxes were actually
		// rendered (data.groups.length > 0, see the template below) -- if the
		// caller can't see the full group list (e.g. missing groups:read),
		// there's no way for them to have specified a real value here, and
		// omitting the field leaves the backend to keep the user's existing
		// group membership unchanged instead of silently clearing it.
		const payload: Record<string, unknown> = {
			username: (formData.get('username') ?? '').toString(),
			password: (formData.get('password') ?? '').toString(),
			active: formData.get('active') === 'on'
		};
		if (data.groups.length > 0) {
			payload.groupIds = formData.getAll('groupIds').map(String);
		}

		const result = await apiRequest(`/bff/users/${encodeURIComponent(user.id)}`, {
			method: 'PUT',
			body: JSON.stringify(payload)
		});

		// Only clear updatingId if it's still this row's -- otherwise a
		// slower-to-resolve edit on another row (submitted after this one but
		// finishing later) would have already overwritten it, and clearing it
		// here would incorrectly re-enable that other row's submit button
		// while its request is still in flight.
		if (updatingId === user.id) {
			updatingId = null;
		}
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
	<title>{m.users_page_title()} • Aerendil</title>
</svelte:head>

<div class="grid gap-6 p-7">
	<div>
		<h1 class="text-xl font-semibold text-ink">{m.users_page_title()}</h1>
		<p class="mt-1 text-ink-muted">{m.users_page_subtitle()}</p>
	</div>

	<div class="rounded-xl border border-line-1 bg-surface">
		{#if data.users.length === 0}
			<p class="p-6 text-sm text-ink-muted">{m.users_empty()}</p>
		{:else}
			{#each data.users as user, i (user.id)}
				<div class="p-5 {i > 0 ? 'border-t border-line-4' : ''}">
					<div class="flex items-center justify-between gap-3">
						<div class="flex items-center gap-2">
							<strong class="text-ink">{user.username}</strong>
							<span
								class="rounded-full px-2 py-0.5 text-xs font-semibold tracking-wide uppercase {user.active
									? 'bg-success-bg text-success'
									: 'bg-control text-ink-muted'}"
							>
								{user.active ? m.users_status_active() : m.users_status_inactive()}
							</span>
						</div>
						<button
							type="button"
							class="cursor-pointer text-sm font-medium text-error hover:underline"
							onclick={() => requestDelete(user)}
						>
							{m.users_delete_button()}
						</button>
					</div>

					<form onsubmit={(event) => handleUpdate(event, user)} class="mt-3 grid gap-3">
						<div class="grid gap-3 sm:grid-cols-2">
							<label class="grid gap-1.5 text-sm text-ink">
								<span class="font-medium">{m.users_table_username()}</span>
								<input
									name="username"
									value={user.username}
									class="w-full rounded-lg border border-line-1 bg-page px-4 py-2 text-sm text-ink focus:border-gold focus:ring-2 focus:ring-gold/40 focus:outline-none"
								/>
							</label>
							<label class="grid gap-1.5 text-sm text-ink">
								<span class="font-medium">{m.users_edit_password_label()}</span>
								<input
									name="password"
									type="password"
									placeholder={m.users_edit_password_placeholder()}
									class="w-full rounded-lg border border-line-1 bg-page px-4 py-2 text-sm text-ink focus:border-gold focus:ring-2 focus:ring-gold/40 focus:outline-none"
								/>
							</label>
						</div>

						{#if data.groups.length > 0}
							<div class="flex flex-wrap gap-3">
								{#each data.groups as group (group.id)}
									<label class="flex items-center gap-1.5 text-sm text-ink">
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

						<label class="flex items-center gap-1.5 text-sm text-ink">
							<input type="checkbox" name="active" checked={user.active} />
							{m.users_edit_active_label()}
						</label>

						{#if updateErrors[user.id]}
							<p class="flex items-center gap-2 text-sm text-error">
								<span class="icon-[lucide--circle-alert] size-4 shrink-0" aria-hidden="true"></span>
								{updateErrors[user.id]}
							</p>
						{/if}

						<button
							type="submit"
							disabled={updatingId === user.id}
							class="cursor-pointer justify-self-start rounded-lg border border-line-1 px-4 py-2 text-sm font-medium text-ink hover:bg-line-3 disabled:cursor-wait disabled:opacity-70"
						>
							{m.users_edit_submit()}
						</button>
					</form>
				</div>
			{/each}
		{/if}
	</div>

	<div class="rounded-xl border border-line-1 bg-surface p-6">
		<h2 class="mb-4 text-base font-semibold text-ink">{m.users_create_button()}</h2>
		<form onsubmit={handleCreate} class="grid gap-4">
			<label class="grid gap-1.5 text-sm font-medium text-ink">
				<span>{m.users_create_username_label()}</span>
				<input
					name="username"
					type="text"
					required
					class="w-full rounded-lg border border-line-1 bg-page px-4 py-2.5 text-base text-ink focus:border-gold focus:ring-2 focus:ring-gold/40 focus:outline-none"
				/>
			</label>

			<label class="grid gap-1.5 text-sm font-medium text-ink">
				<span>{m.users_create_password_label()}</span>
				<input
					name="password"
					type="password"
					required
					class="w-full rounded-lg border border-line-1 bg-page px-4 py-2.5 text-base text-ink focus:border-gold focus:ring-2 focus:ring-gold/40 focus:outline-none"
				/>
			</label>

			{#if data.groups.length > 0}
				<fieldset class="grid gap-1.5">
					<legend class="text-sm font-medium text-ink">{m.users_create_groups_label()}</legend>
					<div class="flex flex-wrap gap-3">
						{#each data.groups as group (group.id)}
							<label class="flex items-center gap-1.5 text-sm text-ink">
								<input type="checkbox" name="groupIds" value={group.id} />
								{group.name}
							</label>
						{/each}
					</div>
				</fieldset>
			{/if}

			{#if createError}
				<p class="flex items-center gap-2 text-sm text-error">
					<span class="icon-[lucide--circle-alert] size-4 shrink-0" aria-hidden="true"></span>
					{createError}
				</p>
			{/if}

			<button
				type="submit"
				disabled={isCreating}
				class="cursor-pointer justify-self-start rounded-lg bg-gold px-5 py-2.5 font-semibold text-navy hover:opacity-90 disabled:cursor-wait disabled:opacity-70"
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
