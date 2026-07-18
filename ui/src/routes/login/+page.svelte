<script lang="ts">
	import { goto } from '$app/navigation';
	import { authenticateUser, getStoredUserRole } from '$lib/auth';

	let username = $state('');
	let password = $state('');
	let error = $state('');
	let isSubmitting = $state(false);

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		error = '';
		isSubmitting = true;

		try {
			const result = await authenticateUser(username, password);
			if (!result.ok) {
				error = result.message;
				isSubmitting = false;
				return;
			}

			const role = getStoredUserRole() ?? result.role;
			await goto(`/dashboard?role=${role}`);
		} catch {
			error = 'Unable to reach the authentication service.';
		} finally {
			isSubmitting = false;
		}
	}

	function clearError() {
		error = '';
	}
</script>

<svelte:head>
	<title>Login • Toggly</title>
</svelte:head>

<div class="page-shell">
	<div class="card">
		<div class="brand-block">
			<p class="eyebrow">Secure access</p>
			<h1>Welcome back</h1>
			<p class="subtext">Sign in to continue to your workspace.</p>
		</div>

		<form onsubmit={handleSubmit} class="form-stack">
			<label class="field">
				<span>Username</span>
				<input bind:value={username} name="username" type="text" autocomplete="username" required oninput={clearError} />
			</label>

			<label class="field">
				<span>Password</span>
				<input bind:value={password} name="password" type="password" autocomplete="current-password" required oninput={clearError} />
			</label>

			{#if error}
				<p class="error">{error}</p>
			{/if}

			<button type="submit" disabled={isSubmitting}>
				{isSubmitting ? 'Signing in…' : 'Sign in'}
			</button>
		</form>

		<div class="hint">
			<p>Demo credentials:</p>
			<p>Admin: admin / admin123</p>
			<p>User: user / user123</p>
		</div>
	</div>
</div>

<style>
	:global(body) {
		margin: 0;
		background: linear-gradient(135deg, oklch(0.96 0.015 300) 0%, oklch(0.99 0.008 240) 100%);
		font-family: Inter, system-ui, sans-serif;
	}

	.page-shell {
		display: grid;
		place-items: center;
		min-height: 100vh;
		padding: 2rem;
		background: linear-gradient(135deg, oklch(0.96 0.015 300) 0%, oklch(0.99 0.008 240) 100%);
	}

	.card {
		width: min(100%, 26rem);
		padding: 2rem;
		border-radius: 1.5rem;
		background: white;
		box-shadow: 0 24px 60px oklch(0.25 0.04 300 / 0.14);
		border: 1px solid oklch(0.92 0.02 300);
	}

	.brand-block {
		margin-bottom: 1.5rem;
	}

	.eyebrow {
		margin: 0 0 0.4rem;
		font-size: 0.8rem;
		font-weight: 700;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: oklch(0.62 0.18 300);
	}

	h1 {
		margin: 0;
		font-size: 2rem;
		color: oklch(0.28 0.08 300);
	}

	.subtext {
		margin: 0.4rem 0 0;
		color: oklch(0.45 0.04 240);
	}

	.form-stack {
		display: grid;
		gap: 1rem;
	}

	.field {
		display: grid;
		gap: 0.4rem;
		font-weight: 600;
		color: oklch(0.35 0.06 300);
	}

	input {
		padding: 0.8rem 0.9rem;
		border: 1px solid oklch(0.86 0.02 300);
		border-radius: 0.9rem;
		background: oklch(0.99 0.005 300);
		font-size: 1rem;
	}

	input:focus {
		outline: 2px solid oklch(0.62 0.18 300);
		outline-offset: 2px;
	}

	button {
		padding: 0.9rem 1rem;
		border: 0;
		border-radius: 999px;
		background: linear-gradient(135deg, oklch(0.62 0.18 300) 0%, oklch(0.76 0.06 300) 100%);
		color: white;
		font-weight: 700;
		cursor: pointer;
	}

	button:disabled {
		opacity: 0.7;
		cursor: wait;
	}

	.error {
		margin: 0;
		color: oklch(0.68 0.22 15);
		font-size: 0.95rem;
	}

	.hint {
		margin-top: 1.2rem;
		padding-top: 1rem;
		border-top: 1px solid oklch(0.9 0.01 300);
		color: oklch(0.45 0.04 240);
		font-size: 0.95rem;
	}
</style>
