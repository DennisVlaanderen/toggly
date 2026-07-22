<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { apiRequest } from '$lib/client/api';
	import { toast } from '$lib/toast.svelte';
	import { m } from '$lib/paraglide/messages.js';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import type { GroupSummary } from '$lib/server/groups';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let isCreating = $state(false);
	let createError = $state('');

	let updatingId: string | null = $state(null);
	let updateErrors: Record<string, string> = $state({});

	let deleteModal: { show: () => void } | undefined = $state();
	let pendingDeleteId: string | undefined;
	let pendingDeleteGroupName = $state('');

	function requestDelete(group: GroupSummary) {
		pendingDeleteId = group.id;
		pendingDeleteGroupName = group.name;
		deleteModal?.show();
	}

	async function confirmDelete() {
		const id = pendingDeleteId;
		pendingDeleteId = undefined;
		if (!id) return;

		const result = await apiRequest(`/bff/groups/${encodeURIComponent(id)}`, {
			method: 'DELETE'
		});
		if (result.error) {
			toast.show(result.error);
			return;
		}
		await invalidateAll();
	}

	async function handleUpdate(event: SubmitEvent, group: GroupSummary) {
		event.preventDefault();
		const data = new FormData(event.currentTarget as HTMLFormElement);

		updatingId = group.id;
		updateErrors[group.id] = '';

		const result = await apiRequest(`/bff/groups/${encodeURIComponent(group.id)}`, {
			method: 'PUT',
			body: JSON.stringify({
				name: (data.get('name') ?? '').toString(),
				permissions: data.getAll('permissions').map(String)
			})
		});

		// Only clear updatingId if it's still this row's -- see the identical
		// comment in dashboard/users/+page.svelte's handleUpdate.
		if (updatingId === group.id) {
			updatingId = null;
		}
		if (result.error) {
			updateErrors[group.id] = result.error;
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

		const result = await apiRequest('/bff/groups', {
			method: 'POST',
			body: JSON.stringify({
				name: (data.get('name') ?? '').toString(),
				permissions: data.getAll('permissions').map(String)
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
	<title>{m.groups_page_title()} • Toggly</title>
</svelte:head>

<div class="grid gap-6 p-8">
	<div>
		<h1 class="text-2xl font-bold text-brand-900">{m.groups_page_title()}</h1>
		<p class="mt-1 text-accent-900/70">{m.groups_page_subtitle()}</p>
	</div>

	<div class="rounded-3xl border border-brand-100 bg-white p-6 shadow-xl">
		{#if data.groups.length === 0}
			<p class="text-sm text-accent-900/60">{m.groups_empty()}</p>
		{:else}
			<div class="grid gap-4">
				{#each data.groups as group (group.id)}
					<div class="rounded-2xl border border-brand-100 bg-accent-50/40 p-4">
						<div class="flex items-center justify-between gap-3">
							<div class="flex items-center gap-2">
								<strong class="text-brand-900">{group.name}</strong>
								{#if group.system}
									<span
										class="rounded-full bg-brand-100 px-2 py-0.5 text-xs font-bold tracking-wide text-brand-700 uppercase"
									>
										{m.groups_system_badge()}
									</span>
								{/if}
							</div>
							{#if !group.system}
								<button
									type="button"
									class="cursor-pointer text-sm font-semibold text-error-600 hover:underline"
									onclick={() => requestDelete(group)}
								>
									{m.groups_delete_button()}
								</button>
							{/if}
						</div>

						{#if group.system}
							<p class="mt-1 text-sm text-accent-900/60">{m.groups_admin_protected_hint()}</p>
						{:else}
							<form onsubmit={(event) => handleUpdate(event, group)} class="mt-3 grid gap-3">
								<input
									name="name"
									value={group.name}
									class="w-full rounded-2xl border border-brand-200 bg-white px-4 py-2 text-sm focus:border-brand-500 focus:ring-2 focus:ring-brand-500 focus:outline-none"
								/>
								<div class="flex flex-wrap gap-3">
									{#each data.allPermissions as perm (perm)}
										<label class="flex items-center gap-1.5 text-sm text-brand-800">
											<input
												type="checkbox"
												name="permissions"
												value={perm}
												checked={group.permissions.includes(perm)}
											/>
											{perm}
										</label>
									{/each}
								</div>

								{#if updateErrors[group.id]}
									<p class="flex items-center gap-2 text-sm text-error-600">
										<span class="icon-[lucide--circle-alert] size-4 shrink-0" aria-hidden="true"
										></span>
										{updateErrors[group.id]}
									</p>
								{/if}

								<button
									type="submit"
									disabled={updatingId === group.id}
									class="cursor-pointer justify-self-start rounded-full bg-brand-600 px-4 py-2 text-sm font-bold text-white disabled:cursor-wait disabled:opacity-70"
								>
									{m.groups_edit_submit()}
								</button>
							</form>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<div class="rounded-3xl border border-brand-100 bg-white p-6 shadow-xl">
		<h2 class="mb-4 text-lg font-bold text-brand-900">{m.groups_create_button()}</h2>
		<form onsubmit={handleCreate} class="grid gap-4">
			<label class="grid gap-1.5 font-semibold text-brand-800">
				<span>{m.groups_create_name_label()}</span>
				<input
					name="name"
					type="text"
					required
					class="w-full rounded-2xl border border-brand-200 bg-accent-50/40 px-4 py-2.5 text-base focus:border-brand-500 focus:ring-2 focus:ring-brand-500 focus:outline-none"
				/>
			</label>

			<fieldset class="grid gap-1.5">
				<legend class="font-semibold text-brand-800">{m.groups_create_permissions_label()}</legend>
				<div class="flex flex-wrap gap-3">
					{#each data.allPermissions as perm (perm)}
						<label class="flex items-center gap-1.5 text-sm text-brand-800">
							<input type="checkbox" name="permissions" value={perm} />
							{perm}
						</label>
					{/each}
				</div>
			</fieldset>

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
				{m.groups_create_submit()}
			</button>
		</form>
	</div>
</div>

<ConfirmModal
	bind:this={deleteModal}
	title={m.groups_delete_confirm_title()}
	description={m.groups_delete_confirm_description({ name: pendingDeleteGroupName })}
	confirmLabel={m.groups_delete_button()}
	cancelLabel={m.modal_cancel()}
	variant="danger"
	onconfirm={confirmDelete}
/>
