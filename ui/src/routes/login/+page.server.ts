import { fail, redirect } from '@sveltejs/kit';
import { getSession, login, setAuthCookie } from '$lib/server/auth';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	if (await getSession(cookies)) {
		redirect(303, '/dashboard');
	}
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const data = await request.formData();
		const username = (data.get('username') ?? '').toString().trim();
		const password = (data.get('password') ?? '').toString().trim();

		if (!username || !password) {
			return fail(400, { message: 'Invalid username or password.' });
		}

		const result = await login(username, password);
		if (!result) {
			return fail(401, { message: 'Invalid username or password.' });
		}

		setAuthCookie(cookies, result.token);
		redirect(303, '/dashboard');
	}
};
