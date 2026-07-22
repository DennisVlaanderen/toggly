import { json } from '@sveltejs/kit';
import { login, setAuthCookie } from '$lib/server/auth';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request, cookies }) => {
	const body = await request.json().catch(() => null);
	const username = typeof body?.username === 'string' ? body.username.trim() : '';
	// Password is intentionally NOT trimmed -- user creation/update never
	// trims it either (backend/internal/api/users.go), so trimming only
	// here would silently lock out anyone whose password has meaningful
	// leading/trailing whitespace.
	const password = typeof body?.password === 'string' ? body.password : '';

	if (!username || !password) {
		return json({ error: 'Invalid username or password.' }, { status: 400 });
	}

	const result = await login(username, password);
	if (!result) {
		return json({ error: 'Invalid username or password.' }, { status: 401 });
	}

	setAuthCookie(cookies, result.token);
	return json({ success: true });
};
