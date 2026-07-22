import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { apiRequest } from './api';

function jsonResponse(status: number, body: unknown): Response {
	return {
		ok: status >= 200 && status < 300,
		status,
		json: async () => body
	} as Response;
}

beforeEach(() => {
	vi.stubGlobal('fetch', vi.fn());
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('apiRequest', () => {
	it('returns the parsed payload as data on a real 2xx response', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(200, { id: 'u1' }));

		const result = await apiRequest<{ id: string }>('/dashboard/users/u1', { method: 'DELETE' });

		expect(result).toEqual({ data: { id: 'u1' } });
	});

	it('returns the backend error message on a real non-2xx response', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(
			jsonResponse(403, { error: 'cannot remove the last remaining admin account' })
		);

		const result = await apiRequest('/dashboard/users/u1', { method: 'DELETE' });

		expect(result).toEqual({ error: 'cannot remove the last remaining admin account' });
	});

	it('falls back to a generic message when the error response has none', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(500, {}));

		const result = await apiRequest('/dashboard/users/u1', { method: 'DELETE' });

		expect(result).toEqual({ error: 'Something went wrong.' });
	});

	it('always sends a JSON content-type header, overridable by the caller', async () => {
		vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(200, {}));

		await apiRequest('/dashboard/users', {
			method: 'POST',
			body: JSON.stringify({ username: 'alice' })
		});

		expect(fetch).toHaveBeenCalledWith(
			'/dashboard/users',
			expect.objectContaining({
				headers: { 'content-type': 'application/json' }
			})
		);
	});
});
