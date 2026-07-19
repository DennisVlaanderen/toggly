const DEFAULT_DURATION_MS = 5000;

class ToastStore {
	message: string | null = $state(null);
	#timeout: ReturnType<typeof setTimeout> | undefined;

	show(message: string, duration = DEFAULT_DURATION_MS) {
		this.message = message;
		clearTimeout(this.#timeout);
		this.#timeout = setTimeout(() => {
			this.message = null;
		}, duration);
	}

	dismiss() {
		this.message = null;
		clearTimeout(this.#timeout);
	}
}

export const toast = new ToastStore();
