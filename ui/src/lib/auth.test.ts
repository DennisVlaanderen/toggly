import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { authenticateUser } from './auth';

// authenticateUser calls fetch('/api/auth/login') with a relative URL, which
// only resolves against a page origin in a real browser. This suite runs in
// a plain Node environment, so fetch is mocked -- these are unit tests for
// the response-handling logic, not integration tests against a live backend.
function fakeToken(role: string): string {
	const payload = Buffer.from(JSON.stringify({ role })).toString('base64url');
	return `header.${payload}.signature`;
}

function jsonResponse(status: number, body: unknown): Response {
	return {
		ok: status >= 200 && status < 300,
		status,
		json: async () => body
	} as Response;
}

describe('authenticateUser', () => {
	beforeEach(() => {
		vi.stubGlobal('fetch', vi.fn());
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it('accepts the admin credentials', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(200, { token: fakeToken('admin') }));

		const result = await authenticateUser('admin', 'admin123');

		expect(result.ok).toBe(true);
		expect(result.role).toBe('admin');
	});

	it('accepts the regular user credentials', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(200, { token: fakeToken('user') }));

		const result = await authenticateUser('user', 'user123');

		expect(result.ok).toBe(true);
		expect(result.role).toBe('user');
	});

	it('rejects empty credentials with a generic message', async () => {
		const result = await authenticateUser('', '');

		if (result.ok) {
			throw new Error('Expected authentication to fail');
		}

		expect(result.ok).toBe(false);
		expect(result.role).toBeNull();
		expect(result.message).toBe('Invalid username or password.');
		expect(fetch).not.toHaveBeenCalled();
	});

	it('rejects unknown credentials', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(
			jsonResponse(401, { error: 'invalid username or password' })
		);

		const result = await authenticateUser('guest', 'wrong');

		expect(result.ok).toBe(false);
		expect(result.role).toBeNull();
	});
});
