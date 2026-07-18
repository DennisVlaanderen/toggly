import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Cookies } from '@sveltejs/kit';
import { getSession, login } from './auth';

// login()/getSession() call the backend over TOGGLY_API_ORIGIN (default
// http://127.0.0.1:8080), which only exists in a real deployment. This suite
// runs in a plain Node environment, so fetch is mocked -- these are unit
// tests for the response-handling logic, not integration tests against a
// live backend.
function jsonResponse(status: number, body: unknown): Response {
	return {
		ok: status >= 200 && status < 300,
		status,
		json: async () => body
	} as Response;
}

function fakeCookies(store: Record<string, string> = {}): Cookies {
	return {
		get: (name: string) => store[name],
		getAll: () => Object.entries(store).map(([name, value]) => ({ name, value })),
		set: vi.fn(),
		delete: vi.fn(),
		serialize: vi.fn()
	} as unknown as Cookies;
}

describe('login', () => {
	beforeEach(() => {
		vi.stubGlobal('fetch', vi.fn());
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it('returns the token and user for the admin credentials', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(
			jsonResponse(200, { token: 'abc', user: { username: 'admin', role: 'admin' } })
		);

		const result = await login('admin', 'admin123');

		expect(result).toEqual({ token: 'abc', user: { username: 'admin', role: 'admin' } });
	});

	it('returns null for rejected credentials', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(401, { error: 'invalid username or password' }));

		const result = await login('guest', 'wrong');

		expect(result).toBeNull();
	});

	it('returns null when the backend response is missing a valid role', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(
			jsonResponse(200, { token: 'abc', user: { username: 'admin', role: 'superuser' } })
		);

		const result = await login('admin', 'admin123');

		expect(result).toBeNull();
	});
});

describe('getSession', () => {
	beforeEach(() => {
		vi.stubGlobal('fetch', vi.fn());
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it('returns null when there is no auth cookie', async () => {
		const result = await getSession(fakeCookies());

		expect(result).toBeNull();
		expect(fetch).not.toHaveBeenCalled();
	});

	it('returns the session when the backend confirms the token', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(200, { user: { username: 'user', role: 'user' } }));

		const result = await getSession(fakeCookies({ 'toggly.auth': 'a-token' }));

		expect(result).toEqual({ username: 'user', role: 'user' });
		expect(fetch).toHaveBeenCalledWith(
			expect.stringContaining('/api/auth/me'),
			expect.objectContaining({ headers: { Authorization: 'Bearer a-token' } })
		);
	});

	it('returns null when the backend rejects the token', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(401, { error: 'invalid token' }));

		const result = await getSession(fakeCookies({ 'toggly.auth': 'stale-token' }));

		expect(result).toBeNull();
	});
});
