import { describe, expect, it } from 'vitest';
import { hasPermission } from './permissions';

describe('hasPermission', () => {
	it('always returns true for an admin, regardless of permissions', () => {
		expect(hasPermission({ isAdmin: true, permissions: [] }, 'users:write')).toBe(true);
	});

	it('returns true when the permission is present in the set', () => {
		expect(hasPermission({ isAdmin: false, permissions: ['flags:read'] }, 'flags:read')).toBe(true);
	});

	it('returns false when the permission is absent', () => {
		expect(hasPermission({ isAdmin: false, permissions: ['flags:read'] }, 'flags:write')).toBe(
			false
		);
	});

	it('returns false for a non-admin with no permissions', () => {
		expect(hasPermission({ isAdmin: false, permissions: [] }, 'flags:read')).toBe(false);
	});
});
