import { redirect } from '@sveltejs/kit';
import { getSession } from '$lib/server/auth';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	const session = await getSession(cookies);
	if (!session) {
		redirect(303, '/login');
	}

	return session;
};
